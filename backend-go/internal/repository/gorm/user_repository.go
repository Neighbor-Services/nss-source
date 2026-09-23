package gorm

import (
	"context"
	"errors"
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
	return r.db.WithContext(ctx).Delete(&entity.User{}, "id = ?", id).Error
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
	var profiles []entity.Profile
	query := r.db.WithContext(ctx).
		Preload("User").
		Preload("CatalogServices").
		Preload("CatalogServices.Category").
		Preload("PortfolioItems").
		Preload("ServicePackages")

	if params.UserType != "" {
		query = query.Where("accounts_profile.user_type = ?", params.UserType)
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
	if params.CategorySlug != "" {
		query = query.Joins("JOIN accounts_profile_catalog_services apcs ON apcs.profile_id = accounts_profile.id").
			Joins("JOIN services_catalogservice scs ON scs.id = apcs.catalog_service_id").
			Joins("JOIN services_category sc ON sc.id = scs.category_id").
			Where("sc.slug = ? OR sc.name ILIKE ?", params.CategorySlug, "%"+params.CategorySlug+"%").
			Distinct()
	} else if params.CategoryName != "" {
		query = query.Joins("JOIN accounts_profile_catalog_services apcs ON apcs.profile_id = accounts_profile.id").
			Joins("JOIN services_catalogservice scs ON scs.id = apcs.catalog_service_id").
			Joins("JOIN services_category sc ON sc.id = scs.category_id").
			Where("sc.name ILIKE ?", "%"+params.CategoryName+"%").
			Distinct()
	}
	if params.ServiceID != "" {
		if sid, err := uuid.Parse(params.ServiceID); err == nil {
			query = query.Joins("JOIN accounts_profile_catalog_services apcs_srv ON apcs_srv.profile_id = accounts_profile.id").
				Where("apcs_srv.catalog_service_id = ?", sid).
				Distinct()
		}
	} else if params.ServiceName != "" {
		query = query.Joins("JOIN accounts_profile_catalog_services apcs_srv ON apcs_srv.profile_id = accounts_profile.id").
			Joins("JOIN services_catalogservice scs_srv ON scs_srv.id = apcs_srv.catalog_service_id").
			Where("scs_srv.name ILIKE ?", "%"+params.ServiceName+"%").
			Distinct()
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
	}

	if params.Popular {
		query = query.Order("accounts_profile.average_rating DESC, accounts_profile.total_reviews DESC")
	} else if !hasLocation {
		query = query.Order("accounts_profile.created_at DESC")
	}

	if !hasLocation {
		if params.Limit > 0 {
			query = query.Limit(params.Limit)
		}
		if params.Offset > 0 {
			query = query.Offset(params.Offset)
		}
	}

	err := query.Find(&profiles).Error
	if err != nil {
		return nil, err
	}

	if hasLocation {
		lat := *params.Latitude
		lng := *params.Longitude
		maxRadius := *params.RadiusKm

		type distProfile struct {
			profile  entity.Profile
			distance float64
		}
		var distList []distProfile

		for _, p := range profiles {
			if p.Latitude == 0 && p.Longitude == 0 {
				continue
			}
			dist := geo.HaversineDistance(lat, lng, p.Latitude, p.Longitude)
			if dist <= maxRadius {
				distList = append(distList, distProfile{profile: p, distance: dist})
			}
		}

		if params.Popular {
			// When popular is requested with location: sort by highest rating & reviews first, then closest distance
			sort.Slice(distList, func(i, j int) bool {
				if distList[i].profile.AverageRating != distList[j].profile.AverageRating {
					return distList[i].profile.AverageRating > distList[j].profile.AverageRating
				}
				if distList[i].profile.TotalReviews != distList[j].profile.TotalReviews {
					return distList[i].profile.TotalReviews > distList[j].profile.TotalReviews
				}
				return distList[i].distance < distList[j].distance
			})
		} else {
			// Sort by distance ascending
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

		var filtered []entity.Profile
		for i := start; i < end; i++ {
			filtered = append(filtered, distList[i].profile)
		}
		return filtered, nil
	}

	return profiles, nil
}

func (r *profileRepository) ListProviders(ctx context.Context, categorySlug string, limit int, offset int) ([]entity.Profile, error) {
	return r.List(ctx, repository.ProfileFilterParams{
		UserType:     "PROVIDER",
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
	return r.db.WithContext(ctx).Where("id = ? AND profile_id = ?", id, profileID).Delete(&entity.Portfolio{}).Error
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
	return r.db.WithContext(ctx).Where("id = ? AND profile_id = ?", id, profileID).Delete(&entity.ServicePackage{}).Error
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
