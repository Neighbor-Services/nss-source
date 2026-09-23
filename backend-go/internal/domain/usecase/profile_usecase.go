package usecase

import (
	"context"

	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	"github.com/google/uuid"
)

type ProfileUseCase interface {
	GetProfile(ctx context.Context, userID uuid.UUID, targetID *uuid.UUID) (*entity.Profile, error)
	ListProfiles(ctx context.Context, params repository.ProfileFilterParams) ([]entity.Profile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) (*entity.Profile, error)
	UpdateProfilePicture(ctx context.Context, userID uuid.UUID, imageURL string) (*entity.Profile, error)
	GetAbout(ctx context.Context, userID uuid.UUID) (*entity.About, error)
	UpdateAbout(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) (*entity.About, error)
	GetPortfolios(ctx context.Context, userID uuid.UUID) ([]entity.Portfolio, error)
	CreatePortfolio(ctx context.Context, userID uuid.UUID, item *entity.Portfolio) (*entity.Portfolio, error)
	DeletePortfolio(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error
	GetServicePackages(ctx context.Context, userID uuid.UUID) ([]entity.ServicePackage, error)
	CreateServicePackage(ctx context.Context, userID uuid.UUID, pkg *entity.ServicePackage) (*entity.ServicePackage, error)
	DeleteServicePackage(ctx context.Context, userID uuid.UUID, pkgID uuid.UUID) error
	GetLegalDocuments(ctx context.Context, docType string) ([]entity.LegalDocument, error)
}
