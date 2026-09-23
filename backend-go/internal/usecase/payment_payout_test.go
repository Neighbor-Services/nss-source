package usecase_test

import (
	"context"
	"testing"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/internal/usecase"
	"github.com/google/uuid"
)

func TestPaymentUseCase_PayoutAndOnboarding(t *testing.T) {
	userID := uuid.New()
	planRepo := &mockPlanRepo{}
	subRepo := &mockUserSubRepo{subs: make(map[uuid.UUID]*entity.UserSubscription)}
	walletRepo := &mockWalletRepo{wallets: map[uuid.UUID]*entity.Wallet{
		userID: {ID: uuid.New(), UserID: userID, Balance: 200.0, Currency: "USD", StripeConnectID: "acct_test123"},
	}}
	txRepo := &mockWalletTxRepo{}
	payoutRepo := &mockPayoutRepo{}
	custRepo := &mockCustomerRepo{custs: make(map[uuid.UUID]*entity.Customer)}
	profileRepo := &mockProfileRepo{profiles: map[uuid.UUID]*entity.Profile{
		userID: {ID: uuid.New(), UserID: userID, SubscriptionTier: "PLATINUM"},
	}}
	cfg := &config.Config{
		StripeSecretKey:      "",
		StripePublishableKey: "pk_test_12345",
	}

	payUC := usecase.NewPaymentUseCase(planRepo, subRepo, walletRepo, txRepo, payoutRepo, custRepo, profileRepo, nil, nil, nil, cfg)
	ctx := context.Background()

	// 1. Test Request Payout
	res, err := payUC.RequestPayout(ctx, userID, 50.0)
	if err != nil {
		t.Fatalf("Failed to request payout: %v", err)
	}
	if res["status"] != "Payout successful" {
		t.Errorf("Expected Payout successful, got %v", res["status"])
	}

	// Verify balance deduction
	wallet, _ := walletRepo.GetByUserID(ctx, userID)
	if wallet.Balance != 150.0 {
		t.Errorf("Expected balance 150.0, got %f", wallet.Balance)
	}

	// 2. Test Onboarding status
	status, err := payUC.GetOnboardingStatus(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get onboarding status: %v", err)
	}
	if !status {
		t.Errorf("Expected onboarding status true")
	}

	// 3. Test Fund Background Check
	bgFund, err := payUC.FundBackgroundCheck(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to fund background check: %v", err)
	}
	if bgFund["paymentIntent"] == "" {
		t.Errorf("Expected paymentIntent client secret")
	}
}
