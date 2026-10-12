package worker

import (
	"backend-go/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// isEmailNotificationEnabled checks whether the given user has email notifications turned on.
// Returns true by default if DB is nil or if no profile is found.
func isEmailNotificationEnabled(db *gorm.DB, userID uuid.UUID) bool {
	if db == nil {
		return true
	}
	var prof entity.Profile
	if err := db.Select("email_notifications_enabled").Where("user_id = ?", userID).First(&prof).Error; err != nil {
		return true
	}
	return prof.EmailNotificationsEnabled
}

// isPushNotificationEnabled checks whether the given user has push notifications turned on.
// Returns true by default if DB is nil or if no profile is found.
func isPushNotificationEnabled(db *gorm.DB, userID uuid.UUID) bool {
	if db == nil {
		return true
	}
	var prof entity.Profile
	if err := db.Select("push_notifications_enabled").Where("user_id = ?", userID).First(&prof).Error; err != nil {
		return true
	}
	return prof.PushNotificationsEnabled
}
