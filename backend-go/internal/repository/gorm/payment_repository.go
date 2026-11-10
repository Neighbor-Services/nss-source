package gorm

import (
	"context"
	"errors"
	"strings"
	"time"

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
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Unscoped().Delete(&entity.UserSubscription{}).Error
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

type ledgerRepository struct {
	db *gorm.DB
}

func NewLedgerRepository(db *gorm.DB) repository.LedgerRepository {
	return &ledgerRepository{db: db}
}

func (r *ledgerRepository) GetOrCreateAccount(ctx context.Context, code, name, accType, currency string) (*entity.LedgerAccount, error) {
	var acc entity.LedgerAccount
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&acc).Error
	if err == nil {
		return &acc, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if currency == "" {
			currency = "USD"
		}
		newAcc := entity.LedgerAccount{
			ID:       uuid.New(),
			Code:     code,
			Name:     name,
			Type:     accType,
			Currency: currency,
			Balance:  0.0,
		}
		if err := r.db.WithContext(ctx).Create(&newAcc).Error; err != nil {
			return nil, err
		}
		return &newAcc, nil
	}
	return nil, err
}

func (r *ledgerRepository) RecordEntry(ctx context.Context, entry *entity.LedgerEntry) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if entry.ID == uuid.Nil {
			entry.ID = uuid.New()
		}
		if entry.TransactionID == uuid.Nil {
			entry.TransactionID = uuid.New()
		}

		if err := tx.Create(entry).Error; err != nil {
			return err
		}

		// Update Debit Account Balance (+ balance for Assets/Expenses, - balance for Liabilities/Equity/Revenue)
		if err := tx.Model(&entity.LedgerAccount{}).Where("id = ?", entry.DebitAccountID).
			UpdateColumn("balance", gorm.Expr("balance + ?", entry.Amount)).Error; err != nil {
			return err
		}

		// Update Credit Account Balance
		if err := tx.Model(&entity.LedgerAccount{}).Where("id = ?", entry.CreditAccountID).
			UpdateColumn("balance", gorm.Expr("balance - ?", entry.Amount)).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *ledgerRepository) ListEntriesByAccount(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]entity.LedgerEntry, int64, error) {
	var entries []entity.LedgerEntry
	var total int64

	q := r.db.WithContext(ctx).Model(&entity.LedgerEntry{}).
		Where("debit_account_id = ? OR credit_account_id = ?", accountID, accountID)

	_ = q.Count(&total).Error
	err := q.Preload("DebitAccount").Preload("CreditAccount").
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&entries).Error

	return entries, total, err
}

func (r *ledgerRepository) GetPlatformFinancialSummary(ctx context.Context) (map[string]float64, error) {
	summary := make(map[string]float64)
	var accounts []entity.LedgerAccount
	if err := r.db.WithContext(ctx).Find(&accounts).Error; err != nil {
		return nil, err
	}
	for _, acc := range accounts {
		summary[acc.Code] = acc.Balance
	}
	return summary, nil
}

type refundRequestRepository struct {
	db *gorm.DB
}

func NewRefundRequestRepository(db *gorm.DB) repository.RefundRequestRepository {
	return &refundRequestRepository{db: db}
}

