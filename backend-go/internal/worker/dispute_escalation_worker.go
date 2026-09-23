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
	"backend-go/pkg/recovery"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DisputeEscalationWorker escalates OPEN disputes that have been idle for 48+ hours.
type DisputeEscalationWorker struct {
	db        *gorm.DB
	cfg       *config.Config
	emailCfg  *email.Config
	fcmClient fcm.Client
	stopChan  chan struct{}
}

func NewDisputeEscalationWorker(db *gorm.DB, cfg *config.Config, fcmClient fcm.Client) *DisputeEscalationWorker {
	return &DisputeEscalationWorker{
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

func (w *DisputeEscalationWorker) Start() {
	log.Println("⚖️  Starting Dispute Escalation Background Worker...")

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		recovery.SafeRun("DisputeEscalationWorker.Initial", w.processDisputeEscalations)

		for {
			select {
			case <-w.stopChan:
				log.Println("⚖️  Stopping Dispute Escalation Worker...")
				return
			case <-ticker.C:
				recovery.SafeRun("DisputeEscalationWorker.Tick", w.processDisputeEscalations)
			}
		}
	}()
}

func (w *DisputeEscalationWorker) Stop() {
	close(w.stopChan)
}

func (w *DisputeEscalationWorker) processDisputeEscalations() {
	idleThreshold := time.Now().Add(-48 * time.Hour)

	var staleDisputes []entity.Dispute
	err := w.db.Preload("RaisedBy").Preload("Defendant").
		Where("status = ? AND updated_at <= ?", "OPEN", idleThreshold).
		Find(&staleDisputes).Error
	if err != nil {
		log.Printf("[DisputeWorker] Error fetching stale disputes: %v", err)
		return
	}

	if len(staleDisputes) == 0 {
		return
	}

	log.Printf("[DisputeWorker] Escalating %d idle dispute(s)", len(staleDisputes))

	for _, dispute := range staleDisputes {
		// Escalate status
		notes := fmt.Sprintf("Auto-escalated after 48h inactivity on %s.", time.Now().Format("Jan 02, 2006 15:04"))
		if err := w.db.Model(&dispute).Updates(map[string]interface{}{
			"status":           "ESCALATED",
			"resolution_notes": notes,
		}).Error; err != nil {
			log.Printf("[DisputeWorker] Failed to escalate dispute %s: %v", dispute.ID, err)
			continue
		}

		// Create a FraudRiskAlert so admins can see it in the dashboard
		w.db.Create(&entity.FraudRiskAlert{
			ID:        uuid.New(),
			UserID:    dispute.RaisedByID,
			RiskScore: 30,
			RiskLevel: "MEDIUM",
			Flags:     entity.JSONSlice{"UNRESOLVED_DISPUTE_48H"},
			Details: entity.JSONMap{
				"dispute_id": dispute.ID.String(),
				"reason":     dispute.Reason,
			},
			Status:    "OPEN",
			CreatedAt: time.Now(),
		})

		disputeMsg := fmt.Sprintf("Dispute #%s regarding '%s' has been escalated to our support team after 48 hours.", dispute.ID.String()[:8], dispute.Reason)

		// Notify raiser
		if dispute.RaisedBy != nil {
			_ = w.db.Create(&entity.Notification{
				ID:               uuid.New(),
				UserID:           dispute.RaisedByID,
				NotificationType: "DISPUTE_ESCALATED",
				Title:            "Dispute Escalated to Support",
				Message:          disputeMsg,
				Data:             entity.JSONMap{"dispute_id": dispute.ID.String()},
				CreatedAt:        time.Now(),
			})
			w.sendPush(dispute.RaisedByID, "Dispute Escalated", disputeMsg, map[string]string{
				"notification_type": "dispute_escalated",
				"dispute_id":        dispute.ID.String(),
			})
			_ = email.SendDisputeEscalatedEmail(w.emailCfg, dispute.RaisedBy.Email, dispute.RaisedBy.Email, dispute.ID.String()[:8], dispute.Reason)
		}

		// Notify defendant
		if dispute.DefendantID != nil && dispute.Defendant != nil {
			_ = w.db.Create(&entity.Notification{
				ID:               uuid.New(),
				UserID:           *dispute.DefendantID,
				NotificationType: "DISPUTE_ESCALATED",
				Title:            "Dispute Escalated to Support",
				Message:          disputeMsg,
				Data:             entity.JSONMap{"dispute_id": dispute.ID.String()},
				CreatedAt:        time.Now(),
			})
			w.sendPush(*dispute.DefendantID, "Dispute Escalated", disputeMsg, map[string]string{
				"notification_type": "dispute_escalated",
				"dispute_id":        dispute.ID.String(),
			})
			_ = email.SendDisputeEscalatedEmail(w.emailCfg, dispute.Defendant.Email, dispute.Defendant.Email, dispute.ID.String()[:8], dispute.Reason)
		}

		log.Printf("[DisputeWorker] Escalated dispute %s", dispute.ID)
	}
}

func (w *DisputeEscalationWorker) sendPush(userID uuid.UUID, title, body string, data map[string]string) {
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
