package worker

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/pkg/cache"
	"backend-go/pkg/email"
	"backend-go/pkg/fcm"
	"backend-go/pkg/recovery"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NearbyRequestWorker struct {
	db        *gorm.DB
	cfg       *config.Config
	emailCfg  *email.Config
	fcmClient fcm.Client
	cache     cache.Cache
	stopChan  chan struct{}
}

func NewNearbyRequestWorker(db *gorm.DB, cfg *config.Config, fcmClient fcm.Client, cache cache.Cache) *NearbyRequestWorker {
	return &NearbyRequestWorker{
		db:        db,
		cfg:       cfg,
		fcmClient: fcmClient,
		cache:     cache,
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

func (w *NearbyRequestWorker) Start() {
	log.Println("🔍 Starting Nearby Request Matching & Redis Pre-caching Background Worker...")

	go func() {
		// Run every 5 minutes
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		// Run immediately on worker start
		recovery.SafeRun("NearbyRequestWorker.Initial", w.processNearbyRequestsForUsers)

		for {
			select {
			case <-w.stopChan:
				log.Println("🔍 Stopping Nearby Request Background Worker...")
				return
			case <-ticker.C:
				recovery.SafeRun("NearbyRequestWorker.Tick", w.processNearbyRequestsForUsers)
			}
		}
	}()
}

func (w *NearbyRequestWorker) Stop() {
	close(w.stopChan)
}

func (w *NearbyRequestWorker) processNearbyRequestsForUsers() {
	// 1. Fetch users and profiles that have coordinates registered
	var profiles []entity.Profile
	err := w.db.Preload("User").
		Where("(latitude != 0 OR longitude != 0)").
		Find(&profiles).Error
	if err != nil {
		log.Printf("[NearbyRequestWorker] Error fetching user profiles with coordinates: %v", err)
		return
	}

	if len(profiles) == 0 {
		return
	}

	// 2. Fetch all OPEN service requests created in the last 14 days
	twoWeeksAgo := time.Now().Add(-14 * 24 * time.Hour)
	var openRequests []entity.ServiceRequest
	err = w.db.Preload("User").
		Preload("User.Profile").
		Preload("CatalogService").
		Where("status = ? AND created_at >= ? AND latitude IS NOT NULL AND longitude IS NOT NULL AND (latitude != 0 OR longitude != 0)",
			"OPEN", twoWeeksAgo).
		Order("created_at DESC").
		Find(&openRequests).Error
	if err != nil {
		log.Printf("[NearbyRequestWorker] Error fetching open service requests: %v", err)
		return
	}

	if len(openRequests) == 0 {
		return
	}

	for _, profile := range profiles {
		if profile.User == nil {
			continue
		}
		if profile.Latitude == 0 && profile.Longitude == 0 {
			continue
		}

		userLat := profile.Latitude
		userLon := profile.Longitude
		userID := profile.UserID

		// Calculate matching nearby requests within max 25 km (or custom radius)
		maxRadiusKm := 25.0
		var nearbyList []entity.ServiceRequest

		for _, req := range openRequests {
			// Do not match user to their own request
			if req.UserID == userID {
				continue
			}
			if req.Latitude == nil || req.Longitude == nil {
				continue
			}

			dist := haversineDistanceKm(userLat, userLon, *req.Latitude, *req.Longitude)
			if dist <= maxRadiusKm {
				reqCopy := req
				reqCopy.Distance = &dist
				nearbyList = append(nearbyList, reqCopy)
			}
		}

		// Sort by nearest distance
		sort.SliceStable(nearbyList, func(i, j int) bool {
			if nearbyList[i].Distance == nil {
				return false
			}
			if nearbyList[j].Distance == nil {
				return true
			}
			return *nearbyList[i].Distance < *nearbyList[j].Distance
		})

		// 3. Precompute and store in Redis Cache for lightning fast app loading!
		if w.cache != nil && len(nearbyList) > 0 {
			cacheKey := fmt.Sprintf("cache:nearby_requests:user:%s", userID.String())
			_ = w.cache.Set(context.Background(), cacheKey, nearbyList, 10*time.Minute)
		}

		// 4. Notify user about newly discovered nearby requests (top 3 closest)
		limitNotify := 3
		if len(nearbyList) < limitNotify {
			limitNotify = len(nearbyList)
		}

		for i := 0; i < limitNotify; i++ {
			matchedReq := nearbyList[i]
			refID := fmt.Sprintf("nearby_%s_%s", userID.String(), matchedReq.ID.String())

			// Check if already notified
			var count int64
			w.db.Model(&entity.EmailCampaignLog{}).
				Where("campaign_type = ? AND reference_id = ?", "NEARBY_REQUEST_ALERT", refID).
				Count(&count)

			if count > 0 {
				continue
			}

			distVal := 0.0
			if matchedReq.Distance != nil {
				distVal = *matchedReq.Distance
			}

			// In-app Notification
			notif := entity.Notification{
				ID:               uuid.New(),
				UserID:           userID,
				SenderID:         &matchedReq.UserID,
				NotificationType: "NEARBY_REQUEST",
				Title:            "New Job Opportunity Nearby",
				Message:          fmt.Sprintf("A client nearby needs '%s' (~%.1f km away).", matchedReq.Title, distVal),
				Data:             entity.JSONMap{"request_id": matchedReq.ID.String(), "distance": fmt.Sprintf("%.1f", distVal)},
				CreatedAt:        time.Now(),
			}
			_ = w.db.Create(&notif)

			// FCM Push Notification
			w.sendPush(
				userID,
				"New Job Nearby",
				fmt.Sprintf("A client requested '%s' (~%.1f km away). Tap to view!", matchedReq.Title, distVal),
				map[string]string{
					"notification_type": "nearby_request",
					"request_id":        matchedReq.ID.String(),
					"distance":          fmt.Sprintf("%.1f", distVal),
				},
			)

			// Send Email if Provider
			userName := profile.User.Email
			if profile.FirstName != "" {
				userName = profile.FirstName
			}
			serviceName := matchedReq.ServiceType
			if matchedReq.CatalogService != nil && matchedReq.CatalogService.Name != "" {
				serviceName = matchedReq.CatalogService.Name
			}

			priceVal := 0.0
			if matchedReq.Price != nil {
				priceVal = *matchedReq.Price
			}
			_ = email.SendNearbyJobAlertEmail(
				w.emailCfg,
				profile.User.Email,
				userName,
				matchedReq.Title,
				serviceName,
				distVal,
				priceVal,
			)

			// Record log to prevent duplicate notifications
			w.db.Create(&entity.EmailCampaignLog{
				ID:           uuid.New(),
				UserID:       userID,
				TargetEmail:  profile.User.Email,
				CampaignType: "NEARBY_REQUEST_ALERT",
				ReferenceID:  &refID,
				SentAt:       time.Now(),
			})
		}
	}
}

func (w *NearbyRequestWorker) sendPush(userID uuid.UUID, title, body string, data map[string]string) {
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
		for _, invToken := range invalidTokens {
			w.db.Where("token = ?", invToken).Delete(&entity.DeviceToken{})
		}
	}()
}

func haversineDistanceKm(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0

	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	rLat1 := lat1 * math.Pi / 180.0
	rLat2 := lat2 * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(rLat1)*math.Cos(rLat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}
