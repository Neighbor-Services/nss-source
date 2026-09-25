package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend-go/internal/delivery/http/handler"
	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	domainUsecase "backend-go/internal/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockAdminUseCase struct {
	stats       *entity.AdminDashboardStats
	users       []entity.User
	user        *entity.User
	verifs      []entity.ProviderVerification
	checks      []entity.BackgroundCheck
	reports     []entity.Report
	disputes    []entity.Dispute
	payouts     []entity.PayoutRequest
	wallets     []entity.Wallet
	subs        []entity.Subscription
	category    *entity.Category
	service     *entity.CatalogService
	settings    *entity.ModerationSetting
	auditLogs   []entity.AuditLog
	updateInput domainUsecase.AdminUpdateUserInput
	err         error
}

func (m *mockAdminUseCase) GetDashboardStats(ctx context.Context, adminID uuid.UUID) (*entity.AdminDashboardStats, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.stats, nil
}

func (m *mockAdminUseCase) ListUsers(ctx context.Context, adminID uuid.UUID, filter repository.AdminUserFilter) ([]entity.User, int64, error) {
	return m.users, int64(len(m.users)), m.err
}

func (m *mockAdminUseCase) GetUserByID(ctx context.Context, adminID, userID uuid.UUID) (*entity.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

func (m *mockAdminUseCase) UpdateUser(ctx context.Context, adminID, userID uuid.UUID, input domainUsecase.AdminUpdateUserInput) (*entity.User, error) {
	m.updateInput = input
	return m.user, m.err
}

func (m *mockAdminUseCase) DeleteUser(ctx context.Context, adminID, userID uuid.UUID) error {
	return m.err
}

func (m *mockAdminUseCase) ListVerifications(ctx context.Context, adminID uuid.UUID, status string, limit, offset int) ([]entity.ProviderVerification, int64, error) {
	return m.verifs, int64(len(m.verifs)), m.err
}

func (m *mockAdminUseCase) ApproveVerification(ctx context.Context, adminID, verificationID uuid.UUID) error {
	return m.err
}

func (m *mockAdminUseCase) RejectVerification(ctx context.Context, adminID, verificationID uuid.UUID) error {
	return m.err
}

func (m *mockAdminUseCase) ListBackgroundChecks(ctx context.Context, adminID uuid.UUID, status string, limit, offset int) ([]entity.BackgroundCheck, int64, error) {
	return m.checks, int64(len(m.checks)), m.err
}

func (m *mockAdminUseCase) OverrideBackgroundCheck(ctx context.Context, adminID, checkID uuid.UUID, status, notes string) error {
	return m.err
}

func (m *mockAdminUseCase) ListReports(ctx context.Context, adminID uuid.UUID, status string, limit, offset int) ([]entity.Report, int64, error) {
	return m.reports, int64(len(m.reports)), m.err
}

func (m *mockAdminUseCase) ResolveReport(ctx context.Context, adminID, reportID uuid.UUID, resolution string) error {
	return m.err
}

func (m *mockAdminUseCase) ListDisputes(ctx context.Context, adminID uuid.UUID, status string, limit, offset int) ([]entity.Dispute, int64, error) {
	return m.disputes, int64(len(m.disputes)), m.err
}

func (m *mockAdminUseCase) ResolveDispute(ctx context.Context, adminID, disputeID uuid.UUID, notes string) error {
	return m.err
}

func (m *mockAdminUseCase) RejectDispute(ctx context.Context, adminID, disputeID uuid.UUID, notes string) error {
	return m.err
}

func (m *mockAdminUseCase) ListPayoutRequests(ctx context.Context, adminID uuid.UUID, status string, limit, offset int) ([]entity.PayoutRequest, int64, error) {
	return m.payouts, int64(len(m.payouts)), m.err
}

func (m *mockAdminUseCase) ApprovePayout(ctx context.Context, adminID, payoutID uuid.UUID) error {
	return m.err
}

func (m *mockAdminUseCase) RejectPayout(ctx context.Context, adminID, payoutID uuid.UUID, notes string) error {
	return m.err
}

func (m *mockAdminUseCase) ListWallets(ctx context.Context, adminID uuid.UUID, limit, offset int) ([]entity.Wallet, int64, error) {
	return m.wallets, int64(len(m.wallets)), m.err
}

func (m *mockAdminUseCase) ListSubscriptions(ctx context.Context, adminID uuid.UUID, isActive *bool, limit, offset int) ([]entity.Subscription, int64, error) {
	return m.subs, int64(len(m.subs)), m.err
}

func (m *mockAdminUseCase) ToggleSubscription(ctx context.Context, adminID, subID uuid.UUID, isActive bool) error {
	return m.err
}

func (m *mockAdminUseCase) AssignSubscription(ctx context.Context, adminID uuid.UUID, input domainUsecase.AssignSubscriptionInput) (*entity.Subscription, error) {
	if len(m.subs) > 0 {
		return &m.subs[0], m.err
	}
	return nil, m.err
}

func (m *mockAdminUseCase) CreateCategory(ctx context.Context, adminID uuid.UUID, cat *entity.Category) (*entity.Category, error) {
	return m.category, m.err
}

func (m *mockAdminUseCase) UpdateCategory(ctx context.Context, adminID, catID uuid.UUID, name, description, icon string) (*entity.Category, error) {
	return m.category, m.err
}

func (m *mockAdminUseCase) DeleteCategory(ctx context.Context, adminID, catID uuid.UUID) error {
	return m.err
}

func (m *mockAdminUseCase) CreateCatalogService(ctx context.Context, adminID uuid.UUID, cs *entity.CatalogService) (*entity.CatalogService, error) {
	return m.service, m.err
}

func (m *mockAdminUseCase) UpdateCatalogService(ctx context.Context, adminID, csID uuid.UUID, name, description string, basePrice float64, categoryID *uuid.UUID) (*entity.CatalogService, error) {
	return m.service, m.err
}

func (m *mockAdminUseCase) DeleteCatalogService(ctx context.Context, adminID, csID uuid.UUID) error {
	return m.err
}

func (m *mockAdminUseCase) GetSettings(ctx context.Context, adminID uuid.UUID) (*entity.ModerationSetting, error) {
	return m.settings, m.err
}

func (m *mockAdminUseCase) UpdateSettings(ctx context.Context, adminID uuid.UUID, paymentMode string, fee float64, broadcastRadiusKm, matchRadiusKm float64) (*entity.ModerationSetting, error) {
	return m.settings, m.err
}

func (m *mockAdminUseCase) ListAuditLogs(ctx context.Context, adminID uuid.UUID, filter repository.AdminAuditFilter) ([]entity.AuditLog, int64, error) {
	return m.auditLogs, int64(len(m.auditLogs)), m.err
}

func (m *mockAdminUseCase) RestoreUser(ctx context.Context, adminID, userID uuid.UUID) error {
	return m.err
}

func (m *mockAdminUseCase) AdjustWalletBalance(ctx context.Context, adminID, walletID uuid.UUID, amount float64, reason string) (*entity.Wallet, error) {
	return &entity.Wallet{ID: walletID, Balance: 150.0}, m.err
}

func (m *mockAdminUseCase) BatchVerifications(ctx context.Context, adminID uuid.UUID, ids []uuid.UUID, action, notes string) error {
	return m.err
}

func (m *mockAdminUseCase) BroadcastNotification(ctx context.Context, adminID uuid.UUID, targetUserType, title, message string) (int, error) {
	return 25, m.err
}

func (m *mockAdminUseCase) ListFeatureFlags(ctx context.Context, adminID uuid.UUID) ([]entity.FeatureFlag, error) {
	return []entity.FeatureFlag{{Key: "instant_payouts", IsEnabled: true}}, m.err
}

func (m *mockAdminUseCase) SetFeatureFlag(ctx context.Context, adminID uuid.UUID, key string, isEnabled bool) (*entity.FeatureFlag, error) {
	return &entity.FeatureFlag{Key: key, IsEnabled: isEnabled}, m.err
}

func (m *mockAdminUseCase) GetFinancialReport(ctx context.Context, adminID uuid.UUID, startDate, endDate string) (*entity.FinancialReportSummary, error) {
	return &entity.FinancialReportSummary{TotalGMV: 50000, TotalPlatformFeeRevenue: 7500}, m.err
}

func (m *mockAdminUseCase) GetSystemHealth(ctx context.Context, adminID uuid.UUID) (*entity.SystemHealthStatus, error) {
	return &entity.SystemHealthStatus{Status: "HEALTHY", DatabaseStatus: "CONNECTED"}, m.err
}

func (m *mockAdminUseCase) GetGDPRUserData(ctx context.Context, adminID, userID uuid.UUID) (*entity.GDPRUserData, error) {
	return &entity.GDPRUserData{User: entity.User{ID: userID, Email: "user@example.com"}}, m.err
}

func (m *mockAdminUseCase) ExportUsersCSV(ctx context.Context, adminID uuid.UUID, filter repository.AdminUserFilter) ([]byte, error) {
	return []byte("ID,Email\n1,test@example.com"), m.err
}

func (m *mockAdminUseCase) ExportPayoutsCSV(ctx context.Context, adminID uuid.UUID, status string) ([]byte, error) {
	return []byte("ID,Amount\n1,100.00"), m.err
}

func (m *mockAdminUseCase) ExportDisputesCSV(ctx context.Context, adminID uuid.UUID, status string) ([]byte, error) {
	return []byte("ID,Reason\n1,Service not rendered"), m.err
}

func (m *mockAdminUseCase) Setup2FA(ctx context.Context, adminID uuid.UUID) (*entity.TOTPSetupResponse, error) {
	return &entity.TOTPSetupResponse{Secret: "JBSWY3DPEHPK3PXP", OTPAuthURL: "otpauth://totp/NS:admin?secret=JBSWY3DPEHPK3PXP"}, m.err
}

func (m *mockAdminUseCase) VerifyAndEnable2FA(ctx context.Context, adminID uuid.UUID, code string) error {
	return m.err
}

func (m *mockAdminUseCase) Disable2FA(ctx context.Context, adminID uuid.UUID, code string) error {
	return m.err
}

func (m *mockAdminUseCase) ImpersonateUser(ctx context.Context, adminID, targetUserID uuid.UUID) (*entity.ImpersonationResult, error) {
	return &entity.ImpersonationResult{AccessToken: "mock-jwt-impersonation", ExpiresIn: 3600}, m.err
}

func (m *mockAdminUseCase) ListRoles(ctx context.Context, adminID uuid.UUID) ([]entity.AdminRole, error) {
	return []entity.AdminRole{{Name: "Support", Slug: "support"}}, m.err
}

func (m *mockAdminUseCase) CreateRole(ctx context.Context, adminID uuid.UUID, role *entity.AdminRole) (*entity.AdminRole, error) {
	return role, m.err
}

func (m *mockAdminUseCase) AssignUserRole(ctx context.Context, adminID, userID, roleID uuid.UUID) error {
	return m.err
}

func (m *mockAdminUseCase) ListFraudRiskAlerts(ctx context.Context, adminID uuid.UUID, status string, limit, offset int) ([]entity.FraudRiskAlert, int64, error) {
	return []entity.FraudRiskAlert{{RiskScore: 75, RiskLevel: "HIGH"}}, 1, m.err
}

func (m *mockAdminUseCase) EvaluateUserRisk(ctx context.Context, adminID, targetUserID uuid.UUID) (*entity.FraudRiskAlert, error) {
	return &entity.FraudRiskAlert{RiskScore: 80, RiskLevel: "CRITICAL"}, m.err
}

func (m *mockAdminUseCase) ResolveRiskAlert(ctx context.Context, adminID, alertID uuid.UUID, action string) error {
	return m.err
}

func (m *mockAdminUseCase) ListNotificationTemplates(ctx context.Context, adminID uuid.UUID) ([]entity.NotificationTemplate, error) {
	return []entity.NotificationTemplate{{Key: "welcome_email", Subject: "Welcome"}}, m.err
}

func (m *mockAdminUseCase) UpdateNotificationTemplate(ctx context.Context, adminID uuid.UUID, key, subject, bodyHTML, bodyText string) (*entity.NotificationTemplate, error) {
	return &entity.NotificationTemplate{Key: key, Subject: subject, BodyHTML: bodyHTML, BodyText: bodyText}, m.err
}

func (m *mockAdminUseCase) TestSendEmailTemplate(ctx context.Context, adminID uuid.UUID, key, recipientEmail string) error {
	return m.err
}

func (m *mockAdminUseCase) TriggerDatabaseBackup(ctx context.Context, adminID uuid.UUID) (*entity.BackupSnapshot, error) {
	return &entity.BackupSnapshot{Filename: "backup_test.dump", Status: "COMPLETED"}, m.err
}

func (m *mockAdminUseCase) ListBackupSnapshots(ctx context.Context, adminID uuid.UUID) ([]entity.BackupSnapshot, error) {
	return []entity.BackupSnapshot{{Filename: "backup_test.dump", Status: "COMPLETED"}}, m.err
}

func (m *mockAdminUseCase) CreateUser(ctx context.Context, adminID uuid.UUID, input domainUsecase.AdminCreateUserInput) (*entity.User, error) {
	return m.user, m.err
}

func (m *mockAdminUseCase) ListSubscriptionPlans(ctx context.Context, adminID uuid.UUID) ([]entity.SubscriptionPlan, error) {
	return []entity.SubscriptionPlan{}, m.err
}

func (m *mockAdminUseCase) CreateSubscriptionPlan(ctx context.Context, adminID uuid.UUID, plan *entity.SubscriptionPlan) (*entity.SubscriptionPlan, error) {
	return plan, m.err
}

func (m *mockAdminUseCase) UpdateSubscriptionPlan(ctx context.Context, adminID uuid.UUID, plan *entity.SubscriptionPlan) (*entity.SubscriptionPlan, error) {
	return plan, m.err
}

func (m *mockAdminUseCase) DeleteSubscriptionPlan(ctx context.Context, adminID, planID uuid.UUID) error {
	return m.err
}

func (m *mockAdminUseCase) UpdateRole(ctx context.Context, adminID, roleID uuid.UUID, role *entity.AdminRole) (*entity.AdminRole, error) {
	return role, m.err
}

func (m *mockAdminUseCase) DeleteRole(ctx context.Context, adminID, roleID uuid.UUID) error {
	return m.err
}

func (m *mockAdminUseCase) BatchApprovePayouts(ctx context.Context, adminID uuid.UUID, ids []uuid.UUID) (*entity.BatchPayoutResponse, error) {
	return &entity.BatchPayoutResponse{ProcessedCount: int64(len(ids)), Message: "Approved"}, m.err
}

func (m *mockAdminUseCase) BatchRejectPayouts(ctx context.Context, adminID uuid.UUID, ids []uuid.UUID, reason string) (*entity.BatchPayoutResponse, error) {
	return &entity.BatchPayoutResponse{ProcessedCount: int64(len(ids)), Message: "Rejected"}, m.err
}

func (m *mockAdminUseCase) GetAdminNotificationFeed(ctx context.Context, adminID uuid.UUID) ([]entity.AdminNotification, error) {
	return []entity.AdminNotification{}, m.err
}

func (m *mockAdminUseCase) GetProviderOnboardingFunnel(ctx context.Context, adminID uuid.UUID) (*entity.ProviderOnboardingFunnel, error) {
	return &entity.ProviderOnboardingFunnel{}, m.err
}

func (m *mockAdminUseCase) ListStaffNotes(ctx context.Context, adminID, userID uuid.UUID) ([]entity.StaffNote, error) {
	return []entity.StaffNote{}, m.err
}

func (m *mockAdminUseCase) CreateStaffNote(ctx context.Context, adminID, userID uuid.UUID, content, category string, isPinned bool) (*entity.StaffNote, error) {
	return &entity.StaffNote{}, m.err
}

func (m *mockAdminUseCase) DeleteStaffNote(ctx context.Context, adminID, noteID uuid.UUID) error {
	return m.err
}

func (m *mockAdminUseCase) ListAppointments(ctx context.Context, adminID uuid.UUID, status, search string, limit, offset int) ([]entity.Appointment, int64, error) {
	return []entity.Appointment{}, 0, m.err
}

func (m *mockAdminUseCase) GetAppointmentByID(ctx context.Context, adminID, id uuid.UUID) (*entity.Appointment, error) {
	return &entity.Appointment{ID: id}, m.err
}

func (m *mockAdminUseCase) UpdateAppointmentStatus(ctx context.Context, adminID, id uuid.UUID, status string) error {
	return m.err
}

func (m *mockAdminUseCase) ExportAppointmentsCSV(ctx context.Context, adminID uuid.UUID, status string) ([]byte, error) {
	return []byte("ID,Status\n"), m.err
}

func (m *mockAdminUseCase) ListReviews(ctx context.Context, adminID uuid.UUID, rating *float64, isHidden *bool, limit, offset int) ([]entity.Review, int64, error) {
	return []entity.Review{}, 0, m.err
}

func (m *mockAdminUseCase) CreateReview(ctx context.Context, adminID uuid.UUID, review *entity.Review) (*entity.Review, error) {
	return review, m.err
}

func (m *mockAdminUseCase) ToggleReviewVisibility(ctx context.Context, adminID, id uuid.UUID, isHidden bool) error {
	return m.err
}

func (m *mockAdminUseCase) DeleteReview(ctx context.Context, adminID, id uuid.UUID) error {
	return m.err
}

func (m *mockAdminUseCase) ListPromoCodes(ctx context.Context, adminID uuid.UUID) ([]entity.PromoCode, error) {
	return []entity.PromoCode{}, m.err
}

func (m *mockAdminUseCase) CreatePromoCode(ctx context.Context, adminID uuid.UUID, promo *entity.PromoCode) (*entity.PromoCode, error) {
	return promo, m.err
}

func (m *mockAdminUseCase) UpdatePromoCode(ctx context.Context, adminID uuid.UUID, promo *entity.PromoCode) (*entity.PromoCode, error) {
	return promo, m.err
}

func (m *mockAdminUseCase) DeletePromoCode(ctx context.Context, adminID, id uuid.UUID) error {
	return m.err
}

func (m *mockAdminUseCase) ExecuteDisputeRefund(ctx context.Context, adminID, disputeID uuid.UUID, amount float64, reason string) error {
	return m.err
}

func (m *mockAdminUseCase) ExportVerificationsCSV(ctx context.Context, adminID uuid.UUID, status string) ([]byte, error) {
	return []byte("ID,Status\n"), m.err
}

func (m *mockAdminUseCase) ExportBackgroundChecksCSV(ctx context.Context, adminID uuid.UUID, status string) ([]byte, error) {
	return []byte("ID,Status\n"), m.err
}

func (m *mockAdminUseCase) ListLegalDocuments(ctx context.Context, adminID uuid.UUID, docType string, isActive *bool) ([]entity.LegalDocument, error) {
	return []entity.LegalDocument{}, m.err
}

func (m *mockAdminUseCase) GetLegalDocumentByID(ctx context.Context, adminID, id uuid.UUID) (*entity.LegalDocument, error) {
	return &entity.LegalDocument{ID: id}, m.err
}

func (m *mockAdminUseCase) CreateLegalDocument(ctx context.Context, adminID uuid.UUID, doc *entity.LegalDocument) (*entity.LegalDocument, error) {
	return doc, m.err
}

func (m *mockAdminUseCase) UpdateLegalDocument(ctx context.Context, adminID uuid.UUID, doc *entity.LegalDocument) (*entity.LegalDocument, error) {
	return doc, m.err
}

func (m *mockAdminUseCase) DeleteLegalDocument(ctx context.Context, adminID, id uuid.UUID) error {
	return m.err
}

func (m *mockAdminUseCase) ListContactMessages(ctx context.Context, adminID uuid.UUID, isResolved *bool, search string, limit, offset int) ([]entity.ContactMessage, int64, error) {
	return []entity.ContactMessage{}, 0, m.err
}

func (m *mockAdminUseCase) ToggleContactMessageResolved(ctx context.Context, adminID, id uuid.UUID, isResolved bool) error {
	return m.err
}

func (m *mockAdminUseCase) DeleteContactMessage(ctx context.Context, adminID, id uuid.UUID) error {
	return m.err
}

func (m *mockAdminUseCase) ListResolutionReports(ctx context.Context, adminID uuid.UUID, isReviewed *bool, search string, limit, offset int) ([]entity.ResolutionReport, int64, error) {
	return []entity.ResolutionReport{}, 0, m.err
}

func (m *mockAdminUseCase) ToggleResolutionReportReviewed(ctx context.Context, adminID, id uuid.UUID, isReviewed bool) error {
	return m.err
}

func (m *mockAdminUseCase) DeleteResolutionReport(ctx context.Context, adminID, id uuid.UUID) error {
	return m.err
}

func (m *mockAdminUseCase) ListFAQs(ctx context.Context, adminID uuid.UUID, category string, isActive *bool) ([]entity.FAQ, error) {
	return []entity.FAQ{}, m.err
}

func (m *mockAdminUseCase) CreateFAQ(ctx context.Context, adminID uuid.UUID, faq *entity.FAQ) (*entity.FAQ, error) {
	return faq, m.err
}

func (m *mockAdminUseCase) UpdateFAQ(ctx context.Context, adminID uuid.UUID, faq *entity.FAQ) (*entity.FAQ, error) {
	return faq, m.err
}

func (m *mockAdminUseCase) DeleteFAQ(ctx context.Context, adminID, id uuid.UUID) error {
	return m.err
}

func (m *mockAdminUseCase) ListTestimonials(ctx context.Context, adminID uuid.UUID, isActive *bool) ([]entity.Testimonial, error) {
	return []entity.Testimonial{}, m.err
}

func (m *mockAdminUseCase) CreateTestimonial(ctx context.Context, adminID uuid.UUID, t *entity.Testimonial) (*entity.Testimonial, error) {
	return t, m.err
}

func (m *mockAdminUseCase) UpdateTestimonial(ctx context.Context, adminID uuid.UUID, t *entity.Testimonial) (*entity.Testimonial, error) {
	return t, m.err
}

func (m *mockAdminUseCase) DeleteTestimonial(ctx context.Context, adminID, id uuid.UUID) error {
	return m.err
}

func (m *mockAdminUseCase) GetHeroSection(ctx context.Context, adminID uuid.UUID) (*entity.HeroSection, error) {
	return &entity.HeroSection{}, m.err
}

func (m *mockAdminUseCase) UpdateHeroSection(ctx context.Context, adminID uuid.UUID, hero *entity.HeroSection) (*entity.HeroSection, error) {
	return hero, m.err
}

func (m *mockAdminUseCase) ListSiteStats(ctx context.Context, adminID uuid.UUID) ([]entity.SiteStat, error) {
	return []entity.SiteStat{}, m.err
}

func (m *mockAdminUseCase) CreateSiteStat(ctx context.Context, adminID uuid.UUID, stat *entity.SiteStat) (*entity.SiteStat, error) {
	return stat, m.err
}

func (m *mockAdminUseCase) UpdateSiteStat(ctx context.Context, adminID uuid.UUID, stat *entity.SiteStat) (*entity.SiteStat, error) {
	return stat, m.err
}

func (m *mockAdminUseCase) DeleteSiteStat(ctx context.Context, adminID, id uuid.UUID) error {
	return m.err
}

func (m *mockAdminUseCase) GetAboutContent(ctx context.Context, adminID uuid.UUID) (*entity.AboutContent, error) {
	return &entity.AboutContent{}, m.err
}

func (m *mockAdminUseCase) UpdateAboutContent(ctx context.Context, adminID uuid.UUID, about *entity.AboutContent) (*entity.AboutContent, error) {
	return about, m.err
}

func (m *mockAdminUseCase) GetBackupSnapshot(ctx context.Context, adminID uuid.UUID, idOrFilename string) (*entity.BackupSnapshot, error) {
	return &entity.BackupSnapshot{ID: uuid.New(), Filename: idOrFilename}, m.err
}

func setupAdminTestRouter(uc domainUsecase.AdminUseCase) (*gin.Engine, *handler.AdminHandler) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	adminH := handler.NewAdminHandler(uc)

	// Inject test admin user context
	adminID := uuid.New().String()
	r.Use(func(c *gin.Context) {
		c.Set("userID", adminID)
		c.Set("isStaff", true)
		c.Set("isSuperuser", true)
		c.Next()
	})

	api := r.Group("/api/v1/admin")
	{
		api.GET("/dashboard/stats", adminH.GetDashboardStats)
		api.GET("/users", adminH.ListUsers)
		api.GET("/users/:id", adminH.GetUserByID)
		api.PATCH("/users/:id", adminH.UpdateUser)
		api.DELETE("/users/:id", adminH.DeleteUser)
		api.POST("/users/:id/restore", adminH.RestoreUser)
		api.POST("/wallets/:id/adjust", adminH.AdjustWallet)
		api.GET("/verifications", adminH.ListVerifications)
		api.POST("/verifications/:id/approve", adminH.ApproveVerification)
		api.POST("/verifications/:id/reject", adminH.RejectVerification)
		api.POST("/verifications/batch", adminH.BatchVerifications)
		api.GET("/background-checks", adminH.ListBackgroundChecks)
		api.POST("/background-checks/:id/override", adminH.OverrideBackgroundCheck)
		api.GET("/reports", adminH.ListReports)
		api.POST("/reports/:id/resolve", adminH.ResolveReport)
		api.GET("/disputes", adminH.ListDisputes)
		api.POST("/disputes/:id/resolve", adminH.ResolveDispute)
		api.POST("/disputes/:id/reject", adminH.RejectDispute)
		api.GET("/payouts", adminH.ListPayoutRequests)
		api.POST("/payouts/:id/approve", adminH.ApprovePayout)
		api.POST("/payouts/:id/reject", adminH.RejectPayout)

		// 2FA, Impersonation, Roles, Fraud, Templates, Backups, Docs
		api.POST("/2fa/setup", adminH.Setup2FA)
		api.POST("/2fa/verify", adminH.Verify2FA)
		api.POST("/2fa/disable", adminH.Disable2FA)
		api.POST("/users/:id/impersonate", adminH.ImpersonateUser)
		api.GET("/roles", adminH.ListRoles)
		api.POST("/roles", adminH.CreateRole)
		api.POST("/users/:id/role", adminH.AssignUserRole)
		api.GET("/fraud/risk-alerts", adminH.ListFraudRiskAlerts)
		api.POST("/fraud/evaluate/:id", adminH.EvaluateUserRisk)
		api.POST("/fraud/risk-alerts/:id/resolve", adminH.ResolveRiskAlert)
		api.GET("/templates/emails", adminH.ListNotificationTemplates)
		api.PUT("/templates/emails/:key", adminH.UpdateNotificationTemplate)
		api.POST("/templates/emails/test-send", adminH.TestSendEmailTemplate)
		api.POST("/system/backup", adminH.TriggerBackup)
		api.GET("/system/backups", adminH.ListBackups)
		api.GET("/openapi.json", adminH.GetOpenAPISpec)
		api.GET("/docs", adminH.GetSwaggerUI)
		api.GET("/wallets", adminH.ListWallets)
		api.GET("/subscriptions", adminH.ListSubscriptions)
		api.POST("/subscriptions/:id/toggle", adminH.ToggleSubscription)
		api.POST("/categories", adminH.CreateCategory)
		api.PUT("/categories/:id", adminH.UpdateCategory)
		api.DELETE("/categories/:id", adminH.DeleteCategory)
		api.POST("/catalog-services", adminH.CreateCatalogService)
		api.PUT("/catalog-services/:id", adminH.UpdateCatalogService)
		api.DELETE("/catalog-services/:id", adminH.DeleteCatalogService)
		api.GET("/settings", adminH.GetSettings)
		api.PUT("/settings", adminH.UpdateSettings)
		api.GET("/audit-logs", adminH.ListAuditLogs)
		api.POST("/notifications/broadcast", adminH.BroadcastNotification)
		api.GET("/feature-flags", adminH.ListFeatureFlags)
		api.PUT("/feature-flags/:key", adminH.SetFeatureFlag)
		api.POST("/cache/clear", adminH.ClearCache)
		api.GET("/system/health", adminH.GetSystemHealth)
		api.GET("/reports/financial", adminH.GetFinancialReport)
		api.GET("/users/:id/gdpr-export", adminH.GetGDPRUserData)
		api.GET("/export/users", adminH.ExportUsers)
		api.GET("/export/payouts", adminH.ExportPayouts)
		api.GET("/export/disputes", adminH.ExportDisputes)
	}

	return r, adminH
}

