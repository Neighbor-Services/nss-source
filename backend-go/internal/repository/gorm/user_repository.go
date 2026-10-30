package gorm

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"sort"

	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	"backend-go/pkg/geo"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var profileIDs []uuid.UUID
		_ = tx.Table("accounts_profile").Where("user_id = ?", id).Pluck("id", &profileIDs).Error
		if len(profileIDs) > 0 {
			_ = tx.Exec("DELETE FROM accounts_profile_catalog_services WHERE profile_id IN ?", profileIDs).Error
			_ = tx.Table("accounts_portfolio").Where("profile_id IN ?", profileIDs).Unscoped().Delete(&entity.Portfolio{}).Error
			_ = tx.Table("accounts_servicepackage").Where("profile_id IN ?", profileIDs).Unscoped().Delete(&entity.ServicePackage{}).Error
			_ = tx.Table("accounts_profile").Where("id IN ?", profileIDs).Unscoped().Delete(&entity.Profile{}).Error
		}
		_ = tx.Table("subscriptions_usersubscription").Where("user_id = ?", id).Unscoped().Delete(&entity.UserSubscription{}).Error
		_ = tx.Table("accounts_devicetoken").Where("user_id = ?", id).Unscoped().Delete(&entity.DeviceToken{}).Error
		_ = tx.Table("wallets_wallet").Where("user_id = ?", id).Unscoped().Delete(&entity.Wallet{}).Error
		return tx.Unscoped().Delete(&entity.User{}, "id = ?", id).Error
	})
}

func (r *userRepository) CountByEmail(ctx context.Context, email string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where("email = ?", email).Count(&count).Error
	return count, err
}

// Profile Repository Implementation
type profileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) repository.ProfileRepository {
	return &profileRepository{db: db}
}

