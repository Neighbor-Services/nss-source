package worker

import (
	"log/slog"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/pkg/recovery"

	"gorm.io/gorm"
)

// ExpiredOTPCleanupWorker routinely purges stale and expired OTP records and temporary tokens.
type ExpiredOTPCleanupWorker struct {
	db       *gorm.DB
	cfg      *config.Config
	stopChan chan struct{}
}

func NewExpiredOTPCleanupWorker(db *gorm.DB, cfg *config.Config) *ExpiredOTPCleanupWorker {
	return &ExpiredOTPCleanupWorker{
		db:       db,
		cfg:      cfg,
		stopChan: make(chan struct{}),
	}
}

func (w *ExpiredOTPCleanupWorker) Start() {
	slog.Info("🧹 Starting Expired OTP Cleanup Background Worker...")

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		recovery.SafeRun("ExpiredOTPCleanupWorker.Initial", w.cleanupExpiredOTPs)

		for {
			select {
			case <-w.stopChan:
				slog.Info("🧹 Stopping Expired OTP Cleanup Worker...")
				return
			case <-ticker.C:
				recovery.SafeRun("ExpiredOTPCleanupWorker.Tick", w.cleanupExpiredOTPs)
			}
		}
	}()
}

func (w *ExpiredOTPCleanupWorker) Stop() {
	close(w.stopChan)
}

func (w *ExpiredOTPCleanupWorker) cleanupExpiredOTPs() {
	now := time.Now().UTC()

	// 1. Delete expired records from OTPVerification table
	res := w.db.Where("expires_at < ?", now).Delete(&entity.OTPVerification{})
	if res.Error != nil {
		slog.Error("Error cleaning expired OTP verification records", slog.Any("error", res.Error))
	} else if res.RowsAffected > 0 {
		slog.Info("Cleaned up expired OTP verification entries", slog.Int64("deleted_count", res.RowsAffected))
	}

	// 2. Clear expired OTP codes in User table
	userRes := w.db.Model(&entity.User{}).
		Where("otp_expiry IS NOT NULL AND otp_expiry < ?", now).
		Updates(map[string]interface{}{
			"otp_code":   "",
			"otp_expiry": nil,
		})
	if userRes.Error != nil {
		slog.Error("Error clearing user expired OTPs", slog.Any("error", userRes.Error))
	} else if userRes.RowsAffected > 0 {
		slog.Info("Cleared expired user OTP codes", slog.Int64("cleared_count", userRes.RowsAffected))
	}
}
