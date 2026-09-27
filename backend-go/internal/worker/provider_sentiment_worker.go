package worker

import (
	"context"
	"log/slog"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/pkg/aimatcher"
	"backend-go/pkg/recovery"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProviderSentimentWorker periodically analyzes review sentiment and pre-computes semantic search indices for paid providers.
type ProviderSentimentWorker struct {
	db       *gorm.DB
	cfg      *config.Config
	stopChan chan struct{}
}

func NewProviderSentimentWorker(db *gorm.DB, cfg *config.Config) *ProviderSentimentWorker {
	return &ProviderSentimentWorker{
		db:       db,
		cfg:      cfg,
		stopChan: make(chan struct{}),
	}
}

func (w *ProviderSentimentWorker) Start() {
	slog.Info("🤖 Starting AI Matcher & Provider Review Sentiment Background Worker...")

	go func() {
		// Run every 15 minutes to keep sentiment analysis and fast search cache fresh
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()

		// Initial indexing run on startup
		recovery.SafeRun("ProviderSentimentWorker.Initial", w.indexAndAnalyzeSentiment)

		for {
			select {
			case <-w.stopChan:
				slog.Info("🤖 Stopping Provider Sentiment Worker...")
				return
			case <-ticker.C:
				recovery.SafeRun("ProviderSentimentWorker.Tick", w.indexAndAnalyzeSentiment)
			}
		}
	}()
}

func (w *ProviderSentimentWorker) Stop() {
	close(w.stopChan)
}

func (w *ProviderSentimentWorker) indexAndAnalyzeSentiment() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 1. Fetch all active paid providers
	var providers []entity.Profile
	err := w.db.WithContext(ctx).
		Preload("User").
		Preload("CatalogServices").
		Preload("CatalogServices.Category").
		Preload("PortfolioItems").
		Preload("ServicePackages").
		Where("accounts_profile.user_type = ?", "PROVIDER").
		Where("(accounts_profile.subscription_tier IN ('SILVER', 'GOLD', 'PLATINUM', 'DIAMOND', 'PRO') OR EXISTS (SELECT 1 FROM payments_subscription ps WHERE ps.user_id = accounts_profile.user_id AND ps.is_active = true AND (ps.next_payment IS NULL OR ps.next_payment > NOW())))").
		Find(&providers).Error

	if err != nil {
		slog.Error("ProviderSentimentWorker: Failed to fetch paid providers", slog.Any("error", err))
		return
	}

	if len(providers) == 0 {
		slog.Info("ProviderSentimentWorker: No active paid providers found to index.")
		return
	}

	// 2. Fetch all reviews for these providers
	providerUserIDs := make([]uuid.UUID, 0, len(providers))
	for _, p := range providers {
		providerUserIDs = append(providerUserIDs, p.UserID)
	}

	var reviews []entity.Review
	if err := w.db.WithContext(ctx).
		Where("provider_id IN ?", providerUserIDs).
		Where("is_hidden = false").
		Order("created_at DESC").
		Find(&reviews).Error; err != nil {
		slog.Warn("ProviderSentimentWorker: Failed to fetch reviews", slog.Any("error", err))
	}

	reviewsMap := make(map[uuid.UUID][]entity.Review)
	for _, r := range reviews {
		reviewsMap[r.ProviderID] = append(reviewsMap[r.ProviderID], r)
	}

	// 3. Pre-index all providers into AI Matcher Fast Cache
	aimatcher.GlobalIndex.BulkIndex(providers, reviewsMap)

	// 4. Update provider performance & review sentiment scores
	updatedCount := 0
	for _, p := range providers {
		summary := aimatcher.SummarizeProviderReviews(reviewsMap[p.UserID], p)

		// Dynamic NeighborScore adjustment with sentiment factor
		if summary.TotalReviews > 0 {
			sentimentScoreAdjustment := int(summary.PerformanceScore * 5.0) // 0 to 500
			if sentimentScoreAdjustment > 0 {
				_ = w.db.Model(&entity.Profile{}).
					Where("id = ?", p.ID).
					Update("neighbor_score", sentimentScoreAdjustment)
				updatedCount++
			}
		}
	}

	slog.Info("⚡ AI Matcher & Provider Review Sentiment Index successfully updated",
		slog.Int("providers_indexed", len(providers)),
		slog.Int("total_reviews_analyzed", len(reviews)),
		slog.Int("performance_scores_updated", updatedCount),
	)
}
