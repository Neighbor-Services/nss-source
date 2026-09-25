package http

import (
	"net/http"

	"backend-go/internal/config"
	"backend-go/internal/delivery/http/handler"
	"backend-go/internal/domain/repository"
	"backend-go/internal/middleware"
	"backend-go/pkg/media"
	"github.com/gin-gonic/gin"
)

type RouterDependencies struct {
	Config          *config.Config
	UserRepo        repository.UserRepository
	AdminRepo       repository.AdminRepository
	AuditRepo       repository.AuditLogRepository
	AuthHandler     *handler.AuthHandler
	ProfileHandler  *handler.ProfileHandler
	ServiceHandler  *handler.ServiceHandler
	InterHandler    *handler.InteractionHandler
	ChatHandler     *handler.ChatHandler
	NotifHandler    *handler.NotificationHandler
	PaymentHandler  *handler.PaymentHandler
	ConsultHandler  *handler.ConsultationHandler
	ModerateHandler *handler.ModerationHandler
	AuditHandler    *handler.AuditHandler
	AdminHandler    *handler.AdminHandler
	PublicHandler   *handler.PublicHandler
	DocsHandler     *handler.DocsHandler
}

func SetupRouter(deps RouterDependencies) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.ErrorRecovery())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS(deps.Config))

	// Static Media Serving with fuzzy fallback
	mediaHandler := func(c *gin.Context) {
		reqPath := c.Param("filepath")
		baseDir := media.GetBaseMediaDir()
		if foundPath, ok := media.FindMediaFile(baseDir, reqPath); ok {
			c.File(foundPath)
			return
		}
		c.Status(http.StatusNotFound)
	}
	r.GET("/media/*filepath", mediaHandler)
	r.HEAD("/media/*filepath", mediaHandler)
	r.GET("/api/v1/media/*filepath", mediaHandler)
	r.HEAD("/api/v1/media/*filepath", mediaHandler)

	// Interactive OpenAPI Documentation & Swagger UI for ALL endpoints
	if deps.DocsHandler != nil {
		r.GET("/docs", deps.DocsHandler.GetSwaggerUI)
		r.GET("/docs/", deps.DocsHandler.GetSwaggerUI)
		r.GET("/swagger", deps.DocsHandler.GetSwaggerUI)
		r.GET("/swagger/", deps.DocsHandler.GetSwaggerUI)
		r.GET("/swagger/index.html", deps.DocsHandler.GetSwaggerUI)
		r.GET("/openapi.json", deps.DocsHandler.GetOpenAPISpec)
		r.GET("/api/docs", deps.DocsHandler.GetSwaggerUI)
		r.GET("/api/docs/", deps.DocsHandler.GetSwaggerUI)
		r.GET("/api/openapi.json", deps.DocsHandler.GetOpenAPISpec)
	}

	// Health & Readiness Probes (Kubernetes / Docker Observability)
	r.GET("/health", deps.PublicHandler.HealthCheck)
	r.GET("/health/", deps.PublicHandler.HealthCheck)
	r.GET("/healthz", deps.PublicHandler.HealthCheck)
	r.GET("/readyz", deps.PublicHandler.ReadyCheck)
	r.GET("/api/health", deps.PublicHandler.HealthCheck)
	r.GET("/api/health/", deps.PublicHandler.HealthCheck)

	// Public Site Forms & CMS (Root Level Django Match)
	r.POST("/contact", deps.PublicHandler.SubmitContactMessage)
	r.POST("/contact/", deps.PublicHandler.SubmitContactMessage)
	r.POST("/resolution", deps.PublicHandler.SubmitResolutionReport)
	r.POST("/resolution/", deps.PublicHandler.SubmitResolutionReport)
	r.GET("/legal/", deps.ProfileHandler.GetLegalDocuments)
	r.GET("/legal", deps.ProfileHandler.GetLegalDocuments)

	// Public Site Static Web Pages (Django HTML template compatibility)
	r.GET("/", deps.PublicHandler.RenderStaticPage("landing"))
	r.GET("/about", deps.PublicHandler.RenderStaticPage("about"))
	r.GET("/about/", deps.PublicHandler.RenderStaticPage("about"))
	r.GET("/services", deps.PublicHandler.RenderStaticPage("services"))
	r.GET("/services/", deps.PublicHandler.RenderStaticPage("services"))
	r.GET("/contact", deps.PublicHandler.RenderStaticPage("contact"))
	r.GET("/contact/", deps.PublicHandler.RenderStaticPage("contact"))
	r.GET("/privacy", deps.PublicHandler.RenderStaticPage("privacy"))
	r.GET("/privacy/", deps.PublicHandler.RenderStaticPage("privacy"))
	r.GET("/terms", deps.PublicHandler.RenderStaticPage("terms"))
	r.GET("/terms/", deps.PublicHandler.RenderStaticPage("terms"))
	r.GET("/support", deps.PublicHandler.RenderStaticPage("support"))
	r.GET("/support/", deps.PublicHandler.RenderStaticPage("support"))
	r.GET("/resolution", deps.PublicHandler.RenderStaticPage("resolution"))
	r.GET("/resolution/", deps.PublicHandler.RenderStaticPage("resolution"))

	// Apple Callbacks & Webhooks (Root level matching Django urls.py)
	r.POST("/callbacks/apple", deps.AuthHandler.AppleCallback)
	r.POST("/payments/webhook/apple-s2s/", deps.PaymentHandler.AppleS2SWebhook)
	r.POST("/payments/webhook/apple-s2s", deps.PaymentHandler.AppleS2SWebhook)

	// API V1 Routes
	v1 := r.Group("/api/v1")
	if deps.AuditRepo != nil {
		v1.Use(middleware.UserAuditLogger(deps.AuditRepo))
	}
	{
		// ─── ROOT OVERRIDE PATHS (DJANGO COMPATIBILITY) ──────────────────────
		v1.POST("/subscription/create/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.ValidateAppleReceipt)

		// ─── ACCOUNTS ────────────────────────────────────────────────────────
		acc := v1.Group("/accounts")
		{
			acc.POST("/register/", middleware.AuthRateLimit(), deps.AuthHandler.Register)
			acc.POST("/verify-otp/", middleware.AuthRateLimit(), deps.AuthHandler.VerifyOTP)
			acc.POST("/resend-otp/", middleware.AuthRateLimit(), deps.AuthHandler.ResendOTP)
			acc.GET("/verify/:uidb64/:token/", deps.AuthHandler.VerifyOTP)
			acc.POST("/login/", middleware.AuthRateLimit(), deps.AuthHandler.Login)
			acc.POST("/token/refresh/", middleware.AuthRateLimit(), deps.AuthHandler.TokenRefresh)
			acc.POST("/token/rotate/", middleware.AuthRateLimit(), deps.AuthHandler.RotateToken)
			acc.POST("/change-password/", middleware.AuthRequired(deps.Config), deps.AuthHandler.ChangePassword)
			acc.PUT("/change-password/", middleware.AuthRequired(deps.Config), deps.AuthHandler.ChangePassword)
			acc.PATCH("/change-password/", middleware.AuthRequired(deps.Config), deps.AuthHandler.ChangePassword)
			acc.POST("/password-reset/", middleware.AuthRateLimit(), deps.AuthHandler.PasswordResetRequest)
			acc.POST("/password-reset-confirm/", middleware.AuthRateLimit(), deps.AuthHandler.PasswordResetConfirm)
			acc.POST("/password-reset-otp-confirm/", middleware.AuthRateLimit(), deps.AuthHandler.PasswordResetConfirm)
			acc.POST("/login-google/", middleware.AuthRateLimit(), deps.AuthHandler.GoogleLogin)
			acc.POST("/login-apple/", middleware.AuthRateLimit(), deps.AuthHandler.AppleLogin)
			acc.POST("/logout/", deps.AuthHandler.Logout)
			acc.GET("/export/", middleware.AuthRequired(deps.Config), deps.AuthHandler.ExportAccountData)
			acc.POST("/delete-account/", middleware.AuthRequired(deps.Config), deps.AuthHandler.DeleteAccount)
			acc.DELETE("/delete-account/", middleware.AuthRequired(deps.Config), deps.AuthHandler.DeleteAccount)
			acc.GET("/legal/", deps.ProfileHandler.GetLegalDocuments)

			// Profile ViewSet
			// GET routes are public (AuthOptional) so guests can browse provider profiles
			acc.GET("/profile/", middleware.AuthOptional(deps.Config), deps.ProfileHandler.GetProfile)
			acc.GET("/profile/me/", middleware.AuthRequired(deps.Config), deps.ProfileHandler.GetProfileMe)
			acc.GET("/profile/popular/", deps.ProfileHandler.GetProfile)
			acc.POST("/profile/", middleware.AuthRequired(deps.Config), deps.ProfileHandler.UpdateProfile)
			acc.PATCH("/profile/update_me/", middleware.AuthRequired(deps.Config), deps.ProfileHandler.UpdateProfile)
			acc.PUT("/profile/update_me/", middleware.AuthRequired(deps.Config), deps.ProfileHandler.UpdateProfile)
			acc.PATCH("/profile/picture/", middleware.AuthRequired(deps.Config), deps.ProfileHandler.UploadPicture)
			acc.GET("/profile/:id/", deps.ProfileHandler.GetProfile)
			acc.PUT("/profile/:id/", middleware.AuthRequired(deps.Config), deps.ProfileHandler.UpdateProfile)
			acc.PATCH("/profile/", middleware.AuthRequired(deps.Config), deps.ProfileHandler.UpdateProfile)
			acc.PATCH("/profile/:id/", middleware.AuthRequired(deps.Config), deps.ProfileHandler.UpdateProfile)

			// About ViewSet — GET routes are public for guest browsing
			acc.GET("/about/", middleware.AuthOptional(deps.Config), deps.ProfileHandler.GetAbout)
			acc.GET("/about/user/", middleware.AuthRequired(deps.Config), deps.ProfileHandler.GetAboutUser)
			acc.POST("/about/", middleware.AuthRequired(deps.Config), deps.ProfileHandler.UpdateAbout)
			acc.GET("/about/:id/", middleware.AuthOptional(deps.Config), deps.ProfileHandler.GetAbout)
			acc.PATCH("/about/:id/", middleware.AuthRequired(deps.Config), deps.ProfileHandler.UpdateAbout)

			// Portfolio ViewSet — GET routes are public for guest browsing
			acc.GET("/portfolio/", middleware.AuthOptional(deps.Config), deps.ProfileHandler.GetPortfolios)
			acc.POST("/portfolio/", middleware.AuthRequired(deps.Config), deps.ProfileHandler.CreatePortfolio)
			acc.DELETE("/portfolio/:id/", middleware.AuthRequired(deps.Config), deps.ProfileHandler.DeletePortfolio)

			// Service Packages ViewSet — GET routes are public for guest browsing
			acc.GET("/service-packages/", middleware.AuthOptional(deps.Config), deps.ProfileHandler.GetServicePackages)
			acc.POST("/service-packages/", middleware.AuthRequired(deps.Config), deps.ProfileHandler.CreateServicePackage)
			acc.DELETE("/service-packages/:id/", middleware.AuthRequired(deps.Config), deps.ProfileHandler.DeleteServicePackage)
		}

		// Also mount directly under v1 for mobile client route variants
		v1.GET("/service-packages/", middleware.AuthOptional(deps.Config), deps.ProfileHandler.GetServicePackages)
		v1.POST("/service-packages/", middleware.AuthRequired(deps.Config), deps.ProfileHandler.CreateServicePackage)
		v1.DELETE("/service-packages/:id/", middleware.AuthRequired(deps.Config), deps.ProfileHandler.DeleteServicePackage)

		// ─── SERVICES ────────────────────────────────────────────────────────
		srv := v1.Group("/services")
		{
			srv.GET("/categories/", deps.ServiceHandler.GetCategories)
			srv.GET("/categories/:id/", deps.ServiceHandler.GetCategories)
			srv.GET("/catalog-services/", deps.ServiceHandler.GetCatalogServices)
			srv.GET("/catalog-services/:id/", deps.ServiceHandler.GetCatalogServices)
			srv.POST("/match-providers/", deps.ServiceHandler.MatchProviders)

			// Requests ViewSet
			srv.GET("/requests/", middleware.AuthRequired(deps.Config), deps.ServiceHandler.GetRequests)
			srv.POST("/requests/", middleware.AuthRequired(deps.Config), deps.ServiceHandler.CreateRequest)
			srv.GET("/requests/:id/", middleware.AuthRequired(deps.Config), deps.ServiceHandler.GetRequestByID)
			srv.PATCH("/requests/:id/", middleware.AuthRequired(deps.Config), deps.ServiceHandler.UpdateRequest)
			srv.DELETE("/requests/:id/", middleware.AuthRequired(deps.Config), deps.ServiceHandler.DeleteRequest)
			srv.POST("/requests/:id/approve_proposal/", middleware.AuthRequired(deps.Config), deps.ServiceHandler.ApproveProposal)
			srv.POST("/requests/:id/accept_proposal/", middleware.AuthRequired(deps.Config), deps.ServiceHandler.ApproveProposal)
			srv.POST("/requests/:id/cancel_approval/", middleware.AuthRequired(deps.Config), deps.ServiceHandler.CancelApproval)
			srv.PATCH("/requests/image/", middleware.AuthRequired(deps.Config), deps.ServiceHandler.UploadRequestImage)
			srv.POST("/requests/:id/upload_image/", middleware.AuthRequired(deps.Config), deps.ServiceHandler.UploadRequestImage)

			// Proposals ViewSet
			srv.GET("/proposals/", middleware.AuthRequired(deps.Config), deps.ServiceHandler.GetProposals)
			srv.POST("/proposals/", middleware.AuthRequired(deps.Config), deps.ServiceHandler.CreateProposal)
			srv.DELETE("/proposals/:id/", middleware.AuthRequired(deps.Config), deps.ServiceHandler.DeleteProposal)
		}

		// ─── INTERACTIONS ────────────────────────────────────────────────────
		inter := v1.Group("/interactions")

		// Public: anyone can read reviews (no auth required).
		inter.GET("/reviews/", deps.InterHandler.GetReviews)

		inter.Use(middleware.AuthRequired(deps.Config))
		{
			// Favorites
			inter.GET("/favorites/", deps.InterHandler.GetFavorites)
			inter.POST("/favorites/", deps.InterHandler.CreateFavorite)
			inter.DELETE("/favorites/:provider_id/", deps.InterHandler.DeleteFavorite)

			// Reviews (write-only requires auth)
			inter.POST("/reviews/", deps.InterHandler.CreateReview)

			// Appointments
			inter.GET("/appointments/", deps.InterHandler.GetAppointments)
			inter.POST("/appointments/", deps.InterHandler.CreateAppointment)
			inter.POST("/appointments/:id/verify-code/", deps.InterHandler.VerifyArrivalCode)
			inter.POST("/appointments/:id/verify_code/", deps.InterHandler.VerifyArrivalCode)
			inter.POST("/appointments/:id/on-the-way/", deps.InterHandler.NotifyOnTheWay)
			inter.POST("/appointments/:id/on_the_way/", deps.InterHandler.NotifyOnTheWay)
			inter.POST("/appointments/:id/complete/", deps.InterHandler.CompleteAppointment)
			inter.POST("/appointments/:id/cancel/", deps.InterHandler.CancelAppointment)
			inter.DELETE("/appointments/:id/", deps.InterHandler.CancelAppointment)

			// Disputes
			inter.GET("/disputes/", deps.InterHandler.GetDisputes)
			inter.POST("/disputes/", deps.InterHandler.CreateDispute)
			inter.POST("/disputes/:id/upload_evidence/", deps.InterHandler.UploadDisputeEvidence)
			inter.POST("/disputes/:id/upload-evidence/", deps.InterHandler.UploadDisputeEvidence)
		}

		// ─── CHAT ────────────────────────────────────────────────────────────
		cht := v1.Group("/chat")
		cht.Use(middleware.AuthRequired(deps.Config))
		{
			cht.GET("/conversations/", deps.ChatHandler.GetConversations)
			cht.POST("/conversations/", deps.ChatHandler.CreateConversation)
			cht.POST("/conversations/set_seen/", deps.ChatHandler.SetSeen)
			cht.POST("/conversations/block_chat/", deps.ChatHandler.BlockChat)
			cht.POST("/conversations/unblock_chat/", deps.ChatHandler.UnblockChat)
			cht.GET("/conversations/blocked_users/", deps.ChatHandler.GetBlockedUsers)
			cht.GET("/messages/", deps.ChatHandler.GetMessages)
			cht.POST("/messages/", deps.ChatHandler.SendMessage)
			cht.POST("/messages/set_seen/", deps.ChatHandler.SetSeen)
			cht.POST("/upload/", deps.ChatHandler.UploadChatMedia)
		}

		// ─── NOTIFICATIONS ───────────────────────────────────────────────────
		notif := v1.Group("/notifications")
		notif.Use(middleware.AuthRequired(deps.Config))
		{
			notif.GET("", deps.NotifHandler.GetNotifications)
			notif.GET("/", deps.NotifHandler.GetNotifications)
			notif.PATCH("/:id/read", deps.NotifHandler.MarkAsRead)
			notif.PATCH("/:id/read/", deps.NotifHandler.MarkAsRead)
			notif.POST("/:id/mark_as_read", deps.NotifHandler.MarkAsRead)
			notif.POST("/:id/mark_as_read/", deps.NotifHandler.MarkAsRead)
			notif.POST("/mark_all_as_read", deps.NotifHandler.MarkAllAsRead)
			notif.POST("/mark_all_as_read/", deps.NotifHandler.MarkAllAsRead)
			notif.POST("/device-token", deps.NotifHandler.RegisterDeviceToken)
			notif.POST("/device-token/", deps.NotifHandler.RegisterDeviceToken)
			notif.GET("/tokens", deps.NotifHandler.GetDeviceTokens)
			notif.GET("/tokens/", deps.NotifHandler.GetDeviceTokens)
			notif.POST("/tokens", deps.NotifHandler.RegisterDeviceToken)
			notif.POST("/tokens/", deps.NotifHandler.RegisterDeviceToken)
			notif.DELETE("/tokens", deps.NotifHandler.UnregisterDeviceToken)
			notif.DELETE("/tokens/", deps.NotifHandler.UnregisterDeviceToken)
			notif.POST("/tokens/unregister", deps.NotifHandler.UnregisterDeviceToken)
			notif.POST("/tokens/unregister/", deps.NotifHandler.UnregisterDeviceToken)
			notif.DELETE("/device-token", deps.NotifHandler.UnregisterDeviceToken)
			notif.DELETE("/device-token/", deps.NotifHandler.UnregisterDeviceToken)
		}


		// ─── PAYMENTS ────────────────────────────────────────────────────────
		pay := v1.Group("/payments")
		{
			pay.GET("/subscription-plans/", deps.PaymentHandler.GetSubscriptionPlans)
			pay.POST("/validate/apple/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.ValidateAppleReceipt)
			pay.POST("/validate/google/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.ValidateGooglePlay)
			pay.GET("/wallet/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.GetWallet)
			pay.GET("/wallet/my_wallet/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.MyWallet)
			pay.GET("/wallet/transactions/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.Transactions)
			pay.POST("/wallet/request_payout/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.RequestPayout)
			pay.POST("/wallet/onboard/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.Onboard)
			pay.GET("/wallet/onboarding-status/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.OnboardingStatus)
			pay.GET("/wallet/stripe-dashboard/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.StripeDashboard)
			pay.GET("/user/get/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.GetUserSubscription)
			pay.DELETE("/user/delete/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.DeleteUserSubscription)

			// Customer endpoints
			pay.GET("/user/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.GetCustomerUser)
			pay.GET("/customers/user/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.GetCustomerUser)
			pay.POST("/create/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.CreateCustomer)
			pay.POST("/customers/create/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.CreateCustomer)
			pay.PATCH("/ephmeral/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.CreateEphemeralKey)
			pay.PATCH("/paymentmethod/update/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.UpdatePaymentMethod)
			pay.POST("/account/connect/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.AccountConnect)
			pay.POST("/transfer/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.Transfer)
			pay.POST("/fund-appointment/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.FundAppointment)
			pay.POST("/fund-background-check/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.FundBackgroundCheck)
			pay.POST("/payment-sheet/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.PaymentSheet)
			pay.POST("/tip/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.TipProvider)
			pay.POST("/tip", middleware.AuthRequired(deps.Config), deps.PaymentHandler.TipProvider)

			// Webhooks
			pay.POST("/webhook/", deps.PaymentHandler.StripeWebhook)
			pay.POST("/webhook/stripe/", deps.PaymentHandler.StripeWebhook)
			pay.POST("/webhook/stripe", deps.PaymentHandler.StripeWebhook)
			pay.POST("/webhook/apple-s2s/", deps.PaymentHandler.AppleS2SWebhook)
			pay.POST("/webhook/apple-s2s", deps.PaymentHandler.AppleS2SWebhook)
			pay.POST("/webhook/google-pubsub/", deps.PaymentHandler.GooglePubSubWebhook)
		}

		// Alternate mapped payment paths (Django URL multi-include compatibility)
		v1.GET("/subscription/user/get/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.GetUserSubscription)
		v1.DELETE("/subscription/user/delete/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.DeleteUserSubscription)
		v1.GET("/subscription/subscription-plans/", deps.PaymentHandler.GetSubscriptionPlans)
		v1.GET("/wallet/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.GetWallet)
		v1.GET("/wallet/my_wallet/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.MyWallet)
		v1.GET("/wallet/transactions/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.Transactions)
		v1.POST("/wallet/request_payout/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.RequestPayout)
		v1.POST("/wallet/onboard/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.Onboard)
		v1.GET("/wallet/onboarding-status/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.OnboardingStatus)
		v1.GET("/wallet/stripe-dashboard/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.StripeDashboard)
		v1.GET("/customer/user/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.GetCustomerUser)
		v1.POST("/customer/create/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.CreateCustomer)
		v1.PATCH("/customer/ephmeral/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.CreateEphemeralKey)
		v1.PATCH("/customer/paymentmethod/update/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.UpdatePaymentMethod)
		v1.POST("/customer/account/connect/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.AccountConnect)
		v1.POST("/customer/transfer/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.Transfer)
		v1.POST("/customer/fund-appointment/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.FundAppointment)
		v1.POST("/customer/fund-background-check/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.FundBackgroundCheck)
		v1.POST("/customer/payment-sheet/", middleware.AuthRequired(deps.Config), deps.PaymentHandler.PaymentSheet)

		// ─── CONSULTATIONS ───────────────────────────────────────────────────
		cons := v1.Group("/consultations")
		cons.Use(middleware.AuthRequired(deps.Config))
		{
			cons.GET("/rtc-token/", deps.ConsultHandler.GetRTCToken)
			cons.POST("/rtc-token/", deps.ConsultHandler.GetRTCToken)
			cons.GET("/rtm-token/", deps.ConsultHandler.GetRTMToken)
			cons.POST("/rtm-token/", deps.ConsultHandler.GetRTMToken)
		}

		// ─── MODERATION & AUDIT ──────────────────────────────────────────────
		v1.POST("/moderation/reports/", middleware.AuthRequired(deps.Config), deps.ModerateHandler.SubmitReport)
		v1.GET("/moderation/reports/", middleware.AuthRequired(deps.Config), deps.ModerateHandler.SubmitReport)
		v1.GET("/moderation/verifications/", middleware.AuthRequired(deps.Config), deps.ModerateHandler.GetVerifications)
		v1.GET("/moderation/background-checks/", middleware.AuthRequired(deps.Config), deps.ModerateHandler.GetBackgroundChecks)
		v1.GET("/moderation/background-checks/config/", deps.ModerateHandler.GetBackgroundCheckConfig)
		v1.POST("/moderation/background-checks/initiate/", middleware.AuthRequired(deps.Config), deps.ModerateHandler.InitiateBackgroundCheck)
		v1.POST("/moderation/background-checks/:id/resync/", middleware.AuthRequired(deps.Config), deps.ModerateHandler.ResyncBackgroundCheck)
		v1.POST("/moderation/checkr-webhook/", deps.ModerateHandler.CheckrWebhook)
		v1.GET("/audit/logs/", middleware.AuthRequired(deps.Config), deps.AuditHandler.GetLogs)
		v1.GET("/audit/", middleware.AuthRequired(deps.Config), deps.AuditHandler.GetLogs)
		// ─── PUBLIC SITE CMS API ─────────────────────────────────────────────
		v1.GET("/public/cms/", deps.PublicHandler.GetCMSContent)
		v1.GET("/public/cms", deps.PublicHandler.GetCMSContent)
		v1.POST("/public/contact/", deps.PublicHandler.SubmitContactMessage)
		v1.POST("/public/contact", deps.PublicHandler.SubmitContactMessage)
		v1.POST("/public/resolution/", deps.PublicHandler.SubmitResolutionReport)
		v1.POST("/public/resolution", deps.PublicHandler.SubmitResolutionReport)

		// ─── SUPERUSER & ADMIN MANAGEMENT ──────────────────────────────────
		admin := v1.Group("/admin")
		admin.Use(middleware.AdminRequired(deps.Config, deps.UserRepo))
		admin.Use(middleware.AdminIPWhitelist())
		if deps.AdminRepo != nil {
			admin.Use(middleware.AdminAuditLogger(deps.AdminRepo))
		}
		{
			// Dashboard Stats
			admin.GET("/dashboard/stats", deps.AdminHandler.GetDashboardStats)
			admin.GET("/dashboard/stats/", deps.AdminHandler.GetDashboardStats)
			admin.GET("/stats", deps.AdminHandler.GetDashboardStats)
			admin.GET("/stats/", deps.AdminHandler.GetDashboardStats)

			// Users Management
			admin.GET("/users", deps.AdminHandler.ListUsers)
			admin.GET("/users/", deps.AdminHandler.ListUsers)
			admin.POST("/users", deps.AdminHandler.CreateUser)
			admin.POST("/users/", deps.AdminHandler.CreateUser)
			admin.GET("/users/:id", deps.AdminHandler.GetUserByID)
			admin.GET("/users/:id/", deps.AdminHandler.GetUserByID)
			admin.PATCH("/users/:id", deps.AdminHandler.UpdateUser)
			admin.PATCH("/users/:id/", deps.AdminHandler.UpdateUser)
			admin.PUT("/users/:id", deps.AdminHandler.UpdateUser)
			admin.PUT("/users/:id/", deps.AdminHandler.UpdateUser)
			admin.DELETE("/users/:id", deps.AdminHandler.DeleteUser)
			admin.DELETE("/users/:id/", deps.AdminHandler.DeleteUser)

			// Provider Identity Verifications
			admin.GET("/verifications", deps.AdminHandler.ListVerifications)
			admin.GET("/verifications/", deps.AdminHandler.ListVerifications)
			admin.POST("/verifications/:id/approve", deps.AdminHandler.ApproveVerification)
			admin.POST("/verifications/:id/approve/", deps.AdminHandler.ApproveVerification)
			admin.POST("/verifications/:id/reject", deps.AdminHandler.RejectVerification)
			admin.POST("/verifications/:id/reject/", deps.AdminHandler.RejectVerification)

			// Background Checks
			admin.GET("/background-checks", deps.AdminHandler.ListBackgroundChecks)
			admin.GET("/background-checks/", deps.AdminHandler.ListBackgroundChecks)
			admin.POST("/background-checks/:id/override", deps.AdminHandler.OverrideBackgroundCheck)
			admin.POST("/background-checks/:id/override/", deps.AdminHandler.OverrideBackgroundCheck)

			// Reports & Moderation
			admin.GET("/reports", deps.AdminHandler.ListReports)
			admin.GET("/reports/", deps.AdminHandler.ListReports)
			admin.POST("/reports/:id/resolve", deps.AdminHandler.ResolveReport)
			admin.POST("/reports/:id/resolve/", deps.AdminHandler.ResolveReport)

			// Disputes
			admin.GET("/disputes", deps.AdminHandler.ListDisputes)
			admin.GET("/disputes/", deps.AdminHandler.ListDisputes)
			admin.POST("/disputes/:id/resolve", deps.AdminHandler.ResolveDispute)
			admin.POST("/disputes/:id/resolve/", deps.AdminHandler.ResolveDispute)
			admin.POST("/disputes/:id/reject", deps.AdminHandler.RejectDispute)
			admin.POST("/disputes/:id/reject/", deps.AdminHandler.RejectDispute)

			// Payouts & Wallets
			admin.GET("/payouts", deps.AdminHandler.ListPayoutRequests)
			admin.GET("/payouts/", deps.AdminHandler.ListPayoutRequests)
			admin.POST("/payouts/:id/approve", deps.AdminHandler.ApprovePayout)
			admin.POST("/payouts/:id/approve/", deps.AdminHandler.ApprovePayout)
			admin.POST("/payouts/:id/reject", deps.AdminHandler.RejectPayout)
			admin.POST("/payouts/:id/reject/", deps.AdminHandler.RejectPayout)
			admin.POST("/payouts/batch-approve", deps.AdminHandler.BatchApprovePayouts)
			admin.POST("/payouts/batch-approve/", deps.AdminHandler.BatchApprovePayouts)
			admin.POST("/payouts/batch-reject", deps.AdminHandler.BatchRejectPayouts)
			admin.POST("/payouts/batch-reject/", deps.AdminHandler.BatchRejectPayouts)
			admin.GET("/wallets", deps.AdminHandler.ListWallets)
			admin.GET("/wallets/", deps.AdminHandler.ListWallets)

			// Live Notification Feed & Funnel
			admin.GET("/notifications/feed", deps.AdminHandler.GetNotificationFeed)
			admin.GET("/notifications/feed/", deps.AdminHandler.GetNotificationFeed)
			admin.GET("/funnel/providers", deps.AdminHandler.GetProviderOnboardingFunnel)
			admin.GET("/funnel/providers/", deps.AdminHandler.GetProviderOnboardingFunnel)

			// Staff Internal Notes
			admin.GET("/users/:id/notes", deps.AdminHandler.ListStaffNotes)
			admin.GET("/users/:id/notes/", deps.AdminHandler.ListStaffNotes)
			admin.POST("/users/:id/notes", deps.AdminHandler.CreateStaffNote)
			admin.POST("/users/:id/notes/", deps.AdminHandler.CreateStaffNote)
			admin.DELETE("/users/:id/notes/:noteId", deps.AdminHandler.DeleteStaffNote)
			admin.DELETE("/users/:id/notes/:noteId/", deps.AdminHandler.DeleteStaffNote)

			// Subscriptions & Plans
			admin.GET("/subscriptions", deps.AdminHandler.ListSubscriptions)
			admin.GET("/subscriptions/", deps.AdminHandler.ListSubscriptions)
			admin.POST("/subscriptions", deps.AdminHandler.AssignSubscription)
			admin.POST("/subscriptions/", deps.AdminHandler.AssignSubscription)
			admin.POST("/subscriptions/assign", deps.AdminHandler.AssignSubscription)
			admin.POST("/subscriptions/assign/", deps.AdminHandler.AssignSubscription)
			admin.POST("/subscriptions/:id/toggle", deps.AdminHandler.ToggleSubscription)
			admin.POST("/subscriptions/:id/toggle/", deps.AdminHandler.ToggleSubscription)
			admin.GET("/subscription-plans", deps.AdminHandler.ListSubscriptionPlans)
			admin.GET("/subscription-plans/", deps.AdminHandler.ListSubscriptionPlans)
			admin.POST("/subscription-plans", deps.AdminHandler.CreateSubscriptionPlan)
			admin.POST("/subscription-plans/", deps.AdminHandler.CreateSubscriptionPlan)
			admin.PUT("/subscription-plans/:id", deps.AdminHandler.UpdateSubscriptionPlan)
			admin.PUT("/subscription-plans/:id/", deps.AdminHandler.UpdateSubscriptionPlan)
			admin.PATCH("/subscription-plans/:id", deps.AdminHandler.UpdateSubscriptionPlan)
			admin.PATCH("/subscription-plans/:id/", deps.AdminHandler.UpdateSubscriptionPlan)
			admin.DELETE("/subscription-plans/:id", deps.AdminHandler.DeleteSubscriptionPlan)
			admin.DELETE("/subscription-plans/:id/", deps.AdminHandler.DeleteSubscriptionPlan)

			// Categories & Catalog Services
			admin.GET("/categories", deps.ServiceHandler.GetCategories)
			admin.GET("/categories/", deps.ServiceHandler.GetCategories)
			admin.POST("/categories", deps.AdminHandler.CreateCategory)
			admin.POST("/categories/", deps.AdminHandler.CreateCategory)
			admin.PUT("/categories/:id", deps.AdminHandler.UpdateCategory)
			admin.PUT("/categories/:id/", deps.AdminHandler.UpdateCategory)
			admin.PATCH("/categories/:id", deps.AdminHandler.UpdateCategory)
			admin.PATCH("/categories/:id/", deps.AdminHandler.UpdateCategory)
			admin.DELETE("/categories/:id", deps.AdminHandler.DeleteCategory)
			admin.DELETE("/categories/:id/", deps.AdminHandler.DeleteCategory)

			admin.GET("/catalog-services", deps.ServiceHandler.GetCatalogServices)
			admin.GET("/catalog-services/", deps.ServiceHandler.GetCatalogServices)
			admin.POST("/catalog-services", deps.AdminHandler.CreateCatalogService)
			admin.POST("/catalog-services/", deps.AdminHandler.CreateCatalogService)
			admin.PUT("/catalog-services/:id", deps.AdminHandler.UpdateCatalogService)
			admin.PUT("/catalog-services/:id/", deps.AdminHandler.UpdateCatalogService)
			admin.PATCH("/catalog-services/:id", deps.AdminHandler.UpdateCatalogService)
			admin.PATCH("/catalog-services/:id/", deps.AdminHandler.UpdateCatalogService)
			admin.DELETE("/catalog-services/:id", deps.AdminHandler.DeleteCatalogService)
			admin.DELETE("/catalog-services/:id/", deps.AdminHandler.DeleteCatalogService)

			// Settings & Audit Logs
			admin.GET("/settings", deps.AdminHandler.GetSettings)
			admin.GET("/settings/", deps.AdminHandler.GetSettings)
			admin.PUT("/settings", deps.AdminHandler.UpdateSettings)
			admin.PUT("/settings/", deps.AdminHandler.UpdateSettings)
			admin.PATCH("/settings", deps.AdminHandler.UpdateSettings)
			admin.PATCH("/settings/", deps.AdminHandler.UpdateSettings)
			admin.GET("/audit-logs", deps.AdminHandler.ListAuditLogs)
			admin.GET("/audit-logs/", deps.AdminHandler.ListAuditLogs)

			// User Restoration & Wallet Adjustment
			admin.POST("/users/:id/restore", deps.AdminHandler.RestoreUser)
			admin.POST("/users/:id/restore/", deps.AdminHandler.RestoreUser)
			admin.POST("/wallets/:id/adjust", deps.AdminHandler.AdjustWallet)
			admin.POST("/wallets/:id/adjust/", deps.AdminHandler.AdjustWallet)

			// Batch Moderation & Broadcast Notifications
			admin.POST("/verifications/batch", deps.AdminHandler.BatchVerifications)
			admin.POST("/verifications/batch/", deps.AdminHandler.BatchVerifications)
			admin.POST("/notifications/broadcast", deps.AdminHandler.BroadcastNotification)
			admin.POST("/notifications/broadcast/", deps.AdminHandler.BroadcastNotification)
			admin.POST("/change-password", deps.AuthHandler.ChangePassword)
			admin.POST("/change-password/", deps.AuthHandler.ChangePassword)

			// Feature Flags & System Operations
			admin.GET("/feature-flags", deps.AdminHandler.ListFeatureFlags)
			admin.GET("/feature-flags/", deps.AdminHandler.ListFeatureFlags)
			admin.PUT("/feature-flags/:key", deps.AdminHandler.SetFeatureFlag)
			admin.PUT("/feature-flags/:key/", deps.AdminHandler.SetFeatureFlag)
			admin.PATCH("/feature-flags/:key", deps.AdminHandler.SetFeatureFlag)
			admin.PATCH("/feature-flags/:key/", deps.AdminHandler.SetFeatureFlag)
			admin.POST("/cache/clear", deps.AdminHandler.ClearCache)
			admin.POST("/cache/clear/", deps.AdminHandler.ClearCache)

			// Reports, System Health & GDPR
			admin.GET("/system/health", deps.AdminHandler.GetSystemHealth)
			admin.GET("/system/health/", deps.AdminHandler.GetSystemHealth)
			admin.GET("/system/workers", deps.AdminHandler.GetWorkersStatus)
			admin.GET("/system/workers/", deps.AdminHandler.GetWorkersStatus)
			admin.GET("/financial/stripe-balance", deps.AdminHandler.GetStripeLiveBalance)
			admin.GET("/financial/stripe-balance/", deps.AdminHandler.GetStripeLiveBalance)
			admin.POST("/users/:id/message", deps.AdminHandler.SendDirectUserMessage)
			admin.POST("/users/:id/message/", deps.AdminHandler.SendDirectUserMessage)
			admin.GET("/reports/financial", deps.AdminHandler.GetFinancialReport)
			admin.GET("/reports/financial/", deps.AdminHandler.GetFinancialReport)
			admin.GET("/users/:id/gdpr-export", deps.AdminHandler.GetGDPRUserData)
			admin.GET("/users/:id/gdpr-export/", deps.AdminHandler.GetGDPRUserData)

			// Data Exports (CSV)
			admin.GET("/export/users", deps.AdminHandler.ExportUsers)
			admin.GET("/export/users/", deps.AdminHandler.ExportUsers)
			admin.GET("/export/payouts", deps.AdminHandler.ExportPayouts)
			admin.GET("/export/payouts/", deps.AdminHandler.ExportPayouts)
			admin.GET("/export/disputes", deps.AdminHandler.ExportDisputes)
			admin.GET("/export/disputes/", deps.AdminHandler.ExportDisputes)
			admin.GET("/export/appointments", deps.AdminHandler.ExportAppointments)
			admin.GET("/export/appointments/", deps.AdminHandler.ExportAppointments)
			admin.GET("/export/verifications", deps.AdminHandler.ExportVerifications)
			admin.GET("/export/verifications/", deps.AdminHandler.ExportVerifications)
			admin.GET("/export/background-checks", deps.AdminHandler.ExportBackgroundChecks)
			admin.GET("/export/background-checks/", deps.AdminHandler.ExportBackgroundChecks)

			// Appointments & Bookings Console
			admin.GET("/appointments", deps.AdminHandler.ListAppointments)
			admin.GET("/appointments/", deps.AdminHandler.ListAppointments)
			admin.GET("/appointments/:id", deps.AdminHandler.GetAppointmentByID)
			admin.GET("/appointments/:id/", deps.AdminHandler.GetAppointmentByID)
			admin.POST("/appointments/:id/status", deps.AdminHandler.UpdateAppointmentStatus)
			admin.POST("/appointments/:id/status/", deps.AdminHandler.UpdateAppointmentStatus)

			// Reviews Moderation Queue
			admin.GET("/reviews", deps.AdminHandler.ListReviews)
			admin.GET("/reviews/", deps.AdminHandler.ListReviews)
			admin.POST("/reviews", deps.AdminHandler.CreateReview)
			admin.POST("/reviews/", deps.AdminHandler.CreateReview)
			admin.POST("/reviews/:id/toggle-hide", deps.AdminHandler.ToggleReviewVisibility)
			admin.POST("/reviews/:id/toggle-hide/", deps.AdminHandler.ToggleReviewVisibility)
			admin.DELETE("/reviews/:id", deps.AdminHandler.DeleteReview)
			admin.DELETE("/reviews/:id/", deps.AdminHandler.DeleteReview)

			// Promo Codes & Marketing Campaigns
			admin.GET("/promo-codes", deps.AdminHandler.ListPromoCodes)
			admin.GET("/promo-codes/", deps.AdminHandler.ListPromoCodes)
			admin.POST("/promo-codes", deps.AdminHandler.CreatePromoCode)
			admin.POST("/promo-codes/", deps.AdminHandler.CreatePromoCode)
			admin.PUT("/promo-codes/:id", deps.AdminHandler.UpdatePromoCode)
			admin.PUT("/promo-codes/:id/", deps.AdminHandler.UpdatePromoCode)
			admin.DELETE("/promo-codes/:id", deps.AdminHandler.DeletePromoCode)
			admin.DELETE("/promo-codes/:id/", deps.AdminHandler.DeletePromoCode)

			// Dispute Refund Execution
			admin.POST("/disputes/:id/refund", deps.AdminHandler.ExecuteDisputeRefund)
			admin.POST("/disputes/:id/refund/", deps.AdminHandler.ExecuteDisputeRefund)


			// TOTP 2FA Engine
			admin.POST("/2fa/setup", deps.AdminHandler.Setup2FA)
			admin.POST("/2fa/setup/", deps.AdminHandler.Setup2FA)
			admin.POST("/2fa/verify", middleware.TwoFactorRateLimit(), deps.AdminHandler.Verify2FA)
			admin.POST("/2fa/verify/", middleware.TwoFactorRateLimit(), deps.AdminHandler.Verify2FA)
			admin.POST("/2fa/disable", middleware.TwoFactorRateLimit(), deps.AdminHandler.Disable2FA)
			admin.POST("/2fa/disable/", middleware.TwoFactorRateLimit(), deps.AdminHandler.Disable2FA)

			// Staff Impersonation
			admin.POST("/users/:id/impersonate", deps.AdminHandler.ImpersonateUser)
			admin.POST("/users/:id/impersonate/", deps.AdminHandler.ImpersonateUser)

			// RBAC Roles & Permissions
			admin.GET("/roles", deps.AdminHandler.ListRoles)
			admin.GET("/roles/", deps.AdminHandler.ListRoles)
			admin.POST("/roles", deps.AdminHandler.CreateRole)
			admin.POST("/roles/", deps.AdminHandler.CreateRole)
			admin.PUT("/roles/:id", deps.AdminHandler.UpdateRole)
			admin.PUT("/roles/:id/", deps.AdminHandler.UpdateRole)
			admin.DELETE("/roles/:id", deps.AdminHandler.DeleteRole)
			admin.DELETE("/roles/:id/", deps.AdminHandler.DeleteRole)
			admin.POST("/users/:id/role", deps.AdminHandler.AssignUserRole)
			admin.POST("/users/:id/role/", deps.AdminHandler.AssignUserRole)

			// Fraud & Risk Scoring Detection
			admin.GET("/fraud/risk-alerts", deps.AdminHandler.ListFraudRiskAlerts)
			admin.GET("/fraud/risk-alerts/", deps.AdminHandler.ListFraudRiskAlerts)
			admin.POST("/fraud/evaluate/:id", deps.AdminHandler.EvaluateUserRisk)
			admin.POST("/fraud/evaluate/:id/", deps.AdminHandler.EvaluateUserRisk)
			admin.POST("/fraud/risk-alerts/:id/resolve", deps.AdminHandler.ResolveRiskAlert)
			admin.POST("/fraud/risk-alerts/:id/resolve/", deps.AdminHandler.ResolveRiskAlert)

			// Notification & Email Templates
			admin.GET("/templates/emails", deps.AdminHandler.ListNotificationTemplates)
			admin.GET("/templates/emails/", deps.AdminHandler.ListNotificationTemplates)
			admin.PUT("/templates/emails/:key", deps.AdminHandler.UpdateNotificationTemplate)
			admin.PUT("/templates/emails/:key/", deps.AdminHandler.UpdateNotificationTemplate)
			admin.POST("/templates/emails/test-send", deps.AdminHandler.TestSendEmailTemplate)
			admin.POST("/templates/emails/test-send/", deps.AdminHandler.TestSendEmailTemplate)

			// System Backups
			admin.POST("/system/backup", deps.AdminHandler.TriggerBackup)
			admin.POST("/system/backup/", deps.AdminHandler.TriggerBackup)
			admin.GET("/system/backups", deps.AdminHandler.ListBackups)
			admin.GET("/system/backups/", deps.AdminHandler.ListBackups)
			admin.GET("/system/backups/:id/download", deps.AdminHandler.DownloadBackup)
			admin.GET("/system/backups/:id/download/", deps.AdminHandler.DownloadBackup)

			// Legal & Compliance Documents
			admin.GET("/legal", deps.AdminHandler.ListLegalDocuments)
			admin.GET("/legal/", deps.AdminHandler.ListLegalDocuments)
			admin.POST("/legal", deps.AdminHandler.CreateLegalDocument)
			admin.POST("/legal/", deps.AdminHandler.CreateLegalDocument)
			admin.GET("/legal/:id", deps.AdminHandler.GetLegalDocument)
			admin.GET("/legal/:id/", deps.AdminHandler.GetLegalDocument)
			admin.PUT("/legal/:id", deps.AdminHandler.UpdateLegalDocument)
			admin.PUT("/legal/:id/", deps.AdminHandler.UpdateLegalDocument)
			admin.DELETE("/legal/:id", deps.AdminHandler.DeleteLegalDocument)
			admin.DELETE("/legal/:id/", deps.AdminHandler.DeleteLegalDocument)

			// Support & Resolution Inbox
			admin.GET("/support/messages", deps.AdminHandler.ListContactMessages)
			admin.GET("/support/messages/", deps.AdminHandler.ListContactMessages)
			admin.POST("/support/messages/:id/toggle-resolved", deps.AdminHandler.ToggleContactMessageResolved)
			admin.POST("/support/messages/:id/toggle-resolved/", deps.AdminHandler.ToggleContactMessageResolved)
			admin.DELETE("/support/messages/:id", deps.AdminHandler.DeleteContactMessage)
			admin.DELETE("/support/messages/:id/", deps.AdminHandler.DeleteContactMessage)

			admin.GET("/support/resolutions", deps.AdminHandler.ListResolutionReports)
			admin.GET("/support/resolutions/", deps.AdminHandler.ListResolutionReports)
			admin.POST("/support/resolutions/:id/toggle-reviewed", deps.AdminHandler.ToggleResolutionReportReviewed)
			admin.POST("/support/resolutions/:id/toggle-reviewed/", deps.AdminHandler.ToggleResolutionReportReviewed)
			admin.DELETE("/support/resolutions/:id", deps.AdminHandler.DeleteResolutionReport)
			admin.DELETE("/support/resolutions/:id/", deps.AdminHandler.DeleteResolutionReport)

			// Public Site Marketing & CMS
			admin.GET("/cms/faqs", deps.AdminHandler.ListFAQs)
			admin.GET("/cms/faqs/", deps.AdminHandler.ListFAQs)
			admin.POST("/cms/faqs", deps.AdminHandler.CreateFAQ)
			admin.POST("/cms/faqs/", deps.AdminHandler.CreateFAQ)
			admin.PUT("/cms/faqs/:id", deps.AdminHandler.UpdateFAQ)
			admin.PUT("/cms/faqs/:id/", deps.AdminHandler.UpdateFAQ)
			admin.DELETE("/cms/faqs/:id", deps.AdminHandler.DeleteFAQ)
			admin.DELETE("/cms/faqs/:id/", deps.AdminHandler.DeleteFAQ)

			admin.GET("/cms/testimonials", deps.AdminHandler.ListTestimonials)
			admin.GET("/cms/testimonials/", deps.AdminHandler.ListTestimonials)
			admin.POST("/cms/testimonials", deps.AdminHandler.CreateTestimonial)
			admin.POST("/cms/testimonials/", deps.AdminHandler.CreateTestimonial)
			admin.PUT("/cms/testimonials/:id", deps.AdminHandler.UpdateTestimonial)
			admin.PUT("/cms/testimonials/:id/", deps.AdminHandler.UpdateTestimonial)
			admin.DELETE("/cms/testimonials/:id", deps.AdminHandler.DeleteTestimonial)
			admin.DELETE("/cms/testimonials/:id/", deps.AdminHandler.DeleteTestimonial)

			admin.GET("/cms/hero", deps.AdminHandler.GetHeroSection)
			admin.GET("/cms/hero/", deps.AdminHandler.GetHeroSection)
			admin.PUT("/cms/hero", deps.AdminHandler.UpdateHeroSection)
			admin.PUT("/cms/hero/", deps.AdminHandler.UpdateHeroSection)

			admin.GET("/cms/stats", deps.AdminHandler.ListSiteStats)
			admin.GET("/cms/stats/", deps.AdminHandler.ListSiteStats)
			admin.POST("/cms/stats", deps.AdminHandler.CreateSiteStat)
			admin.POST("/cms/stats/", deps.AdminHandler.CreateSiteStat)
			admin.PUT("/cms/stats/:id", deps.AdminHandler.UpdateSiteStat)
			admin.PUT("/cms/stats/:id/", deps.AdminHandler.UpdateSiteStat)
			admin.DELETE("/cms/stats/:id", deps.AdminHandler.DeleteSiteStat)
			admin.DELETE("/cms/stats/:id/", deps.AdminHandler.DeleteSiteStat)

			admin.GET("/cms/about", deps.AdminHandler.GetAboutContent)
			admin.GET("/cms/about/", deps.AdminHandler.GetAboutContent)
			admin.PUT("/cms/about", deps.AdminHandler.UpdateAboutContent)
			admin.PUT("/cms/about/", deps.AdminHandler.UpdateAboutContent)

			// Interactive OpenAPI 3.0 & Swagger UI
			admin.GET("/openapi.json", deps.AdminHandler.GetOpenAPISpec)
			admin.GET("/docs", deps.AdminHandler.GetSwaggerUI)
			admin.GET("/docs/", deps.AdminHandler.GetSwaggerUI)

			// Maintenance Mode
			admin.GET("/maintenance", deps.AdminHandler.GetMaintenanceMode)
			admin.GET("/maintenance/", deps.AdminHandler.GetMaintenanceMode)
			admin.POST("/maintenance", deps.AdminHandler.SetMaintenanceMode)
			admin.POST("/maintenance/", deps.AdminHandler.SetMaintenanceMode)

			// Webhook Event Log
			admin.GET("/webhooks/events", deps.AdminHandler.ListWebhookEvents)
			admin.GET("/webhooks/events/", deps.AdminHandler.ListWebhookEvents)
		}
	}

	// Flutter client direct endpoints (/api/notifications/)
	notifDirect := r.Group("/api/notifications")
	notifDirect.Use(middleware.AuthRequired(deps.Config))
	{
		notifDirect.GET("", deps.NotifHandler.GetNotifications)
		notifDirect.GET("/", deps.NotifHandler.GetNotifications)
		notifDirect.PATCH("/:id/read", deps.NotifHandler.MarkAsRead)
		notifDirect.PATCH("/:id/read/", deps.NotifHandler.MarkAsRead)
		notifDirect.POST("/:id/mark_as_read", deps.NotifHandler.MarkAsRead)
		notifDirect.POST("/:id/mark_as_read/", deps.NotifHandler.MarkAsRead)
		notifDirect.POST("/mark_all_as_read", deps.NotifHandler.MarkAllAsRead)
		notifDirect.POST("/mark_all_as_read/", deps.NotifHandler.MarkAllAsRead)
		notifDirect.POST("/device-token", deps.NotifHandler.RegisterDeviceToken)
		notifDirect.POST("/device-token/", deps.NotifHandler.RegisterDeviceToken)
		notifDirect.GET("/tokens", deps.NotifHandler.GetDeviceTokens)
		notifDirect.GET("/tokens/", deps.NotifHandler.GetDeviceTokens)
		notifDirect.POST("/tokens", deps.NotifHandler.RegisterDeviceToken)
		notifDirect.POST("/tokens/", deps.NotifHandler.RegisterDeviceToken)
		notifDirect.DELETE("/tokens", deps.NotifHandler.UnregisterDeviceToken)
		notifDirect.DELETE("/tokens/", deps.NotifHandler.UnregisterDeviceToken)
		notifDirect.POST("/tokens/unregister", deps.NotifHandler.UnregisterDeviceToken)
		notifDirect.POST("/tokens/unregister/", deps.NotifHandler.UnregisterDeviceToken)
		notifDirect.DELETE("/device-token", deps.NotifHandler.UnregisterDeviceToken)
		notifDirect.DELETE("/device-token/", deps.NotifHandler.UnregisterDeviceToken)
	}


	// WebSockets
	r.GET("/ws/presence/", deps.ChatHandler.HandleWebSocket)
	r.GET("/ws/presence", deps.ChatHandler.HandleWebSocket)
	r.GET("/ws/chat/:conversation_id/", deps.ChatHandler.HandleWebSocket)
	r.GET("/ws/chat/:conversation_id", deps.ChatHandler.HandleWebSocket)
	r.GET("/ws/:conversation_id/", deps.ChatHandler.HandleWebSocket)
	r.GET("/ws/:conversation_id", deps.ChatHandler.HandleWebSocket)
	r.GET("/ws/notifications/", deps.NotifHandler.HandleWebSocket)
	r.GET("/ws/notifications", deps.NotifHandler.HandleWebSocket)
	r.GET("/ws/tracking/:appointment_id/", deps.ChatHandler.HandleWebSocket)
	r.GET("/ws/tracking/:appointment_id", deps.ChatHandler.HandleWebSocket)
	r.GET("/ws/admin/events/", deps.ChatHandler.HandleWebSocket)
	r.GET("/ws/admin/events", deps.ChatHandler.HandleWebSocket)
	r.GET("/ws/", deps.ChatHandler.HandleWebSocket)
	r.GET("/ws", deps.ChatHandler.HandleWebSocket)

	return r
}
