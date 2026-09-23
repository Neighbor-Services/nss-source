package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/pkg/email"
	"backend-go/pkg/fcm"
	"backend-go/pkg/recovery"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReviewReminderWorker gently prompts seekers to review completed services 2 hours post completion.
type ReviewReminderWorker struct {
	db        *gorm.DB
	cfg       *config.Config
	emailCfg  *email.Config
	fcmClient fcm.Client
	stopChan  chan struct{}
}

func NewReviewReminderWorker(db *gorm.DB, cfg *config.Config, fcmClient fcm.Client) *ReviewReminderWorker {
	return &ReviewReminderWorker{
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

func (w *ReviewReminderWorker) Start() {
	slog.Info("⭐ Starting Review Reminder Background Worker...")

	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		recovery.SafeRun("ReviewReminderWorker.Initial", w.processReviewReminders)

		for {
			select {
			case <-w.stopChan:
				slog.Info("⭐ Stopping Review Reminder Worker...")
				return
			case <-ticker.C:
				recovery.SafeRun("ReviewReminderWorker.Tick", w.processReviewReminders)
			}
		}
	}()
}

func (w *ReviewReminderWorker) Stop() {
	close(w.stopChan)
}

func (w *ReviewReminderWorker) processReviewReminders() {
	now := time.Now().UTC()
	twoHoursAgo := now.Add(-2 * time.Hour)

	var completedApts []entity.Appointment
	err := w.db.Preload("Seeker").Preload("Provider").Preload("Seeker.Profile").Preload("Provider.Profile").
		Where("status = ? AND updated_at <= ? AND review_reminder_sent = ?", "COMPLETED", twoHoursAgo, false).
		Find(&completedApts).Error

	if err != nil {
		slog.Error("Error fetching completed appointments for review reminder", slog.Any("error", err))
		return
	}

	if len(completedApts) == 0 {
		return
	}

	slog.Info("Processing review reminders", slog.Int("count", len(completedApts)))

	for _, apt := range completedApts {
		// Mark reminder sent to avoid duplicate processing
		_ = w.db.Model(&entity.Appointment{}).Where("id = ?", apt.ID).Update("review_reminder_sent", true)

		// Check if seeker already reviewed this provider recently
		var existingReviewCount int64
		w.db.Model(&entity.Review{}).
			Where("provider_id = ? AND reviewer_id = ? AND created_at >= ?", apt.ProviderID, apt.SeekerID, apt.UpdatedAt.Add(-24*time.Hour)).
			Count(&existingReviewCount)

		if existingReviewCount > 0 {
			continue
		}

		seekerName := "there"
		if apt.Seeker != nil && apt.Seeker.Profile != nil && apt.Seeker.Profile.FirstName != "" {
			seekerName = apt.Seeker.Profile.FirstName
		}

		providerName := "your service provider"
		if apt.Provider != nil && apt.Provider.Profile != nil && apt.Provider.Profile.FirstName != "" {
			providerName = apt.Provider.Profile.FirstName
		}

		pushTitle := "How was your service?"
		pushMsg := fmt.Sprintf("Your service with %s is complete. Tap to leave a rating & review!", providerName)

		// 1. Create In-App Notification
		_ = w.db.Create(&entity.Notification{
			ID:               uuid.New(),
			UserID:           apt.SeekerID,
			NotificationType: "REVIEW_REMINDER",
			Title:            pushTitle,
			Message:          pushMsg,
			Data: entity.JSONMap{
				"appointment_id": apt.ID.String(),
				"provider_id":    apt.ProviderID.String(),
			},
			CreatedAt: now,
		})

		// 2. FCM Push
		w.sendPush(apt.SeekerID, pushTitle, pushMsg, map[string]string{
			"notification_type": "review_reminder",
			"appointment_id":    apt.ID.String(),
			"provider_id":       apt.ProviderID.String(),
		})

		// 3. Email
		if apt.Seeker != nil && apt.Seeker.Email != "" {
			_ = email.SendReviewReminderEmail(w.emailCfg, apt.Seeker.Email, seekerName, providerName, apt.Title)
		}
	}
}

func (w *ReviewReminderWorker) sendPush(userID uuid.UUID, title, body string, data map[string]string) {
	if w.fcmClient == nil || w.db == nil {
		return
	}
	go func() {
		recovery.SafeRun("ReviewReminderWorker.sendPush", func() {
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
		})
	}()
}
