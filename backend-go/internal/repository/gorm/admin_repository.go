package gorm

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) repository.AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) GetDashboardStats(ctx context.Context) (*entity.AdminDashboardStats, error) {
	var stats entity.AdminDashboardStats

	// Counts
	_ = r.db.WithContext(ctx).Model(&entity.User{}).Count(&stats.TotalUsers)
	_ = r.db.WithContext(ctx).Model(&entity.Profile{}).Where("user_type ILIKE ?", "SEEKER").Count(&stats.TotalSeekers)
	_ = r.db.WithContext(ctx).Model(&entity.Profile{}).Where("user_type ILIKE ?", "PROVIDER").Count(&stats.TotalProviders)
	_ = r.db.WithContext(ctx).Model(&entity.User{}).Where("is_staff = ?", true).Count(&stats.TotalStaff)
	_ = r.db.WithContext(ctx).Model(&entity.Profile{}).Where("is_identity_verified = ?", true).Count(&stats.TotalVerified)
	_ = r.db.WithContext(ctx).Model(&entity.User{}).Where("is_active = ?", false).Count(&stats.TotalSuspended)
	_ = r.db.WithContext(ctx).Model(&entity.Subscription{}).Where("is_active = ?", true).Count(&stats.ActiveSubscriptions)

	// Wallet & Payout Stats
	var totalBal struct {
		Total float64
	}
	_ = r.db.WithContext(ctx).Model(&entity.Wallet{}).Select("COALESCE(SUM(balance), 0) as total").Scan(&totalBal)
	stats.TotalWalletBalance = totalBal.Total

	_ = r.db.WithContext(ctx).Model(&entity.PayoutRequest{}).Where("status = ?", "PENDING").Count(&stats.PendingPayoutsCount)
	var pendingPayoutAmt struct {
		Total float64
	}
	_ = r.db.WithContext(ctx).Model(&entity.PayoutRequest{}).Where("status = ?", "PENDING").Select("COALESCE(SUM(amount), 0) as total").Scan(&pendingPayoutAmt)
	stats.PendingPayoutsAmount = pendingPayoutAmt.Total

	// Appointments
	_ = r.db.WithContext(ctx).Model(&entity.Appointment{}).Count(&stats.AppointmentsCount)
	_ = r.db.WithContext(ctx).Model(&entity.Appointment{}).Where("status = ?", "COMPLETED").Count(&stats.AppointmentsCompleted)
	_ = r.db.WithContext(ctx).Model(&entity.Appointment{}).Where("status IN ?", []string{"SCHEDULED", "IN_PROGRESS"}).Count(&stats.AppointmentsPending)

	// Moderation & Disputes
	_ = r.db.WithContext(ctx).Model(&entity.Dispute{}).Where("status = ?", "OPEN").Count(&stats.OpenDisputesCount)
	_ = r.db.WithContext(ctx).Model(&entity.ProviderVerification{}).Where("status = ?", "PENDING").Count(&stats.PendingVerifications)
	_ = r.db.WithContext(ctx).Model(&entity.BackgroundCheck{}).Where("status = ?", "PENDING").Count(&stats.PendingBackground)

	// Recent Audit Logs
	_ = r.db.WithContext(ctx).Preload("User").Order("created_at DESC").Limit(10).Find(&stats.RecentAuditLogs)

	// Signups past 30 days
	stats.SignupsLast30Days = make(map[string]int64)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	type DateCount struct {
		Day   string
		Count int64
	}
	var dateCounts []DateCount
	_ = r.db.WithContext(ctx).Model(&entity.User{}).
		Select("TO_CHAR(created_at, 'YYYY-MM-DD') as day, count(*) as count").
		Where("created_at >= ?", thirtyDaysAgo).
		Group("TO_CHAR(created_at, 'YYYY-MM-DD')").
		Scan(&dateCounts)
	for _, dc := range dateCounts {
		stats.SignupsLast30Days[dc.Day] = dc.Count
	}

	return &stats, nil
}

func (r *adminRepository) ListUsers(ctx context.Context, filter repository.AdminUserFilter) ([]entity.User, int64, error) {
	var users []entity.User
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.User{}).
		Preload("Profile").
		Preload("Profile.CatalogServices").
		Preload("Profile.PortfolioItems").
		Preload("Profile.ServicePackages").
		Preload("Wallet").
		Preload("BackgroundCheck").
		Preload("AdminRole")

	needProfileJoin := filter.Search != "" || filter.UserType != "" || filter.IsIdentityVerified != nil || filter.SubscriptionTier != ""
	if needProfileJoin {
		query = query.Joins("LEFT JOIN accounts_profile ON accounts_profile.user_id = accounts_user.id")
	}

	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		query = query.Where("accounts_user.email ILIKE ? OR accounts_profile.first_name ILIKE ? OR accounts_profile.last_name ILIKE ?", s, s, s)
	}

	if filter.UserType != "" {
		query = query.Where("accounts_profile.user_type ILIKE ?", filter.UserType)
	}

	if filter.IsStaff != nil {
		query = query.Where("accounts_user.is_staff = ?", *filter.IsStaff)
	}
	if filter.IsActive != nil {
		query = query.Where("accounts_user.is_active = ?", *filter.IsActive)
	}
	if filter.IsVerified != nil {
		query = query.Where("accounts_user.is_verified = ?", *filter.IsVerified)
	}
	if filter.IsIdentityVerified != nil {
		query = query.Where("accounts_profile.is_identity_verified = ?", *filter.IsIdentityVerified)
	}
	if filter.SubscriptionTier != "" {
		query = query.Where("accounts_profile.subscription_tier ILIKE ?", filter.SubscriptionTier)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	err = query.Order("accounts_user.created_at DESC").Find(&users).Error
	return users, total, err
}

func (r *adminRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Profile").
		Preload("Profile.CatalogServices").
		Preload("Profile.PortfolioItems").
		Preload("Profile.ServicePackages").
		Preload("Wallet").
		Preload("BackgroundCheck").
		Preload("AdminRole").
		First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *adminRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *adminRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, "id = ?", id).Error
}

func (r *adminRepository) ListVerifications(ctx context.Context, status string, limit, offset int) ([]entity.ProviderVerification, int64, error) {
	var list []entity.ProviderVerification
	var total int64
	query := r.db.WithContext(ctx).Model(&entity.ProviderVerification{}).Preload("Provider").Preload("Provider.Profile")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	_ = query.Count(&total)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Order("created_at DESC").Find(&list).Error
	return list, total, err
}

