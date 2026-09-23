package usecase_test

import (
	"context"
	"testing"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/internal/usecase"
	"github.com/google/uuid"
)

type mockPlanRepo struct {
	plans []entity.SubscriptionPlan
}

func (m *mockPlanRepo) ListActive(ctx context.Context) ([]entity.SubscriptionPlan, error) {
	return m.plans, nil
}
func (m *mockPlanRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.SubscriptionPlan, error) {
	for _, p := range m.plans {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, nil
}
func (m *mockPlanRepo) GetByProductID(ctx context.Context, productID string) (*entity.SubscriptionPlan, error) {
	for _, p := range m.plans {
		if p.AppleProductID == productID || p.GoogleProductID == productID {
			return &p, nil
		}
	}
	return nil, nil
}

type mockUserSubRepo struct {
	subs map[uuid.UUID]*entity.UserSubscription
}

func (m *mockUserSubRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.UserSubscription, error) {
	return m.subs[userID], nil
}
func (m *mockUserSubRepo) GetByOriginalTxID(ctx context.Context, txID string) (*entity.UserSubscription, error) {
	for _, s := range m.subs {
		if s.StoreTransactionID == txID {
			return s, nil
		}
	}
	return nil, nil
}
func (m *mockUserSubRepo) GetByPurchaseToken(ctx context.Context, token string) (*entity.UserSubscription, error) {
	return m.GetByOriginalTxID(ctx, token)
}
func (m *mockUserSubRepo) Upsert(ctx context.Context, sub *entity.UserSubscription) error {
	m.subs[sub.UserID] = sub
	return nil
}
func (m *mockUserSubRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	delete(m.subs, userID)
	return nil
}

type mockCustomerRepo struct {
	custs map[uuid.UUID]*entity.Customer
}

func (m *mockCustomerRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Customer, error) {
	return m.custs[userID], nil
}
func (m *mockCustomerRepo) Upsert(ctx context.Context, customer *entity.Customer) error {
	m.custs[customer.UserID] = customer
	return nil
}

type mockPayoutRepo struct {
	payouts []*entity.PayoutRequest
}

func (m *mockPayoutRepo) Create(ctx context.Context, req *entity.PayoutRequest) error {
	m.payouts = append(m.payouts, req)
	return nil
}
func (m *mockPayoutRepo) ListByWalletID(ctx context.Context, walletID uuid.UUID) ([]entity.PayoutRequest, error) {
	return nil, nil
}

func TestPaymentUseCase_PaymentSheetAndCustomer(t *testing.T) {
	userID := uuid.New()
	planRepo := &mockPlanRepo{}
	subRepo := &mockUserSubRepo{subs: make(map[uuid.UUID]*entity.UserSubscription)}
	walletRepo := &mockWalletRepo{}
	txRepo := &mockWalletTxRepo{}
	payoutRepo := &mockPayoutRepo{}
	custRepo := &mockCustomerRepo{custs: make(map[uuid.UUID]*entity.Customer)}
	profileRepo := &mockProfileRepo{profiles: map[uuid.UUID]*entity.Profile{
		userID: {ID: uuid.New(), UserID: userID, SubscriptionTier: "NONE"},
	}}
	cfg := &config.Config{
		StripeSecretKey:      "",
		StripePublishableKey: "pk_test_12345",
	}

	payUC := usecase.NewPaymentUseCase(planRepo, subRepo, walletRepo, txRepo, payoutRepo, custRepo, profileRepo, nil, nil, nil, cfg)
	ctx := context.Background()

	// 1. Create Customer
	cust, err := payUC.CreateCustomer(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to create customer: %v", err)
	}
	if cust.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, cust.UserID)
	}

	// 2. Generate PaymentSheet
	sheet, err := payUC.PaymentSheet(ctx, userID, 25.0)
	if err != nil {
		t.Fatalf("Failed to create payment sheet: %v", err)
	}
	if sheet["paymentIntent"] == "" {
		t.Errorf("Expected paymentIntent client secret")
	}
	if sheet["ephemeralKey"] == "" {
		t.Errorf("Expected ephemeralKey")
	}

	// 3. Validate Apple JWS / StoreKit token
	sub, err := payUC.ValidateAppleReceipt(ctx, userID, domainUsecase.AppleValidationInput{
		ReceiptData: "eyJhbGciOiJSUzI1NiJ9.test.payload",
		ProductID:   "silver_monthly",
	})
	if err != nil {
		t.Fatalf("Failed to validate Apple receipt: %v", err)
	}
	if !sub.IsActive {
		t.Errorf("Expected active subscription")
	}
}
