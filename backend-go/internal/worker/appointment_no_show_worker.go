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

// AppointmentNoShowWorker checks for scheduled appointments that passed without completion or start,
// alerting parties and tracking attendance reliability.
type AppointmentNoShowWorker struct {
	db        *gorm.DB
	cfg       *config.Config
	emailCfg  *email.Config
	fcmClient fcm.Client
	stopChan  chan struct{}
}

func NewAppointmentNoShowWorker(db *gorm.DB, cfg *config.Config, fcmClient fcm.Client) *AppointmentNoShowWorker {
	return &AppointmentNoShowWorker{
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

func (w *AppointmentNoShowWorker) Start() {
	slog.Info("⏰ Starting Appointment No-Show & Attendance Worker...")

	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		recovery.SafeRun("AppointmentNoShowWorker.Initial", w.processNoShows)

		for {
			select {
			case <-w.stopChan:
				slog.Info("⏰ Stopping Appointment No-Show Worker...")
				return
			case <-ticker.C:
				recovery.SafeRun("AppointmentNoShowWorker.Tick", w.processNoShows)
			}
		}
	}()
}

func (w *AppointmentNoShowWorker) Stop() {
	close(w.stopChan)
}

func (w *AppointmentNoShowWorker) processNoShows() {
	now := time.Now().UTC()
	// Check appointments that were scheduled for > 4 hours ago and still marked SCHEDULED
	cutoff := now.Add(-4 * time.Hour)

	var pastDueApts []entity.Appointment
	err := w.db.Preload("Seeker").Preload("Provider").Preload("Seeker.Profile").Preload("Provider.Profile").
		Where("status IN ? AND appointment_date <= ? AND no_show_processed = ?", []string{"SCHEDULED", "ACCEPTED", "CONFIRMED"}, cutoff, false).
		Find(&pastDueApts).Error

	if err != nil {
		slog.Error("Error fetching past-due appointments", slog.Any("error", err))
		return
	}

	if len(pastDueApts) == 0 {
		return
	}

	slog.Info("Processing past due appointments", slog.Int("count", len(pastDueApts)))

	for _, apt := range pastDueApts {
		if err := w.db.Model(&entity.Appointment{}).Where("id = ?", apt.ID).Update("no_show_processed", true).Error; err != nil {
			slog.Error("Failed to update no_show_processed", slog.String("apt_id", apt.ID.String()), slog.Any("error", err))
			continue
		}

		aptDate := now
		if apt.AppointmentDate != nil {
			aptDate = *apt.AppointmentDate
		}

		msg := fmt.Sprintf("Your appointment for '%s' scheduled for %s has not been marked completed. Please confirm status or reschedule.", apt.Title, aptDate.Format("Jan 02 3:04 PM"))

		// 1. Notify Seeker
		if apt.Seeker != nil {
			_ = w.db.Create(&entity.Notification{
				ID:               uuid.New(),
				UserID:           apt.SeekerID,
				NotificationType: "APPOINTMENT_NO_SHOW",
				Title:            "Appointment Status Check",
				Message:          msg,
				Data:             entity.JSONMap{"appointment_id": apt.ID.String()},
				CreatedAt:        now,
			})
			w.sendPush(apt.SeekerID, "Appointment Follow-up", msg, map[string]string{
				"notification_type": "appointment_no_show",
				"appointment_id":    apt.ID.String(),
			})
			seekerName := apt.Seeker.Email
			if apt.Seeker.Profile != nil && apt.Seeker.Profile.FirstName != "" {
				seekerName = apt.Seeker.Profile.FirstName
			}
			_ = email.SendAppointmentNoShowEmail(w.emailCfg, apt.Seeker.Email, seekerName, "Client", apt.Title, aptDate)
		}

		// 2. Notify Provider
		if apt.Provider != nil {
			_ = w.db.Create(&entity.Notification{
				ID:               uuid.New(),
				UserID:           apt.ProviderID,
				NotificationType: "APPOINTMENT_NO_SHOW",
				Title:            "Appointment Status Check",
				Message:          msg,
				Data:             entity.JSONMap{"appointment_id": apt.ID.String()},
				CreatedAt:        now,
			})
			w.sendPush(apt.ProviderID, "Appointment Follow-up", msg, map[string]string{
				"notification_type": "appointment_no_show",
				"appointment_id":    apt.ID.String(),
			})
			providerName := apt.Provider.Email
			if apt.Provider.Profile != nil && apt.Provider.Profile.FirstName != "" {
				providerName = apt.Provider.Profile.FirstName
			}
			_ = email.SendAppointmentNoShowEmail(w.emailCfg, apt.Provider.Email, providerName, "Provider", apt.Title, aptDate)
		}
	}
}

func (w *AppointmentNoShowWorker) sendPush(userID uuid.UUID, title, body string, data map[string]string) {
	if w.fcmClient == nil || w.db == nil {
		return
	}
	go func() {
		recovery.SafeRun("AppointmentNoShowWorker.sendPush", func() {
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
