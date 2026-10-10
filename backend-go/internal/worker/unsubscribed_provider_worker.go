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

// UnsubscribedProviderWorker periodically checks for registered service providers
// who do not have an active subscription, sending them push notifications, in-app
// alerts, and promotional emails to encourage subscribing and boosting their client patronage.
type UnsubscribedProviderWorker struct {
	db        *gorm.DB
	cfg       *config.Config
	emailCfg  *email.Config
	fcmClient fcm.Client
	stopChan  chan struct{}
}

func NewUnsubscribedProviderWorker(db *gorm.DB, cfg *config.Config, fcmClient fcm.Client) *UnsubscribedProviderWorker {
	return &UnsubscribedProviderWorker{
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

func (w *UnsubscribedProviderWorker) Start() {
	log.Println("🚀 Starting Unsubscribed Provider Engagement Worker...")

	go func() {
		// Initial delay before first run
		time.Sleep(1 * time.Minute)
		w.processUnsubscribedProviders()

		// Run every 24 hours
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-w.stopChan:
				log.Println("🛑 Stopping Unsubscribed Provider Worker...")
				return
			case <-ticker.C:
				w.processUnsubscribedProviders()
			}
		}
	}()
}

func (w *UnsubscribedProviderWorker) Stop() {
	close(w.stopChan)
}

func (w *UnsubscribedProviderWorker) processUnsubscribedProviders() {
	now := time.Now()
	registeredBefore := now.Add(-24 * time.Hour) // Must have registered at least 24h ago
	coolDownDate := now.Add(-5 * 24 * time.Hour)  // Don't send more frequently than once every 5 days

	// Query providers who don't have an active subscription
	var providers []entity.Profile
	err := w.db.Preload("User").
		Joins("JOIN accounts_user ON accounts_user.id = accounts_profile.user_id").
		Where("accounts_profile.user_type = ? AND accounts_user.is_active = ? AND accounts_profile.created_at <= ?", "PROVIDER", true, registeredBefore).
		Where("accounts_profile.user_id NOT IN (SELECT user_id FROM payments_subscription WHERE is_active = true)").
		Limit(100).
		Find(&providers).Error

	if err != nil {
		log.Printf("[UnsubscribedProviderWorker] Error fetching unsubscribed providers: %v", err)
		return
	}

	if len(providers) == 0 {
		log.Println("[UnsubscribedProviderWorker] No unsubscribed providers found requiring reminders.")
		return
	}

	log.Printf("[UnsubscribedProviderWorker] Processing %d unsubscribed provider(s)...", len(providers))

	for _, p := range providers {
		if p.User == nil || p.User.Email == "" {
			continue
		}

		// Check cooldown: ensure we haven't sent a reminder in the last 5 days
		var recentCampaignCount int64
		w.db.Model(&entity.EmailCampaignLog{}).
			Where("user_id = ? AND campaign_type = ? AND sent_at >= ?", p.UserID, "UNSUBSCRIBED_PROVIDER_REMINDER", coolDownDate).
			Count(&recentCampaignCount)

		if recentCampaignCount > 0 {
			continue
		}

		displayName := p.FirstName
		if displayName == "" {
			displayName = "Service Provider"
		}

		title := "🚀 Boost Your Client Bookings & Patronage!"
		msg := fmt.Sprintf("Hi %s, local clients are looking for services in your area. Subscribe to activate Live Request Radar, submit direct proposals, and boost your monthly earnings!", displayName)

		// 1. In-App Notification
		_ = w.db.Create(&entity.Notification{
			ID:               uuid.New(),
			UserID:           p.UserID,
			NotificationType: "PROVIDER_SUBSCRIPTION_PROMO",
			Title:            title,
			Message:          msg,
			Data: entity.JSONMap{
				"action":    "open_subscription",
				"route":     "/subscription",
				"user_type": "PROVIDER",
			},
			CreatedAt: time.Now(),
		})

		// 2. Push Notification via FCM
		w.sendPush(p.UserID, title, msg, map[string]string{
			"notification_type": "provider_subscription_promo",
			"route":             "/subscription",
		})

		// 3. Marketing Email
		serviceName := p.Service
		_ = email.SendUnsubscribedProviderReminderEmail(w.emailCfg, p.User.Email, displayName, serviceName)

		// 4. Log Campaign Send to prevent spam
		refID := fmt.Sprintf("unsub_prov_%s_%s", p.UserID.String(), now.Format("20060102"))
		w.db.Create(&entity.EmailCampaignLog{
			ID:           uuid.New(),
			UserID:       p.UserID,
			TargetEmail:  p.User.Email,
			CampaignType: "UNSUBSCRIBED_PROVIDER_REMINDER",
			ReferenceID:  &refID,
			SentAt:       time.Now(),
		})

		log.Printf("[UnsubscribedProviderWorker] Sent subscription reminder to provider %s (%s)", displayName, p.User.Email)
	}
}

func (w *UnsubscribedProviderWorker) sendPush(userID uuid.UUID, title, body string, data map[string]string) {
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
