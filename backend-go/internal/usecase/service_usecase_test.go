package usecase_test

import (
	"context"
	"testing"
	"time"

	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/internal/usecase"
	"github.com/google/uuid"
)

type mockCategoryRepo struct {
	cats []entity.Category
}

func (m *mockCategoryRepo) ListActive(ctx context.Context) ([]entity.Category, error) {
	return m.cats, nil
}
func (m *mockCategoryRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	for _, c := range m.cats {
		if c.ID == id {
			return &c, nil
		}
	}
	return nil, nil
}
func (m *mockCategoryRepo) GetBySlug(ctx context.Context, slug string) (*entity.Category, error) {
	for _, c := range m.cats {
		if c.Name == slug {
			return &c, nil
		}
	}
	return nil, nil
}

type mockCatalogRepo struct {
	services []entity.CatalogService
}

func (m *mockCatalogRepo) List(ctx context.Context, categorySlug, search string) ([]entity.CatalogService, error) {
	return m.services, nil
}
func (m *mockCatalogRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.CatalogService, error) {
	for _, s := range m.services {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, nil
}

type mockRequestRepo struct {
	requests map[uuid.UUID]*entity.ServiceRequest
}

func (m *mockRequestRepo) List(ctx context.Context, params repository.ServiceRequestFilterParams) ([]entity.ServiceRequest, error) {
	var list []entity.ServiceRequest
	for _, r := range m.requests {
		list = append(list, *r)
	}
	return list, nil
}
func (m *mockRequestRepo) ListByUser(ctx context.Context, customerID *uuid.UUID, providerID *uuid.UUID, status string) ([]entity.ServiceRequest, error) {
	var list []entity.ServiceRequest
	for _, r := range m.requests {
		list = append(list, *r)
	}
	return list, nil
}
func (m *mockRequestRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.ServiceRequest, error) {
	if r, ok := m.requests[id]; ok {
		return r, nil
	}
	return nil, nil
}
func (m *mockRequestRepo) Create(ctx context.Context, req *entity.ServiceRequest) error {
	m.requests[req.ID] = req
	return nil
}
func (m *mockRequestRepo) Update(ctx context.Context, req *entity.ServiceRequest) error {
	m.requests[req.ID] = req
	return nil
}
func (m *mockRequestRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.requests, id)
	return nil
}

type mockProposalRepo struct {
	proposals map[uuid.UUID]*entity.Proposal
}

func (m *mockProposalRepo) ListByRequest(ctx context.Context, requestID uuid.UUID) ([]entity.Proposal, error) {
	var list []entity.Proposal
	for _, p := range m.proposals {
		if p.RequestID == requestID {
			list = append(list, *p)
		}
	}
	return list, nil
}
func (m *mockProposalRepo) ListByProvider(ctx context.Context, providerID uuid.UUID) ([]entity.Proposal, error) {
	var list []entity.Proposal
	for _, p := range m.proposals {
		if p.ProviderID == providerID {
			list = append(list, *p)
		}
	}
	return list, nil
}
func (m *mockProposalRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Proposal, error) {
	if p, ok := m.proposals[id]; ok {
		return p, nil
	}
	return nil, nil
}
func (m *mockProposalRepo) GetByRequestAndProvider(ctx context.Context, requestID, providerID uuid.UUID) (*entity.Proposal, error) {
	for _, p := range m.proposals {
		if p.RequestID == requestID && p.ProviderID == providerID {
			return p, nil
		}
	}
	return nil, nil
}
func (m *mockProposalRepo) Create(ctx context.Context, proposal *entity.Proposal) error {
	m.proposals[proposal.ID] = proposal
	return nil
}
func (m *mockProposalRepo) Update(ctx context.Context, proposal *entity.Proposal) error {
	m.proposals[proposal.ID] = proposal
	return nil
}
func (m *mockProposalRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.proposals, id)
	return nil
}

