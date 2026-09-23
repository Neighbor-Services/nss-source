package repository

import (
	"context"

	"backend-go/internal/domain/entity"
	"github.com/google/uuid"
)

type AdminUserFilter struct {
	Search             string
	UserType           string
	IsStaff            *bool
	IsActive           *bool
	IsVerified         *bool
	IsIdentityVerified *bool
	SubscriptionTier   string
	Limit              int
	Offset             int
}

type AdminAuditFilter struct {
	UserID       *uuid.UUID
	Action       string
	ResourceType string
	Limit        int
	Offset       int
}

type AdminRepository interface {
	GetDashboardStats(ctx context.Context) (*entity.AdminDashboardStats, error)
	ListUsers(ctx context.Context, filter AdminUserFilter) ([]entity.User, int64, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error

	ListVerifications(ctx context.Context, status string, limit, offset int) ([]entity.ProviderVerification, int64, error)
	GetVerificationByID(ctx context.Context, id uuid.UUID) (*entity.ProviderVerification, error)
	UpdateVerification(ctx context.Context, v *entity.ProviderVerification) error

	ListBackgroundChecks(ctx context.Context, status string, limit, offset int) ([]entity.BackgroundCheck, int64, error)
	GetBackgroundCheckByID(ctx context.Context, id uuid.UUID) (*entity.BackgroundCheck, error)
	UpdateBackgroundCheck(ctx context.Context, bc *entity.BackgroundCheck) error

	ListReports(ctx context.Context, status string, limit, offset int) ([]entity.Report, int64, error)
	GetReportByID(ctx context.Context, id uuid.UUID) (*entity.Report, error)
	UpdateReport(ctx context.Context, r *entity.Report) error

	ListDisputes(ctx context.Context, status string, limit, offset int) ([]entity.Dispute, int64, error)
	GetDisputeByID(ctx context.Context, id uuid.UUID) (*entity.Dispute, error)
	UpdateDispute(ctx context.Context, dispute *entity.Dispute) error

	ListPayoutRequests(ctx context.Context, status string, limit, offset int) ([]entity.PayoutRequest, int64, error)
	GetPayoutRequestByID(ctx context.Context, id uuid.UUID) (*entity.PayoutRequest, error)
	UpdatePayoutRequest(ctx context.Context, req *entity.PayoutRequest) error

	ListSubscriptions(ctx context.Context, isActive *bool, limit, offset int) ([]entity.Subscription, int64, error)
	GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error)
	GetSubscriptionByUserID(ctx context.Context, userID uuid.UUID) (*entity.Subscription, error)
	CreateSubscription(ctx context.Context, sub *entity.Subscription) error
	UpdateSubscription(ctx context.Context, sub *entity.Subscription) error

	ListSubscriptionPlans(ctx context.Context) ([]entity.SubscriptionPlan, error)
	GetSubscriptionPlanByID(ctx context.Context, id uuid.UUID) (*entity.SubscriptionPlan, error)
	CreateSubscriptionPlan(ctx context.Context, plan *entity.SubscriptionPlan) error
	UpdateSubscriptionPlan(ctx context.Context, plan *entity.SubscriptionPlan) error
	DeleteSubscriptionPlan(ctx context.Context, id uuid.UUID) error

	ListWallets(ctx context.Context, limit, offset int) ([]entity.Wallet, int64, error)
	GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error)
	UpdateWallet(ctx context.Context, wallet *entity.Wallet) error

	CreateCategory(ctx context.Context, cat *entity.Category) error
	UpdateCategory(ctx context.Context, cat *entity.Category) error
	DeleteCategory(ctx context.Context, id uuid.UUID) error

	CreateCatalogService(ctx context.Context, cs *entity.CatalogService) error
	UpdateCatalogService(ctx context.Context, cs *entity.CatalogService) error
	DeleteCatalogService(ctx context.Context, id uuid.UUID) error

	GetSettings(ctx context.Context) (*entity.ModerationSetting, error)
	UpdateSettings(ctx context.Context, setting *entity.ModerationSetting) error

	CreateAuditLog(ctx context.Context, log *entity.AuditLog) error
	ListAuditLogs(ctx context.Context, filter AdminAuditFilter) ([]entity.AuditLog, int64, error)

	// User Restoration & Wallet Transaction
	RestoreUser(ctx context.Context, id uuid.UUID) error
	GetWalletByID(ctx context.Context, id uuid.UUID) (*entity.Wallet, error)
	CreateWalletTransaction(ctx context.Context, tx *entity.WalletTransaction) error

	// Batch Moderation
	BatchUpdateVerifications(ctx context.Context, ids []uuid.UUID, status, notes string) error

	// Feature Flags & Operations
	ListFeatureFlags(ctx context.Context) ([]entity.FeatureFlag, error)
	GetFeatureFlag(ctx context.Context, key string) (*entity.FeatureFlag, error)
	SetFeatureFlag(ctx context.Context, key string, isEnabled bool) (*entity.FeatureFlag, error)

	// Reports & Analytics
	GetFinancialReport(ctx context.Context, startDate, endDate string) (*entity.FinancialReportSummary, error)
	GetSystemHealth(ctx context.Context) (*entity.SystemHealthStatus, error)
	GetGDPRUserData(ctx context.Context, userID uuid.UUID) (*entity.GDPRUserData, error)

	// Data Exports
	ExportUsers(ctx context.Context, filter AdminUserFilter) ([]entity.User, error)
	ExportPayouts(ctx context.Context, status string) ([]entity.PayoutRequest, error)
	ExportDisputes(ctx context.Context, status string) ([]entity.Dispute, error)

	// RBAC & Roles
	ListRoles(ctx context.Context) ([]entity.AdminRole, error)
	GetRoleByID(ctx context.Context, id uuid.UUID) (*entity.AdminRole, error)
	CreateRole(ctx context.Context, role *entity.AdminRole) error
	UpdateRole(ctx context.Context, role *entity.AdminRole) error
	DeleteRole(ctx context.Context, id uuid.UUID) error

	// Fraud & Risk Detection
	ListFraudRiskAlerts(ctx context.Context, status string, limit, offset int) ([]entity.FraudRiskAlert, int64, error)
	CreateFraudRiskAlert(ctx context.Context, alert *entity.FraudRiskAlert) error
	UpdateFraudRiskAlert(ctx context.Context, alert *entity.FraudRiskAlert) error
	EvaluateUserRisk(ctx context.Context, userID uuid.UUID) (*entity.FraudRiskAlert, error)

	// Notification & Email Templates
	ListNotificationTemplates(ctx context.Context) ([]entity.NotificationTemplate, error)
	GetNotificationTemplate(ctx context.Context, key string) (*entity.NotificationTemplate, error)
	UpdateNotificationTemplate(ctx context.Context, tpl *entity.NotificationTemplate) error

	// Backup Snapshots
	ListBackupSnapshots(ctx context.Context) ([]entity.BackupSnapshot, error)
	GetBackupSnapshotByID(ctx context.Context, id uuid.UUID) (*entity.BackupSnapshot, error)
	GetBackupSnapshotByFilename(ctx context.Context, filename string) (*entity.BackupSnapshot, error)
	CreateBackupSnapshot(ctx context.Context, snap *entity.BackupSnapshot) error

	// Batch Payouts
	BatchApprovePayouts(ctx context.Context, ids []uuid.UUID, adminID uuid.UUID) (*entity.BatchPayoutResponse, error)
	BatchRejectPayouts(ctx context.Context, ids []uuid.UUID, reason string, adminID uuid.UUID) (*entity.BatchPayoutResponse, error)

	// Staff Internal Notes
	ListStaffNotes(ctx context.Context, userID uuid.UUID) ([]entity.StaffNote, error)
	CreateStaffNote(ctx context.Context, note *entity.StaffNote) error
	DeleteStaffNote(ctx context.Context, noteID uuid.UUID) error

	// Live Notification Feed & Telemetry
	GetAdminNotificationFeed(ctx context.Context) ([]entity.AdminNotification, error)
	GetProviderOnboardingFunnel(ctx context.Context) (*entity.ProviderOnboardingFunnel, error)

	// Bookings & Appointments Console
	ListAppointments(ctx context.Context, status string, search string, limit, offset int) ([]entity.Appointment, int64, error)
	GetAppointmentByID(ctx context.Context, id uuid.UUID) (*entity.Appointment, error)
	UpdateAppointmentStatus(ctx context.Context, id uuid.UUID, status string) error
	ExportAppointments(ctx context.Context, status string) ([]entity.Appointment, error)

	// Reviews & Ratings Moderation
	ListReviews(ctx context.Context, rating *float64, isHidden *bool, limit, offset int) ([]entity.Review, int64, error)
	CreateReview(ctx context.Context, review *entity.Review) error
	ToggleReviewVisibility(ctx context.Context, id uuid.UUID, isHidden bool) error
	DeleteReview(ctx context.Context, id uuid.UUID) error

	// Promo Codes & Marketing Campaigns
	ListPromoCodes(ctx context.Context) ([]entity.PromoCode, error)
	CreatePromoCode(ctx context.Context, promo *entity.PromoCode) error
	UpdatePromoCode(ctx context.Context, promo *entity.PromoCode) error
	DeletePromoCode(ctx context.Context, id uuid.UUID) error

	// Dispute Refund Execution
	ExecuteDisputeRefund(ctx context.Context, disputeID uuid.UUID, amount float64, reason string, adminID uuid.UUID) error

	// Additional Data Exports
	ExportVerifications(ctx context.Context, status string) ([]entity.ProviderVerification, error)
	ExportBackgroundChecks(ctx context.Context, status string) ([]entity.BackgroundCheck, error)

	// Legal & Compliance Documents
	ListLegalDocuments(ctx context.Context, docType string, isActive *bool) ([]entity.LegalDocument, error)
	GetLegalDocumentByID(ctx context.Context, id uuid.UUID) (*entity.LegalDocument, error)
	CreateLegalDocument(ctx context.Context, doc *entity.LegalDocument) error
	UpdateLegalDocument(ctx context.Context, doc *entity.LegalDocument) error
	DeleteLegalDocument(ctx context.Context, id uuid.UUID) error

	// Support Messages & Inquiries
	ListContactMessages(ctx context.Context, isResolved *bool, search string, limit, offset int) ([]entity.ContactMessage, int64, error)
	ToggleContactMessageResolved(ctx context.Context, id uuid.UUID, isResolved bool) error
	DeleteContactMessage(ctx context.Context, id uuid.UUID) error

	// Resolution Reports
	ListResolutionReports(ctx context.Context, isReviewed *bool, search string, limit, offset int) ([]entity.ResolutionReport, int64, error)
	ToggleResolutionReportReviewed(ctx context.Context, id uuid.UUID, isReviewed bool) error
	DeleteResolutionReport(ctx context.Context, id uuid.UUID) error

	// Public Site CMS
	ListFAQs(ctx context.Context, category string, isActive *bool) ([]entity.FAQ, error)
	CreateFAQ(ctx context.Context, faq *entity.FAQ) error
	UpdateFAQ(ctx context.Context, faq *entity.FAQ) error
	DeleteFAQ(ctx context.Context, id uuid.UUID) error

	ListTestimonials(ctx context.Context, isActive *bool) ([]entity.Testimonial, error)
	CreateTestimonial(ctx context.Context, t *entity.Testimonial) error
	UpdateTestimonial(ctx context.Context, t *entity.Testimonial) error
	DeleteTestimonial(ctx context.Context, id uuid.UUID) error

	GetHeroSection(ctx context.Context) (*entity.HeroSection, error)
	UpdateHeroSection(ctx context.Context, hero *entity.HeroSection) error

	ListSiteStats(ctx context.Context) ([]entity.SiteStat, error)
	CreateSiteStat(ctx context.Context, stat *entity.SiteStat) error
	UpdateSiteStat(ctx context.Context, stat *entity.SiteStat) error
	DeleteSiteStat(ctx context.Context, id uuid.UUID) error

	GetAboutContent(ctx context.Context) (*entity.AboutContent, error)
	UpdateAboutContent(ctx context.Context, about *entity.AboutContent) error
}

