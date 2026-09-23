package usecase

import (
	"context"
	"time"

	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"

	"github.com/google/uuid"
)

type AdminUpdateUserInput struct {
	Email                *string  `json:"email,omitempty"`
	FirstName            *string  `json:"first_name,omitempty"`
	LastName             *string  `json:"last_name,omitempty"`
	Phone                *string  `json:"phone,omitempty"`
	Password             *string  `json:"password,omitempty"`
	IsActive             *bool    `json:"is_active,omitempty"`
	IsStaff              *bool    `json:"is_staff,omitempty"`
	IsVerified           *bool    `json:"is_verified,omitempty"`
	UserType             *string  `json:"user_type,omitempty"`
	SubscriptionTier     *string  `json:"subscription_tier,omitempty"`
	SubscriptionInterval *string  `json:"subscription_interval,omitempty"`
	PreferredPaymentMode *string  `json:"preferred_payment_mode,omitempty"`
	IsIdentityVerified   *bool    `json:"is_identity_verified,omitempty"`
	Bio                  *string  `json:"bio,omitempty"`
	City                 *string  `json:"city,omitempty"`
	State                *string  `json:"state,omitempty"`
	ZipCode              *string  `json:"zip_code,omitempty"`
	Address              *string  `json:"address,omitempty"`
	Country              *string  `json:"country,omitempty"`
	Gender               *string  `json:"gender,omitempty"`
	Service              *string  `json:"service,omitempty"`
	WalletBalanceDelta   *float64 `json:"wallet_balance_delta,omitempty"`
	RoleID               *string  `json:"role_id,omitempty"`
}

type AdminCreateUserInput struct {
	Email                string  `json:"email" binding:"required"`
	Password             string  `json:"password" binding:"required"`
	FirstName            string  `json:"first_name"`
	LastName             string  `json:"last_name"`
	Phone                string  `json:"phone"`
	UserType             string  `json:"user_type"` // seeker or provider
	IsActive             bool    `json:"is_active"`
	IsStaff              bool    `json:"is_staff"`
	IsVerified           bool    `json:"is_verified"`
	IsIdentityVerified   bool    `json:"is_identity_verified"`
	SubscriptionTier     string  `json:"subscription_tier"`
	SubscriptionInterval string  `json:"subscription_interval"`
	PreferredPaymentMode string  `json:"preferred_payment_mode"`
	Bio                  string  `json:"bio"`
	City                 string  `json:"city"`
	State                string  `json:"state"`
	ZipCode              string  `json:"zip_code"`
	Address              string  `json:"address"`
	Country              string  `json:"country"`
	Gender               string  `json:"gender"`
	Service              string  `json:"service"`
	RoleID               *string `json:"role_id,omitempty"`
}

