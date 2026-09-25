package usecase_test

import (
	"context"
	"errors"
	"testing"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockAdminRepo struct {
	stats         *entity.AdminDashboardStats
	users         map[uuid.UUID]*entity.User
	verifications map[uuid.UUID]*entity.ProviderVerification
	checks        map[uuid.UUID]*entity.BackgroundCheck
	reports       map[uuid.UUID]*entity.Report
	disputes      map[uuid.UUID]*entity.Dispute
	payouts       map[uuid.UUID]*entity.PayoutRequest
	subs          map[uuid.UUID]*entity.Subscription
	plans         map[uuid.UUID]*entity.SubscriptionPlan
	wallets       map[uuid.UUID]*entity.Wallet
	categories    map[uuid.UUID]*entity.Category
	catalog       map[uuid.UUID]*entity.CatalogService
	settings      *entity.ModerationSetting
	auditLogs     []entity.AuditLog
	roles         map[uuid.UUID]*entity.AdminRole
	alerts        map[uuid.UUID]*entity.FraudRiskAlert
	templates     map[string]*entity.NotificationTemplate
	snapshots     []entity.BackupSnapshot
}

func newMockAdminRepo() *mockAdminRepo {
	return &mockAdminRepo{
		stats:         &entity.AdminDashboardStats{TotalUsers: 10, TotalSeekers: 6, TotalProviders: 4},
		users:         make(map[uuid.UUID]*entity.User),
		verifications: make(map[uuid.UUID]*entity.ProviderVerification),
		checks:        make(map[uuid.UUID]*entity.BackgroundCheck),
		reports:       make(map[uuid.UUID]*entity.Report),
		disputes:      make(map[uuid.UUID]*entity.Dispute),
		payouts:       make(map[uuid.UUID]*entity.PayoutRequest),
		subs:          make(map[uuid.UUID]*entity.Subscription),
		plans:         make(map[uuid.UUID]*entity.SubscriptionPlan),
		wallets:       make(map[uuid.UUID]*entity.Wallet),
		categories:    make(map[uuid.UUID]*entity.Category),
		catalog:       make(map[uuid.UUID]*entity.CatalogService),
		settings:      &entity.ModerationSetting{BackgroundCheckPaymentMode: "IN_APP_STRIPE", BackgroundCheckFee: 29.99},
		roles:         make(map[uuid.UUID]*entity.AdminRole),
		alerts:        make(map[uuid.UUID]*entity.FraudRiskAlert),
		templates: map[string]*entity.NotificationTemplate{
			"welcome_email": {
				Key:      "welcome_email",
				Subject:  "Welcome {{name}}",
				BodyHTML: "<p>Hello {{name}}</p>",
			},
		},
	}
}

func (m *mockAdminRepo) GetDashboardStats(ctx context.Context) (*entity.AdminDashboardStats, error) {
	return m.stats, nil
}
func (m *mockAdminRepo) ListUsers(ctx context.Context, filter repository.AdminUserFilter) ([]entity.User, int64, error) {
	var list []entity.User
	for _, u := range m.users {
		list = append(list, *u)
	}
	return list, int64(len(list)), nil
}
func (m *mockAdminRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return m.users[id], nil
}
func (m *mockAdminRepo) UpdateUser(ctx context.Context, user *entity.User) error {
	m.users[user.ID] = user
	return nil
}
func (m *mockAdminRepo) DeleteUser(ctx context.Context, id uuid.UUID) error {
	delete(m.users, id)
	return nil
}
func (m *mockAdminRepo) ListVerifications(ctx context.Context, status string, limit, offset int) ([]entity.ProviderVerification, int64, error) {
	var list []entity.ProviderVerification
	for _, v := range m.verifications {
		list = append(list, *v)
	}
	return list, int64(len(list)), nil
}
func (m *mockAdminRepo) GetVerificationByID(ctx context.Context, id uuid.UUID) (*entity.ProviderVerification, error) {
	return m.verifications[id], nil
}
func (m *mockAdminRepo) UpdateVerification(ctx context.Context, v *entity.ProviderVerification) error {
	m.verifications[v.ID] = v
	return nil
}
func (m *mockAdminRepo) ListBackgroundChecks(ctx context.Context, status string, limit, offset int) ([]entity.BackgroundCheck, int64, error) {
	var list []entity.BackgroundCheck
	for _, bc := range m.checks {
		list = append(list, *bc)
	}
	return list, int64(len(list)), nil
}
func (m *mockAdminRepo) GetBackgroundCheckByID(ctx context.Context, id uuid.UUID) (*entity.BackgroundCheck, error) {
	return m.checks[id], nil
}
func (m *mockAdminRepo) UpdateBackgroundCheck(ctx context.Context, bc *entity.BackgroundCheck) error {
	m.checks[bc.ID] = bc
	return nil
}
func (m *mockAdminRepo) ListReports(ctx context.Context, status string, limit, offset int) ([]entity.Report, int64, error) {
	var list []entity.Report
	for _, r := range m.reports {
		list = append(list, *r)
	}
	return list, int64(len(list)), nil
}
func (m *mockAdminRepo) GetReportByID(ctx context.Context, id uuid.UUID) (*entity.Report, error) {
	return m.reports[id], nil
}
func (m *mockAdminRepo) UpdateReport(ctx context.Context, r *entity.Report) error {
	m.reports[r.ID] = r
	return nil
}
func (m *mockAdminRepo) ListDisputes(ctx context.Context, status string, limit, offset int) ([]entity.Dispute, int64, error) {
	var list []entity.Dispute
	for _, d := range m.disputes {
		list = append(list, *d)
	}
	return list, int64(len(list)), nil
}
func (m *mockAdminRepo) GetDisputeByID(ctx context.Context, id uuid.UUID) (*entity.Dispute, error) {
	return m.disputes[id], nil
}
func (m *mockAdminRepo) UpdateDispute(ctx context.Context, dispute *entity.Dispute) error {
	m.disputes[dispute.ID] = dispute
	return nil
}
func (m *mockAdminRepo) ListPayoutRequests(ctx context.Context, status string, limit, offset int) ([]entity.PayoutRequest, int64, error) {
	var list []entity.PayoutRequest
	for _, pr := range m.payouts {
		list = append(list, *pr)
	}
	return list, int64(len(list)), nil
}
func (m *mockAdminRepo) GetPayoutRequestByID(ctx context.Context, id uuid.UUID) (*entity.PayoutRequest, error) {
	return m.payouts[id], nil
}
func (m *mockAdminRepo) UpdatePayoutRequest(ctx context.Context, req *entity.PayoutRequest) error {
	m.payouts[req.ID] = req
	return nil
}
func (m *mockAdminRepo) ListSubscriptions(ctx context.Context, isActive *bool, limit, offset int) ([]entity.Subscription, int64, error) {
	var list []entity.Subscription
	for _, s := range m.subs {
		list = append(list, *s)
	}
	return list, int64(len(list)), nil
}
func (m *mockAdminRepo) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error) {
	return m.subs[id], nil
}
func (m *mockAdminRepo) GetSubscriptionByUserID(ctx context.Context, userID uuid.UUID) (*entity.Subscription, error) {
	for _, s := range m.subs {
		if s.UserID == userID {
			return s, nil
		}
	}
	return nil, nil
}
func (m *mockAdminRepo) CreateSubscription(ctx context.Context, sub *entity.Subscription) error {
	m.subs[sub.ID] = sub
	return nil
}
func (m *mockAdminRepo) UpdateSubscription(ctx context.Context, sub *entity.Subscription) error {
	m.subs[sub.ID] = sub
	return nil
}
func (m *mockAdminRepo) GetSubscriptionPlanByID(ctx context.Context, id uuid.UUID) (*entity.SubscriptionPlan, error) {
	for _, p := range m.plans {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, nil
}
func (m *mockAdminRepo) ListWallets(ctx context.Context, limit, offset int) ([]entity.Wallet, int64, error) {
	var list []entity.Wallet
	for _, w := range m.wallets {
		list = append(list, *w)
	}
	return list, int64(len(list)), nil
}
func (m *mockAdminRepo) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error) {
	return m.wallets[userID], nil
}
func (m *mockAdminRepo) UpdateWallet(ctx context.Context, wallet *entity.Wallet) error {
	m.wallets[wallet.UserID] = wallet
	return nil
}
func (m *mockAdminRepo) CreateCategory(ctx context.Context, cat *entity.Category) error {
	m.categories[cat.ID] = cat
	return nil
}
func (m *mockAdminRepo) UpdateCategory(ctx context.Context, cat *entity.Category) error {
	m.categories[cat.ID] = cat
	return nil
}
func (m *mockAdminRepo) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	delete(m.categories, id)
	return nil
}
func (m *mockAdminRepo) CreateCatalogService(ctx context.Context, cs *entity.CatalogService) error {
	m.catalog[cs.ID] = cs
	return nil
}
func (m *mockAdminRepo) UpdateCatalogService(ctx context.Context, cs *entity.CatalogService) error {
	m.catalog[cs.ID] = cs
	return nil
}
func (m *mockAdminRepo) DeleteCatalogService(ctx context.Context, id uuid.UUID) error {
	delete(m.catalog, id)
	return nil
}
func (m *mockAdminRepo) GetSettings(ctx context.Context) (*entity.ModerationSetting, error) {
	return m.settings, nil
}
func (m *mockAdminRepo) UpdateSettings(ctx context.Context, setting *entity.ModerationSetting) error {
	m.settings = setting
	return nil
}
func (m *mockAdminRepo) CreateAuditLog(ctx context.Context, log *entity.AuditLog) error {
	m.auditLogs = append(m.auditLogs, *log)
	return nil
}
func (m *mockAdminRepo) RestoreUser(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockAdminRepo) GetWalletByID(ctx context.Context, id uuid.UUID) (*entity.Wallet, error) {
	for _, w := range m.wallets {
		if w.ID == id {
			return w, nil
		}
	}
	return &entity.Wallet{ID: id, Balance: 100.0, Currency: "USD"}, nil
}
func (m *mockAdminRepo) CreateWalletTransaction(ctx context.Context, tx *entity.WalletTransaction) error {
	return nil
}
func (m *mockAdminRepo) BatchUpdateVerifications(ctx context.Context, ids []uuid.UUID, status, notes string) error {
	for _, id := range ids {
		if v, ok := m.verifications[id]; ok {
			v.Status = status
			v.ReviewerNotes = notes
		}
	}
	return nil
}
func (m *mockAdminRepo) ListFeatureFlags(ctx context.Context) ([]entity.FeatureFlag, error) {
	return []entity.FeatureFlag{{Key: "instant_payouts", IsEnabled: true}}, nil
}
func (m *mockAdminRepo) GetFeatureFlag(ctx context.Context, key string) (*entity.FeatureFlag, error) {
	return &entity.FeatureFlag{Key: key, IsEnabled: true}, nil
}
func (m *mockAdminRepo) SetFeatureFlag(ctx context.Context, key string, isEnabled bool) (*entity.FeatureFlag, error) {
	return &entity.FeatureFlag{Key: key, IsEnabled: isEnabled}, nil
}
func (m *mockAdminRepo) GetFinancialReport(ctx context.Context, startDate, endDate string) (*entity.FinancialReportSummary, error) {
	return &entity.FinancialReportSummary{
		TotalGMV:                50000,
		TotalPlatformFeeRevenue: 7500,
		TotalPayoutsProcessed:   40000,
	}, nil
}
func (m *mockAdminRepo) GetSystemHealth(ctx context.Context) (*entity.SystemHealthStatus, error) {
	return &entity.SystemHealthStatus{Status: "HEALTHY", DatabaseStatus: "CONNECTED"}, nil
}
func (m *mockAdminRepo) GetGDPRUserData(ctx context.Context, userID uuid.UUID) (*entity.GDPRUserData, error) {
	return &entity.GDPRUserData{User: entity.User{ID: userID, Email: "user@example.com"}}, nil
}
func (m *mockAdminRepo) ExportUsers(ctx context.Context, filter repository.AdminUserFilter) ([]entity.User, error) {
	var list []entity.User
	for _, u := range m.users {
		list = append(list, *u)
	}
	return list, nil
}
func (m *mockAdminRepo) ExportPayouts(ctx context.Context, status string) ([]entity.PayoutRequest, error) {
	var list []entity.PayoutRequest
	for _, p := range m.payouts {
		list = append(list, *p)
	}
	return list, nil
}
func (m *mockAdminRepo) ExportDisputes(ctx context.Context, status string) ([]entity.Dispute, error) {
	var list []entity.Dispute
	for _, d := range m.disputes {
		list = append(list, *d)
	}
	return list, nil
}
func (m *mockAdminRepo) ListAuditLogs(ctx context.Context, filter repository.AdminAuditFilter) ([]entity.AuditLog, int64, error) {
	return m.auditLogs, int64(len(m.auditLogs)), nil
}
func (m *mockAdminRepo) ListRoles(ctx context.Context) ([]entity.AdminRole, error) {
	var list []entity.AdminRole
	for _, r := range m.roles {
		list = append(list, *r)
	}
	return list, nil
}
func (m *mockAdminRepo) GetRoleByID(ctx context.Context, id uuid.UUID) (*entity.AdminRole, error) {
	return m.roles[id], nil
}
func (m *mockAdminRepo) CreateRole(ctx context.Context, role *entity.AdminRole) error {
	m.roles[role.ID] = role
	return nil
}
func (m *mockAdminRepo) UpdateRole(ctx context.Context, role *entity.AdminRole) error {
	m.roles[role.ID] = role
	return nil
}
func (m *mockAdminRepo) DeleteRole(ctx context.Context, id uuid.UUID) error {
	delete(m.roles, id)
	return nil
}
func (m *mockAdminRepo) ListFraudRiskAlerts(ctx context.Context, status string, limit, offset int) ([]entity.FraudRiskAlert, int64, error) {
	var list []entity.FraudRiskAlert
	for _, a := range m.alerts {
		list = append(list, *a)
	}
	return list, int64(len(list)), nil
}
func (m *mockAdminRepo) CreateFraudRiskAlert(ctx context.Context, alert *entity.FraudRiskAlert) error {
	m.alerts[alert.ID] = alert
	return nil
}
func (m *mockAdminRepo) UpdateFraudRiskAlert(ctx context.Context, alert *entity.FraudRiskAlert) error {
	m.alerts[alert.ID] = alert
	return nil
}
func (m *mockAdminRepo) EvaluateUserRisk(ctx context.Context, userID uuid.UUID) (*entity.FraudRiskAlert, error) {
	alert := &entity.FraudRiskAlert{
		ID:        uuid.New(),
		UserID:    userID,
		RiskScore: 35,
		RiskLevel: "MEDIUM",
		Status:    "OPEN",
	}
	m.alerts[alert.ID] = alert
	return alert, nil
}
func (m *mockAdminRepo) ListNotificationTemplates(ctx context.Context) ([]entity.NotificationTemplate, error) {
	var list []entity.NotificationTemplate
	for _, t := range m.templates {
		list = append(list, *t)
	}
	return list, nil
}
func (m *mockAdminRepo) GetNotificationTemplate(ctx context.Context, key string) (*entity.NotificationTemplate, error) {
	return m.templates[key], nil
}
func (m *mockAdminRepo) UpdateNotificationTemplate(ctx context.Context, tpl *entity.NotificationTemplate) error {
	m.templates[tpl.Key] = tpl
	return nil
}
func (m *mockAdminRepo) ListBackupSnapshots(ctx context.Context) ([]entity.BackupSnapshot, error) {
	return m.snapshots, nil
}
func (m *mockAdminRepo) GetBackupSnapshotByID(ctx context.Context, id uuid.UUID) (*entity.BackupSnapshot, error) {
	for _, s := range m.snapshots {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, errors.New("not found")
}
func (m *mockAdminRepo) GetBackupSnapshotByFilename(ctx context.Context, filename string) (*entity.BackupSnapshot, error) {
	for _, s := range m.snapshots {
		if s.Filename == filename {
			return &s, nil
		}
	}
	return nil, errors.New("not found")
}
func (m *mockAdminRepo) CreateBackupSnapshot(ctx context.Context, snapshot *entity.BackupSnapshot) error {
	m.snapshots = append(m.snapshots, *snapshot)
	return nil
}
func (m *mockAdminRepo) ListSubscriptionPlans(ctx context.Context) ([]entity.SubscriptionPlan, error) {
	return []entity.SubscriptionPlan{}, nil
}
func (m *mockAdminRepo) CreateSubscriptionPlan(ctx context.Context, plan *entity.SubscriptionPlan) error {
	return nil
}
func (m *mockAdminRepo) UpdateSubscriptionPlan(ctx context.Context, plan *entity.SubscriptionPlan) error {
	return nil
}
func (m *mockAdminRepo) DeleteSubscriptionPlan(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockAdminRepo) BatchApprovePayouts(ctx context.Context, ids []uuid.UUID, adminID uuid.UUID) (*entity.BatchPayoutResponse, error) {
	return &entity.BatchPayoutResponse{ProcessedCount: int64(len(ids)), Message: "Approved"}, nil
}
func (m *mockAdminRepo) BatchRejectPayouts(ctx context.Context, ids []uuid.UUID, reason string, adminID uuid.UUID) (*entity.BatchPayoutResponse, error) {
	return &entity.BatchPayoutResponse{ProcessedCount: int64(len(ids)), Message: "Rejected"}, nil
}
func (m *mockAdminRepo) ListStaffNotes(ctx context.Context, userID uuid.UUID) ([]entity.StaffNote, error) {
	return []entity.StaffNote{}, nil
}
func (m *mockAdminRepo) CreateStaffNote(ctx context.Context, note *entity.StaffNote) error {
	return nil
}
func (m *mockAdminRepo) DeleteStaffNote(ctx context.Context, noteID uuid.UUID) error {
	return nil
}
func (m *mockAdminRepo) GetAdminNotificationFeed(ctx context.Context) ([]entity.AdminNotification, error) {
	return []entity.AdminNotification{}, nil
}
func (m *mockAdminRepo) GetProviderOnboardingFunnel(ctx context.Context) (*entity.ProviderOnboardingFunnel, error) {
	return &entity.ProviderOnboardingFunnel{}, nil
}
func (m *mockAdminRepo) ListAppointments(ctx context.Context, status string, search string, limit, offset int) ([]entity.Appointment, int64, error) {
	return []entity.Appointment{}, 0, nil
}
func (m *mockAdminRepo) GetAppointmentByID(ctx context.Context, id uuid.UUID) (*entity.Appointment, error) {
	return &entity.Appointment{ID: id}, nil
}
func (m *mockAdminRepo) UpdateAppointmentStatus(ctx context.Context, id uuid.UUID, status string) error {
	return nil
}
func (m *mockAdminRepo) ExportAppointments(ctx context.Context, status string) ([]entity.Appointment, error) {
	return []entity.Appointment{}, nil
}
func (m *mockAdminRepo) ListReviews(ctx context.Context, rating *float64, isHidden *bool, limit, offset int) ([]entity.Review, int64, error) {
	return []entity.Review{}, 0, nil
}
func (m *mockAdminRepo) CreateReview(ctx context.Context, review *entity.Review) error {
	return nil
}
func (m *mockAdminRepo) ToggleReviewVisibility(ctx context.Context, id uuid.UUID, isHidden bool) error {
	return nil
}
func (m *mockAdminRepo) DeleteReview(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockAdminRepo) ListPromoCodes(ctx context.Context) ([]entity.PromoCode, error) {
	return []entity.PromoCode{}, nil
}
func (m *mockAdminRepo) CreatePromoCode(ctx context.Context, promo *entity.PromoCode) error {
	return nil
}
func (m *mockAdminRepo) UpdatePromoCode(ctx context.Context, promo *entity.PromoCode) error {
	return nil
}
func (m *mockAdminRepo) DeletePromoCode(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockAdminRepo) ExecuteDisputeRefund(ctx context.Context, disputeID uuid.UUID, amount float64, reason string, adminID uuid.UUID) error {
	return nil
}
func (m *mockAdminRepo) ExportVerifications(ctx context.Context, status string) ([]entity.ProviderVerification, error) {
	return []entity.ProviderVerification{}, nil
}
func (m *mockAdminRepo) ExportBackgroundChecks(ctx context.Context, status string) ([]entity.BackgroundCheck, error) {
	return []entity.BackgroundCheck{}, nil
}
func (m *mockAdminRepo) ListLegalDocuments(ctx context.Context, docType string, isActive *bool) ([]entity.LegalDocument, error) {
	return []entity.LegalDocument{}, nil
}
func (m *mockAdminRepo) GetLegalDocumentByID(ctx context.Context, id uuid.UUID) (*entity.LegalDocument, error) {
	return &entity.LegalDocument{ID: id}, nil
}
func (m *mockAdminRepo) CreateLegalDocument(ctx context.Context, doc *entity.LegalDocument) error {
	return nil
}
func (m *mockAdminRepo) UpdateLegalDocument(ctx context.Context, doc *entity.LegalDocument) error {
	return nil
}
func (m *mockAdminRepo) DeleteLegalDocument(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockAdminRepo) ListContactMessages(ctx context.Context, isResolved *bool, search string, limit, offset int) ([]entity.ContactMessage, int64, error) {
	return []entity.ContactMessage{}, 0, nil
}
func (m *mockAdminRepo) ToggleContactMessageResolved(ctx context.Context, id uuid.UUID, isResolved bool) error {
	return nil
}
func (m *mockAdminRepo) DeleteContactMessage(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockAdminRepo) ListResolutionReports(ctx context.Context, isReviewed *bool, search string, limit, offset int) ([]entity.ResolutionReport, int64, error) {
	return []entity.ResolutionReport{}, 0, nil
}
func (m *mockAdminRepo) ToggleResolutionReportReviewed(ctx context.Context, id uuid.UUID, isReviewed bool) error {
	return nil
}
func (m *mockAdminRepo) DeleteResolutionReport(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockAdminRepo) ListFAQs(ctx context.Context, category string, isActive *bool) ([]entity.FAQ, error) {
	return []entity.FAQ{}, nil
}
func (m *mockAdminRepo) CreateFAQ(ctx context.Context, faq *entity.FAQ) error {
	return nil
}
func (m *mockAdminRepo) UpdateFAQ(ctx context.Context, faq *entity.FAQ) error {
	return nil
}
func (m *mockAdminRepo) DeleteFAQ(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockAdminRepo) ListTestimonials(ctx context.Context, isActive *bool) ([]entity.Testimonial, error) {
	return []entity.Testimonial{}, nil
}
func (m *mockAdminRepo) CreateTestimonial(ctx context.Context, t *entity.Testimonial) error {
	return nil
}
func (m *mockAdminRepo) UpdateTestimonial(ctx context.Context, t *entity.Testimonial) error {
	return nil
}
func (m *mockAdminRepo) DeleteTestimonial(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockAdminRepo) GetHeroSection(ctx context.Context) (*entity.HeroSection, error) {
	return &entity.HeroSection{}, nil
}
func (m *mockAdminRepo) UpdateHeroSection(ctx context.Context, hero *entity.HeroSection) error {
	return nil
}
func (m *mockAdminRepo) ListSiteStats(ctx context.Context) ([]entity.SiteStat, error) {
	return []entity.SiteStat{}, nil
}
func (m *mockAdminRepo) CreateSiteStat(ctx context.Context, stat *entity.SiteStat) error {
	return nil
}
func (m *mockAdminRepo) UpdateSiteStat(ctx context.Context, stat *entity.SiteStat) error {
	return nil
}
func (m *mockAdminRepo) DeleteSiteStat(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockAdminRepo) GetAboutContent(ctx context.Context) (*entity.AboutContent, error) {
	return &entity.AboutContent{}, nil
}
func (m *mockAdminRepo) UpdateAboutContent(ctx context.Context, about *entity.AboutContent) error {
	return nil
}

func TestAdminUseCase_DashboardAndUserOperations(t *testing.T) {
	adminID := uuid.New()
	targetUserID := uuid.New()

	adminRepo := newMockAdminRepo()
	user := &entity.User{
		ID:         targetUserID,
		Email:      "provider@example.com",
		IsActive:   true,
		IsStaff:    false,
		IsVerified: true,
		Profile: &entity.Profile{
			ID:               uuid.New(),
			UserID:           targetUserID,
			UserType:         "SEEKER",
			SubscriptionTier: "NONE",
		},
	}
	adminRepo.users[targetUserID] = user
	profileRepo := &mockProfileRepo{profiles: map[uuid.UUID]*entity.Profile{targetUserID: user.Profile}}
	userRepo := &mockUserRepo{users: map[string]*entity.User{user.Email: user}}
	walletRepo := &mockWalletRepo{wallets: map[uuid.UUID]*entity.Wallet{}}

	cfg := &config.Config{JWTSecret: "test-super-secret-jwt-key-for-admin-tests-12345"}
	adminUC := usecase.NewAdminUseCase(adminRepo, profileRepo, userRepo, walletRepo, cfg)
	ctx := context.Background()

	// 1. Test Dashboard Stats
	stats, err := adminUC.GetDashboardStats(ctx, adminID)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(10), stats.TotalUsers)

	// 2. Test User Update (Role upgrade to PROVIDER and GOLD tier)
	newType := "PROVIDER"
	newTier := "GOLD"
	updatedUser, err := adminUC.UpdateUser(ctx, adminID, targetUserID, domainUsecase.AdminUpdateUserInput{
		UserType:         &newType,
		SubscriptionTier: &newTier,
	})
	assert.NoError(t, err)
	assert.Equal(t, "PROVIDER", updatedUser.Profile.UserType)
	assert.Equal(t, "GOLD", updatedUser.Profile.SubscriptionTier)
	assert.NotEmpty(t, adminRepo.auditLogs)
}

