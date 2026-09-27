package repository

import (
	"context"

	"backend-go/internal/domain/entity"
	"github.com/google/uuid"
)

type CategoryRepository interface {
	ListActive(ctx context.Context) ([]entity.Category, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Category, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)
}

type CatalogServiceRepository interface {
	List(ctx context.Context, categorySlug string, search string) ([]entity.CatalogService, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.CatalogService, error)
}

type ServiceRequestFilterParams struct {
	UserID             *uuid.UUID
	TargetProviderID   *uuid.UUID
	CatalogServiceID   *uuid.UUID
	CatalogServiceName string
	OfferedServices    []string
	Status             string
	Latitude           *float64
	Longitude          *float64
	RadiusKm           *float64
	Limit              int
	Offset             int
}

type ServiceRequestRepository interface {
	Create(ctx context.Context, req *entity.ServiceRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ServiceRequest, error)
	List(ctx context.Context, params ServiceRequestFilterParams) ([]entity.ServiceRequest, error)
	ListByUser(ctx context.Context, customerID *uuid.UUID, providerID *uuid.UUID, status string) ([]entity.ServiceRequest, error)
	Update(ctx context.Context, req *entity.ServiceRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ProposalRepository interface {
	Create(ctx context.Context, proposal *entity.Proposal) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Proposal, error)
	GetByRequestAndProvider(ctx context.Context, requestID, providerID uuid.UUID) (*entity.Proposal, error)
	ListByRequest(ctx context.Context, requestID uuid.UUID) ([]entity.Proposal, error)
	ListByProvider(ctx context.Context, providerID uuid.UUID) ([]entity.Proposal, error)
	Update(ctx context.Context, proposal *entity.Proposal) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type FlashDispatchRepository interface {
	Create(ctx context.Context, dispatch *entity.FlashDispatch) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.FlashDispatch, error)
	Update(ctx context.Context, dispatch *entity.FlashDispatch) error
}

type AIKnowledgeRepository interface {
	UpsertKnowledge(ctx context.Context, item *entity.CatalogKnowledgeIndex) error
	BatchUpsertKnowledge(ctx context.Context, items []entity.CatalogKnowledgeIndex) error
	GetAllKnowledge(ctx context.Context) ([]entity.CatalogKnowledgeIndex, error)
	GetKnowledgeByConcept(ctx context.Context, concept string) ([]entity.CatalogKnowledgeIndex, error)
	GetDynamicSuggestions(ctx context.Context, limit int) ([]entity.CatalogKnowledgeIndex, error)
	RecordSearchLog(ctx context.Context, log *entity.AISearchLog) error
	IncrementSearchUsage(ctx context.Context, queryOrKeyword string) error
}


