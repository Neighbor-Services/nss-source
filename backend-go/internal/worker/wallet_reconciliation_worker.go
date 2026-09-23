package worker

import (
	"fmt"
	"log"
	"math"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/pkg/email"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WalletReconciliationWorker checks that each wallet's balance matches
// the sum of its transactions. Flags discrepancies as fraud risk alerts.
type WalletReconciliationWorker struct {
	db       *gorm.DB
	cfg      *config.Config
	emailCfg *email.Config
	stopChan chan struct{}
}

func NewWalletReconciliationWorker(db *gorm.DB, cfg *config.Config) *WalletReconciliationWorker {
	return &WalletReconciliationWorker{
		db:  db,
		cfg: cfg,
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

func (w *WalletReconciliationWorker) Start() {
	log.Println("🏦 Starting Wallet Reconciliation Background Worker...")

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		w.reconcile()

		for {
			select {
			case <-w.stopChan:
				log.Println("🏦 Stopping Wallet Reconciliation Worker...")
				return
			case <-ticker.C:
				w.reconcile()
			}
		}
	}()
}

func (w *WalletReconciliationWorker) Stop() {
	close(w.stopChan)
}

func (w *WalletReconciliationWorker) reconcile() {
	var wallets []entity.Wallet
	if err := w.db.Find(&wallets).Error; err != nil {
		log.Printf("[WalletReconciliation] Error fetching wallets: %v", err)
		return
	}

	discrepancies := 0

	for _, wallet := range wallets {
		// Sum all completed transactions for this wallet
		var txSum float64
		w.db.Model(&entity.WalletTransaction{}).
			Where("wallet_id = ? AND status = ?", wallet.ID, "COMPLETED").
			Select("COALESCE(SUM(amount), 0)").
			Scan(&txSum)

		diff := math.Abs(wallet.Balance - txSum)

		// Tolerance: ignore differences under 1 cent (floating point noise)
		if diff <= 0.01 {
			continue
		}

		discrepancies++
		log.Printf("[WalletReconciliation] ⚠️  Discrepancy for wallet %s (user %s): balance=%.2f, tx_sum=%.2f, diff=%.2f",
			wallet.ID, wallet.UserID, wallet.Balance, txSum, diff)

		// Check if we've already flagged this wallet today
		today := time.Now().Truncate(24 * time.Hour)
		var existing int64
		w.db.Model(&entity.FraudRiskAlert{}).
			Where("user_id = ? AND created_at >= ? AND status = 'OPEN'", wallet.UserID, today).
			Where("JSON_CONTAINS(flags, ?)", `"WALLET_BALANCE_MISMATCH"`).
			Count(&existing)
		if existing > 0 {
			continue
		}

		riskScore := 40
		riskLevel := "MEDIUM"
		if diff > 100 {
			riskScore = 70
			riskLevel = "HIGH"
		}

		w.db.Create(&entity.FraudRiskAlert{
			ID:        uuid.New(),
			UserID:    wallet.UserID,
			RiskScore: riskScore,
			RiskLevel: riskLevel,
			Flags:     entity.JSONSlice{"WALLET_BALANCE_MISMATCH"},
			Details: entity.JSONMap{
				"wallet_id":    wallet.ID.String(),
				"book_balance": fmt.Sprintf("%.2f", wallet.Balance),
				"tx_sum":       fmt.Sprintf("%.2f", txSum),
				"difference":   fmt.Sprintf("%.2f", diff),
				"detected_at":  time.Now().Format(time.RFC3339),
			},
			Status:    "OPEN",
			CreatedAt: time.Now(),
		})

		// Alert admin via email
		_ = email.SendWalletDiscrepancyAlertEmail(
			w.emailCfg,
			w.cfg.AdminEmail,
			"Admin",
			wallet.UserID.String(),
			wallet.ID.String(),
			wallet.Balance,
			txSum,
			diff,
		)
	}

	if discrepancies > 0 {
		log.Printf("[WalletReconciliation] Cycle complete — %d discrepancy/discrepancies flagged", discrepancies)
	} else {
		log.Printf("[WalletReconciliation] Cycle complete — all %d wallet(s) balanced ✅", len(wallets))
	}
}