func (r *refundRequestRepository) Create(ctx context.Context, req *entity.RefundRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *refundRequestRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.RefundRequest, error) {
	var req entity.RefundRequest
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("User.Profile").
		Preload("Appointment").
		Preload("Admin").
		First(&req, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *refundRequestRepository) GetByPaymentIntentID(ctx context.Context, piID string) (*entity.RefundRequest, error) {
	var req entity.RefundRequest
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("User.Profile").
		Preload("Appointment").
		Preload("Admin").
		Order("created_at DESC").
		First(&req, "payment_intent_id = ?", piID).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *refundRequestRepository) GetByAppointmentID(ctx context.Context, appointmentID uuid.UUID) (*entity.RefundRequest, error) {
	var req entity.RefundRequest
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("User.Profile").
		Preload("Appointment").
		Preload("Admin").
		Order("created_at DESC").
		First(&req, "appointment_id = ?", appointmentID).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *refundRequestRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]entity.RefundRequest, error) {
	var list []entity.RefundRequest
	err := r.db.WithContext(ctx).
		Preload("Appointment").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

func (r *refundRequestRepository) List(ctx context.Context, status string, limit, offset int) ([]entity.RefundRequest, int64, error) {
	var list []entity.RefundRequest
	var total int64

	q := r.db.WithContext(ctx).Model(&entity.RefundRequest{})
	if status != "" {
		q = q.Where("status = ?", status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	err := q.Preload("User").
		Preload("User.Profile").
		Preload("Appointment").
		Preload("Admin").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&list).Error

	return list, total, err
}

func (r *refundRequestRepository) Update(ctx context.Context, req *entity.RefundRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}

type stripePaymentRepository struct {
	db *gorm.DB
}

func NewStripePaymentRepository(db *gorm.DB) repository.StripePaymentRepository {
	return &stripePaymentRepository{db: db}
}

func (r *stripePaymentRepository) Create(ctx context.Context, p *entity.StripePayment) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *stripePaymentRepository) Update(ctx context.Context, p *entity.StripePayment) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *stripePaymentRepository) Upsert(ctx context.Context, p *entity.StripePayment) error {
	var existing entity.StripePayment
	err := r.db.WithContext(ctx).Where("payment_intent_id = ?", p.PaymentIntentID).First(&existing).Error
	if err == nil {
		p.ID = existing.ID
		if p.Amount <= 0 {
			p.Amount = existing.Amount
		}
		if p.AmountRefunded == 0 && existing.AmountRefunded > 0 {
			p.AmountRefunded = existing.AmountRefunded
		}
		if p.UserID == nil && existing.UserID != nil {
			p.UserID = existing.UserID
		}
		if p.UserEmail == "" && existing.UserEmail != "" {
			p.UserEmail = existing.UserEmail
		}
		if p.UserName == "" && existing.UserName != "" {
			p.UserName = existing.UserName
		}
		if p.AppointmentID == nil && existing.AppointmentID != nil {
			p.AppointmentID = existing.AppointmentID
		}
		if p.ReceiptURL == "" && existing.ReceiptURL != "" {
			p.ReceiptURL = existing.ReceiptURL
		}
		if p.CardBrand == "" && existing.CardBrand != "" {
			p.CardBrand = existing.CardBrand
		}
		if p.CardLast4 == "" && existing.CardLast4 != "" {
			p.CardLast4 = existing.CardLast4
		}
		p.CreatedAt = existing.CreatedAt
		p.UpdatedAt = time.Now()
		return r.db.WithContext(ctx).Save(p).Error
	}
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *stripePaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.StripePayment, error) {
	var p entity.StripePayment
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("User.Profile").
		Preload("Appointment").
		Preload("Appointment.Seeker").
		Preload("Appointment.Seeker.Profile").
		Preload("Appointment.Provider").
		Preload("Appointment.Provider.Profile").
		Where("id = ?", id).
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *stripePaymentRepository) GetByPaymentIntentID(ctx context.Context, piID string) (*entity.StripePayment, error) {
	var p entity.StripePayment
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("User.Profile").
		Preload("Appointment").
		Where("payment_intent_id = ?", piID).
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *stripePaymentRepository) List(ctx context.Context, status, paymentType, search string, limit, offset int) ([]entity.StripePayment, int64, error) {
	var list []entity.StripePayment
	var total int64

	q := r.db.WithContext(ctx).Model(&entity.StripePayment{})
	if status != "" && status != "ALL" {
		q = q.Where("status = ?", status)
	}
	if paymentType != "" && paymentType != "ALL" {
		q = q.Where("payment_type = ?", paymentType)
	}
	if search != "" {
		s := "%" + strings.ToLower(search) + "%"
		q = q.Where("LOWER(payment_intent_id) LIKE ? OR LOWER(user_email) LIKE ? OR LOWER(user_name) LIKE ? OR LOWER(description) LIKE ?", s, s, s, s)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	err := q.Preload("User").
		Preload("User.Profile").
		Preload("Appointment").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&list).Error

	return list, total, err
}

func (r *stripePaymentRepository) GetStats(ctx context.Context) (*entity.StripePaymentStats, error) {
	var stats entity.StripePaymentStats

	// Total count
	_ = r.db.WithContext(ctx).Model(&entity.StripePayment{}).Count(&stats.TotalCount).Error

	// Total volume
	_ = r.db.WithContext(ctx).Model(&entity.StripePayment{}).
		Where("status IN ('succeeded', 'refunded', 'partially_refunded')").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.TotalVolume).Error

	// Succeeded count
	_ = r.db.WithContext(ctx).Model(&entity.StripePayment{}).
		Where("status = 'succeeded'").
		Count(&stats.SucceededCount).Error

	// Refunded count
	_ = r.db.WithContext(ctx).Model(&entity.StripePayment{}).
		Where("status IN ('refunded', 'partially_refunded')").
		Count(&stats.RefundedCount).Error

	// Total refunded amount
	_ = r.db.WithContext(ctx).Model(&entity.StripePayment{}).
		Select("COALESCE(SUM(amount_refunded), 0)").
		Scan(&stats.TotalRefundedAmount).Error

	// Pending count
	_ = r.db.WithContext(ctx).Model(&entity.StripePayment{}).
		Where("status IN ('processing', 'requires_payment_method', 'requires_action')").
		Count(&stats.PendingCount).Error

	return &stats, nil
}