func (r *adminRepository) GetVerificationByID(ctx context.Context, id uuid.UUID) (*entity.ProviderVerification, error) {
	var v entity.ProviderVerification
	err := r.db.WithContext(ctx).Preload("Provider").Preload("Provider.Profile").First(&v, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *adminRepository) UpdateVerification(ctx context.Context, v *entity.ProviderVerification) error {
	return r.db.WithContext(ctx).Save(v).Error
}

func (r *adminRepository) ListBackgroundChecks(ctx context.Context, status string, limit, offset int) ([]entity.BackgroundCheck, int64, error) {
	var list []entity.BackgroundCheck
	var total int64
	query := r.db.WithContext(ctx).Model(&entity.BackgroundCheck{}).Preload("Provider").Preload("Provider.Profile")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	_ = query.Count(&total)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Order("created_at DESC").Find(&list).Error
	return list, total, err
}

func (r *adminRepository) GetBackgroundCheckByID(ctx context.Context, id uuid.UUID) (*entity.BackgroundCheck, error) {
	var bc entity.BackgroundCheck
	err := r.db.WithContext(ctx).Preload("Provider").Preload("Provider.Profile").First(&bc, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &bc, nil
}

func (r *adminRepository) UpdateBackgroundCheck(ctx context.Context, bc *entity.BackgroundCheck) error {
	return r.db.WithContext(ctx).Save(bc).Error
}

func (r *adminRepository) ListReports(ctx context.Context, status string, limit, offset int) ([]entity.Report, int64, error) {
	var list []entity.Report
	var total int64
	query := r.db.WithContext(ctx).Model(&entity.Report{}).
		Preload("Reporter").
		Preload("Reporter.Profile").
		Preload("ReportedUser").
		Preload("ReportedUser.Profile")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	_ = query.Count(&total)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Order("created_at DESC").Find(&list).Error
	return list, total, err
}

func (r *adminRepository) GetReportByID(ctx context.Context, id uuid.UUID) (*entity.Report, error) {
	var rep entity.Report
	err := r.db.WithContext(ctx).
		Preload("Reporter").
		Preload("Reporter.Profile").
		Preload("ReportedUser").
		Preload("ReportedUser.Profile").
		First(&rep, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &rep, nil
}

func (r *adminRepository) UpdateReport(ctx context.Context, rep *entity.Report) error {
	return r.db.WithContext(ctx).Save(rep).Error
}

func (r *adminRepository) ListDisputes(ctx context.Context, status string, limit, offset int) ([]entity.Dispute, int64, error) {
	var list []entity.Dispute
	var total int64
	query := r.db.WithContext(ctx).Model(&entity.Dispute{}).
		Preload("RaisedBy").Preload("RaisedBy.Profile").
		Preload("Defendant").Preload("Defendant.Profile").
		Preload("Appointment")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	_ = query.Count(&total)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Order("created_at DESC").Find(&list).Error
	return list, total, err
}

func (r *adminRepository) GetDisputeByID(ctx context.Context, id uuid.UUID) (*entity.Dispute, error) {
	var d entity.Dispute
	err := r.db.WithContext(ctx).
		Preload("RaisedBy").Preload("RaisedBy.Profile").
		Preload("Defendant").Preload("Defendant.Profile").
		Preload("Appointment").
		First(&d, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *adminRepository) UpdateDispute(ctx context.Context, dispute *entity.Dispute) error {
	return r.db.WithContext(ctx).Save(dispute).Error
}

func (r *adminRepository) ListPayoutRequests(ctx context.Context, status string, limit, offset int) ([]entity.PayoutRequest, int64, error) {
	var list []entity.PayoutRequest
	var total int64
	query := r.db.WithContext(ctx).Model(&entity.PayoutRequest{}).Preload("Wallet").Preload("Wallet.User").Preload("Wallet.User.Profile")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	_ = query.Count(&total)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Order("created_at DESC").Find(&list).Error
	return list, total, err
}

func (r *adminRepository) GetPayoutRequestByID(ctx context.Context, id uuid.UUID) (*entity.PayoutRequest, error) {
	var pr entity.PayoutRequest
	err := r.db.WithContext(ctx).Preload("Wallet").Preload("Wallet.User").Preload("Wallet.User.Profile").First(&pr, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *adminRepository) UpdatePayoutRequest(ctx context.Context, req *entity.PayoutRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}

func (r *adminRepository) ListSubscriptions(ctx context.Context, isActive *bool, limit, offset int) ([]entity.Subscription, int64, error) {
	var list []entity.Subscription
	var total int64
	query := r.db.WithContext(ctx).Model(&entity.Subscription{}).Preload("User").Preload("User.Profile").Preload("Plan")
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}
	_ = query.Count(&total)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Order("created_at DESC").Find(&list).Error
	return list, total, err
}

func (r *adminRepository) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error) {
	var sub entity.Subscription
	err := r.db.WithContext(ctx).Preload("User").Preload("User.Profile").Preload("Plan").First(&sub, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *adminRepository) GetSubscriptionByUserID(ctx context.Context, userID uuid.UUID) (*entity.Subscription, error) {
	var sub entity.Subscription
	err := r.db.WithContext(ctx).Preload("User").Preload("User.Profile").Preload("Plan").First(&sub, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *adminRepository) CreateSubscription(ctx context.Context, sub *entity.Subscription) error {
	_ = r.db.AutoMigrate(&entity.Subscription{})
	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *adminRepository) UpdateSubscription(ctx context.Context, sub *entity.Subscription) error {
	return r.db.WithContext(ctx).Save(sub).Error
}

func (r *adminRepository) ListSubscriptionPlans(ctx context.Context) ([]entity.SubscriptionPlan, error) {
	_ = r.db.AutoMigrate(&entity.SubscriptionPlan{})
	var plans []entity.SubscriptionPlan
	err := r.db.WithContext(ctx).Order("display_order ASC, price ASC").Find(&plans).Error
	return plans, err
}

func (r *adminRepository) GetSubscriptionPlanByID(ctx context.Context, id uuid.UUID) (*entity.SubscriptionPlan, error) {
	var plan entity.SubscriptionPlan
	err := r.db.WithContext(ctx).First(&plan, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *adminRepository) CreateSubscriptionPlan(ctx context.Context, plan *entity.SubscriptionPlan) error {
	_ = r.db.AutoMigrate(&entity.SubscriptionPlan{})
	if plan.ID == uuid.Nil {
		plan.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *adminRepository) UpdateSubscriptionPlan(ctx context.Context, plan *entity.SubscriptionPlan) error {
	_ = r.db.AutoMigrate(&entity.SubscriptionPlan{})
	return r.db.WithContext(ctx).Save(plan).Error
}

func (r *adminRepository) DeleteSubscriptionPlan(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.SubscriptionPlan{}, "id = ?", id).Error
}

func (r *adminRepository) ListWallets(ctx context.Context, limit, offset int) ([]entity.Wallet, int64, error) {
	var list []entity.Wallet
	var total int64
	query := r.db.WithContext(ctx).Model(&entity.Wallet{}).Preload("User").Preload("User.Profile")
	_ = query.Count(&total)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Order("balance DESC").Find(&list).Error
	return list, total, err
}

func (r *adminRepository) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error) {
	var w entity.Wallet
	err := r.db.WithContext(ctx).Preload("User").Preload("User.Profile").First(&w, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *adminRepository) UpdateWallet(ctx context.Context, wallet *entity.Wallet) error {
	return r.db.WithContext(ctx).Save(wallet).Error
}

func (r *adminRepository) CreateCategory(ctx context.Context, cat *entity.Category) error {
	return r.db.WithContext(ctx).Create(cat).Error
}

func (r *adminRepository) UpdateCategory(ctx context.Context, cat *entity.Category) error {
	return r.db.WithContext(ctx).Save(cat).Error
}

func (r *adminRepository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var catalogServiceIDs []uuid.UUID
		if err := tx.Model(&entity.CatalogService{}).Where("category_id = ?", id).Pluck("id", &catalogServiceIDs).Error; err == nil && len(catalogServiceIDs) > 0 {
			_ = tx.Exec("DELETE FROM accounts_profile_catalog_services WHERE catalog_service_id IN ?", catalogServiceIDs).Error
			_ = tx.Model(&entity.ServiceRequest{}).Where("catalog_service_id IN ?", catalogServiceIDs).Update("catalog_service_id", nil).Error
			if err := tx.Where("id IN ?", catalogServiceIDs).Delete(&entity.CatalogService{}).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&entity.Category{}, "id = ?", id).Error
	})
}

func (r *adminRepository) CreateCatalogService(ctx context.Context, cs *entity.CatalogService) error {
	return r.db.WithContext(ctx).Create(cs).Error
}

func (r *adminRepository) UpdateCatalogService(ctx context.Context, cs *entity.CatalogService) error {
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if cs.Name != "" {
		updates["name"] = cs.Name
	}
	if cs.Description != "" {
		updates["description"] = cs.Description
	}
	if cs.BasePrice != nil {
		updates["base_price"] = *cs.BasePrice
	}
	if cs.CategoryID != uuid.Nil {
		updates["category_id"] = cs.CategoryID
	}
	return r.db.WithContext(ctx).Model(&entity.CatalogService{}).Where("id = ?", cs.ID).Updates(updates).Error
}

func (r *adminRepository) DeleteCatalogService(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		_ = tx.Exec("DELETE FROM accounts_profile_catalog_services WHERE catalog_service_id = ?", id).Error
		_ = tx.Model(&entity.ServiceRequest{}).Where("catalog_service_id = ?", id).Update("catalog_service_id", nil).Error
		return tx.Delete(&entity.CatalogService{}, "id = ?", id).Error
	})
}

func (r *adminRepository) GetSettings(ctx context.Context) (*entity.ModerationSetting, error) {
	var s entity.ModerationSetting
	err := r.db.WithContext(ctx).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s = entity.ModerationSetting{
				ID:                         uuid.New(),
				BackgroundCheckPaymentMode: "IN_APP_STRIPE",
				BackgroundCheckFee:         29.99,
			}
			_ = r.db.WithContext(ctx).Create(&s)
			return &s, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *adminRepository) UpdateSettings(ctx context.Context, setting *entity.ModerationSetting) error {
	return r.db.WithContext(ctx).Save(setting).Error
}

func (r *adminRepository) CreateAuditLog(ctx context.Context, log *entity.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *adminRepository) ListAuditLogs(ctx context.Context, filter repository.AdminAuditFilter) ([]entity.AuditLog, int64, error) {
	var logs []entity.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.AuditLog{}).Preload("User")
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.ResourceType != "" {
		query = query.Where("resource_type = ?", filter.ResourceType)
	}

	_ = query.Count(&total)

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	err := query.Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}

func (r *adminRepository) RestoreUser(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Model(&entity.User{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (r *adminRepository) GetWalletByID(ctx context.Context, id uuid.UUID) (*entity.Wallet, error) {
	var w entity.Wallet
	err := r.db.WithContext(ctx).Preload("User").First(&w, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *adminRepository) CreateWalletTransaction(ctx context.Context, tx *entity.WalletTransaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *adminRepository) BatchUpdateVerifications(ctx context.Context, ids []uuid.UUID, status, notes string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if notes != "" {
		updates["reviewer_notes"] = notes
	}
	return r.db.WithContext(ctx).Model(&entity.ProviderVerification{}).Where("id IN ?", ids).Updates(updates).Error
}

func (r *adminRepository) ListFeatureFlags(ctx context.Context) ([]entity.FeatureFlag, error) {
	var flags []entity.FeatureFlag
	err := r.db.WithContext(ctx).Order("key ASC").Find(&flags).Error
	return flags, err
}

func (r *adminRepository) GetFeatureFlag(ctx context.Context, key string) (*entity.FeatureFlag, error) {
	var flag entity.FeatureFlag
	err := r.db.WithContext(ctx).First(&flag, "key = ?", key).Error
	if err != nil {
		return nil, err
	}
	return &flag, nil
}

func (r *adminRepository) SetFeatureFlag(ctx context.Context, key string, isEnabled bool) (*entity.FeatureFlag, error) {
	var flag entity.FeatureFlag
	err := r.db.WithContext(ctx).First(&flag, "key = ?", key).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			flag = entity.FeatureFlag{
				ID:        uuid.New(),
				Key:       key,
				Name:      key,
				IsEnabled: isEnabled,
			}
			if err := r.db.WithContext(ctx).Create(&flag).Error; err != nil {
				return nil, err
			}
			return &flag, nil
		}
		return nil, err
	}
	flag.IsEnabled = isEnabled
	flag.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Save(&flag).Error; err != nil {
		return nil, err
	}
	return &flag, nil
}

func (r *adminRepository) GetFinancialReport(ctx context.Context, startDate, endDate string) (*entity.FinancialReportSummary, error) {
	summary := entity.FinancialReportSummary{
		StartDate:      startDate,
		EndDate:        endDate,
		DailyBreakdown: []entity.FinancialDayBreakdown{},
	}

	query := r.db.WithContext(ctx)

	// Total GMV from funded appointments
	_ = query.Model(&entity.Appointment{}).
		Where("is_funded = ? AND created_at >= ? AND created_at <= ?", true, startDate, endDate).
		Select("COALESCE(SUM(total_price), 0)").Scan(&summary.TotalGMV)

	// Platform Commission Estimate (avg 15%)
	summary.TotalPlatformFeeRevenue = summary.TotalGMV * 0.15

	// Processed Payouts
	_ = query.Model(&entity.PayoutRequest{}).
		Where("status = ? AND created_at >= ? AND created_at <= ?", "PROCESSED", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").Scan(&summary.TotalPayoutsProcessed)

	// Pending Payouts
	_ = query.Model(&entity.PayoutRequest{}).
		Where("status = ?", "PENDING").
		Select("COALESCE(SUM(amount), 0)").Scan(&summary.PendingPayoutsAmount)

	// Active Subscriptions count & estimate
	_ = query.Model(&entity.Subscription{}).Where("is_active = ?", true).Count(&summary.ActiveSubscriptionsCount)
	summary.SubscriptionRevenueEstimate = float64(summary.ActiveSubscriptionsCount) * 19.99

	// Daily Breakdown
	type DaySum struct {
		Day   string  `json:"day"`
		Total float64 `json:"total"`
	}
	var gmvDays []DaySum
	_ = query.Model(&entity.Appointment{}).
		Select("TO_CHAR(created_at, 'YYYY-MM-DD') as day, COALESCE(SUM(total_price), 0) as total").
		Where("is_funded = ? AND created_at >= ? AND created_at <= ?", true, startDate, endDate).
		Group("TO_CHAR(created_at, 'YYYY-MM-DD')").
		Scan(&gmvDays)

	for _, gd := range gmvDays {
		summary.DailyBreakdown = append(summary.DailyBreakdown, entity.FinancialDayBreakdown{
			Day:         gd.Day,
			GMV:         gd.Total,
			PlatformFee: gd.Total * 0.15,
		})
	}

	return &summary, nil
}

func (r *adminRepository) GetSystemHealth(ctx context.Context) (*entity.SystemHealthStatus, error) {
	status := entity.SystemHealthStatus{
		Status:         "HEALTHY",
		DatabaseStatus: "CONNECTED",
		Timestamp:      time.Now(),
	}

	sqlDB, err := r.db.DB()
	if err != nil {
		status.Status = "DEGRADED"
		status.DatabaseStatus = "ERROR: " + err.Error()
	} else {
		dbStats := sqlDB.Stats()
		status.ActiveDBConnections = dbStats.InUse
		status.IdleDBConnections = dbStats.Idle
		status.OpenDBConnections = dbStats.OpenConnections
	}

	_ = r.db.WithContext(ctx).Model(&entity.User{}).Count(&status.TotalUsersCount)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	status.GoRoutinesCount = runtime.NumGoroutine()
	status.MemoryAllocMB = float64(m.Alloc) / 1024 / 1024

	return &status, nil
}

func (r *adminRepository) GetGDPRUserData(ctx context.Context, userID uuid.UUID) (*entity.GDPRUserData, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).
		Preload("Profile").
		Preload("Profile.CatalogServices").
		Preload("Profile.PortfolioItems").
		Preload("Profile.ServicePackages").
		Preload("Wallet").
		Preload("BackgroundCheck").
		Preload("AdminRole").
		First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}

	var profile entity.Profile
	_ = r.db.WithContext(ctx).Preload("CatalogServices").Preload("PortfolioItems").Preload("ServicePackages").First(&profile, "user_id = ?", userID)

	var wallet entity.Wallet
	_ = r.db.WithContext(ctx).First(&wallet, "user_id = ?", userID)

	var txs []entity.WalletTransaction
	if wallet.ID != uuid.Nil {
		_ = r.db.WithContext(ctx).Where("wallet_id = ?", wallet.ID).Order("created_at DESC").Find(&txs)
	}

	var apts []entity.Appointment
	_ = r.db.WithContext(ctx).
		Preload("Seeker").
		Preload("Seeker.Profile").
		Preload("Provider").
		Preload("Provider.Profile").
		Preload("ServiceRequest").
		Where("seeker_id = ? OR provider_id = ?", userID, userID).
		Order("created_at DESC").
		Find(&apts)

	var reviews []entity.Review
	_ = r.db.WithContext(ctx).
		Preload("Reviewer").
		Preload("Reviewer.Profile").
		Preload("Provider").
		Preload("Provider.Profile").
		Where("reviewer_id = ? OR provider_id = ?", userID, userID).
		Order("created_at DESC").
		Find(&reviews)

	var disputes []entity.Dispute
	_ = r.db.WithContext(ctx).
		Preload("RaisedBy").
		Preload("RaisedBy.Profile").
		Preload("Defendant").
		Preload("Defendant.Profile").
		Preload("Appointment").
		Where("raised_by_id = ? OR defendant_id = ?", userID, userID).
		Order("created_at DESC").
		Find(&disputes)

	var msgCount int64
	_ = r.db.WithContext(ctx).Model(&entity.Message{}).Where("sender_id = ?", userID).Count(&msgCount)

	return &entity.GDPRUserData{
		User:          user,
		Profile:       &profile,
		Wallet:        &wallet,
		Transactions:  txs,
		Appointments:  apts,
		Reviews:       reviews,
		Disputes:      disputes,
		MessagesCount: msgCount,
		ExportedAt:    time.Now(),
	}, nil
}

func (r *adminRepository) ExportUsers(ctx context.Context, filter repository.AdminUserFilter) ([]entity.User, error) {
	var users []entity.User
	query := r.db.WithContext(ctx).Model(&entity.User{}).Preload("Profile")
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		query = query.Joins("LEFT JOIN accounts_profile ON accounts_profile.user_id = accounts_user.id").
			Where("accounts_user.email ILIKE ? OR accounts_profile.first_name ILIKE ? OR accounts_profile.last_name ILIKE ?", s, s, s)
	}
	if filter.UserType != "" {
		query = query.Joins("JOIN accounts_profile ON accounts_profile.user_id = accounts_user.id").
			Where("accounts_profile.user_type = ?", filter.UserType)
	}
	err := query.Order("created_at DESC").Find(&users).Error
	return users, err
}

func (r *adminRepository) ExportPayouts(ctx context.Context, status string) ([]entity.PayoutRequest, error) {
	var payouts []entity.PayoutRequest
	query := r.db.WithContext(ctx).Model(&entity.PayoutRequest{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at DESC").Find(&payouts).Error
	return payouts, err
}

func (r *adminRepository) ExportDisputes(ctx context.Context, status string) ([]entity.Dispute, error) {
	var disputes []entity.Dispute
	query := r.db.WithContext(ctx).Model(&entity.Dispute{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at DESC").Find(&disputes).Error
	return disputes, err
}

// ─── RBAC & ROLES ───────────────────────────────────────────────────────────

func (r *adminRepository) ListRoles(ctx context.Context) ([]entity.AdminRole, error) {
	var roles []entity.AdminRole
	err := r.db.WithContext(ctx).Order("name ASC").Find(&roles).Error
	return roles, err
}

func (r *adminRepository) GetRoleByID(ctx context.Context, id uuid.UUID) (*entity.AdminRole, error) {
	var role entity.AdminRole
	err := r.db.WithContext(ctx).First(&role, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *adminRepository) CreateRole(ctx context.Context, role *entity.AdminRole) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *adminRepository) UpdateRole(ctx context.Context, role *entity.AdminRole) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *adminRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.AdminRole{}, "id = ?", id).Error
}

// ─── FRAUD & RISK DETECTION ─────────────────────────────────────────────────

func (r *adminRepository) ListFraudRiskAlerts(ctx context.Context, status string, limit, offset int) ([]entity.FraudRiskAlert, int64, error) {
	var alerts []entity.FraudRiskAlert
	var total int64
	query := r.db.WithContext(ctx).Model(&entity.FraudRiskAlert{}).Preload("User").Preload("User.Profile")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	_ = query.Count(&total)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Order("risk_score DESC, created_at DESC").Find(&alerts).Error
	return alerts, total, err
}

func (r *adminRepository) CreateFraudRiskAlert(ctx context.Context, alert *entity.FraudRiskAlert) error {
	return r.db.WithContext(ctx).Create(alert).Error
}

func (r *adminRepository) UpdateFraudRiskAlert(ctx context.Context, alert *entity.FraudRiskAlert) error {
	return r.db.WithContext(ctx).Save(alert).Error
}

func (r *adminRepository) EvaluateUserRisk(ctx context.Context, userID uuid.UUID) (*entity.FraudRiskAlert, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).Preload("Profile").First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}

	score := 0
	var flags entity.JSONSlice
	details := entity.JSONMap{}

	// 1. Check duplicate device token
	if user.Profile != nil && user.Profile.DeviceToken != "" {
		var tokenCount int64
		_ = r.db.WithContext(ctx).Model(&entity.Profile{}).
			Where("device_token = ? AND user_id != ?", user.Profile.DeviceToken, userID).
			Count(&tokenCount)
		if tokenCount > 0 {
			score += 30
			flags = append(flags, "DUPLICATE_DEVICE_TOKEN_FOUND")
			details["duplicate_device_accounts"] = tokenCount
		}
	}

	// 2. Check dispute history
	var disputeCount int64
	_ = r.db.WithContext(ctx).Model(&entity.Dispute{}).
		Where("defendant_id = ?", userID).
		Count(&disputeCount)
	if disputeCount >= 2 {
		score += 35
		flags = append(flags, "HIGH_DISPUTE_INCIDENTS")
		details["dispute_count"] = disputeCount
	}

	// 3. Provider unverified identity check
	if user.Profile != nil && user.Profile.UserType == "PROVIDER" && !user.Profile.IsIdentityVerified {
		score += 20
		flags = append(flags, "UNVERIFIED_PROVIDER_ACTIVITY")
	}

	level := "LOW"
	if score >= 70 {
		level = "CRITICAL"
	} else if score >= 50 {
		level = "HIGH"
	} else if score >= 25 {
		level = "MEDIUM"
	}

	alert := entity.FraudRiskAlert{
		ID:        uuid.New(),
		UserID:    userID,
		RiskScore: score,
		RiskLevel: level,
		Flags:     flags,
		Details:   details,
		Status:    "OPEN",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_ = r.db.WithContext(ctx).Create(&alert)
	alert.User = &user
	return &alert, nil
}

// ─── NOTIFICATION & EMAIL TEMPLATES ─────────────────────────────────────────

func (r *adminRepository) ListNotificationTemplates(ctx context.Context) ([]entity.NotificationTemplate, error) {
	var templates []entity.NotificationTemplate
	err := r.db.WithContext(ctx).Order("name ASC").Find(&templates).Error
	return templates, err
}

func (r *adminRepository) GetNotificationTemplate(ctx context.Context, key string) (*entity.NotificationTemplate, error) {
	var tpl entity.NotificationTemplate
	err := r.db.WithContext(ctx).First(&tpl, "key = ?", key).Error
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

func (r *adminRepository) UpdateNotificationTemplate(ctx context.Context, tpl *entity.NotificationTemplate) error {
	return r.db.WithContext(ctx).Save(tpl).Error
}

// ─── BACKUP SNAPSHOTS ───────────────────────────────────────────────────────

func (r *adminRepository) ListBackupSnapshots(ctx context.Context) ([]entity.BackupSnapshot, error) {
	var snapshots []entity.BackupSnapshot
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&snapshots).Error
	return snapshots, err
}

func (r *adminRepository) GetBackupSnapshotByID(ctx context.Context, id uuid.UUID) (*entity.BackupSnapshot, error) {
	var snapshot entity.BackupSnapshot
	err := r.db.WithContext(ctx).First(&snapshot, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (r *adminRepository) GetBackupSnapshotByFilename(ctx context.Context, filename string) (*entity.BackupSnapshot, error) {
	var snapshot entity.BackupSnapshot
	err := r.db.WithContext(ctx).First(&snapshot, "filename = ?", filename).Error
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (r *adminRepository) CreateBackupSnapshot(ctx context.Context, snap *entity.BackupSnapshot) error {
	return r.db.WithContext(ctx).Create(snap).Error
}

// ─── BATCH PAYOUT DISBURSEMENTS ─────────────────────────────────────────────

func (r *adminRepository) BatchApprovePayouts(ctx context.Context, ids []uuid.UUID, adminID uuid.UUID) (*entity.BatchPayoutResponse, error) {
	if len(ids) == 0 {
		return &entity.BatchPayoutResponse{Message: "No payout IDs provided"}, nil
	}

	var payouts []entity.PayoutRequest
	err := r.db.WithContext(ctx).Where("id IN ? AND status = ?", ids, "PENDING").Find(&payouts).Error
	if err != nil {
		return nil, err
	}

	var processedCount int64
	var totalAmount float64

	now := time.Now()
	for _, p := range payouts {
		p.Status = "APPROVED"
		p.ProcessedAt = &now
		if err := r.db.WithContext(ctx).Save(&p).Error; err == nil {
			processedCount++
			totalAmount += p.Amount
			// Log audit event
			_ = r.CreateAuditLog(ctx, &entity.AuditLog{
				ID:           uuid.New(),
				UserID:       &adminID,
				Action:       "BATCH_PAYOUT_APPROVED",
				ResourceType: "PAYOUT_REQUEST",
				ResourceID:   p.ID.String(),
				Details:      entity.JSONMap{"amount": p.Amount, "batch": true},
				CreatedAt:    now,
			})
		}
	}

	return &entity.BatchPayoutResponse{
		ProcessedCount: processedCount,
		TotalAmount:    totalAmount,
		FailedCount:    int64(len(ids)) - processedCount,
		Message:        "Batch payout approval completed successfully",
	}, nil
}

func (r *adminRepository) BatchRejectPayouts(ctx context.Context, ids []uuid.UUID, reason string, adminID uuid.UUID) (*entity.BatchPayoutResponse, error) {
	if len(ids) == 0 {
		return &entity.BatchPayoutResponse{Message: "No payout IDs provided"}, nil
	}

	var payouts []entity.PayoutRequest
	err := r.db.WithContext(ctx).Where("id IN ? AND status = ?", ids, "PENDING").Find(&payouts).Error
	if err != nil {
		return nil, err
	}

	var processedCount int64
	now := time.Now()
	for _, p := range payouts {
		p.Status = "REJECTED"
		p.AdminNotes = reason
		p.ProcessedAt = &now
		if err := r.db.WithContext(ctx).Save(&p).Error; err == nil {
			processedCount++
			_ = r.CreateAuditLog(ctx, &entity.AuditLog{
				ID:           uuid.New(),
				UserID:       &adminID,
				Action:       "BATCH_PAYOUT_REJECTED",
				ResourceType: "PAYOUT_REQUEST",
				ResourceID:   p.ID.String(),
				Details:      entity.JSONMap{"reason": reason, "batch": true},
				CreatedAt:    now,
			})
		}
	}

	return &entity.BatchPayoutResponse{
		ProcessedCount: processedCount,
		FailedCount:    int64(len(ids)) - processedCount,
		Message:        "Batch payout rejection completed",
	}, nil
}

// ─── STAFF INTERNAL NOTES ───────────────────────────────────────────────────

func (r *adminRepository) ListStaffNotes(ctx context.Context, userID uuid.UUID) ([]entity.StaffNote, error) {
	_ = r.db.AutoMigrate(&entity.StaffNote{})
	var notes []entity.StaffNote
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("is_pinned DESC, created_at DESC").Find(&notes).Error
	return notes, err
}

func (r *adminRepository) CreateStaffNote(ctx context.Context, note *entity.StaffNote) error {
	_ = r.db.AutoMigrate(&entity.StaffNote{})
	if note.ID == uuid.Nil {
		note.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(note).Error
}

func (r *adminRepository) DeleteStaffNote(ctx context.Context, noteID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.StaffNote{}, "id = ?", noteID).Error
}

// ─── LIVE NOTIFICATION FEED ──────────────────────────────────────────────────

func (r *adminRepository) GetAdminNotificationFeed(ctx context.Context) ([]entity.AdminNotification, error) {
	var notifications []entity.AdminNotification

	// 1. Unresolved Disputes
	var openDisputes []entity.Dispute
	_ = r.db.WithContext(ctx).Where("status = ?", "OPEN").Order("created_at DESC").Limit(5).Find(&openDisputes)
	for _, d := range openDisputes {
		notifications = append(notifications, entity.AdminNotification{
			ID:         "disp-" + d.ID.String(),
			Type:       "DISPUTE",
			Severity:   "HIGH",
			Title:      "Dispute Escalation",
			Message:    fmt.Sprintf("Dispute #%s: %s", d.ID.String()[:8], d.Reason),
			ResourceID: d.ID.String(),
			Route:      "/admin/disputes",
			IsRead:     false,
			CreatedAt:  d.CreatedAt,
		})
	}

	// 2. High-Risk Fraud Alerts
	var fraudAlerts []entity.FraudRiskAlert
	_ = r.db.WithContext(ctx).Where("status = ?", "OPEN").Order("risk_score DESC, created_at DESC").Limit(5).Find(&fraudAlerts)
	for _, f := range fraudAlerts {
		notifications = append(notifications, entity.AdminNotification{
			ID:         "fraud-" + f.ID.String(),
			Type:       "FRAUD_VELOCITY",
			Severity:   "HIGH",
			Title:      fmt.Sprintf("Risk Alert: Score %d/100 (%s)", f.RiskScore, f.RiskLevel),
			Message:    "High anomalous activity detected requiring staff review.",
			ResourceID: f.UserID.String(),
			Route:      "/admin/fraud",
			IsRead:     false,
			CreatedAt:  f.CreatedAt,
		})
	}

	// 3. Flagged / Consider Background Checks
	var flaggedChecks []entity.BackgroundCheck
	_ = r.db.WithContext(ctx).Where("status IN ('FLAGGED', 'CONSIDER', 'SUSPENDED') OR adjudication IN ('ENGAGE', 'CONSIDER')").Order("created_at DESC").Limit(5).Find(&flaggedChecks)
	for _, b := range flaggedChecks {
		notifications = append(notifications, entity.AdminNotification{
			ID:         "bg-" + b.ID.String(),
			Type:       "CHECKR_FLAG",
			Severity:   "MEDIUM",
			Title:      "Checkr Screening Requires Review",
			Message:    "Candidate screening returned 'CONSIDER' or manual review flag.",
			ResourceID: b.ID.String(),
			Route:      "/admin/background-checks",
			IsRead:     false,
			CreatedAt:  b.CreatedAt,
		})
	}

	// 4. Large Pending Payouts
	var largePayouts []entity.PayoutRequest
	_ = r.db.WithContext(ctx).Where("status = ? AND amount >= ?", "PENDING", 200.0).Order("created_at DESC").Limit(5).Find(&largePayouts)
	for _, p := range largePayouts {
		notifications = append(notifications, entity.AdminNotification{
			ID:         "payout-" + p.ID.String(),
			Type:       "PAYOUT_HOLD",
			Severity:   "MEDIUM",
			Title:      "High-Value Payout Request",
			Message:    fmt.Sprintf("Disbursement request of $%.2f awaiting operator approval.", p.Amount),
			ResourceID: p.ID.String(),
			Route:      "/admin/payouts",
			IsRead:     false,
			CreatedAt:  p.CreatedAt,
		})
	}

	// Sort by created_at descending
	if len(notifications) == 0 {
		notifications = append(notifications, entity.AdminNotification{
			ID:        "sys-ok",
			Type:      "SYSTEM_ALERT",
			Severity:  "INFO",
			Title:     "Platform Systems Nominal",
			Message:   "Zero critical disputes or unhandled fraud triggers at this time.",
			Route:     "/admin/dashboard",
			IsRead:    true,
			CreatedAt: time.Now(),
		})
	}

	return notifications, nil
}

// ─── PROVIDER ONBOARDING FUNNEL ─────────────────────────────────────────────

func (r *adminRepository) GetProviderOnboardingFunnel(ctx context.Context) (*entity.ProviderOnboardingFunnel, error) {
	var funnel entity.ProviderOnboardingFunnel

	_ = r.db.WithContext(ctx).Model(&entity.Profile{}).Where("user_type = ?", "PROVIDER").Count(&funnel.TotalSignedUp)
	_ = r.db.WithContext(ctx).Model(&entity.ProviderVerification{}).Where("status = ?", "APPROVED").Count(&funnel.IDUploaded)
	_ = r.db.WithContext(ctx).Model(&entity.BackgroundCheck{}).Where("status ILIKE ? OR status = ? OR adjudication ILIKE ?", "clear%", "APPROVED", "pass%").Count(&funnel.CheckrCompleted)
	_ = r.db.WithContext(ctx).Table("payments_wallet").
		Joins("JOIN accounts_profile ON accounts_profile.user_id = payments_wallet.user_id").
		Where("accounts_profile.user_type = ? AND (payments_wallet.stripe_connect_id != '' OR EXISTS (SELECT 1 FROM payments_customer WHERE payments_customer.user_id = accounts_profile.user_id AND (payments_customer.stripe_account_id != '' OR payments_customer.stripe_customer_id != '')))", "PROVIDER").
		Count(&funnel.StripeConnected)
	_ = r.db.WithContext(ctx).Model(&entity.Appointment{}).Where("status = ?", "COMPLETED").Distinct("provider_id").Count(&funnel.FirstBookingDone)

	if funnel.TotalSignedUp > 0 {
		funnel.ConversionRatePct = float64(funnel.FirstBookingDone) / float64(funnel.TotalSignedUp) * 100.0
	}

	return &funnel, nil
}

// ─── BOOKINGS & APPOINTMENTS CONSOLE ─────────────────────────────────────────

func (r *adminRepository) ListAppointments(ctx context.Context, status string, search string, limit, offset int) ([]entity.Appointment, int64, error) {
	var list []entity.Appointment
	var total int64
	query := r.db.WithContext(ctx).Model(&entity.Appointment{}).
		Preload("Seeker").Preload("Seeker.Profile").
		Preload("Provider").Preload("Provider.Profile").
		Preload("ServiceRequest")

	if status != "" && status != "ALL" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ? OR CAST(id AS TEXT) ILIKE ?", like, like, like)
	}

	_ = query.Count(&total)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Order("created_at DESC").Find(&list).Error
	return list, total, err
}

func (r *adminRepository) GetAppointmentByID(ctx context.Context, id uuid.UUID) (*entity.Appointment, error) {
	var apt entity.Appointment
	err := r.db.WithContext(ctx).
		Preload("Seeker").Preload("Seeker.Profile").
		Preload("Provider").Preload("Provider.Profile").
		Preload("ServiceRequest").
		First(&apt, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &apt, nil
}

func (r *adminRepository) UpdateAppointmentStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&entity.Appointment{}).Where("id = ?", id).Update("status", status).Error
}

func (r *adminRepository) ExportAppointments(ctx context.Context, status string) ([]entity.Appointment, error) {
	var list []entity.Appointment
	query := r.db.WithContext(ctx).Model(&entity.Appointment{}).
		Preload("Seeker").Preload("Provider")
	if status != "" && status != "ALL" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at DESC").Limit(5000).Find(&list).Error
	return list, err
}

// ─── REVIEWS & RATINGS MODERATION ───────────────────────────────────────────

func (r *adminRepository) ListReviews(ctx context.Context, rating *float64, isHidden *bool, limit, offset int) ([]entity.Review, int64, error) {
	_ = r.db.AutoMigrate(&entity.Review{})
	var list []entity.Review
	var total int64
	query := r.db.WithContext(ctx).Model(&entity.Review{}).
		Preload("Reviewer").Preload("Reviewer.Profile").
		Preload("Provider").Preload("Provider.Profile")

	if rating != nil && *rating > 0 {
		query = query.Where("rating = ?", *rating)
	}
	if isHidden != nil {
		query = query.Where("is_hidden = ?", *isHidden)
	}

	_ = query.Count(&total)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Order("created_at DESC").Find(&list).Error
	return list, total, err
}

func (r *adminRepository) CreateReview(ctx context.Context, review *entity.Review) error {
	_ = r.db.AutoMigrate(&entity.Review{})
	if review.ID == uuid.Nil {
		review.ID = uuid.New()
	}
	now := time.Now()
	if review.CreatedAt.IsZero() {
		review.CreatedAt = now
	}
	review.UpdatedAt = now
	return r.db.WithContext(ctx).Create(review).Error
}

func (r *adminRepository) ToggleReviewVisibility(ctx context.Context, id uuid.UUID, isHidden bool) error {
	_ = r.db.AutoMigrate(&entity.Review{})
	return r.db.WithContext(ctx).Model(&entity.Review{}).Where("id = ?", id).Update("is_hidden", isHidden).Error
}

func (r *adminRepository) DeleteReview(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Review{}, "id = ?", id).Error
}

// ─── PROMO CODES & MARKETING CAMPAIGNS ───────────────────────────────────────

func (r *adminRepository) ListPromoCodes(ctx context.Context) ([]entity.PromoCode, error) {
	_ = r.db.AutoMigrate(&entity.PromoCode{})
	var list []entity.PromoCode
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&list).Error
	return list, err
}

func (r *adminRepository) CreatePromoCode(ctx context.Context, promo *entity.PromoCode) error {
	_ = r.db.AutoMigrate(&entity.PromoCode{})
	if promo.ID == uuid.Nil {
		promo.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(promo).Error
}

func (r *adminRepository) UpdatePromoCode(ctx context.Context, promo *entity.PromoCode) error {
	_ = r.db.AutoMigrate(&entity.PromoCode{})
	return r.db.WithContext(ctx).Save(promo).Error
}

func (r *adminRepository) DeletePromoCode(ctx context.Context, id uuid.UUID) error {
	_ = r.db.AutoMigrate(&entity.PromoCode{})
	return r.db.WithContext(ctx).Delete(&entity.PromoCode{}, "id = ?", id).Error
}

// ─── DISPUTE REFUND EXECUTION ────────────────────────────────────────────────

func (r *adminRepository) ExecuteDisputeRefund(ctx context.Context, disputeID uuid.UUID, amount float64, reason string, adminID uuid.UUID) error {
	var dispute entity.Dispute
	if err := r.db.WithContext(ctx).First(&dispute, "id = ?", disputeID).Error; err != nil {
		return err
	}

	// Update dispute status
	now := time.Now()
	dispute.Status = "RESOLVED"
	dispute.ResolutionNotes = fmt.Sprintf("Refund of $%.2f issued by operator. Reason: %s", amount, reason)
	dispute.UpdatedAt = now
	if err := r.db.WithContext(ctx).Save(&dispute).Error; err != nil {
		return err
	}

	// Credit seeker wallet if dispute is linked to user
	var wallet entity.Wallet
	if err := r.db.WithContext(ctx).Where("user_id = ?", dispute.RaisedByID).First(&wallet).Error; err == nil {
		wallet.Balance += amount
		wallet.UpdatedAt = now
		_ = r.db.WithContext(ctx).Save(&wallet)

		_ = r.CreateWalletTransaction(ctx, &entity.WalletTransaction{
			ID:              uuid.New(),
			WalletID:        wallet.ID,
			Amount:          amount,
			TransactionType: "CREDIT",
			Description:     fmt.Sprintf("Dispute #%s Refund Resolution", dispute.ID.String()[:8]),
			ReferenceID:     dispute.ID.String(),
			CreatedAt:       now,
		})
	}

	// Audit Log
	_ = r.CreateAuditLog(ctx, &entity.AuditLog{
		ID:           uuid.New(),
		UserID:       &adminID,
		Action:       "DISPUTE_REFUND_EXECUTED",
		ResourceType: "DISPUTE",
		ResourceID:   dispute.ID.String(),
		Details:      entity.JSONMap{"refund_amount": amount, "reason": reason},
		CreatedAt:    now,
	})

	return nil
}

// ─── ADDITIONAL DATA EXPORTS ─────────────────────────────────────────────────

func (r *adminRepository) ExportVerifications(ctx context.Context, status string) ([]entity.ProviderVerification, error) {
	var list []entity.ProviderVerification
	query := r.db.WithContext(ctx).Model(&entity.ProviderVerification{}).Preload("Provider")
	if status != "" && status != "ALL" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at DESC").Limit(5000).Find(&list).Error
	return list, err
}

func (r *adminRepository) ExportBackgroundChecks(ctx context.Context, status string) ([]entity.BackgroundCheck, error) {
	var list []entity.BackgroundCheck
	query := r.db.WithContext(ctx).Model(&entity.BackgroundCheck{}).Preload("Provider")
	if status != "" && status != "ALL" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at DESC").Limit(5000).Find(&list).Error
	return list, err
}

// ─── LEGAL & COMPLIANCE DOCUMENTS ───────────────────────────────────────────

func (r *adminRepository) ListLegalDocuments(ctx context.Context, docType string, isActive *bool) ([]entity.LegalDocument, error) {
	_ = r.db.AutoMigrate(&entity.LegalDocument{})
	var docs []entity.LegalDocument
	q := r.db.WithContext(ctx).Model(&entity.LegalDocument{})
	if docType != "" {
		q = q.Where("doc_type = ?", docType)
	}
	if isActive != nil {
		q = q.Where("is_active = ?", *isActive)
	}
	err := q.Order("doc_type ASC, version DESC").Find(&docs).Error
	return docs, err
}

func (r *adminRepository) GetLegalDocumentByID(ctx context.Context, id uuid.UUID) (*entity.LegalDocument, error) {
	_ = r.db.AutoMigrate(&entity.LegalDocument{})
	var doc entity.LegalDocument
	err := r.db.WithContext(ctx).First(&doc, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *adminRepository) CreateLegalDocument(ctx context.Context, doc *entity.LegalDocument) error {
	_ = r.db.AutoMigrate(&entity.LegalDocument{})
	if doc.ID == uuid.Nil {
		doc.ID = uuid.New()
	}
	now := time.Now()
	doc.CreatedAt = now
	doc.UpdatedAt = now
	return r.db.WithContext(ctx).Create(doc).Error
}

func (r *adminRepository) UpdateLegalDocument(ctx context.Context, doc *entity.LegalDocument) error {
	_ = r.db.AutoMigrate(&entity.LegalDocument{})
	doc.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(doc).Error
}

func (r *adminRepository) DeleteLegalDocument(ctx context.Context, id uuid.UUID) error {
	_ = r.db.AutoMigrate(&entity.LegalDocument{})
	return r.db.WithContext(ctx).Delete(&entity.LegalDocument{}, "id = ?", id).Error
}

// ─── SUPPORT MESSAGES & INQUIRIES ───────────────────────────────────────────

func (r *adminRepository) ListContactMessages(ctx context.Context, isResolved *bool, search string, limit, offset int) ([]entity.ContactMessage, int64, error) {
	_ = r.db.AutoMigrate(&entity.ContactMessage{})
	var list []entity.ContactMessage
	var total int64
	q := r.db.WithContext(ctx).Model(&entity.ContactMessage{})
	if isResolved != nil {
		q = q.Where("is_resolved = ?", *isResolved)
	}
	if search != "" {
		s := "%" + strings.ToLower(search) + "%"
		q = q.Where("LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(message) LIKE ?", s, s, s, s)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 50
	}
	err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&list).Error
	return list, total, err
}

func (r *adminRepository) ToggleContactMessageResolved(ctx context.Context, id uuid.UUID, isResolved bool) error {
	_ = r.db.AutoMigrate(&entity.ContactMessage{})
	return r.db.WithContext(ctx).Model(&entity.ContactMessage{}).Where("id = ?", id).Update("is_resolved", isResolved).Error
}

func (r *adminRepository) DeleteContactMessage(ctx context.Context, id uuid.UUID) error {
	_ = r.db.AutoMigrate(&entity.ContactMessage{})
	return r.db.WithContext(ctx).Delete(&entity.ContactMessage{}, "id = ?", id).Error
}

// ─── RESOLUTION REPORTS ─────────────────────────────────────────────────────

func (r *adminRepository) ListResolutionReports(ctx context.Context, isReviewed *bool, search string, limit, offset int) ([]entity.ResolutionReport, int64, error) {
	_ = r.db.AutoMigrate(&entity.ResolutionReport{})
	var list []entity.ResolutionReport
	var total int64
	q := r.db.WithContext(ctx).Model(&entity.ResolutionReport{})
	if isReviewed != nil {
		q = q.Where("is_reviewed = ?", *isReviewed)
	}
	if search != "" {
		s := "%" + strings.ToLower(search) + "%"
		q = q.Where("LOWER(booking_ref) LIKE ? OR LOWER(description) LIKE ? OR LOWER(expected_outcome) LIKE ? OR LOWER(other_neighbor) LIKE ?", s, s, s, s)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 50
	}
	err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&list).Error
	return list, total, err
}

func (r *adminRepository) ToggleResolutionReportReviewed(ctx context.Context, id uuid.UUID, isReviewed bool) error {
	_ = r.db.AutoMigrate(&entity.ResolutionReport{})
	return r.db.WithContext(ctx).Model(&entity.ResolutionReport{}).Where("id = ?", id).Update("is_reviewed", isReviewed).Error
}

func (r *adminRepository) DeleteResolutionReport(ctx context.Context, id uuid.UUID) error {
	_ = r.db.AutoMigrate(&entity.ResolutionReport{})
	return r.db.WithContext(ctx).Delete(&entity.ResolutionReport{}, "id = ?", id).Error
}

// ─── PUBLIC SITE CMS ────────────────────────────────────────────────────────

func (r *adminRepository) ListFAQs(ctx context.Context, category string, isActive *bool) ([]entity.FAQ, error) {
	_ = r.db.AutoMigrate(&entity.FAQ{})
	var list []entity.FAQ
	q := r.db.WithContext(ctx).Model(&entity.FAQ{})
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if isActive != nil {
		q = q.Where("is_active = ?", *isActive)
	}
	err := q.Order("`order` ASC, id ASC").Find(&list).Error
	if err != nil {
		// Postgres doesn't use backticks, retry standard order
		err = r.db.WithContext(ctx).Model(&entity.FAQ{}).Order("\"order\" ASC").Find(&list).Error
	}
	return list, err
}

func (r *adminRepository) CreateFAQ(ctx context.Context, faq *entity.FAQ) error {
	_ = r.db.AutoMigrate(&entity.FAQ{})
	if faq.ID == uuid.Nil {
		faq.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(faq).Error
}

func (r *adminRepository) UpdateFAQ(ctx context.Context, faq *entity.FAQ) error {
	_ = r.db.AutoMigrate(&entity.FAQ{})
	return r.db.WithContext(ctx).Save(faq).Error
}

func (r *adminRepository) DeleteFAQ(ctx context.Context, id uuid.UUID) error {
	_ = r.db.AutoMigrate(&entity.FAQ{})
	return r.db.WithContext(ctx).Delete(&entity.FAQ{}, "id = ?", id).Error
}

func (r *adminRepository) ListTestimonials(ctx context.Context, isActive *bool) ([]entity.Testimonial, error) {
	_ = r.db.AutoMigrate(&entity.Testimonial{})
	var list []entity.Testimonial
	q := r.db.WithContext(ctx).Model(&entity.Testimonial{})
	if isActive != nil {
		q = q.Where("is_active = ?", *isActive)
	}
	err := q.Order("created_at DESC").Find(&list).Error
	return list, err
}

func (r *adminRepository) CreateTestimonial(ctx context.Context, t *entity.Testimonial) error {
	_ = r.db.AutoMigrate(&entity.Testimonial{})
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	t.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *adminRepository) UpdateTestimonial(ctx context.Context, t *entity.Testimonial) error {
	_ = r.db.AutoMigrate(&entity.Testimonial{})
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *adminRepository) DeleteTestimonial(ctx context.Context, id uuid.UUID) error {
	_ = r.db.AutoMigrate(&entity.Testimonial{})
	return r.db.WithContext(ctx).Delete(&entity.Testimonial{}, "id = ?", id).Error
}

func (r *adminRepository) GetHeroSection(ctx context.Context) (*entity.HeroSection, error) {
	_ = r.db.AutoMigrate(&entity.HeroSection{})
	var hero entity.HeroSection
	err := r.db.WithContext(ctx).First(&hero).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &entity.HeroSection{IsActive: false}, nil
		}
		return nil, err
	}
	return &hero, nil
}

