package gorm

import (
	"context"
	"math"
	"sort"
	"strings"
	"time"

	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) repository.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) ListActive(ctx context.Context) ([]entity.Category, error) {
	var categories []entity.Category
	err := r.db.WithContext(ctx).Order("name ASC").Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) GetBySlug(ctx context.Context, slug string) (*entity.Category, error) {
	var cat entity.Category
	var err error
	if id, parseErr := uuid.Parse(slug); parseErr == nil {
		err = r.db.WithContext(ctx).First(&cat, "id = ?", id).Error
	} else {
		err = r.db.WithContext(ctx).First(&cat, "name = ?", slug).Error
	}
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	var cat entity.Category
	err := r.db.WithContext(ctx).First(&cat, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

type catalogServiceRepository struct {
	db *gorm.DB
}

func NewCatalogServiceRepository(db *gorm.DB) repository.CatalogServiceRepository {
	return &catalogServiceRepository{db: db}
}

func (r *catalogServiceRepository) List(ctx context.Context, categorySlug string, search string) ([]entity.CatalogService, error) {
	var services []entity.CatalogService
	query := r.db.WithContext(ctx).Preload("Category")

	if categorySlug != "" {
		if catUUID, err := uuid.Parse(categorySlug); err == nil {
			query = query.Where("category_id = ?", catUUID)
		} else {
			query = query.Joins("JOIN services_category ON services_category.id = services_catalogservice.category_id").
				Where("services_category.name ILIKE ?", categorySlug)
		}
	}
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("services_catalogservice.name ILIKE ? OR services_catalogservice.description ILIKE ?", searchPattern, searchPattern)
	}

	err := query.Order("name ASC").Find(&services).Error
	return services, err
}

func (r *catalogServiceRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.CatalogService, error) {
	var srv entity.CatalogService
	err := r.db.WithContext(ctx).Preload("Category").First(&srv, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &srv, nil
}

type serviceRequestRepository struct {
	db *gorm.DB
}

func NewServiceRequestRepository(db *gorm.DB) repository.ServiceRequestRepository {
	return &serviceRequestRepository{db: db}
}

func (r *serviceRequestRepository) Create(ctx context.Context, req *entity.ServiceRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *serviceRequestRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ServiceRequest, error) {
	var req entity.ServiceRequest
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("User.Profile").
		Preload("TargetProvider").
		Preload("CatalogService").
		Preload("Proposals").
		Preload("Proposals.Provider").
		First(&req, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *serviceRequestRepository) List(ctx context.Context, params repository.ServiceRequestFilterParams) ([]entity.ServiceRequest, error) {
	var list []entity.ServiceRequest
	query := r.db.WithContext(ctx).
		Preload("User").
		Preload("User.Profile").
		Preload("TargetProvider").
		Preload("CatalogService").
		Preload("Proposals").
		Order("services_servicerequest.created_at DESC")

	if params.UserID != nil {
		query = query.Where("services_servicerequest.user_id = ?", *params.UserID)
	}
	if params.TargetProviderID != nil && len(params.OfferedServices) == 0 {
		query = query.Where("services_servicerequest.target_provider_id = ?", *params.TargetProviderID)
	}
	if params.CatalogServiceID != nil {
		query = query.Where("services_servicerequest.catalog_service_id = ?", *params.CatalogServiceID)
	}
	if params.Status != "" {
		query = query.Where("services_servicerequest.status = ?", params.Status)
	}

	// Filter by provider's offered services
	if len(params.OfferedServices) > 0 {
		var serviceConditions []string
		var serviceArgs []interface{}
		for _, s := range params.OfferedServices {
			s = strings.TrimSpace(s)
			if s != "" {
				serviceConditions = append(serviceConditions, "services_servicerequest.service_type ILIKE ? OR services_servicerequest.title ILIKE ? OR services_catalogservice.name ILIKE ?")
				pattern := "%" + s + "%"
				serviceArgs = append(serviceArgs, pattern, pattern, pattern)
			}
		}
		if len(serviceConditions) > 0 {
			combinedSQL := "(" + strings.Join(serviceConditions, " OR ")
			if params.TargetProviderID != nil {
				combinedSQL += " OR services_servicerequest.target_provider_id = ?"
				serviceArgs = append(serviceArgs, *params.TargetProviderID)
			}
			combinedSQL += ")"
			query = query.Joins("LEFT JOIN services_catalogservice ON services_catalogservice.id = services_servicerequest.catalog_service_id").
				Where(combinedSQL, serviceArgs...)
		}
	}

	if params.Limit > 0 && (params.Latitude == nil || params.Longitude == nil) {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 && (params.Latitude == nil || params.Longitude == nil) {
		query = query.Offset(params.Offset)
	}

	err := query.Find(&list).Error
	if err != nil {
		return nil, err
	}

	// Geo-radius & Haversine distance calculation
	if params.Latitude != nil && params.Longitude != nil {
		var filtered []entity.ServiceRequest
		for i := range list {
			req := &list[i]
			if req.Latitude != nil && req.Longitude != nil && (*req.Latitude != 0 || *req.Longitude != 0) {
				dist := haversineDistance(*params.Latitude, *params.Longitude, *req.Latitude, *req.Longitude)
				req.Distance = &dist

				if params.RadiusKm != nil && *params.RadiusKm > 0 {
					if dist <= *params.RadiusKm {
						filtered = append(filtered, *req)
					}
				} else {
					filtered = append(filtered, *req)
				}
			} else {
				// Include requests without coordinates if no strict radius is enforced
				if params.RadiusKm == nil || *params.RadiusKm <= 0 {
					filtered = append(filtered, *req)
				}
			}
		}

		// Sort by distance ascending
		sort.SliceStable(filtered, func(i, j int) bool {
			if filtered[i].Distance == nil {
				return false
			}
			if filtered[j].Distance == nil {
				return true
			}
			return *filtered[i].Distance < *filtered[j].Distance
		})

		if params.Offset > 0 {
			if params.Offset < len(filtered) {
				filtered = filtered[params.Offset:]
			} else {
				filtered = []entity.ServiceRequest{}
			}
		}
		if params.Limit > 0 && len(filtered) > params.Limit {
			filtered = filtered[:params.Limit]
		}
		list = filtered
	}

	return list, nil
}

func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	rLat1 := lat1 * (math.Pi / 180.0)
	rLat2 := lat2 * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(rLat1)*math.Cos(rLat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

func (r *serviceRequestRepository) ListByUser(ctx context.Context, userID *uuid.UUID, targetProviderID *uuid.UUID, status string) ([]entity.ServiceRequest, error) {
	return r.List(ctx, repository.ServiceRequestFilterParams{
		UserID:           userID,
		TargetProviderID: targetProviderID,
		Status:           status,
	})
}

func (r *serviceRequestRepository) Update(ctx context.Context, req *entity.ServiceRequest) error {
	return r.db.WithContext(ctx).
		Session(&gorm.Session{FullSaveAssociations: false}).
		Model(&entity.ServiceRequest{}).
		Where("id = ?", req.ID).
		Select(
			"user_id", "target_provider_id", "catalog_service_id",
			"title", "description", "price", "status",
			"preferred_payment_mode", "service_type",
			"with_image", "image",
			"longitude", "latitude", "scheduled_time",
			"updated_at", "deleted_at",
		).
		Updates(req).Error
}

func (r *serviceRequestRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.ServiceRequest{}, "id = ?", id).Error
}

type proposalRepository struct {
	db *gorm.DB
}

func NewProposalRepository(db *gorm.DB) repository.ProposalRepository {
	return &proposalRepository{db: db}
}

func (r *proposalRepository) Create(ctx context.Context, proposal *entity.Proposal) error {
	return r.db.WithContext(ctx).Create(proposal).Error
}

func (r *proposalRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Proposal, error) {
	var p entity.Proposal
	err := r.db.WithContext(ctx).Preload("Provider").Preload("Provider.Profile").Preload("Request").First(&p, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *proposalRepository) GetByRequestAndProvider(ctx context.Context, requestID, providerID uuid.UUID) (*entity.Proposal, error) {
	var p entity.Proposal
	err := r.db.WithContext(ctx).Preload("Provider").Preload("Provider.Profile").Preload("Request").
		Where("request_id = ? AND provider_id = ?", requestID, providerID).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *proposalRepository) ListByRequest(ctx context.Context, requestID uuid.UUID) ([]entity.Proposal, error) {
	var list []entity.Proposal
	err := r.db.WithContext(ctx).
		Preload("Provider").
		Preload("Provider.Profile").
		Preload("Request").
		Preload("Request.User").
		Preload("Request.User.Profile").
		Where("request_id = ?", requestID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

func (r *proposalRepository) ListByProvider(ctx context.Context, providerID uuid.UUID) ([]entity.Proposal, error) {
	var list []entity.Proposal
	err := r.db.WithContext(ctx).
		Preload("Provider").
		Preload("Provider.Profile").
		Preload("Request").
		Preload("Request.User").
		Preload("Request.User.Profile").
		Where("provider_id = ?", providerID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

func (r *proposalRepository) Update(ctx context.Context, proposal *entity.Proposal) error {
	return r.db.WithContext(ctx).Model(&entity.Proposal{}).Where("id = ?", proposal.ID).Updates(map[string]interface{}{
		"is_approved": proposal.IsApproved,
		"updated_at":  time.Now(),
	}).Error
}

func (r *proposalRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Proposal{}, "id = ?", id).Error
}
