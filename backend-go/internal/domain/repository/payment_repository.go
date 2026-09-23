package repository

import (
	"context"

	"backend-go/internal/domain/entity"
	"github.com/google/uuid"
)

type SubscriptionPlanRepository interface {
	ListActive(ctx context.Context) ([]entity.SubscriptionPlan, error)
	GetByProductID(ctx context.Context, productID string) (*entity.SubscriptionPlan, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.SubscriptionPlan, error)
}

type UserSubscriptionRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.UserSubscription, error)
	GetByOriginalTxID(ctx context.Context, originalTxID string) (*entity.UserSubscription, error)
	GetByPurchaseToken(ctx context.Context, purchaseToken string) (*entity.UserSubscription, error)
	Upsert(ctx context.Context, sub *entity.UserSubscription) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

type WalletRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Wallet, error)
	Create(ctx context.Context, wallet *entity.Wallet) error
	Update(ctx context.Context, wallet *entity.Wallet) error
}

type WalletTransactionRepository interface {
	Create(ctx context.Context, tx *entity.WalletTransaction) error
	ListByWalletID(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]entity.WalletTransaction, error)
}

type CustomerRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Customer, error)
	Upsert(ctx context.Context, customer *entity.Customer) error
}

type PayoutRequestRepository interface {
	Create(ctx context.Context, req *entity.PayoutRequest) error
	ListByWalletID(ctx context.Context, walletID uuid.UUID) ([]entity.PayoutRequest, error)
}