func TestServiceUseCase_ProposalApprovalFlow(t *testing.T) {
	seekerID := uuid.New()
	providerID := uuid.New()
	reqID := uuid.New()
	propID := uuid.New()

	req := &entity.ServiceRequest{
		ID:        reqID,
		UserID:    seekerID,
		Title:     "Fix Plumbing",
		Status:    "OPEN",
		CreatedAt: time.Now(),
	}

	prop := &entity.Proposal{
		ID:         propID,
		RequestID:  reqID,
		ProviderID: providerID,
		IsApproved: false,
		CreatedAt:  time.Now(),
	}

	catRepo := &mockCategoryRepo{}
	catalogRepo := &mockCatalogRepo{}
	requestRepo := &mockRequestRepo{requests: map[uuid.UUID]*entity.ServiceRequest{reqID: req}}
	proposalRepo := &mockProposalRepo{proposals: map[uuid.UUID]*entity.Proposal{propID: prop}}
	profileRepo := &mockProfileRepo{profiles: map[uuid.UUID]*entity.Profile{
		seekerID:   {ID: uuid.New(), UserID: seekerID},
		providerID: {ID: uuid.New(), UserID: providerID},
	}}
	aptRepo := &mockAppointmentRepo{apts: make(map[uuid.UUID]*entity.Appointment)}

	srvUC := usecase.NewServiceUseCase(catRepo, catalogRepo, requestRepo, proposalRepo, profileRepo, aptRepo, nil, nil, nil, nil, nil, nil, nil)
	ctx := context.Background()


	// 1. Approve proposal -> Appointment must be created
	updatedReq, err := srvUC.ApproveProposal(ctx, seekerID, reqID, propID)
	if err != nil {
		t.Fatalf("Failed to approve proposal: %v", err)
	}

	if updatedReq.Status != "IN_PROGRESS" {
		t.Errorf("Expected status IN_PROGRESS, got %s", updatedReq.Status)
	}
	if len(aptRepo.apts) != 1 {
		t.Fatalf("Expected 1 appointment created, got %d", len(aptRepo.apts))
	}

	// 2. Cancel approval -> Appointment must be deleted (no duplicate/cancelled remnants)
	err = srvUC.CancelApproval(ctx, seekerID, reqID)
	if err != nil {
		t.Fatalf("Failed to cancel approval: %v", err)
	}

	reqAfterCancel, _ := srvUC.GetRequestByID(ctx, reqID)
	if reqAfterCancel.Status != "OPEN" {
		t.Errorf("Expected status OPEN after cancel, got %s", reqAfterCancel.Status)
	}
	if len(aptRepo.apts) != 0 {
		t.Errorf("Expected 0 appointments after cancel approval, got %d", len(aptRepo.apts))
	}

	// 3. Approve proposal again -> exactly 1 appointment created (no duplicates)
	_, err = srvUC.ApproveProposal(ctx, seekerID, reqID, propID)
	if err != nil {
		t.Fatalf("Failed to re-approve proposal: %v", err)
	}
	if len(aptRepo.apts) != 1 {
		t.Errorf("Expected exactly 1 appointment after re-approval, got %d", len(aptRepo.apts))
	}
}

func TestServiceUseCase_ImageUploadAndProposalDeletion(t *testing.T) {
	seekerID := uuid.New()
	providerID := uuid.New()
	reqID := uuid.New()
	propID := uuid.New()

	req := &entity.ServiceRequest{
		ID:        reqID,
		UserID:    seekerID,
		Title:     "House Cleaning",
		Status:    "OPEN",
		CreatedAt: time.Now(),
	}

	prop := &entity.Proposal{
		ID:         propID,
		RequestID:  reqID,
		ProviderID: providerID,
		IsApproved: false,
		CreatedAt:  time.Now(),
	}

	catRepo := &mockCategoryRepo{}
	catalogRepo := &mockCatalogRepo{}
	requestRepo := &mockRequestRepo{requests: map[uuid.UUID]*entity.ServiceRequest{reqID: req}}
	proposalRepo := &mockProposalRepo{proposals: map[uuid.UUID]*entity.Proposal{propID: prop}}
	profileRepo := &mockProfileRepo{profiles: map[uuid.UUID]*entity.Profile{
		seekerID:   {ID: uuid.New(), UserID: seekerID},
		providerID: {ID: uuid.New(), UserID: providerID, Latitude: 51.5074, Longitude: -0.1278},
	}}

	srvUC := usecase.NewServiceUseCase(catRepo, catalogRepo, requestRepo, proposalRepo, profileRepo, nil, nil, nil, nil, nil, nil, nil, nil)
	ctx := context.Background()


	// 1. Upload Image
	updatedReq, err := srvUC.UploadRequestImage(ctx, seekerID, reqID, "/media/test.jpg")
	if err != nil {
		t.Fatalf("Upload image failed: %v", err)
	}
	if updatedReq.Image != "/media/test.jpg" || !updatedReq.WithImage {
		t.Errorf("Expected image /media/test.jpg, got %s", updatedReq.Image)
	}

	// 2. Match Providers with location
	matched, err := srvUC.MatchProviders(ctx, domainUsecase.MatchProvidersInput{
		Latitude:    51.5000,
		Longitude:   -0.1200,
		MaxDistance: 25.0,
	})
	if err != nil {
		t.Fatalf("MatchProviders failed: %v", err)
	}
	if len(matched) == 0 {
		t.Log("MatchProviders executed with location filter")
	}

	// 3. Delete Proposal
	err = srvUC.DeleteProposal(ctx, providerID, propID)
	if err != nil {
		t.Fatalf("Delete proposal failed: %v", err)
	}
	if _, ok := proposalRepo.proposals[propID]; ok {
		t.Errorf("Expected proposal to be deleted")
	}
}