func TestAdminHandler_DashboardStats(t *testing.T) {
	mockUC := &mockAdminUseCase{
		stats: &entity.AdminDashboardStats{
			TotalUsers:          100,
			TotalSeekers:        60,
			TotalProviders:      40,
			ActiveSubscriptions: 25,
			TotalWalletBalance:  5400.50,
			PendingPayoutsCount: 3,
		},
	}
	r, _ := setupAdminTestRouter(mockUC)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/dashboard/stats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(100), resp["total_users"])
	assert.Equal(t, float64(60), resp["total_seekers"])
}

func TestAdminHandler_UserOperations(t *testing.T) {
	userID := uuid.New()
	mockUC := &mockAdminUseCase{
		users: []entity.User{{ID: userID, Email: "test@example.com", IsActive: true}},
		user:  &entity.User{ID: userID, Email: "test@example.com", IsActive: true},
	}
	r, _ := setupAdminTestRouter(mockUC)

	// List
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/users?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// GetByID
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/admin/users/"+userID.String(), nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Update
	activeFalse := false
	body, _ := json.Marshal(map[string]interface{}{
		"is_active": false,
	})
	req, _ = http.NewRequest(http.MethodPatch, "/api/v1/admin/users/"+userID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, &activeFalse, mockUC.updateInput.IsActive)
}