type AdminUseCase interface {
	GetDashboardStats(ctx context.Context, adminID uuid.UUID) (*entity.AdminDashboardStats, error)

	// Users
	ListUsers(ctx context.Context, adminID uuid.UUID, filter repository.AdminUserFilter) ([]entity.User, int64, error)
	GetUserByID(ctx context.Context, adminID, userID uuid.UUID) (*entity.User, error)
	CreateUser(ctx context.Context, adminID uuid.UUID, input AdminCreateUserInput) (*entity.User, error)
	UpdateUser(ctx context.Context, adminID, userID uuid.UUID, input AdminUpdateUserInput) (*entity.User, error)
	DeleteUser(ctx context.Context, adminID, userID uuid.UUID) error

	// Verifications & Moderation
	ListVerifications(ctx context.Context, adminID uuid.UUID, status string, limit, offset int) ([]entity.ProviderVerification, int64, error)
	ApproveVerification(ctx context.Context, adminID, verificationID uuid.UUID) error
	RejectVerification(ctx context.Context, adminID, verificationID uuid.UUID) error

	// Background Checks
	ListBackgroundChecks(ctx context.Context, adminID uuid.UUID, status string, limit, offset int) ([]entity.BackgroundCheck, int64, error)
	OverrideBackgroundCheck(ctx context.Context, adminID, checkID uuid.UUID, status, notes string) error

	// Reports
	ListReports(ctx context.Context, adminID uuid.UUID, status string, limit, offset int) ([]entity.Report, int64, error)
	ResolveReport(ctx context.Context, adminID, reportID uuid.UUID, resolution string) error

	// Disputes
	ListDisputes(ctx context.Context, adminID uuid.UUID, status string, limit, offset int) ([]entity.Dispute, int64, error)
	ResolveDispute(ctx context.Context, adminID, disputeID uuid.UUID, notes string) error
	RejectDispute(ctx context.Context, adminID, disputeID uuid.UUID, notes string) error

	// Payouts & Wallets
	ListPayoutRequests(ctx context.Context, adminID uuid.UUID, status string, limit, offset int) ([]entity.PayoutRequest, int64, error)
	ApprovePayout(ctx context.Context, adminID, payoutID uuid.UUID) error
	RejectPayout(ctx context.Context, adminID, payoutID uuid.UUID, notes string) error
	ListWallets(ctx context.Context, adminID uuid.UUID, limit, offset int) ([]entity.Wallet, int64, error)

	// Subscriptions & Plans
	ListSubscriptions(ctx context.Context, adminID uuid.UUID, isActive *bool, limit, offset int) ([]entity.Subscription, int64, error)
	ToggleSubscription(ctx context.Context, adminID, subID uuid.UUID, isActive bool) error
	AssignSubscription(ctx context.Context, adminID uuid.UUID, input AssignSubscriptionInput) (*entity.Subscription, error)
	ListSubscriptionPlans(ctx context.Context, adminID uuid.UUID) ([]entity.SubscriptionPlan, error)
	CreateSubscriptionPlan(ctx context.Context, adminID uuid.UUID, plan *entity.SubscriptionPlan) (*entity.SubscriptionPlan, error)
	UpdateSubscriptionPlan(ctx context.Context, adminID uuid.UUID, plan *entity.SubscriptionPlan) (*entity.SubscriptionPlan, error)
	DeleteSubscriptionPlan(ctx context.Context, adminID, planID uuid.UUID) error

	// Categories & Services
	CreateCategory(ctx context.Context, adminID uuid.UUID, cat *entity.Category) (*entity.Category, error)
	UpdateCategory(ctx context.Context, adminID, catID uuid.UUID, name, description, icon string) (*entity.Category, error)
	DeleteCategory(ctx context.Context, adminID, catID uuid.UUID) error

	CreateCatalogService(ctx context.Context, adminID uuid.UUID, cs *entity.CatalogService) (*entity.CatalogService, error)
	UpdateCatalogService(ctx context.Context, adminID, csID uuid.UUID, name, description string, basePrice float64, categoryID *uuid.UUID) (*entity.CatalogService, error)
	DeleteCatalogService(ctx context.Context, adminID, csID uuid.UUID) error

	// Settings & Audit
	GetSettings(ctx context.Context, adminID uuid.UUID) (*entity.ModerationSetting, error)
	UpdateSettings(ctx context.Context, adminID uuid.UUID, paymentMode string, fee float64, broadcastRadiusKm, matchRadiusKm float64) (*entity.ModerationSetting, error)
	ListAuditLogs(ctx context.Context, adminID uuid.UUID, filter repository.AdminAuditFilter) ([]entity.AuditLog, int64, error)

	// User Restoration & Wallet Management
	RestoreUser(ctx context.Context, adminID, userID uuid.UUID) error
	AdjustWalletBalance(ctx context.Context, adminID, walletID uuid.UUID, amount float64, reason string) (*entity.Wallet, error)

	// Batch Moderation & Broadcast Notifications
	BatchVerifications(ctx context.Context, adminID uuid.UUID, ids []uuid.UUID, action, notes string) error
	BroadcastNotification(ctx context.Context, adminID uuid.UUID, targetUserType, title, message string) (int, error)

	// Feature Flags & Operations
	ListFeatureFlags(ctx context.Context, adminID uuid.UUID) ([]entity.FeatureFlag, error)
	SetFeatureFlag(ctx context.Context, adminID uuid.UUID, key string, isEnabled bool) (*entity.FeatureFlag, error)

	// Reports, Health & GDPR
	GetFinancialReport(ctx context.Context, adminID uuid.UUID, startDate, endDate string) (*entity.FinancialReportSummary, error)
	GetSystemHealth(ctx context.Context, adminID uuid.UUID) (*entity.SystemHealthStatus, error)
	GetGDPRUserData(ctx context.Context, adminID, userID uuid.UUID) (*entity.GDPRUserData, error)

	// Data Exports (CSV)
	ExportUsersCSV(ctx context.Context, adminID uuid.UUID, filter repository.AdminUserFilter) ([]byte, error)
	ExportPayoutsCSV(ctx context.Context, adminID uuid.UUID, status string) ([]byte, error)
	ExportDisputesCSV(ctx context.Context, adminID uuid.UUID, status string) ([]byte, error)

	// TOTP 2FA Engine
	Setup2FA(ctx context.Context, adminID uuid.UUID) (*entity.TOTPSetupResponse, error)
	VerifyAndEnable2FA(ctx context.Context, adminID uuid.UUID, code string) error
	Disable2FA(ctx context.Context, adminID uuid.UUID, code string) error

	// Staff Impersonation
	ImpersonateUser(ctx context.Context, adminID, targetUserID uuid.UUID) (*entity.ImpersonationResult, error)

	// RBAC Roles
	ListRoles(ctx context.Context, adminID uuid.UUID) ([]entity.AdminRole, error)
	CreateRole(ctx context.Context, adminID uuid.UUID, role *entity.AdminRole) (*entity.AdminRole, error)
	UpdateRole(ctx context.Context, adminID, roleID uuid.UUID, role *entity.AdminRole) (*entity.AdminRole, error)
	DeleteRole(ctx context.Context, adminID, roleID uuid.UUID) error
	AssignUserRole(ctx context.Context, adminID, targetUserID, roleID uuid.UUID) error

	// Fraud & Risk Detection
	ListFraudRiskAlerts(ctx context.Context, adminID uuid.UUID, status string, limit, offset int) ([]entity.FraudRiskAlert, int64, error)
	EvaluateUserRisk(ctx context.Context, adminID, targetUserID uuid.UUID) (*entity.FraudRiskAlert, error)
	ResolveRiskAlert(ctx context.Context, adminID, alertID uuid.UUID, action string) error

	// Notification Templates
	ListNotificationTemplates(ctx context.Context, adminID uuid.UUID) ([]entity.NotificationTemplate, error)
	UpdateNotificationTemplate(ctx context.Context, adminID uuid.UUID, key, subject, bodyHTML, bodyText string) (*entity.NotificationTemplate, error)
	TestSendEmailTemplate(ctx context.Context, adminID uuid.UUID, key, targetEmail string) error

	// Database Backup Snapshots
	TriggerDatabaseBackup(ctx context.Context, adminID uuid.UUID) (*entity.BackupSnapshot, error)
	ListBackupSnapshots(ctx context.Context, adminID uuid.UUID) ([]entity.BackupSnapshot, error)
	GetBackupSnapshot(ctx context.Context, adminID uuid.UUID, idOrFilename string) (*entity.BackupSnapshot, error)

	// Batch Payouts
	BatchApprovePayouts(ctx context.Context, adminID uuid.UUID, ids []uuid.UUID) (*entity.BatchPayoutResponse, error)
	BatchRejectPayouts(ctx context.Context, adminID uuid.UUID, ids []uuid.UUID, reason string) (*entity.BatchPayoutResponse, error)

	// Live Notification Feed & Telemetry
	GetAdminNotificationFeed(ctx context.Context, adminID uuid.UUID) ([]entity.AdminNotification, error)
	GetProviderOnboardingFunnel(ctx context.Context, adminID uuid.UUID) (*entity.ProviderOnboardingFunnel, error)

	// Staff Internal Notes
	ListStaffNotes(ctx context.Context, adminID, userID uuid.UUID) ([]entity.StaffNote, error)
	CreateStaffNote(ctx context.Context, adminID, userID uuid.UUID, content, category string, isPinned bool) (*entity.StaffNote, error)
	DeleteStaffNote(ctx context.Context, adminID, noteID uuid.UUID) error

	// Bookings & Appointments Console
	ListAppointments(ctx context.Context, adminID uuid.UUID, status, search string, limit, offset int) ([]entity.Appointment, int64, error)
	GetAppointmentByID(ctx context.Context, adminID, id uuid.UUID) (*entity.Appointment, error)
	UpdateAppointmentStatus(ctx context.Context, adminID, id uuid.UUID, status string) error
	ExportAppointmentsCSV(ctx context.Context, adminID uuid.UUID, status string) ([]byte, error)

	// Reviews & Ratings Moderation
	ListReviews(ctx context.Context, adminID uuid.UUID, rating *float64, isHidden *bool, limit, offset int) ([]entity.Review, int64, error)
	CreateReview(ctx context.Context, adminID uuid.UUID, review *entity.Review) (*entity.Review, error)
	ToggleReviewVisibility(ctx context.Context, adminID, id uuid.UUID, isHidden bool) error
	DeleteReview(ctx context.Context, adminID, id uuid.UUID) error

	// Promo Codes
	ListPromoCodes(ctx context.Context, adminID uuid.UUID) ([]entity.PromoCode, error)
	CreatePromoCode(ctx context.Context, adminID uuid.UUID, promo *entity.PromoCode) (*entity.PromoCode, error)
	UpdatePromoCode(ctx context.Context, adminID uuid.UUID, promo *entity.PromoCode) (*entity.PromoCode, error)
	DeletePromoCode(ctx context.Context, adminID, id uuid.UUID) error

	// Dispute Refund Execution
	ExecuteDisputeRefund(ctx context.Context, adminID, disputeID uuid.UUID, amount float64, reason string) error

	// Additional Data Exports
	ExportVerificationsCSV(ctx context.Context, adminID uuid.UUID, status string) ([]byte, error)
	ExportBackgroundChecksCSV(ctx context.Context, adminID uuid.UUID, status string) ([]byte, error)

	// Legal & Compliance Documents
	ListLegalDocuments(ctx context.Context, adminID uuid.UUID, docType string, isActive *bool) ([]entity.LegalDocument, error)
	GetLegalDocumentByID(ctx context.Context, adminID, id uuid.UUID) (*entity.LegalDocument, error)
	CreateLegalDocument(ctx context.Context, adminID uuid.UUID, doc *entity.LegalDocument) (*entity.LegalDocument, error)
	UpdateLegalDocument(ctx context.Context, adminID uuid.UUID, doc *entity.LegalDocument) (*entity.LegalDocument, error)
	DeleteLegalDocument(ctx context.Context, adminID, id uuid.UUID) error

	// Support Messages & Inquiries
	ListContactMessages(ctx context.Context, adminID uuid.UUID, isResolved *bool, search string, limit, offset int) ([]entity.ContactMessage, int64, error)
	ToggleContactMessageResolved(ctx context.Context, adminID, id uuid.UUID, isResolved bool) error
	DeleteContactMessage(ctx context.Context, adminID, id uuid.UUID) error

	// Resolution Incident Reports
	ListResolutionReports(ctx context.Context, adminID uuid.UUID, isReviewed *bool, search string, limit, offset int) ([]entity.ResolutionReport, int64, error)
	ToggleResolutionReportReviewed(ctx context.Context, adminID, id uuid.UUID, isReviewed bool) error
	DeleteResolutionReport(ctx context.Context, adminID, id uuid.UUID) error

	// Public Site CMS (FAQs, Testimonials, Hero, Stats, About)
	ListFAQs(ctx context.Context, adminID uuid.UUID, category string, isActive *bool) ([]entity.FAQ, error)
	CreateFAQ(ctx context.Context, adminID uuid.UUID, faq *entity.FAQ) (*entity.FAQ, error)
	UpdateFAQ(ctx context.Context, adminID uuid.UUID, faq *entity.FAQ) (*entity.FAQ, error)
	DeleteFAQ(ctx context.Context, adminID, id uuid.UUID) error

	ListTestimonials(ctx context.Context, adminID uuid.UUID, isActive *bool) ([]entity.Testimonial, error)
	CreateTestimonial(ctx context.Context, adminID uuid.UUID, t *entity.Testimonial) (*entity.Testimonial, error)
	UpdateTestimonial(ctx context.Context, adminID uuid.UUID, t *entity.Testimonial) (*entity.Testimonial, error)
	DeleteTestimonial(ctx context.Context, adminID, id uuid.UUID) error

	GetHeroSection(ctx context.Context, adminID uuid.UUID) (*entity.HeroSection, error)
	UpdateHeroSection(ctx context.Context, adminID uuid.UUID, hero *entity.HeroSection) (*entity.HeroSection, error)

	ListSiteStats(ctx context.Context, adminID uuid.UUID) ([]entity.SiteStat, error)
	CreateSiteStat(ctx context.Context, adminID uuid.UUID, stat *entity.SiteStat) (*entity.SiteStat, error)
	UpdateSiteStat(ctx context.Context, adminID uuid.UUID, stat *entity.SiteStat) (*entity.SiteStat, error)
	DeleteSiteStat(ctx context.Context, adminID, id uuid.UUID) error

	GetAboutContent(ctx context.Context, adminID uuid.UUID) (*entity.AboutContent, error)
	UpdateAboutContent(ctx context.Context, adminID uuid.UUID, about *entity.AboutContent) (*entity.AboutContent, error)
}

type AssignSubscriptionInput struct {
	UserID             uuid.UUID  `json:"user_id"`
	PlanID             *uuid.UUID `json:"plan_id"`
	Tier               string     `json:"tier"`
	Interval           string     `json:"interval"`
	IsActive           bool       `json:"is_active"`
	NextPayment        *time.Time `json:"next_payment"`
	StoreTransactionID string     `json:"store_transaction_id"`
	Notes              string     `json:"notes"`
}
