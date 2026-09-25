package usecase_test

import (
	"context"
	"testing"
	"time"

	"backend-go/internal/domain/entity"
	"backend-go/internal/usecase"
	"github.com/google/uuid"
)

type mockAppointmentRepo struct {
	apts map[uuid.UUID]*entity.Appointment
}

func (m *mockAppointmentRepo) List(ctx context.Context, seekerID *uuid.UUID, providerID *uuid.UUID, status string) ([]entity.Appointment, error) {
	var list []entity.Appointment
	for _, a := range m.apts {
		list = append(list, *a)
	}
	return list, nil
}

func (m *mockAppointmentRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Appointment, error) {
	if a, ok := m.apts[id]; ok {
		return a, nil
	}
	return nil, nil
}

func (m *mockAppointmentRepo) Create(ctx context.Context, apt *entity.Appointment) error {
	m.apts[apt.ID] = apt
	return nil
}

func (m *mockAppointmentRepo) Update(ctx context.Context, apt *entity.Appointment) error {
	m.apts[apt.ID] = apt
	return nil
}

func (m *mockAppointmentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.apts, id)
	return nil
}

func (m *mockAppointmentRepo) CheckConflict(ctx context.Context, providerID uuid.UUID, scheduledTime time.Time) (bool, error) {
	return false, nil
}

type mockReviewRepo struct{}

func (m *mockReviewRepo) ListByProvider(ctx context.Context, providerID uuid.UUID) ([]entity.Review, error) {
	return nil, nil
}
func (m *mockReviewRepo) Create(ctx context.Context, r *entity.Review) error { return nil }
func (m *mockReviewRepo) CalculateProviderRating(ctx context.Context, providerID uuid.UUID) (float64, int, error) {
	return 5.0, 1, nil
}

type mockFavoriteRepo struct{}

func (m *mockFavoriteRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Favorite, error) {
	return nil, nil
}
func (m *mockFavoriteRepo) Create(ctx context.Context, f *entity.Favorite) error { return nil }
func (m *mockFavoriteRepo) Delete(ctx context.Context, seekerID, providerID uuid.UUID) error {
	return nil
}

type mockDisputeRepo struct{}

func (m *mockDisputeRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Dispute, error) {
	return nil, nil
}
func (m *mockDisputeRepo) Create(ctx context.Context, d *entity.Dispute) error { return nil }
func (m *mockDisputeRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Dispute, error) {
	return nil, nil
}
func (m *mockDisputeRepo) Update(ctx context.Context, d *entity.Dispute) error {
	return nil
}

type mockWalletTxRepo struct {
	txs []*entity.WalletTransaction
}

func (m *mockWalletTxRepo) Create(ctx context.Context, tx *entity.WalletTransaction) error {
	m.txs = append(m.txs, tx)
	return nil
}
func (m *mockWalletTxRepo) ListByWalletID(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]entity.WalletTransaction, error) {
	return nil, nil
}

func TestInteractionUseCase_VerifyArrivalAndCompleteAppointment(t *testing.T) {
	providerID := uuid.New()
	seekerID := uuid.New()
	aptID := uuid.New()
	now := time.Now()

	apt := &entity.Appointment{
		ID:              aptID,
		ProviderID:      providerID,
		SeekerID:        seekerID,
		Status:          "CONFIRMED",
		SecretCode:      "123456",
		IsFunded:        true,
		TotalPrice:      100.0,
		AppointmentDate: &now,
	}

	aptRepo := &mockAppointmentRepo{apts: map[uuid.UUID]*entity.Appointment{aptID: apt}}
	favRepo := &mockFavoriteRepo{}
	reviewRepo := &mockReviewRepo{}
	disputeRepo := &mockDisputeRepo{}
	profileRepo := &mockProfileRepo{profiles: map[uuid.UUID]*entity.Profile{
		providerID: {ID: uuid.New(), UserID: providerID, SubscriptionTier: "SILVER"},
	}}
	walletRepo := &mockWalletRepo{}
	txRepo := &mockWalletTxRepo{}
	userRepo := &mockUserRepo{users: make(map[string]*entity.User)}

	interUC := usecase.NewInteractionUseCase(favRepo, reviewRepo, aptRepo, disputeRepo, profileRepo, walletRepo, txRepo, userRepo, nil, nil, nil, nil)
	ctx := context.Background()

	// 1. Test Notify On The Way
	notifiedApt, err := interUC.NotifyOnTheWay(ctx, providerID, aptID)
	if err != nil {
		t.Fatalf("Failed to notify on the way: %v", err)
	}
	if notifiedApt.ID != aptID {
		t.Fatalf("Expected appointment ID %s, got %s", aptID, notifiedApt.ID)
	}

	// 2. Test Verify Arrival Code
	_, err = interUC.VerifyArrivalCode(ctx, providerID, aptID, "123456")
	if err != nil {
		t.Fatalf("Failed to verify arrival code: %v", err)
	}
	if apt.Status != "IN_PROGRESS" {
		t.Fatalf("Expected status IN_PROGRESS, got %s", apt.Status)
	}

	// 2. Test Complete Appointment with Silver tier (15% commission)
	released, commission, err := interUC.CompleteAppointment(ctx, seekerID, aptID, 100.0)
	if err != nil {
		t.Fatalf("Failed to complete appointment: %v", err)
	}

	if released != 85.0 {
		t.Errorf("Expected released funds 85.0, got %f", released)
	}
	if commission != 15.0 {
		t.Errorf("Expected commission 15.0, got %f", commission)
	}
	if apt.Status != "COMPLETED" {
		t.Errorf("Expected status COMPLETED, got %s", apt.Status)
	}
}
