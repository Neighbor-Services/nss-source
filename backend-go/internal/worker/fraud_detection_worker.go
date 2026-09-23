package worker

import (
	"fmt"
	"log"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/pkg/recovery"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FraudDetectionWorker scans recent transactions for anomalies and flags
// suspicious activity as FraudRiskAlert records for admin review.
type FraudDetectionWorker struct {
	db       *gorm.DB
	cfg      *config.Config
	stopChan chan struct{}
}

func NewFraudDetectionWorker(db *gorm.DB, cfg *config.Config) *FraudDetectionWorker {
	return &FraudDetectionWorker{
		db:       db,
		cfg:      cfg,
		stopChan: make(chan struct{}),
	}
}

func (w *FraudDetectionWorker) Start() {
	log.Println("🕵️  Starting Fraud Detection Background Worker...")

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		recovery.SafeRun("FraudDetectionWorker.Initial", w.scan)

		for {
			select {
			case <-w.stopChan:
				log.Println("🕵️  Stopping Fraud Detection Worker...")
				return
			case <-ticker.C:
				recovery.SafeRun("FraudDetectionWorker.Tick", w.scan)
			}
		}
	}()
}

func (w *FraudDetectionWorker) Stop() {
	close(w.stopChan)
}

func (w *FraudDetectionWorker) scan() {
	window := time.Now().Add(-10 * time.Minute) // scan the last 10-minute window
	day := time.Now().Add(-24 * time.Hour)

	// ── Rule 1: Large single transaction (> $500) ──────────────────────────
	type txRow struct {
		WalletID uuid.UUID
		Amount   float64
	}

	var largeTxns []txRow
	w.db.Model(&entity.WalletTransaction{}).
		Select("wallet_id, amount").
		Where("created_at >= ? AND amount > 500 AND transaction_type = 'DEBIT'", window).
		Scan(&largeTxns)

	for _, tx := range largeTxns {
		var wallet entity.Wallet
		if err := w.db.First(&wallet, "id = ?", tx.WalletID).Error; err != nil {
			continue
		}
		w.createAlertIfNew(wallet.UserID, "HIGH", 65,
			entity.JSONSlice{"LARGE_TRANSACTION"},
			entity.JSONMap{
				"rule":      "large_single_transaction",
				"amount":    fmt.Sprintf("%.2f", tx.Amount),
				"wallet_id": tx.WalletID.String(),
			},
		)
	}

	// ── Rule 2: High velocity — more than 10 transactions in 10 minutes ───
	type velocityRow struct {
		WalletID uuid.UUID
		TxCount  int
	}

	var highVelocity []velocityRow
	w.db.Model(&entity.WalletTransaction{}).
		Select("wallet_id, COUNT(*) as tx_count").
		Where("created_at >= ?", window).
		Group("wallet_id").
		Having("COUNT(*) > 10").
		Scan(&highVelocity)

	for _, row := range highVelocity {
		var wallet entity.Wallet
		if err := w.db.First(&wallet, "id = ?", row.WalletID).Error; err != nil {
			continue
		}
		w.createAlertIfNew(wallet.UserID, "HIGH", 70,
			entity.JSONSlice{"HIGH_TX_VELOCITY"},
			entity.JSONMap{
				"rule":         "high_transaction_velocity",
				"tx_count":     fmt.Sprintf("%d", row.TxCount),
				"window_mins":  "10",
				"wallet_id":    row.WalletID.String(),
			},
		)
	}

	// ── Rule 3: Multiple payout requests in 24h (> 3) ────────────────────
	type payoutRow struct {
		WalletID uuid.UUID
		Count    int
	}

	var multiPayouts []payoutRow
	w.db.Model(&entity.PayoutRequest{}).
		Select("wallet_id, COUNT(*) as count").
		Where("created_at >= ?", day).
		Group("wallet_id").
		Having("COUNT(*) > 3").
		Scan(&multiPayouts)

	for _, row := range multiPayouts {
		var wallet entity.Wallet
		if err := w.db.First(&wallet, "id = ?", row.WalletID).Error; err != nil {
			continue
		}
		w.createAlertIfNew(wallet.UserID, "MEDIUM", 50,
			entity.JSONSlice{"MULTIPLE_PAYOUT_REQUESTS"},
			entity.JSONMap{
				"rule":        "multiple_payouts_24h",
				"payout_count": fmt.Sprintf("%d", row.Count),
				"wallet_id":   row.WalletID.String(),
			},
		)
	}
}

// createAlertIfNew inserts a FraudRiskAlert only if no open alert with the same
// primary flag already exists for this user today.
func (w *FraudDetectionWorker) createAlertIfNew(
	userID uuid.UUID,
	riskLevel string,
	riskScore int,
	flags entity.JSONSlice,
	details entity.JSONMap,
) {
	today := time.Now().Truncate(24 * time.Hour)
	primaryFlag := ""
	if len(flags) > 0 {
		primaryFlag = flags[0]
	}

	var count int64
	w.db.Model(&entity.FraudRiskAlert{}).
		Where("user_id = ? AND status = 'OPEN' AND created_at >= ?", userID, today).
		Where("flags LIKE ?", fmt.Sprintf("%%%s%%", primaryFlag)).
		Count(&count)

	if count > 0 {
		return
	}

	if err := w.db.Create(&entity.FraudRiskAlert{
		ID:        uuid.New(),
		UserID:    userID,
		RiskScore: riskScore,
		RiskLevel: riskLevel,
		Flags:     flags,
		Details:   details,
		Status:    "OPEN",
		CreatedAt: time.Now(),
	}).Error; err != nil {
		log.Printf("[FraudDetection] Failed to create alert for user %s: %v", userID, err)
		return
	}

	log.Printf("[FraudDetection] 🚨 Alert created — user %s | level=%s | flags=%v", userID, riskLevel, flags)
}
