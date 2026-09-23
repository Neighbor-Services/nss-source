package repository

import (
	"context"

	"backend-go/internal/domain/entity"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByEmail(ctx context.Context, email string) (int64, error)
}

type ProfileFilterParams struct {
	UserType     string
	Popular      bool
	Search       string
	CategorySlug string
	CategoryName string
	ServiceName  string
	ServiceID    string
	RatingMin    *float64
	PriceMin     *float64
	PriceMax     *float64
	City         string
	Latitude     *float64
	Longitude    *float64
	RadiusKm     *float64
	Limit        int
	Offset       int
}

type ProfileRepository interface {
	Create(ctx context.Context, profile *entity.Profile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Profile, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Profile, error)
	Update(ctx context.Context, profile *entity.Profile) error
	UpdateCatalogServices(ctx context.Context, profileID uuid.UUID, serviceIDs []uuid.UUID) error
	List(ctx context.Context, params ProfileFilterParams) ([]entity.Profile, error)
	ListProviders(ctx context.Context, categorySlug string, limit int, offset int) ([]entity.Profile, error)
}

type AboutRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.About, error)
	Upsert(ctx context.Context, about *entity.About) error
}

type PortfolioRepository interface {
	ListByProfileID(ctx context.Context, profileID uuid.UUID) ([]entity.Portfolio, error)
	Create(ctx context.Context, portfolio *entity.Portfolio) error
	Delete(ctx context.Context, id uuid.UUID, profileID uuid.UUID) error
}

type ServicePackageRepository interface {
	ListByProfileID(ctx context.Context, profileID uuid.UUID) ([]entity.ServicePackage, error)
	Create(ctx context.Context, pkg *entity.ServicePackage) error
	Update(ctx context.Context, pkg *entity.ServicePackage) error
	Delete(ctx context.Context, id uuid.UUID, profileID uuid.UUID) error
}

type LegalDocumentRepository interface {
	ListActive(ctx context.Context) ([]entity.LegalDocument, error)
	GetByType(ctx context.Context, docType string) ([]entity.LegalDocument, error)
}
