package usecase

import (
	"context"

	"backend-go/internal/domain/entity"
	"github.com/google/uuid"
)

type MatchProvidersInput struct {
	UserID       uuid.UUID `json:"-"`
	Description  string    `json:"description"`
	Query        string    `json:"query"`
	CategoryID   string    `json:"category_id"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	MaxDistance  float64   `json:"max_distance"`
	ScheduledFor string    `json:"scheduled_for"`
}

type ServiceUseCase interface {
	GetCategories(ctx context.Context) ([]entity.Category, error)
	GetCatalogServices(ctx context.Context, categorySlug string, search string) ([]entity.CatalogService, error)
	MatchProviders(ctx context.Context, input MatchProvidersInput) ([]entity.Profile, error)
	GetRequests(ctx context.Context, userID uuid.UUID, userType string, status string) ([]entity.ServiceRequest, error)
	GetRequestByID(ctx context.Context, id uuid.UUID) (*entity.ServiceRequest, error)
	CreateRequest(ctx context.Context, customerID uuid.UUID, req *entity.ServiceRequest) (*entity.ServiceRequest, error)
	UpdateRequest(ctx context.Context, userID, id uuid.UUID, updates map[string]interface{}) (*entity.ServiceRequest, error)
	UploadRequestImage(ctx context.Context, userID, id uuid.UUID, imageURL string) (*entity.ServiceRequest, error)
	DeleteRequest(ctx context.Context, userID, id uuid.UUID) error
	ApproveProposal(ctx context.Context, userID, requestID, proposalID uuid.UUID) (*entity.ServiceRequest, error)
	CancelApproval(ctx context.Context, userID, requestID uuid.UUID) error
	GetProposals(ctx context.Context, userID uuid.UUID, userType string, requestID *uuid.UUID) ([]entity.Proposal, error)
	CreateProposal(ctx context.Context, providerID uuid.UUID, prop *entity.Proposal) (*entity.Proposal, error)
	DeleteProposal(ctx context.Context, providerID, idOrReqID uuid.UUID) error
}