func (r *adminRepository) UpdateHeroSection(ctx context.Context, hero *entity.HeroSection) error {
	_ = r.db.AutoMigrate(&entity.HeroSection{})
	hero.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(hero).Error
}

func (r *adminRepository) ListSiteStats(ctx context.Context) ([]entity.SiteStat, error) {
	_ = r.db.AutoMigrate(&entity.SiteStat{})
	var list []entity.SiteStat
	err := r.db.WithContext(ctx).Find(&list).Error
	return list, err
}

func (r *adminRepository) CreateSiteStat(ctx context.Context, stat *entity.SiteStat) error {
	_ = r.db.AutoMigrate(&entity.SiteStat{})
	if stat.ID == uuid.Nil {
		stat.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(stat).Error
}

func (r *adminRepository) UpdateSiteStat(ctx context.Context, stat *entity.SiteStat) error {
	_ = r.db.AutoMigrate(&entity.SiteStat{})
	return r.db.WithContext(ctx).Save(stat).Error
}

func (r *adminRepository) DeleteSiteStat(ctx context.Context, id uuid.UUID) error {
	_ = r.db.AutoMigrate(&entity.SiteStat{})
	return r.db.WithContext(ctx).Delete(&entity.SiteStat{}, "id = ?", id).Error
}

func (r *adminRepository) GetAboutContent(ctx context.Context) (*entity.AboutContent, error) {
	_ = r.db.AutoMigrate(&entity.AboutContent{})
	var about entity.AboutContent
	err := r.db.WithContext(ctx).First(&about).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &entity.AboutContent{IsActive: false}, nil
		}
		return nil, err
	}
	return &about, nil
}

func (r *adminRepository) UpdateAboutContent(ctx context.Context, about *entity.AboutContent) error {
	_ = r.db.AutoMigrate(&entity.AboutContent{})
	about.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(about).Error
}

