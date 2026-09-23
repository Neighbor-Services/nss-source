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

// PayoutProcessingWorker picks up PENDING payout requests, validates wallet balance,
// transitions them to PROCESSING, and notifies the user.
type PayoutProcessingWorker struct {
	db        *gorm.DB
	cfg       *config.Config
	emailCfg  *email.Config
	fcmClient fcm.Client
	stopChan  chan struct{}
}

func NewPayoutProcessingWorker(db *gorm.DB, cfg *config.Config, fcmClient fcm.Client) *PayoutProcessingWorker {
	return &PayoutProcessingWorker{
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

func (w *PayoutProcessingWorker) Start() {
	log.Println("💸 Starting Payout Processing Background Worker...")

	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		// Run immediately on startup
		w.processPayouts()

		for {
			select {
			case <-w.stopChan:
				log.Println("💸 Stopping Payout Processing Worker...")
				return
			case <-ticker.C:
				w.processPayouts()
			}
		}
	}()
}

func (w *PayoutProcessingWorker) Stop() {
	close(w.stopChan)
}

func (w *PayoutProcessingWorker) processPayouts() {
	var pendingPayouts []entity.PayoutRequest
	err := w.db.Preload("Wallet").Preload("Wallet.User").
		Where("status = ?", "PENDING").
		Order("created_at ASC").
		Find(&pendingPayouts).Error
	if err != nil {
		log.Printf("[PayoutWorker] Error fetching pending payouts: %v", err)
		return
	}

	if len(pendingPayouts) == 0 {
		return
	}

	log.Printf("[PayoutWorker] Processing %d pending payout request(s)", len(pendingPayouts))

	for _, payout := range pendingPayouts {
		if payout.Wallet == nil || payout.Wallet.User == nil {
			log.Printf("[PayoutWorker] Skipping payout %s: missing wallet/user", payout.ID)
			continue
		}

		wallet := payout.Wallet
		user := wallet.User

		// Verify sufficient balance
		if wallet.Balance < payout.Amount {
			notes := fmt.Sprintf("Insufficient balance (%.2f available, %.2f requested). Request held.", wallet.Balance, payout.Amount)
			w.db.Model(&payout).Updates(map[string]interface{}{
				"status":      "FAILED",
				"admin_notes": notes,
			})
			w.notifyUser(
				user.ID,
				"Payout Failed",
				fmt.Sprintf("Your payout of $%.2f could not be processed: insufficient balance.", payout.Amount),
				map[string]string{"notification_type": "payout_failed", "payout_id": payout.ID.String()},
			)
			_ = email.SendPayoutProcessingEmail(w.emailCfg, user.Email, user.Email, payout.Amount, "FAILED", notes)
			continue
		}

		// Transition to PROCESSING (payment gateway would be invoked here)
		now := time.Now()
		if err := w.db.Model(&payout).Updates(map[string]interface{}{
			"status":       "PROCESSING",
			"processed_at": now,
			"admin_notes":  "Queued for bank/Stripe transfer.",
		}).Error; err != nil {
			log.Printf("[PayoutWorker] Failed to update payout %s: %v", payout.ID, err)
			continue
		}

		// Deduct from wallet balance
		w.db.Model(wallet).Update("balance", gorm.Expr("balance - ?", payout.Amount))

		// Record wallet transaction
		w.db.Create(&entity.WalletTransaction{
			ID:              uuid.New(),
			WalletID:        wallet.ID,
			Amount:          -payout.Amount,
			TransactionType: "DEBIT",
			Description:     fmt.Sprintf("Payout withdrawal — request #%s", payout.ID.String()[:8]),
			Status:          "COMPLETED",
			ReferenceID:     payout.ID.String(),
			CreatedAt:       time.Now(),
		})

		// In-app notification
		_ = w.db.Create(&entity.Notification{
			ID:               uuid.New(),
			UserID:           user.ID,
			NotificationType: "PAYOUT_PROCESSING",
			Title:            "Payout in Progress",
			Message:          fmt.Sprintf("Your payout of $%.2f is now being processed. Funds typically arrive within 2–5 business days.", payout.Amount),
			Data:             entity.JSONMap{"payout_id": payout.ID.String(), "amount": fmt.Sprintf("%.2f", payout.Amount)},
			CreatedAt:        time.Now(),
		})

		w.notifyUser(
			user.ID,
			"Payout in Progress 💰",
			fmt.Sprintf("Your $%.2f payout is being processed. Expect funds in 2–5 business days.", payout.Amount),
			map[string]string{"notification_type": "payout_processing", "payout_id": payout.ID.String()},
		)
		_ = email.SendPayoutProcessingEmail(w.emailCfg, user.Email, user.Email, payout.Amount, "PROCESSING", "")

		log.Printf("[PayoutWorker] Payout %s transitioned to PROCESSING ($%.2f for user %s)", payout.ID, payout.Amount, user.ID)
	}
}

func (w *PayoutProcessingWorker) notifyUser(userID uuid.UUID, title, body string, data map[string]string) {
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