func TestAdminHandler_DisputesAndPayouts(t *testing.T) {
	mockUC := &mockAdminUseCase{}
	r, _ := setupAdminTestRouter(mockUC)

	disputeID := uuid.New()
	body, _ := json.Marshal(map[string]string{"notes": "Resolved in favor of seeker"})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/disputes/"+disputeID.String()+"/resolve", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	payoutID := uuid.New()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/admin/payouts/"+payoutID.String()+"/approve", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_CategoriesAndServices(t *testing.T) {
	mockUC := &mockAdminUseCase{
		category: &entity.Category{ID: uuid.New(), Name: "Plumbing"},
		service:  &entity.CatalogService{ID: uuid.New(), Name: "Pipe Repair"},
	}
	r, _ := setupAdminTestRouter(mockUC)

	// Create Category
	catBody, _ := json.Marshal(map[string]string{"name": "Plumbing", "description": "Plumbing services"})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/categories", bytes.NewBuffer(catBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Create Catalog Service
	srvBody, _ := json.Marshal(map[string]interface{}{
		"category_id": uuid.New().String(),
		"name":        "Pipe Repair",
		"base_price":  50.0,
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/admin/catalog-services", bytes.NewBuffer(srvBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAdminHandler_AdvancedEndpoints(t *testing.T) {
	mockUC := &mockAdminUseCase{}
	r, _ := setupAdminTestRouter(mockUC)

	// 1. Restore User
	userID := uuid.New()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/users/"+userID.String()+"/restore", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 2. Adjust Wallet
	walletID := uuid.New()
	adjBody, _ := json.Marshal(map[string]interface{}{
		"amount": 50.0,
		"reason": "Promotion",
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/admin/wallets/"+walletID.String()+"/adjust", bytes.NewBuffer(adjBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Batch Verifications
	batchBody, _ := json.Marshal(map[string]interface{}{
		"ids":    []string{uuid.New().String(), uuid.New().String()},
		"action": "APPROVE",
		"notes":  "Batch ok",
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/admin/verifications/batch", bytes.NewBuffer(batchBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4. Feature Flags
	flagBody, _ := json.Marshal(map[string]interface{}{"is_enabled": true})
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/admin/feature-flags/instant_payouts", bytes.NewBuffer(flagBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 5. System Health & Financial Report
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/admin/system/health", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	req, _ = http.NewRequest(http.MethodGet, "/api/v1/admin/reports/financial?start_date=2026-01-01&end_date=2026-12-31", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 6. CSV Exports
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/admin/export/users", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
}

func TestAdminHandler_EnterpriseAdminFeatures(t *testing.T) {
	mockUC := &mockAdminUseCase{}
	r, _ := setupAdminTestRouter(mockUC)

	// 1. TOTP 2FA Setup
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/2fa/setup", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 2. TOTP 2FA Verify
	verifyBody, _ := json.Marshal(map[string]string{"code": "123456"})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/admin/2fa/verify", bytes.NewBuffer(verifyBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Staff Impersonation
	targetUserID := uuid.New()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/admin/users/"+targetUserID.String()+"/impersonate", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4. RBAC Roles
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/admin/roles", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Fraud Risk Alerts
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/admin/fraud/risk-alerts", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 6. Notification Templates
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/admin/templates/emails", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 7. System Backups
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/admin/system/backup", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 8. OpenAPI & Docs
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/admin/openapi.json", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	req, _ = http.NewRequest(http.MethodGet, "/api/v1/admin/docs", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