func (r *profileRepository) Create(ctx context.Context, profile *entity.Profile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

func (r *profileRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Profile, error) {
	var profile entity.Profile
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("CatalogServices").
		Preload("CatalogServices.Category").
		Preload("PortfolioItems").
		Preload("ServicePackages").
		First(&profile, "user_id = ?", userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

func (r *profileRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Profile, error) {
	var profile entity.Profile
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("CatalogServices").
		Preload("CatalogServices.Category").
		Preload("PortfolioItems").
		Preload("ServicePackages").
		First(&profile, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

func (r *profileRepository) Update(ctx context.Context, profile *entity.Profile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

func (r *profileRepository) UpdateCatalogServices(ctx context.Context, profileID uuid.UUID, serviceIDs []uuid.UUID) error {
	var services []entity.CatalogService
	if len(serviceIDs) > 0 {
		if err := r.db.WithContext(ctx).Where("id IN ?", serviceIDs).Find(&services).Error; err != nil {
			return err
		}
	}
	profile := entity.Profile{ID: profileID}
	return r.db.WithContext(ctx).Model(&profile).Association("CatalogServices").Replace(&services)
}

func (r *profileRepository) List(ctx context.Context, params repository.ProfileFilterParams) ([]entity.Profile, error) {
	query := r.db.WithContext(ctx).Model(&entity.Profile{})

	if params.UserType != "" {
		query = query.Where("accounts_profile.user_type = ?", params.UserType)
	}
	if params.ExcludeUserID != nil && *params.ExcludeUserID != uuid.Nil {
		query = query.Where("accounts_profile.user_id != ?", *params.ExcludeUserID)
	}

	// Providers without an active paid subscription AND cleared Checkr verification must be invisible across AI match, popular listings, and provider searches.
	if params.UserType == "PROVIDER" || params.Popular {
		query = query.Where("accounts_profile.is_identity_verified = true AND (accounts_profile.subscription_tier IN ('SILVER', 'GOLD', 'PLATINUM', 'DIAMOND', 'PRO') OR EXISTS (SELECT 1 FROM payments_subscription ps WHERE ps.user_id = accounts_profile.user_id AND ps.is_active = true AND (ps.next_payment IS NULL OR ps.next_payment > NOW())) OR EXISTS (SELECT 1 FROM moderation_backgroundcheck mbc WHERE mbc.provider_id = accounts_profile.user_id AND mbc.status = 'clear' AND mbc.cleared_at IS NOT NULL AND mbc.cleared_at > NOW() - INTERVAL '30 days'))")
	} else if params.UserType == "" {
		query = query.Where("(accounts_profile.user_type != 'PROVIDER' OR (accounts_profile.is_identity_verified = true AND (accounts_profile.subscription_tier IN ('SILVER', 'GOLD', 'PLATINUM', 'DIAMOND', 'PRO') OR EXISTS (SELECT 1 FROM payments_subscription ps WHERE ps.user_id = accounts_profile.user_id AND ps.is_active = true AND (ps.next_payment IS NULL OR ps.next_payment > NOW())) OR EXISTS (SELECT 1 FROM moderation_backgroundcheck mbc WHERE mbc.provider_id = accounts_profile.user_id AND mbc.status = 'clear' AND mbc.cleared_at IS NOT NULL AND mbc.cleared_at > NOW() - INTERVAL '30 days'))))")
	}
	if params.Search != "" {
		s := "%" + params.Search + "%"
		query = query.Where("(accounts_profile.first_name ILIKE ? OR accounts_profile.last_name ILIKE ? OR accounts_profile.service ILIKE ? OR accounts_profile.city ILIKE ?)", s, s, s, s)
	}
	if params.City != "" {
		query = query.Where("accounts_profile.city ILIKE ?", "%"+params.City+"%")
	}
	if params.RatingMin != nil && *params.RatingMin > 0 {
		query = query.Where("accounts_profile.average_rating >= ?", *params.RatingMin)
	}
	// NOTE: Distinct() is intentionally avoided here — GORM cannot resolve Preload
	// primary keys when Distinct is active, causing panics/500s. We use Group() instead
	// which deduplicates joined rows without breaking the Preload pipeline.
	if params.CategorySlug != "" {
		query = query.Joins("JOIN accounts_profile_catalog_services apcs ON apcs.profile_id = accounts_profile.id").
			Joins("JOIN services_catalogservice scs ON scs.id = apcs.catalog_service_id").
			Joins("JOIN services_category sc ON sc.id = scs.category_id").
			Where("sc.slug = ? OR sc.name ILIKE ?", params.CategorySlug, "%"+params.CategorySlug+"%").
			Group("accounts_profile.id")
	} else if params.CategoryName != "" {
		query = query.Joins("JOIN accounts_profile_catalog_services apcs ON apcs.profile_id = accounts_profile.id").
			Joins("JOIN services_catalogservice scs ON scs.id = apcs.catalog_service_id").
			Joins("JOIN services_category sc ON sc.id = scs.category_id").
			Where("sc.name ILIKE ?", "%"+params.CategoryName+"%").
			Group("accounts_profile.id")
	}
	if params.ServiceID != "" {
		if sid, err := uuid.Parse(params.ServiceID); err == nil {
			query = query.Joins("JOIN accounts_profile_catalog_services apcs_srv ON apcs_srv.profile_id = accounts_profile.id").
				Where("apcs_srv.catalog_service_id = ?", sid).
				Group("accounts_profile.id")
		}
	} else if params.ServiceName != "" {
		query = query.Joins("JOIN accounts_profile_catalog_services apcs_srv ON apcs_srv.profile_id = accounts_profile.id").
			Joins("JOIN services_catalogservice scs_srv ON scs_srv.id = apcs_srv.catalog_service_id").
			Where("scs_srv.name ILIKE ?", "%"+params.ServiceName+"%").
			Group("accounts_profile.id")
	}

	// Geospatial bounding-box optimization when location is provided
	hasLocation := params.Latitude != nil && params.Longitude != nil && params.RadiusKm != nil && *params.RadiusKm > 0
	if hasLocation {
		lat := *params.Latitude
		lng := *params.Longitude
		radius := *params.RadiusKm

		latDelta := (radius / 111.0) * 1.15
		cosLat := math.Cos(lat * math.Pi / 180.0)
		if cosLat < 0.001 {
			cosLat = 0.001
		}
		lngDelta := (radius / (111.0 * cosLat)) * 1.15

		minLat := lat - latDelta
		maxLat := lat + latDelta
		minLng := lng - lngDelta
		maxLng := lng + lngDelta

		query = query.Where("accounts_profile.latitude != 0 AND accounts_profile.longitude != 0").
			Where("accounts_profile.latitude BETWEEN ? AND ?", minLat, maxLat).
			Where("accounts_profile.longitude BETWEEN ? AND ?", minLng, maxLng)

		type lightProfile struct {
			ID            uuid.UUID `gorm:"column:id"`
			Latitude      float64   `gorm:"column:latitude"`
			Longitude     float64   `gorm:"column:longitude"`
			AverageRating float64   `gorm:"column:average_rating"`
			TotalReviews  int       `gorm:"column:total_reviews"`
			NeighborScore int       `gorm:"column:neighbor_score"`
		}
		var lightProfiles []lightProfile
		// Use MAX() aggregates for all non-key columns so the SELECT is valid
		// whether or not a GROUP BY was already applied by a category/service JOIN filter.
		// The values are functionally determined by id, so MAX() is semantically equivalent.
		err := query.
			Select("accounts_profile.id, MAX(accounts_profile.latitude) AS latitude, MAX(accounts_profile.longitude) AS longitude, MAX(accounts_profile.average_rating) AS average_rating, MAX(accounts_profile.total_reviews) AS total_reviews, MAX(accounts_profile.neighbor_score) AS neighbor_score").
			Group("accounts_profile.id").
			Find(&lightProfiles).Error

		if err != nil {
			slog.Error("profileRepository.List: light query failed", "error", err)
			return nil, err
		}

		type distProfile struct {
			id       uuid.UUID
			distance float64
			rating   float64
			reviews  int
			score    float64
		}
		var distList []distProfile
		for _, lp := range lightProfiles {
			if lp.Latitude == 0 && lp.Longitude == 0 {
				continue
			}
			dist := geo.HaversineDistance(lat, lng, lp.Latitude, lp.Longitude)
			if dist <= radius {
				distList = append(distList, distProfile{
					id:       lp.ID,
					distance: dist,
					rating:   lp.AverageRating,
					reviews:  lp.TotalReviews,
					score:    float64(lp.NeighborScore),
				})

			}
		}

		if params.Popular {
			// When popular is requested with location: sort by highest rating & reviews first, then closest distance
			sort.Slice(distList, func(i, j int) bool {
				if distList[i].rating != distList[j].rating {
					return distList[i].rating > distList[j].rating
				}
				if distList[i].reviews != distList[j].reviews {
					return distList[i].reviews > distList[j].reviews
				}
				if distList[i].score != distList[j].score {
					return distList[i].score > distList[j].score
				}
				return distList[i].distance < distList[j].distance
			})
		} else {
			sort.Slice(distList, func(i, j int) bool {
				return distList[i].distance < distList[j].distance
			})
		}

		start := params.Offset
		if start > len(distList) {
			return []entity.Profile{}, nil
		}
		end := len(distList)
		if params.Limit > 0 && start+params.Limit < end {
			end = start + params.Limit
		}
		if start >= end {
			return []entity.Profile{}, nil
		}

		targetIDs := make([]uuid.UUID, 0, end-start)
		orderMap := make(map[uuid.UUID]int)
		for i := start; i < end; i++ {
			targetIDs = append(targetIDs, distList[i].id)
			orderMap[distList[i].id] = i - start
		}

		var hydrated []entity.Profile
		err = r.db.WithContext(ctx).
			Preload("User").
			Preload("CatalogServices").
			Preload("CatalogServices.Category").
			Preload("PortfolioItems").
			Preload("ServicePackages").
			Where("id IN ?", targetIDs).
			Find(&hydrated).Error
		if err != nil {
			return nil, err
		}

		// Reorder hydrated list to match distance/popular sorted order
		sortedProfiles := make([]entity.Profile, len(hydrated))
		for _, p := range hydrated {
			if idx, ok := orderMap[p.ID]; ok && idx < len(sortedProfiles) {
				sortedProfiles[idx] = p
			}
		}
		return sortedProfiles, nil
	}

	// Non-location flow
	var profiles []entity.Profile
	query = query.
		Preload("User").
		Preload("CatalogServices").
		Preload("CatalogServices.Category").
		Preload("PortfolioItems").
		Preload("ServicePackages")

	if params.Popular {
		query = query.Order("accounts_profile.neighbor_score DESC, accounts_profile.average_rating DESC, accounts_profile.total_reviews DESC")
	} else {
		query = query.Order("accounts_profile.created_at DESC")
	}

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	err := query.Find(&profiles).Error
	if err != nil {
		slog.Error("profileRepository.List: query failed",
 "error", err,
 "user_type", params.UserType,
 "service_id", params.ServiceID,
 "service_name", params.ServiceName,
 "category_slug", params.CategorySlug,
 "has_location", false,
		)
		return nil, err
	}

	return profiles, nil
}

func (r *profileRepository) ListProviders(ctx context.Context, categorySlug string, limit int, offset int) ([]entity.Profile, error) {
	return r.List(ctx, repository.ProfileFilterParams{
		UserType: "PROVIDER",
		CategorySlug: categorySlug,
		Limit:        limit,
		Offset:       offset,
	})
}

// About Repository Implementation
type aboutRepository struct {
	db *gorm.DB
}

func NewAboutRepository(db *gorm.DB) repository.AboutRepository {
	return &aboutRepository{db: db}
}

func (r *aboutRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.About, error) {
	var about entity.About
	err := r.db.WithContext(ctx).First(&about, "user_id = ?", userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &about, nil
}

func (r *aboutRepository) Upsert(ctx context.Context, about *entity.About) error {
	var existing entity.About
	err := r.db.WithContext(ctx).Where("user_id = ?", about.UserID).First(&existing).Error
	if err == nil {
		about.ID = existing.ID
		return r.db.WithContext(ctx).Save(about).Error
	}
	return r.db.WithContext(ctx).Create(about).Error
}

// Portfolio Repository Implementation
type portfolioRepository struct {
	db *gorm.DB
}

func NewPortfolioRepository(db *gorm.DB) repository.PortfolioRepository {
	return &portfolioRepository{db: db}
}

func (r *portfolioRepository) ListByProfileID(ctx context.Context, profileID uuid.UUID) ([]entity.Portfolio, error) {
	var list []entity.Portfolio
	err := r.db.WithContext(ctx).Where("profile_id = ?", profileID).Find(&list).Error
	return list, err
}

func (r *portfolioRepository) Create(ctx context.Context, portfolio *entity.Portfolio) error {
	return r.db.WithContext(ctx).Create(portfolio).Error
}

func (r *portfolioRepository) Delete(ctx context.Context, id uuid.UUID, profileID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ? AND profile_id = ?", id, profileID).Unscoped().Delete(&entity.Portfolio{}).Error
}

// Service Package Repository Implementation
type servicePackageRepository struct {
	db *gorm.DB
}

func NewServicePackageRepository(db *gorm.DB) repository.ServicePackageRepository {
	return &servicePackageRepository{db: db}
}

func (r *servicePackageRepository) ListByProfileID(ctx context.Context, profileID uuid.UUID) ([]entity.ServicePackage, error) {
	var list []entity.ServicePackage
	err := r.db.WithContext(ctx).Where("profile_id = ?", profileID).Find(&list).Error
	return list, err
}

func (r *servicePackageRepository) Create(ctx context.Context, pkg *entity.ServicePackage) error {
	return r.db.WithContext(ctx).Create(pkg).Error
}

func (r *servicePackageRepository) Update(ctx context.Context, pkg *entity.ServicePackage) error {
	return r.db.WithContext(ctx).Save(pkg).Error
}

func (r *servicePackageRepository) Delete(ctx context.Context, id uuid.UUID, profileID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ? AND profile_id = ?", id, profileID).Unscoped().Delete(&entity.ServicePackage{}).Error
}

// Legal Document Repository
type legalDocRepository struct {
	db *gorm.DB
}

func NewLegalDocumentRepository(db *gorm.DB) repository.LegalDocumentRepository {
	return &legalDocRepository{db: db}
}

func (r *legalDocRepository) ListActive(ctx context.Context) ([]entity.LegalDocument, error) {
	var docs []entity.LegalDocument
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&docs).Error
	return docs, err
}

func (r *legalDocRepository) GetByType(ctx context.Context, docType string) ([]entity.LegalDocument, error) {
	var docs []entity.LegalDocument
	err := r.db.WithContext(ctx).Where("doc_type = ? AND is_active = ?", docType, true).Order("created_at ASC").Find(&docs).Error
	return docs, err
}
