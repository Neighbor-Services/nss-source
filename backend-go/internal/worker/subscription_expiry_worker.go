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

// SubscriptionExpiryWorker sends renewal reminders 3 days before expiry
// and deactivates subscriptions past their NextPayment date.
type SubscriptionExpiryWorker struct {
	db        *gorm.DB
	cfg       *config.Config
	emailCfg  *email.Config
	fcmClient fcm.Client
	stopChan  chan struct{}
}

func NewSubscriptionExpiryWorker(db *gorm.DB, cfg *config.Config, fcmClient fcm.Client) *SubscriptionExpiryWorker {
	return &SubscriptionExpiryWorker{
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

func (w *SubscriptionExpiryWorker) Start() {
	log.Println("📅 Starting Subscription Expiry Background Worker...")

	go func() {
		// Run once a day
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		// Run immediately on startup
		w.processSubscriptions()

		for {
			select {
			case <-w.stopChan:
				log.Println("📅 Stopping Subscription Expiry Worker...")
				return
			case <-ticker.C:
				w.processSubscriptions()
			}
		}
	}()
}

func (w *SubscriptionExpiryWorker) Stop() {
	close(w.stopChan)
}

func (w *SubscriptionExpiryWorker) processSubscriptions() {
	now := time.Now()
	in3Days := now.Add(3 * 24 * time.Hour)

	// 1. Send renewal reminders for subscriptions expiring in ≤ 3 days
	var expiringSoon []entity.Subscription
	err := w.db.Preload("User").Preload("Plan").
		Where("is_active = ? AND next_payment IS NOT NULL AND next_payment <= ? AND next_payment > ?", true, in3Days, now).
		Find(&expiringSoon).Error
	if err != nil {
		log.Printf("[SubscriptionWorker] Error fetching expiring subscriptions: %v", err)
	}

	for _, sub := range expiringSoon {
		if sub.User == nil {
			continue
		}

		refID := fmt.Sprintf("sub_renewal_%s", sub.ID.String())
		var count int64
		w.db.Model(&entity.EmailCampaignLog{}).
			Where("campaign_type = ? AND reference_id = ?", "SUB_RENEWAL_REMINDER", refID).
			Count(&count)
		if count > 0 {
			continue
		}

		planName := "your subscription plan"
		if sub.Plan != nil {
			planName = sub.Plan.Name
		}
		expiryStr := ""
		if sub.NextPayment != nil {
			expiryStr = sub.NextPayment.Format("Jan 02, 2006")
		}

		msg := fmt.Sprintf("Your '%s' plan expires on %s. Renew now to keep your benefits!", planName, expiryStr)

		_ = w.db.Create(&entity.Notification{
			ID:               uuid.New(),
			UserID:           sub.UserID,
			NotificationType: "SUB_EXPIRY_REMINDER",
			Title:            "Subscription Expiring Soon ⚠️",
			Message:          msg,
			Data:             entity.JSONMap{"subscription_id": sub.ID.String()},
			CreatedAt:        time.Now(),
		})

		w.sendPush(sub.UserID, "Subscription Expiring Soon ⚠️", msg, map[string]string{
			"notification_type": "sub_expiry_reminder",
			"subscription_id":   sub.ID.String(),
		})

		userName := sub.User.Email
		planN := planName
		_ = email.SendSubscriptionExpiryReminderEmail(w.emailCfg, sub.User.Email, userName, planN, expiryStr)

		w.db.Create(&entity.EmailCampaignLog{
			ID:           uuid.New(),
			UserID:       sub.UserID,
			TargetEmail:  sub.User.Email,
			CampaignType: "SUB_RENEWAL_REMINDER",
			ReferenceID:  &refID,
			SentAt:       time.Now(),
		})

		log.Printf("[SubscriptionWorker] Renewal reminder sent for subscription %s (user %s)", sub.ID, sub.UserID)
	}

	// 2. Deactivate expired subscriptions
	var expired []entity.Subscription
	err = w.db.Preload("User").Preload("Plan").
		Where("is_active = ? AND next_payment IS NOT NULL AND next_payment < ?", true, now).
		Find(&expired).Error
	if err != nil {
		log.Printf("[SubscriptionWorker] Error fetching expired subscriptions: %v", err)
		return
	}

	for _, sub := range expired {
		if err := w.db.Model(&sub).Update("is_active", false).Error; err != nil {
			log.Printf("[SubscriptionWorker] Failed to deactivate subscription %s: %v", sub.ID, err)
			continue
		}

		if sub.User == nil {
			continue
		}

		planName := "your subscription"
		if sub.Plan != nil {
			planName = sub.Plan.Name
		}

		msg := fmt.Sprintf("Your '%s' plan has expired. Resubscribe to restore premium features.", planName)

		_ = w.db.Create(&entity.Notification{
			ID:               uuid.New(),
			UserID:           sub.UserID,
			NotificationType: "SUB_EXPIRED",
			Title:            "Subscription Expired",
			Message:          msg,
			Data:             entity.JSONMap{"subscription_id": sub.ID.String()},
			CreatedAt:        time.Now(),
		})

		w.sendPush(sub.UserID, "Subscription Expired", msg, map[string]string{
			"notification_type": "sub_expired",
			"subscription_id":   sub.ID.String(),
		})

		userName := sub.User.Email
		planN := planName
		_ = email.SendSubscriptionExpiredEmail(w.emailCfg, sub.User.Email, userName, planN)

		log.Printf("[SubscriptionWorker] Deactivated expired subscription %s for user %s", sub.ID, sub.UserID)
	}
}

func (w *SubscriptionExpiryWorker) sendPush(userID uuid.UUID, title, body string, data map[string]string) {
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