func TestAdminUseCase_DisputesAndPayouts(t *testing.T) {
	adminID := uuid.New()
	disputeID := uuid.New()
	payoutID := uuid.New()
	walletID := uuid.New()
	providerID := uuid.New()

	adminRepo := newMockAdminRepo()
	dispute := &entity.Dispute{
		ID:     disputeID,
		Status: "OPEN",
	}
	adminRepo.disputes[disputeID] = dispute

	payout := &entity.PayoutRequest{
		ID:       payoutID,
		WalletID: walletID,
		Amount:   150.00,
		Status:   "PENDING",
	}
	adminRepo.payouts[payoutID] = payout

	wallet := &entity.Wallet{
		ID:       walletID,
		UserID:   providerID,
		Balance:  50.00,
		Currency: "USD",
	}
	adminRepo.wallets[providerID] = wallet

	profileRepo := &mockProfileRepo{profiles: make(map[uuid.UUID]*entity.Profile)}
	userRepo := &mockUserRepo{users: make(map[string]*entity.User)}
	walletRepo := &mockWalletRepo{wallets: map[uuid.UUID]*entity.Wallet{walletID: wallet, providerID: wallet}}
	cfg := &config.Config{JWTSecret: "test-super-secret-jwt-key-for-admin-tests-12345"}

	adminUC := usecase.NewAdminUseCase(adminRepo, profileRepo, userRepo, walletRepo, cfg)
	ctx := context.Background()

	// 1. Resolve Dispute
	err := adminUC.ResolveDispute(ctx, adminID, disputeID, "Both parties agreed")
	assert.NoError(t, err)
	assert.Equal(t, "RESOLVED", adminRepo.disputes[disputeID].Status)

	// 2. Approve Payout
	err = adminUC.ApprovePayout(ctx, adminID, payoutID)
	assert.NoError(t, err)
	assert.Equal(t, "PROCESSED", adminRepo.payouts[payoutID].Status)
	assert.NotNil(t, adminRepo.payouts[payoutID].ProcessedAt)

	// 3. Reject Payout with balance refund
	payout2ID := uuid.New()
	adminRepo.payouts[payout2ID] = &entity.PayoutRequest{
		ID:       payout2ID,
		WalletID: walletID,
		Amount:   50.00,
		Status:   "PENDING",
	}
	err = adminUC.RejectPayout(ctx, adminID, payout2ID, "Invalid bank details")
	assert.NoError(t, err)
	assert.Equal(t, "REJECTED", adminRepo.payouts[payout2ID].Status)
	assert.Equal(t, 100.00, wallet.Balance)
}

