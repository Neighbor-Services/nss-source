package worker

import (
	"fmt"
	"log"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/pkg/email"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PerformanceBadgeWorker awards badges to providers based on performance metrics.
type PerformanceBadgeWorker struct {
	db       *gorm.DB
	cfg      *config.Config
	emailCfg *email.Config
	stopChan chan struct{}
}

func NewPerformanceBadgeWorker(db *gorm.DB, cfg *config.Config) *PerformanceBadgeWorker {
	return &PerformanceBadgeWorker{
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

func (w *PerformanceBadgeWorker) Start() {
	log.Println("🏅 Starting Performance Badge Background Worker...")

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		w.processBadges()

		for {
			select {
			case <-w.stopChan:
				log.Println("🏅 Stopping Performance Badge Worker...")
				return
			case <-ticker.C:
				w.processBadges()
			}
		}
	}()
}

func (w *PerformanceBadgeWorker) Stop() {
	close(w.stopChan)
}

type badgeCriteria struct {
	name        string
	iconType    string
	description string
	check       func(profile entity.Profile, completedJobs int64) bool
}

var badgeCriteriaList = []badgeCriteria{
	{
		name:        "Top Rated",
		iconType:    "star",
		description: "Maintained a 4.8+ star rating with at least 10 reviews.",
		check: func(p entity.Profile, jobs int64) bool {
			return p.AverageRating >= 4.8 && p.TotalReviews >= 10
		},
	},
	{
		name:        "Trusted Pro",
		iconType:    "shield",
		description: "Completed 20+ jobs successfully.",
		check: func(p entity.Profile, jobs int64) bool {
			return jobs >= 20
		},
	},
	{
		name:        "Rising Star",
		iconType:    "bolt",
		description: "Completed 5+ jobs with a 4.5+ rating.",
		check: func(p entity.Profile, jobs int64) bool {
			return jobs >= 5 && p.AverageRating >= 4.5
		},
	},
	{
		name:        "Community Favorite",
		iconType:    "heart",
		description: "Received 25+ positive reviews.",
		check: func(p entity.Profile, jobs int64) bool {
			return p.TotalReviews >= 25
		},
	},
	{
		name:        "Elite Provider",
		iconType:    "crown",
		description: "Completed 50+ jobs with a 4.9+ rating.",
		check: func(p entity.Profile, jobs int64) bool {
			return jobs >= 50 && p.AverageRating >= 4.9
		},
	},
}

func (w *PerformanceBadgeWorker) processBadges() {
	// Fetch provider profiles only
	var profiles []entity.Profile
	err := w.db.Where("service != '' AND average_rating > 0").Find(&profiles).Error
	if err != nil {
		log.Printf("[BadgeWorker] Error fetching provider profiles: %v", err)
		return
	}

	thisMonth := time.Now().Truncate(30 * 24 * time.Hour)
	awarded := 0

	for _, profile := range profiles {
		// Count completed jobs this provider has done
		var completedJobs int64
		w.db.Model(&entity.Appointment{}).
			Where("provider_id = ? AND status = ?", profile.UserID, "COMPLETED").
			Count(&completedJobs)

		for _, criteria := range badgeCriteriaList {
			if !criteria.check(profile, completedJobs) {
				continue
			}

			// Already awarded this badge this month?
			var existing int64
			w.db.Model(&entity.PerformanceBadge{}).
				Where("profile_id = ? AND name = ? AND awarded_at >= ?", profile.ID, criteria.name, thisMonth).
				Count(&existing)
			if existing > 0 {
				continue
			}

			// Award badge
			if err := w.db.Create(&entity.PerformanceBadge{
				ID:          uuid.New(),
				ProfileID:   profile.ID,
				Name:        criteria.name,
				IconType:    criteria.iconType,
				Description: criteria.description,
				AwardedAt:   time.Now(),
			}).Error; err != nil {
				log.Printf("[BadgeWorker] Failed to award badge '%s' to profile %s: %v", criteria.name, profile.ID, err)
				continue
			}

			// In-app notification
			_ = w.db.Create(&entity.Notification{
				ID:               uuid.New(),
				UserID:           profile.UserID,
				NotificationType: "BADGE_AWARDED",
				Title:            fmt.Sprintf("🏅 Badge Unlocked: %s", criteria.name),
				Message:          criteria.description,
				Data:             entity.JSONMap{"badge_name": criteria.name, "icon_type": criteria.iconType},
				CreatedAt:        time.Now(),
			})

			// Fetch user email for email notification
			var user entity.User
			if err := w.db.First(&user, "id = ?", profile.UserID).Error; err == nil {
				providerName := user.Email
				if profile.FirstName != "" {
					providerName = profile.FirstName
				}
				_ = email.SendBadgeAwardedEmail(w.emailCfg, user.Email, providerName, criteria.name, criteria.description)
			}

			awarded++
			log.Printf("[BadgeWorker] Awarded '%s' badge to profile %s", criteria.name, profile.ID)
		}
	}

	if awarded > 0 {
		log.Printf("[BadgeWorker] Awarded %d badge(s) this cycle", awarded)
	}
}
