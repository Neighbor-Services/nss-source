package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/database"
	deliveryHttp "backend-go/internal/delivery/http"
	"backend-go/internal/delivery/http/handler"
	gormRepo "backend-go/internal/repository/gorm"
	"backend-go/internal/usecase"
	"backend-go/internal/websocket"
	"backend-go/internal/worker"
	"backend-go/pkg/cache"
	"backend-go/pkg/fcm"
	"backend-go/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load Application Configuration
	cfg := config.Load()
	logger.InitGlobalLogger(cfg.Env)

	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		if cfg.Env == "production" || cfg.Env == "release" {
			ginMode = gin.ReleaseMode
		} else {
			ginMode = gin.ReleaseMode // Default to release mode as requested
		}
	}
	gin.SetMode(ginMode)

	// 2. Connect Database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// 2.1 Connect Redis Cache
	redisCache := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, cfg.RedisURL)
	defer redisCache.Close()


	// Run auto-migrations if configured
	if err := database.AutoMigrate(db); err != nil {
		log.Printf("Warning: Database auto-migration: %v", err)
	}

	// 3. Start Real-time WebSocket Hub
	hub := websocket.GlobalHub
	go hub.Run()

	// 4. Initialize Push Notifications (Firebase Cloud Messaging)
	fcmClient, err := fcm.NewClient(cfg.FirebaseCredentialsFile, cfg.FirebaseCredentialsJSON)
	if err != nil {
		log.Printf("Warning: Firebase Cloud Messaging client initialization: %v", err)
	}

	// 5. Clean Architecture Dependency Injection
	// Layer: Repositories (Infrastructure / Data Access)
	userRepo := gormRepo.NewUserRepository(db)
	profileRepo := gormRepo.NewProfileRepository(db)
	aboutRepo := gormRepo.NewAboutRepository(db)
	portfolioRepo := gormRepo.NewPortfolioRepository(db)
	servicePackageRepo := gormRepo.NewServicePackageRepository(db)
	legalRepo := gormRepo.NewLegalDocumentRepository(db)

	categoryRepo := gormRepo.NewCategoryRepository(db)
	catalogRepo := gormRepo.NewCatalogServiceRepository(db)
	requestRepo := gormRepo.NewServiceRequestRepository(db)
	proposalRepo := gormRepo.NewProposalRepository(db)

	favRepo := gormRepo.NewFavoriteRepository(db)
	reviewRepo := gormRepo.NewReviewRepository(db)
	aptRepo := gormRepo.NewAppointmentRepository(db)
	disputeRepo := gormRepo.NewDisputeRepository(db)

	convRepo := gormRepo.NewConversationRepository(db)
	msgRepo := gormRepo.NewMessageRepository(db)
	chatBlockRepo := gormRepo.NewChatBlockRepository(db)

	notifRepo := gormRepo.NewNotificationRepository(db)
	tokenRepo := gormRepo.NewDeviceTokenRepository(db)

	planRepo := gormRepo.NewSubscriptionPlanRepository(db)
	subRepo := gormRepo.NewUserSubscriptionRepository(db)
	walletRepo := gormRepo.NewWalletRepository(db)
	walletTxRepo := gormRepo.NewWalletTransactionRepository(db)
	payoutRepo := gormRepo.NewPayoutRequestRepository(db)
	customerRepo := gormRepo.NewCustomerRepository(db)

	reportRepo := gormRepo.NewReportRepository(db)
	verificationRepo := gormRepo.NewProviderVerificationRepository(db)
	bgCheckRepo := gormRepo.NewBackgroundCheckRepository(db)
	auditRepo := gormRepo.NewAuditLogRepository(db)
	adminRepo := gormRepo.NewAdminRepository(db)
	publicRepo := gormRepo.NewPublicRepository(db)

	// Layer: UseCases (Application Business Rules)
	authUC := usecase.NewAuthUseCase(userRepo, profileRepo, walletRepo, cfg)
	profileUC := usecase.NewProfileUseCase(profileRepo, aboutRepo, portfolioRepo, servicePackageRepo, legalRepo)
	serviceUC := usecase.NewServiceUseCase(categoryRepo, catalogRepo, requestRepo, proposalRepo, profileRepo, aptRepo, notifRepo, tokenRepo, fcmClient, userRepo, adminRepo, redisCache, cfg)
	interUC := usecase.NewInteractionUseCase(favRepo, reviewRepo, aptRepo, disputeRepo, profileRepo, walletRepo, walletTxRepo, userRepo, notifRepo, tokenRepo, fcmClient, cfg)
	chatUC := usecase.NewChatUseCase(convRepo, msgRepo, chatBlockRepo, tokenRepo, fcmClient)
	notifUC := usecase.NewNotificationUseCase(notifRepo, tokenRepo, fcmClient)
	paymentUC := usecase.NewPaymentUseCase(planRepo, subRepo, walletRepo, walletTxRepo, payoutRepo, customerRepo, profileRepo, notifRepo, tokenRepo, fcmClient, cfg)
	consultUC := usecase.NewConsultationUseCase(cfg)
	moderateUC := usecase.NewModerationUseCase(reportRepo, verificationRepo, bgCheckRepo, profileRepo, userRepo, cfg)
	auditUC := usecase.NewAuditUseCase(auditRepo)
	adminUC := usecase.NewAdminUseCase(adminRepo, profileRepo, userRepo, walletRepo, cfg)
	publicUC := usecase.NewPublicUseCase(publicRepo, catalogRepo, categoryRepo, cfg)

	// Layer: Delivery Handlers (Interface Adapters / Transport)
	authH := handler.NewAuthHandler(authUC)
	profileH := handler.NewProfileHandler(profileUC)
	serviceH := handler.NewServiceHandler(serviceUC)
	interH := handler.NewInteractionHandler(interUC)
	chatH := handler.NewChatHandler(chatUC, hub, cfg)
	notifH := handler.NewNotificationHandler(notifUC, hub, cfg)
	paymentH := handler.NewPaymentHandler(paymentUC)
	consultH := handler.NewConsultationHandler(consultUC)
	moderateH := handler.NewModerationHandler(moderateUC)
	auditH := handler.NewAuditHandler(auditUC)
	adminH := handler.NewAdminHandler(adminUC, cfg)
	publicH := handler.NewPublicHandler(publicUC)
	docsH := handler.NewDocsHandler()

	// 5. Setup HTTP Router with Handlers & Middleware
	router := deliveryHttp.SetupRouter(deliveryHttp.RouterDependencies{
		Config:          cfg,
		UserRepo:        userRepo,
		AdminRepo:       adminRepo,
		AuditRepo:       auditRepo,
		AuthHandler:     authH,
		ProfileHandler:  profileH,
		ServiceHandler:  serviceH,
		InterHandler:    interH,
		ChatHandler:     chatH,
		NotifHandler:    notifH,
		PaymentHandler:  paymentH,
		ConsultHandler:  consultH,
		ModerateHandler: moderateH,
		AuditHandler:    auditH,
		AdminHandler:    adminH,
		PublicHandler:   publicH,
		DocsHandler:     docsH,
	})

	// 6. Start Background Workers
	marketingWorker := worker.NewMarketingWorker(db, cfg, fcmClient)
	marketingWorker.Start()

	nearbyWorker := worker.NewNearbyRequestWorker(db, cfg, fcmClient, redisCache)
	nearbyWorker.Start()

	staleWorker := worker.NewStaleRequestWorker(db, cfg, fcmClient)
	staleWorker.Start()

	payoutWorker := worker.NewPayoutProcessingWorker(db, cfg, fcmClient)
	payoutWorker.Start()

	subExpiryWorker := worker.NewSubscriptionExpiryWorker(db, cfg, fcmClient)
	subExpiryWorker.Start()

	disputeWorker := worker.NewDisputeEscalationWorker(db, cfg, fcmClient)
	disputeWorker.Start()

	inactiveUserWorker := worker.NewInactiveUserWorker(db, cfg, fcmClient)
	inactiveUserWorker.Start()

	badgeWorker := worker.NewPerformanceBadgeWorker(db, cfg)
	badgeWorker.Start()

	reconcileWorker := worker.NewWalletReconciliationWorker(db, cfg)
	reconcileWorker.Start()

	fraudWorker := worker.NewFraudDetectionWorker(db, cfg)
	fraudWorker.Start()

	noShowWorker := worker.NewAppointmentNoShowWorker(db, cfg, fcmClient)
	noShowWorker.Start()

	reviewReminderWorker := worker.NewReviewReminderWorker(db, cfg, fcmClient)
	reviewReminderWorker.Start()

	rankingWorker := worker.NewProviderRankingWorker(db, cfg, fcmClient)
	rankingWorker.Start()

	otpCleanupWorker := worker.NewExpiredOTPCleanupWorker(db, cfg)
	otpCleanupWorker.Start()

	backupWorker := worker.NewDatabaseBackupWorker(db, cfg)
	backupWorker.Start()

	// 7. Start HTTP Server with Graceful Shutdown
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("🚀 Clean Architecture Golang Neighbor Services backend running on http://0.0.0.0:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server listen error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")
	marketingWorker.Stop()
	nearbyWorker.Stop()
	staleWorker.Stop()
	payoutWorker.Stop()
	subExpiryWorker.Stop()
	disputeWorker.Stop()
	inactiveUserWorker.Stop()
	badgeWorker.Stop()
	reconcileWorker.Stop()
	fraudWorker.Stop()
	noShowWorker.Stop()
	reviewReminderWorker.Stop()
	rankingWorker.Stop()
	otpCleanupWorker.Stop()
	backupWorker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly.")
}
