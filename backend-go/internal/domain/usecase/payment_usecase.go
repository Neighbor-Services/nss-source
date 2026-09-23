package usecase

import (
	"context"

	"backend-go/internal/domain/entity"
	"github.com/google/uuid"
)

type AppleValidationInput struct {
	ReceiptData string `json:"receipt_data"`
	ProductID   string `json:"product_id"`
}

type GoogleValidationInput struct {
	PackageName    string `json:"package_name"`
	ProductID      string `json:"product_id"`
	PurchaseToken  string `json:"purchase_token"`
	SubscriptionID string `json:"subscription_id"`
}

type CustomerResponse struct {
	ID                   uuid.UUID `json:"id"`
	UserID               uuid.UUID `json:"user_id"`
	StripeCustomerID     string    `json:"stripe_customer_id"`
	EphemeralSecret      string    `json:"ephemeral_secret,omitempty"`
	DefaultPaymentMethod string    `json:"default_payment_method,omitempty"`
	StripeAccountID      string    `json:"stripe_account_id,omitempty"`
}

type PaymentUseCase interface {
	GetSubscriptionPlans(ctx context.Context) ([]entity.SubscriptionPlan, error)
	ValidateAppleReceipt(ctx context.Context, userID uuid.UUID, input AppleValidationInput) (*entity.UserSubscription, error)
	ValidateGooglePlay(ctx context.Context, userID uuid.UUID, input GoogleValidationInput) (*entity.UserSubscription, error)
	GetWallet(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error)
	GetWalletTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.WalletTransaction, error)
	RequestPayout(ctx context.Context, userID uuid.UUID, amount float64) (map[string]interface{}, error)
	StripeOnboard(ctx context.Context, userID uuid.UUID, email string) (map[string]interface{}, error)
	GetOnboardingStatus(ctx context.Context, userID uuid.UUID) (bool, error)
	GetStripeDashboardLink(ctx context.Context, userID uuid.UUID) (string, error)
	GetUserSubscription(ctx context.Context, userID uuid.UUID) (*entity.UserSubscription, error)
	DeleteUserSubscription(ctx context.Context, userID uuid.UUID) error
	ProcessAppleS2SWebhook(ctx context.Context, payload map[string]interface{}) error
	ProcessGooglePubSubWebhook(ctx context.Context, payload map[string]interface{}) error

	// Customer & Payment Operations
	GetCustomer(ctx context.Context, userID uuid.UUID) (*entity.Customer, error)
	CreateCustomer(ctx context.Context, userID uuid.UUID) (*entity.Customer, error)
	CreateEphemeralKey(ctx context.Context, userID uuid.UUID) (string, error)
	UpdatePaymentMethod(ctx context.Context, userID uuid.UUID, paymentMethodID string) error
	AccountConnect(ctx context.Context, userID uuid.UUID) (string, error)
	Transfer(ctx context.Context, senderID, targetUserID uuid.UUID, amount float64) (string, error)
	PaymentSheet(ctx context.Context, userID uuid.UUID, amount float64) (map[string]string, error)
	FundAppointment(ctx context.Context, userID, appointmentID uuid.UUID, amount float64) (map[string]string, error)
	FundBackgroundCheck(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
	TipProvider(ctx context.Context, senderID, providerID, appointmentID uuid.UUID, amount float64) (string, error)
	ProcessStripeWebhook(ctx context.Context, payload []byte, signature string) error
}