func TestAdminUseCase_VerificationAndBackgroundChecks(t *testing.T) {
	adminID := uuid.New()
	vID := uuid.New()
	bcID := uuid.New()
	providerID := uuid.New()

	adminRepo := newMockAdminRepo()
	verification := &entity.ProviderVerification{
		ID:         vID,
		ProviderID: providerID,
		Status:     "PENDING",
	}
	adminRepo.verifications[vID] = verification

	bc := &entity.BackgroundCheck{
		ID:         bcID,
		ProviderID: providerID,
		Status:     "PENDING",
	}
	adminRepo.checks[bcID] = bc

	profile := &entity.Profile{
		ID:                 uuid.New(),
		UserID:             providerID,
		IsIdentityVerified: false,
	}
	profileRepo := &mockProfileRepo{profiles: map[uuid.UUID]*entity.Profile{providerID: profile}}
	userRepo := &mockUserRepo{users: make(map[string]*entity.User)}
	walletRepo := &mockWalletRepo{wallets: make(map[uuid.UUID]*entity.Wallet)}
	cfg := &config.Config{JWTSecret: "test-super-secret-jwt-key-for-admin-tests-12345"}

	adminUC := usecase.NewAdminUseCase(adminRepo, profileRepo, userRepo, walletRepo, cfg)
	ctx := context.Background()

	// 1. Approve Verification
	err := adminUC.ApproveVerification(ctx, adminID, vID)
	assert.NoError(t, err)
	assert.Equal(t, "APPROVED", adminRepo.verifications[vID].Status)
	assert.True(t, profile.IsIdentityVerified)

	// 2. Override Background Check
	err = adminUC.OverrideBackgroundCheck(ctx, adminID, bcID, "CLEAR", "Manual passport verification confirmed")
	assert.NoError(t, err)
	assert.Equal(t, "CLEAR", adminRepo.checks[bcID].Status)
	assert.True(t, profile.IsIdentityVerified)
}

