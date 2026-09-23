package worker

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/pkg/email"
	"backend-go/pkg/fcm"
	"backend-go/pkg/recovery"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProviderRankingWorker recalculates dynamic provider ranking scores (NeighborScore)
// based on completion rate, average rating, response time, verified identity, and review depth.
type ProviderRankingWorker struct {
	db        *gorm.DB
	cfg       *config.Config
	emailCfg  *email.Config
	fcmClient fcm.Client
	stopChan  chan struct{}
}

func NewProviderRankingWorker(db *gorm.DB, cfg *config.Config, fcmClient fcm.Client) *ProviderRankingWorker {
	return &ProviderRankingWorker{
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

func (w *ProviderRankingWorker) Start() {
	slog.Info("🏆 Starting Provider Ranking & Dynamic Scoring Background Worker...")

	go func() {
		// Run every 6 hours
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()

		recovery.SafeRun("ProviderRankingWorker.Initial", w.recalculateRankings)

		for {
			select {
			case <-w.stopChan:
				slog.Info("🏆 Stopping Provider Ranking Worker...")
				return
			case <-ticker.C:
				recovery.SafeRun("ProviderRankingWorker.Tick", w.recalculateRankings)
			}
		}
	}()
}

func (w *ProviderRankingWorker) Stop() {
	close(w.stopChan)
}

func (w *ProviderRankingWorker) recalculateRankings() {
	var providers []entity.Profile
	err := w.db.Preload("User").Where("user_type = ?", "PROVIDER").Find(&providers).Error
	if err != nil {
		slog.Error("Error fetching providers for ranking recalculation", slog.Any("error", err))
		return
	}

	if len(providers) == 0 {
		return
	}

	slog.Info("Recalculating ranking for providers", slog.Int("count", len(providers)))

	for _, profile := range providers {
		// Calculate Completed Appointments
		var completedCount int64
		w.db.Model(&entity.Appointment{}).
			Where("provider_id = ? AND status = ?", profile.UserID, "COMPLETED").
			Count(&completedCount)

		// Calculate Disputes
		var disputeCount int64
		w.db.Model(&entity.Dispute{}).
			Where("defendant_id = ?", profile.UserID).
			Count(&disputeCount)

		// Base score: 300
		score := 300.0

		// Appointments bonus: +15 per completed job (up to 300 pts)
		score += math.Min(float64(completedCount)*15.0, 300.0)

		// Rating bonus: average rating (0-5) * 40 (up to 200 pts)
		if profile.AverageRating > 0 {
			score += profile.AverageRating * 40.0
		}

		// Reviews bonus: +5 per review (up to 100 pts)
		score += math.Min(float64(profile.TotalReviews)*5.0, 100.0)

		// Verified identity: +100 pts
		if profile.IsIdentityVerified {
			score += 100.0
		}

		// Streak count: +2 pts per day (up to 50 pts)
		score += math.Min(float64(profile.StreakCount)*2.0, 50.0)

		// Dispute penalties: -50 pts each
		score -= float64(disputeCount) * 50.0

		// Clamp between 100 and 1000
		finalScore := int(math.Max(100.0, math.Min(1000.0, score)))

		oldScore := profile.NeighborScore
		tier := getTier(finalScore)

		if finalScore != oldScore {
			_ = w.db.Model(&entity.Profile{}).Where("id = ?", profile.ID).Update("neighbor_score", finalScore)

			// If reached Gold or Platinum or scored high, celebrate via push and email
			if finalScore >= 700 && oldScore < 700 && profile.User != nil {
				title := "Congratulations! Higher Provider Tier Unlocked"
				body := fmt.Sprintf("Your NeighborScore increased to %d (%s Tier)! You will now receive priority placement.", finalScore, tier)

				_ = w.db.Create(&entity.Notification{
					ID:               uuid.New(),
					UserID:           profile.UserID,
					NotificationType: "RANKING_TIER_PROMOTED",
					Title:            title,
					Message:          body,
					Data:             entity.JSONMap{"score": finalScore, "tier": tier},
					CreatedAt:        time.Now().UTC(),
				})

				w.sendPush(profile.UserID, title, body, map[string]string{
					"notification_type": "ranking_tier_promoted",
					"score":             fmt.Sprintf("%d", finalScore),
					"tier":              tier,
				})

				providerName := profile.FirstName
				if providerName == "" {
					providerName = profile.User.Email
				}
				_ = email.SendProviderRankingEmail(w.emailCfg, profile.User.Email, providerName, tier, finalScore)
			}
		}
	}
}

func getTier(score int) string {
	switch {
	case score >= 900:
		return "Platinum"
	case score >= 700:
		return "Gold"
	case score >= 450:
		return "Silver"
	default:
		return "Bronze"
	}
}

func (w *ProviderRankingWorker) sendPush(userID uuid.UUID, title, body string, data map[string]string) {
	if w.fcmClient == nil || w.db == nil {
		return
	}
	go func() {
		recovery.SafeRun("ProviderRankingWorker.sendPush", func() {
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
