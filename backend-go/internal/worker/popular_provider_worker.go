package worker

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/pkg/aimatcher"
	"backend-go/pkg/cache"
	"backend-go/pkg/geo"
	"backend-go/pkg/recovery"

	"gorm.io/gorm"
)

type PopularProviderWorker struct {
	db       *gorm.DB
	cfg      *config.Config
	cache    cache.Cache
	stopChan chan struct{}
}

func NewPopularProviderWorker(db *gorm.DB, cfg *config.Config, cache cache.Cache) *PopularProviderWorker {
	return &PopularProviderWorker{
		db:       db,
		cfg:      cfg,
		cache:    cache,
		stopChan: make(chan struct{}),
	}
}

func (w *PopularProviderWorker) Start() {
	log.Println("⭐ Starting Popular Provider Geospatial Pre-caching Background Worker...")

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		recovery.SafeRun("PopularProviderWorker.Initial", w.processPopularProviders)

		for {
			select {
			case <-w.stopChan:
				log.Println("⭐ Stopping Popular Provider Background Worker...")
				return
			case <-ticker.C:
				recovery.SafeRun("PopularProviderWorker.Tick", w.processPopularProviders)
			}
		}
	}()
}

func (w *PopularProviderWorker) Stop() {
	close(w.stopChan)
}

func (w *PopularProviderWorker) processPopularProviders() {
	if w.cache == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 1. Fetch actively seen seekers who have valid non-zero coordinates
	var seekers []entity.Profile
	recentCutoff := time.Now().Add(-48 * time.Hour)
	err := w.db.WithContext(ctx).
		Where("user_type = ? AND (latitude != 0 OR longitude != 0) AND (last_seen > ? OR updated_at > ?)", "SEEKER", recentCutoff, recentCutoff).
		Limit(250).
		Find(&seekers).Error
	if err != nil || len(seekers) == 0 {
		// Fallback to most recently updated seekers with coordinates if no recent activity
		_ = w.db.WithContext(ctx).
			Where("user_type = ? AND (latitude != 0 OR longitude != 0)", "SEEKER").
			Order("updated_at DESC").
			Limit(50).
			Find(&seekers).Error
	}

	if len(seekers) == 0 {
		return
	}

	// 2. Fetch all providers with active subscriptions and valid coordinates
	var providers []entity.Profile
	err = w.db.WithContext(ctx).
		Preload("User").
		Preload("CatalogServices").
		Preload("CatalogServices.Category").
		Preload("PortfolioItems").
		Preload("ServicePackages").
		Where("accounts_profile.user_type = ? AND (accounts_profile.latitude != 0 OR accounts_profile.longitude != 0)", "PROVIDER").
		Where("(accounts_profile.subscription_tier IN ('SILVER', 'GOLD', 'PLATINUM', 'DIAMOND', 'PRO') OR EXISTS (SELECT 1 FROM payments_subscription ps WHERE ps.user_id = accounts_profile.user_id AND ps.is_active = true AND (ps.next_payment IS NULL OR ps.next_payment > NOW())))").
		Find(&providers).Error
	if err != nil {
		log.Printf("[PopularProviderWorker] Error fetching subscribed providers: %v", err)
		return
	}

	if len(providers) == 0 {
		return
	}

	// Enrich providers with catalog services and sentiment badges
	for i := range providers {
		providers[i].EnrichCatalogServices()
		if summary, ok := aimatcher.GlobalIndex.GetSentimentSummary(providers[i].ID); ok {
			providers[i].PositiveReviewRatio = summary.PositiveRatio
			providers[i].ReviewSentimentBadge = summary.SentimentBadge
			providers[i].PraiseBadges = summary.TopPraiseBadges
		}
	}

	type distProvider struct {
		provider entity.Profile
		distance float64
	}

	cachedCount := 0
	// 3. For each active seeker, map to their distance preference & providers around them
	for _, seeker := range seekers {
		radius := seeker.SearchRadiusKm
		if radius <= 0 {
			radius = 25.0
		}

		// Fast Bounding Box pre-filtering: 1 deg lat is ~69 miles
		maxLatDelta := radius / 69.0
		maxLonDelta := radius / 50.0

		var candidates []distProvider
		for _, p := range providers {
			if p.UserID == seeker.UserID || p.ID == seeker.ID {
				continue
			}

			// Scalar bounding-box rejection before computing trigonometry
			latDiff := seeker.Latitude - p.Latitude
			if latDiff < 0 {
				latDiff = -latDiff
			}
			if latDiff > maxLatDelta {
				continue
			}

			lonDiff := seeker.Longitude - p.Longitude
			if lonDiff < 0 {
				lonDiff = -lonDiff
			}
			if lonDiff > maxLonDelta {
				continue
			}

			dist := geo.HaversineDistance(seeker.Latitude, seeker.Longitude, p.Latitude, p.Longitude)
			if dist <= radius {
				candidates = append(candidates, distProvider{
					provider: p,
					distance: dist,
				})
			}
		}

		// Sort: highest rating, then reviews, then neighbor score, then distance
		sort.Slice(candidates, func(i, j int) bool {
			pi := candidates[i].provider
			pj := candidates[j].provider
			if pi.AverageRating != pj.AverageRating {
				return pi.AverageRating > pj.AverageRating
			}
			if pi.TotalReviews != pj.TotalReviews {
				return pi.TotalReviews > pj.TotalReviews
			}
			if pi.NeighborScore != pj.NeighborScore {
				return pi.NeighborScore > pj.NeighborScore
			}
			return candidates[i].distance < candidates[j].distance
		})

		limit := 20
		if len(candidates) < limit {
			limit = len(candidates)
		}

		topProviders := make([]entity.Profile, 0, limit)
		for i := 0; i < limit; i++ {
			topProviders = append(topProviders, candidates[i].provider)
		}

		if len(topProviders) > 0 {
			// Cache by seeker ID
			seekerKey := fmt.Sprintf("cache:popular_providers:seeker:%s", seeker.UserID.String())
			_ = w.cache.Set(ctx, seekerKey, topProviders, 30*time.Minute)

			// Cache by rounded geo coordinates grid (~1km)
			geoKey := fmt.Sprintf("cache:popular_providers:geo:%.2f:%.2f:rad:%.0f", seeker.Latitude, seeker.Longitude, radius)
			_ = w.cache.Set(ctx, geoKey, topProviders, 30*time.Minute)

			cachedCount++
		}
	}

	if cachedCount > 0 {
		log.Printf("[PopularProviderWorker] Successfully pre-cached popular providers for %d seekers", cachedCount)
	}
}