func TestAdminUseCase_AdvancedOperations(t *testing.T) {
	adminID := uuid.New()
	targetUserID := uuid.New()
	walletID := uuid.New()

	adminRepo := newMockAdminRepo()
	wallet := &entity.Wallet{ID: walletID, UserID: targetUserID, Balance: 100.0, Currency: "USD"}
	adminRepo.wallets[targetUserID] = wallet
	adminRepo.wallets[walletID] = wallet

	user := &entity.User{ID: targetUserID, Email: "target@example.com", IsActive: true}
	adminRepo.users[targetUserID] = user

	profileRepo := &mockProfileRepo{profiles: make(map[uuid.UUID]*entity.Profile)}
	userRepo := &mockUserRepo{users: make(map[string]*entity.User)}
	walletRepo := &mockWalletRepo{wallets: map[uuid.UUID]*entity.Wallet{walletID: wallet, targetUserID: wallet}}
	cfg := &config.Config{JWTSecret: "test-super-secret-jwt-key-for-admin-tests-12345"}

	adminUC := usecase.NewAdminUseCase(adminRepo, profileRepo, userRepo, walletRepo, cfg)
	ctx := context.Background()

	// 1. Wallet Adjustment
	adjWallet, err := adminUC.AdjustWalletBalance(ctx, adminID, walletID, 50.00, "Goodwill customer credit")
	assert.NoError(t, err)
	assert.Equal(t, 150.00, adjWallet.Balance)

	// Negative adjustment beyond balance should fail
	_, err = adminUC.AdjustWalletBalance(ctx, adminID, walletID, -200.00, "Should fail")
	assert.Error(t, err)

	// 2. Batch Verifications
	v1ID := uuid.New()
	v2ID := uuid.New()
	adminRepo.verifications[v1ID] = &entity.ProviderVerification{ID: v1ID, Status: "PENDING"}
	adminRepo.verifications[v2ID] = &entity.ProviderVerification{ID: v2ID, Status: "PENDING"}

	err = adminUC.BatchVerifications(ctx, adminID, []uuid.UUID{v1ID, v2ID}, "APPROVE", "Batch approved")
	assert.NoError(t, err)
	assert.Equal(t, "APPROVED", adminRepo.verifications[v1ID].Status)
	assert.Equal(t, "APPROVED", adminRepo.verifications[v2ID].Status)

	// 3. Feature Flags
	flag, err := adminUC.SetFeatureFlag(ctx, adminID, "instant_payouts", true)
	assert.NoError(t, err)
	assert.True(t, flag.IsEnabled)

	flags, err := adminUC.ListFeatureFlags(ctx, adminID)
	assert.NoError(t, err)
	assert.NotEmpty(t, flags)

	// 4. Financial Report & System Health
	finReport, err := adminUC.GetFinancialReport(ctx, adminID, "2026-01-01", "2026-12-31")
	assert.NoError(t, err)
	assert.Equal(t, float64(50000), finReport.TotalGMV)

	health, err := adminUC.GetSystemHealth(ctx, adminID)
	assert.NoError(t, err)
	assert.Equal(t, "HEALTHY", health.Status)

	// 5. GDPR Export
	gdpr, err := adminUC.GetGDPRUserData(ctx, adminID, targetUserID)
	assert.NoError(t, err)
	assert.Equal(t, targetUserID, gdpr.User.ID)

	// 6. CSV Exports
	usersCSV, err := adminUC.ExportUsersCSV(ctx, adminID, repository.AdminUserFilter{})
	assert.NoError(t, err)
	assert.Contains(t, string(usersCSV), "target@example.com")

	payoutsCSV, err := adminUC.ExportPayoutsCSV(ctx, adminID, "")
	assert.NoError(t, err)
	assert.Contains(t, string(payoutsCSV), "WalletID")

	disputesCSV, err := adminUC.ExportDisputesCSV(ctx, adminID, "")
	assert.NoError(t, err)
	assert.Contains(t, string(disputesCSV), "RaisedByID")
}

