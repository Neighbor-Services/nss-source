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

// InactiveUserWorker re-engages users who haven't been active in 14 or 30 days.
type InactiveUserWorker struct {
	db        *gorm.DB
	cfg       *config.Config
	emailCfg  *email.Config
	fcmClient fcm.Client
	stopChan  chan struct{}
}

func NewInactiveUserWorker(db *gorm.DB, cfg *config.Config, fcmClient fcm.Client) *InactiveUserWorker {
	return &InactiveUserWorker{
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

func (w *InactiveUserWorker) Start() {
	log.Println("💤 Starting Inactive User Re-engagement Background Worker...")

	go func() {
		// Run once a day at startup then every 24h
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		w.processInactiveUsers()

		for {
			select {
			case <-w.stopChan:
				log.Println("💤 Stopping Inactive User Worker...")
				return
			case <-ticker.C:
				w.processInactiveUsers()
			}
		}
	}()
}

func (w *InactiveUserWorker) Stop() {
	close(w.stopChan)
}

type inactiveSegment struct {
	cutoff       time.Time
	campaignType string
	cooldown     time.Duration
	title        string
	message      string
}

func (w *InactiveUserWorker) processInactiveUsers() {
	now := time.Now()

	segments := []inactiveSegment{
		{
			cutoff:       now.Add(-14 * 24 * time.Hour),
			campaignType: "INACTIVE_14D",
			cooldown:     20 * 24 * time.Hour,
			title:        "We miss you! 👋",
			message:      "It's been a while. New jobs are waiting nearby — come check them out!",
		},
		{
			cutoff:       now.Add(-30 * 24 * time.Hour),
			campaignType: "INACTIVE_30D",
			cooldown:     35 * 24 * time.Hour,
			title:        "Still there? 🤔",
			message:      "Providers in your area are booking fast. Don't miss out on great opportunities!",
		},
	}

	for _, seg := range segments {
		var users []entity.User
		err := w.db.
			Where("updated_at <= ? AND is_active = ?", seg.cutoff, true).
			Limit(100).
			Find(&users).Error
		if err != nil {
			log.Printf("[InactiveUserWorker] Error fetching users for %s: %v", seg.campaignType, err)
			continue
		}

		cooldownThreshold := now.Add(-seg.cooldown)

		for _, u := range users {
			// Skip if already notified within cooldown
			var count int64
			w.db.Model(&entity.EmailCampaignLog{}).
				Where("user_id = ? AND campaign_type = ? AND sent_at >= ?", u.ID, seg.campaignType, cooldownThreshold).
				Count(&count)
			if count > 0 {
				continue
			}

			// Push notification
			w.sendPush(u.ID, seg.title, seg.message, map[string]string{
				"notification_type": "reengagement",
				"campaign":          seg.campaignType,
			})

			// Email re-engagement
			userName := u.Email
			_ = email.SendInactiveUserReengagementEmail(w.emailCfg, u.Email, userName, seg.campaignType)

			// In-app notification
			_ = w.db.Create(&entity.Notification{
				ID:               uuid.New(),
				UserID:           u.ID,
				NotificationType: "REENGAGEMENT",
				Title:            seg.title,
				Message:          seg.message,
				Data:             entity.JSONMap{"campaign": seg.campaignType},
				CreatedAt:        time.Now(),
			})

			// Log campaign send
			refID := fmt.Sprintf("%s_%s", seg.campaignType, u.ID.String())
			w.db.Create(&entity.EmailCampaignLog{
				ID:           uuid.New(),
				UserID:       u.ID,
				TargetEmail:  u.Email,
				CampaignType: seg.campaignType,
				ReferenceID:  &refID,
				SentAt:       time.Now(),
			})

			log.Printf("[InactiveUserWorker] Re-engagement push sent (%s) to user %s", seg.campaignType, u.ID)
		}
	}
}

func (w *InactiveUserWorker) sendPush(userID uuid.UUID, title, body string, data map[string]string) {
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
		for _, inv := range invalidTokens {
			w.db.Where("token = ?", inv).Delete(&entity.DeviceToken{})
		}
	}()
}
