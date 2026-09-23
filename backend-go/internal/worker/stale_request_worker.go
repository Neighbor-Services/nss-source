package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
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
	stopChan  chan struct{}
}

func NewStaleRequestWorker(db *gorm.DB, cfg *config.Config, fcmClient fcm.Client) *StaleRequestWorker {
	return &StaleRequestWorker{
		db:        db,
		cfg:       cfg,
		fcmClient: fcmClient,
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

	for _, req := range staleRequests {
		// Mark the request as EXPIRED
		if err := w.db.Model(&req).Update("status", "EXPIRED").Error; err != nil {
			log.Printf("[StaleRequestWorker] Failed to close request %s: %v", req.ID, err)
			continue
		}

		if req.User == nil {
			continue
		}

		// In-app notification
		msg := fmt.Sprintf("Your request '%s' has been automatically closed after 30 days with no activity.", req.Title)
		_ = w.db.Create(&entity.Notification{
			ID:               uuid.New(),
			UserID:           req.UserID,
			NotificationType: "REQUEST_EXPIRED",
			Title:            "Service Request Expired",
			Message:          msg,
			Data:             entity.JSONMap{"request_id": req.ID.String()},
			CreatedAt:        time.Now(),
		})

		// Push notification
		w.sendPush(
			req.UserID,
			"Service Request Expired",
			msg,
			map[string]string{
				"notification_type": "request_expired",
				"request_id":        req.ID.String(),
			},
		)

		// Email notification
		seekerName := req.User.Email
		var profile entity.Profile
		if err := w.db.Where("user_id = ?", req.UserID).First(&profile).Error; err == nil && profile.FirstName != "" {
			seekerName = profile.FirstName
		}
		_ = email.SendStaleRequestExpiredEmail(w.emailCfg, req.User.Email, seekerName, req.Title)

		log.Printf("[StaleRequestWorker] Auto-closed stale request: %s (%s)", req.Title, req.ID)
	}
}

func (w *StaleRequestWorker) sendPush(userID uuid.UUID, title, body string, data map[string]string) {
	if w.fcmClient == nil || w.db == nil {
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