func TestAdminUseCase_EnterpriseFeatures(t *testing.T) {
	adminID := uuid.New()
	targetUserID := uuid.New()

	adminRepo := newMockAdminRepo()
	user := &entity.User{
		ID:         targetUserID,
		Email:      "target@example.com",
		IsActive:   true,
		IsStaff:    false,
		IsVerified: true,
	}
	adminRepo.users[targetUserID] = user
	adminRepo.users[adminID] = &entity.User{
		ID:         adminID,
		Email:      "admin@example.com",
		IsActive:   true,
		IsStaff:    true,
		IsVerified: true,
	}

	profileRepo := &mockProfileRepo{profiles: make(map[uuid.UUID]*entity.Profile)}
	userRepo := &mockUserRepo{users: map[string]*entity.User{"target@example.com": user, "admin@example.com": adminRepo.users[adminID]}}
	walletRepo := &mockWalletRepo{wallets: make(map[uuid.UUID]*entity.Wallet)}
	cfg := &config.Config{JWTSecret: "test-super-secret-jwt-key-for-admin-tests-12345"}

	adminUC := usecase.NewAdminUseCase(adminRepo, profileRepo, userRepo, walletRepo, cfg)
	ctx := context.Background()

	// 1. TOTP 2FA Setup
	setupResp, err := adminUC.Setup2FA(ctx, adminID)
	assert.NoError(t, err)
	assert.NotEmpty(t, setupResp.Secret)
	assert.NotEmpty(t, setupResp.OTPAuthURL)

	// 2. Staff Impersonation
	impResult, err := adminUC.ImpersonateUser(ctx, adminID, targetUserID)
	assert.NoError(t, err)
	assert.NotEmpty(t, impResult.AccessToken)
	assert.Equal(t, 3600, impResult.ExpiresIn)

	// 3. RBAC Roles
	role, err := adminUC.CreateRole(ctx, adminID, &entity.AdminRole{
		Name:        "Risk Manager",
		Slug:        "risk-manager",
		Permissions: entity.JSONSlice{"admin:fraud:manage", "admin:payout:approve"},
	})
	assert.NoError(t, err)
	assert.Equal(t, "Risk Manager", role.Name)

	roles, err := adminUC.ListRoles(ctx, adminID)
	assert.NoError(t, err)
	assert.NotEmpty(t, roles)

	err = adminUC.AssignUserRole(ctx, adminID, targetUserID, role.ID)
	assert.NoError(t, err)

	// 4. Fraud Risk Alerts
	alert, err := adminUC.EvaluateUserRisk(ctx, adminID, targetUserID)
	assert.NoError(t, err)
	assert.NotNil(t, alert)

	alerts, count, err := adminUC.ListFraudRiskAlerts(ctx, adminID, "", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
	assert.NotEmpty(t, alerts)

	err = adminUC.ResolveRiskAlert(ctx, adminID, alert.ID, "ACTIONED")
	assert.NoError(t, err)

	// 5. Notification Templates
	templates, err := adminUC.ListNotificationTemplates(ctx, adminID)
	assert.NoError(t, err)
	assert.NotEmpty(t, templates)

	updatedTpl, err := adminUC.UpdateNotificationTemplate(ctx, adminID, "welcome_email", "Updated Subject", "<p>Updated</p>", "Updated")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Subject", updatedTpl.Subject)

	// 6. Database Backups
	snap, err := adminUC.TriggerDatabaseBackup(ctx, adminID)
	assert.NoError(t, err)
	assert.NotEmpty(t, snap.Filename)

	snaps, err := adminUC.ListBackupSnapshots(ctx, adminID)
	assert.NoError(t, err)
	assert.NotEmpty(t, snaps)
}
