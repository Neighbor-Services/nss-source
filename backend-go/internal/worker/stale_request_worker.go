package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/pkg/cache"
	"backend-go/pkg/email"
	"backend-go/pkg/fcm"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StaleRequestWorker auto-closes OPEN service requests older than 30 days
// and notifies the seeker so they are aware.
type StaleRequestWorker struct {
	db        *gorm.DB
	cfg       *config.Config
	emailCfg  *email.Config
	fcmClient fcm.Client
	cache     cache.Cache
	stopChan  chan struct{}
}

func NewStaleRequestWorker(db *gorm.DB, cfg *config.Config, fcmClient fcm.Client, cache cache.Cache) *StaleRequestWorker {
	return &StaleRequestWorker{
		db:        db,
		cfg:       cfg,
		fcmClient: fcmClient,
		cache:     cache,
		emailCfg: &email.Config{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			User:     cfg.SMTPUser,
			Password: cfg.SMTPPassword,
			From:     cfg.EmailFrom,
		},
		stopChan: make(chan struct{}),
	}
}

func (w *StaleRequestWorker) Start() {
	log.Println("🗑️  Starting Stale Request Auto-Close Background Worker...")

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		// Run immediately on startup
		w.processStaleRequests()

		for {
			select {
			case <-w.stopChan:
				log.Println("🗑️  Stopping Stale Request Worker...")
				return
			case <-ticker.C:
				w.processStaleRequests()
			}
		}
	}()
}

func (w *StaleRequestWorker) Stop() {
	close(w.stopChan)
}

func (w *StaleRequestWorker) processStaleRequests() {
	if w.cache != nil {
		locked, err := w.cache.AcquireLock(context.Background(), "worker:stale_request", 50*time.Minute)
		if err == nil && !locked {
			log.Println("[StaleRequestWorker] Another instance is currently processing stale requests. Skipping tick.")
			return
		}
		if locked {
			defer func() {
				_ = w.cache.ReleaseLock(context.Background(), "worker:stale_request")
			}()
		}
	}

	staleCutoff := time.Now().Add(-30 * 24 * time.Hour)

	var staleRequests []entity.ServiceRequest
	err := w.db.Preload("User").
		Where("status = ? AND created_at <= ?", "OPEN", staleCutoff).
		Find(&staleRequests).Error
	if err != nil {
		log.Printf("[StaleRequestWorker] Error fetching stale requests: %v", err)
		return
	}

	if len(staleRequests) == 0 {
		return
	}

	log.Printf("[StaleRequestWorker] Found %d stale requests to auto-close", len(staleRequests))

	// 1. Batch SQL update to mark all stale requests as EXPIRED in one query
	staleIDs := make([]uuid.UUID, 0, len(staleRequests))
	userIDs := make([]uuid.UUID, 0, len(staleRequests))
	for _, req := range staleRequests {
		staleIDs = append(staleIDs, req.ID)
		if req.UserID != uuid.Nil {
			userIDs = append(userIDs, req.UserID)
		}
	}

	if err := w.db.Model(&entity.ServiceRequest{}).
		Where("id IN ?", staleIDs).
		Update("status", "EXPIRED").Error; err != nil {
		log.Printf("[StaleRequestWorker] Failed to batch expire requests: %v", err)
		return
	}

	// 2. Batch fetch profiles to avoid N+1 queries for notification preferences
	var profiles []entity.Profile
	profileMap := make(map[uuid.UUID]entity.Profile, len(userIDs))
	if len(userIDs) > 0 {
		if err := w.db.Where("user_id IN ?", userIDs).Find(&profiles).Error; err == nil {
			for _, p := range profiles {
				profileMap[p.UserID] = p
			}
		}
	}

	// 3. Batch insert in-app notifications
	notifications := make([]entity.Notification, 0, len(staleRequests))
	for _, req := range staleRequests {
		msg := fmt.Sprintf("Your request '%s' has been automatically closed after 30 days with no activity.", req.Title)
		notifications = append(notifications, entity.Notification{
			ID:               uuid.New(),
			UserID:           req.UserID,
			NotificationType: "REQUEST_EXPIRED",
			Title:            "Service Request Expired",
			Message:          msg,
			Data:             entity.JSONMap{"request_id": req.ID.String()},
			CreatedAt:        time.Now(),
		})
	}
	if len(notifications) > 0 {
		_ = w.db.CreateInBatches(notifications, 100).Error
	}

	// 4. Asynchronous push and email delivery pool (non-blocking)
	go func(requests []entity.ServiceRequest, pMap map[uuid.UUID]entity.Profile) {
		for _, req := range requests {
			if req.User == nil {
				continue
			}
			msg := fmt.Sprintf("Your request '%s' has been automatically closed after 30 days with no activity.", req.Title)
			w.sendPush(
				req.UserID,
				"Service Request Expired",
				msg,
				map[string]string{
					"notification_type": "request_expired",
					"request_id":        req.ID.String(),
				},
			)

			prof, ok := pMap[req.UserID]
			seekerName := req.User.Email
			if ok && prof.FirstName != "" {
				seekerName = prof.FirstName
			}
			if !ok || prof.EmailNotificationsEnabled {
				_ = email.SendStaleRequestExpiredEmail(w.emailCfg, req.User.Email, seekerName, req.Title)
			}
		}
	}(staleRequests, profileMap)

	log.Printf("[StaleRequestWorker] Successfully batch-closed %d stale requests", len(staleRequests))
}

func (w *StaleRequestWorker) sendPush(userID uuid.UUID, title, body string, data map[string]string) {
	if w.fcmClient == nil || w.db == nil {
		return
	}
	if !isPushNotificationEnabled(w.db, userID) {
		return
	}
	go func() {
		var tokens []entity.DeviceToken
		if err := w.db.Where("user_id = ? AND is_active = ?", userID, true).Find(&tokens).Error; err != nil || len(tokens) == 0 {
			return
		}
		var tokenList []string
		for _, t := range tokens {
			if t.Token != "" {
				tokenList = append(tokenList, t.Token)
			}
		}
		if len(tokenList) == 0 {
			return
		}
		invalidTokens, _ := w.fcmClient.SendMulticast(context.Background(), tokenList, title, body, data)
		for _, invToken := range invalidTokens {
			w.db.Where("token = ?", invToken).Delete(&entity.DeviceToken{})
		}
	}()
}
