package gorm

import (
	"context"
	"errors"

	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type subscriptionPlanRepository struct {
	db *gorm.DB
}

func NewSubscriptionPlanRepository(db *gorm.DB) repository.SubscriptionPlanRepository {
	return &subscriptionPlanRepository{db: db}
}

func (r *subscriptionPlanRepository) ListActive(ctx context.Context) ([]entity.SubscriptionPlan, error) {
	var plans []entity.SubscriptionPlan
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("display_order ASC, price ASC").Find(&plans).Error
	return plans, err
}

func (r *subscriptionPlanRepository) GetByProductID(ctx context.Context, productID string) (*entity.SubscriptionPlan, error) {
	var plan entity.SubscriptionPlan
	err := r.db.WithContext(ctx).
		Where("apple_product_id = ? OR google_product_id = ? OR name = ?", productID, productID, productID).
		First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *subscriptionPlanRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.SubscriptionPlan, error) {
	var plan entity.SubscriptionPlan
	err := r.db.WithContext(ctx).First(&plan, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

type userSubscriptionRepository struct {
	db *gorm.DB
}

func NewUserSubscriptionRepository(db *gorm.DB) repository.UserSubscriptionRepository {
	return &userSubscriptionRepository{db: db}
}

func (r *userSubscriptionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.UserSubscription, error) {
	var sub entity.UserSubscription
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Preload("User").
		First(&sub, "user_id = ?", userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

func (r *userSubscriptionRepository) GetByOriginalTxID(ctx context.Context, originalTxID string) (*entity.UserSubscription, error) {
	var sub entity.UserSubscription
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Preload("User").
		First(&sub, "store_transaction_id = ?", originalTxID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

func (r *userSubscriptionRepository) GetByPurchaseToken(ctx context.Context, purchaseToken string) (*entity.UserSubscription, error) {
	var sub entity.UserSubscription
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Preload("User").
		First(&sub, "store_transaction_id = ?", purchaseToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

func (r *userSubscriptionRepository) Upsert(ctx context.Context, sub *entity.UserSubscription) error {
	var existing entity.UserSubscription
	err := r.db.WithContext(ctx).Where("user_id = ?", sub.UserID).First(&existing).Error
	if err == nil {
		sub.ID = existing.ID
		return r.db.WithContext(ctx).Save(sub).Error
	}
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *userSubscriptionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&entity.UserSubscription{}).Error
}

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) repository.WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error) {
	var wallet entity.Wallet
	err := r.db.WithContext(ctx).Preload("User").First(&wallet, "user_id = ?", userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newWallet := entity.Wallet{
				ID:       uuid.New(),
				UserID:   userID,
				Balance:  0.0,
				Currency: "USD",
			}
			if createErr := r.db.WithContext(ctx).Create(&newWallet).Error; createErr == nil {
				return &newWallet, nil
			}
		}
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Wallet, error) {
	var wallet entity.Wallet
	err := r.db.WithContext(ctx).Preload("User").First(&wallet, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) Create(ctx context.Context, wallet *entity.Wallet) error {
	return r.db.WithContext(ctx).Create(wallet).Error
}

func (r *walletRepository) Update(ctx context.Context, wallet *entity.Wallet) error {
	return r.db.WithContext(ctx).Save(wallet).Error
}

type walletTxRepository struct {
	db *gorm.DB
}

func NewWalletTransactionRepository(db *gorm.DB) repository.WalletTransactionRepository {
	return &walletTxRepository{db: db}
}

func (r *walletTxRepository) Create(ctx context.Context, tx *entity.WalletTransaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *walletTxRepository) ListByWalletID(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]entity.WalletTransaction, error) {
	var list []entity.WalletTransaction
	query := r.db.WithContext(ctx).Where("wallet_id = ?", walletID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Find(&list).Error
	return list, err
}

type customerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) repository.CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Customer, error) {
	var cust entity.Customer
	err := r.db.WithContext(ctx).First(&cust, "user_id = ?", userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newCust := entity.Customer{
				ID:     uuid.New(),
				UserID: userID,
			}
			_ = r.db.WithContext(ctx).Create(&newCust)
			return &newCust, nil
		}
		return nil, err
	}
	return &cust, nil
}

func (r *customerRepository) Upsert(ctx context.Context, customer *entity.Customer) error {
	var existing entity.Customer
	err := r.db.WithContext(ctx).Where("user_id = ?", customer.UserID).First(&existing).Error
	if err == nil {
		customer.ID = existing.ID
		return r.db.WithContext(ctx).Save(customer).Error
	}
	return r.db.WithContext(ctx).Create(customer).Error
}

type payoutRequestRepository struct {
	db *gorm.DB
}

func NewPayoutRequestRepository(db *gorm.DB) repository.PayoutRequestRepository {
	return &payoutRequestRepository{db: db}
}

func (r *payoutRequestRepository) Create(ctx context.Context, req *entity.PayoutRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *payoutRequestRepository) ListByWalletID(ctx context.Context, walletID uuid.UUID) ([]entity.PayoutRequest, error) {
	var list []entity.PayoutRequest
	err := r.db.WithContext(ctx).Where("wallet_id = ?", walletID).Order("created_at DESC").Find(&list).Error
	return list, err
}
