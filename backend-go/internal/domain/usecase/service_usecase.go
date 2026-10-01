package usecase

import (
	"context"

	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
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
	GetRequests(ctx context.Context, userID uuid.UUID, userType string, status string, targetedOnly bool, optParams ...repository.ServiceRequestFilterParams) ([]entity.ServiceRequest, error)
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

	// Emergency Flash Dispatch
	CreateFlashDispatch(ctx context.Context, seekerID uuid.UUID, dispatch *entity.FlashDispatch) (*entity.FlashDispatch, error)
	GetFlashDispatch(ctx context.Context, id uuid.UUID) (*entity.FlashDispatch, error)
	AcceptFlashDispatch(ctx context.Context, providerID, dispatchID uuid.UUID) (*entity.FlashDispatch, *entity.Appointment, error)
	CancelFlashDispatch(ctx context.Context, seekerID, dispatchID uuid.UUID) error

	// AI Continuous Learning & Catalog Suggestions
	GetAISuggestions(ctx context.Context) ([]entity.CatalogKnowledgeIndex, error)
	TriggerCatalogReindex(ctx context.Context) error
	ParseAndRefineVoiceSpeech(ctx context.Context, userID *uuid.UUID, rawSpeech string) (*entity.VoiceSpeechLog, error)
}
