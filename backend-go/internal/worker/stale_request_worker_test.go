package worker

import (
	"testing"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/database"
	"backend-go/internal/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestStaleRequestWorker_ProcessStaleRequests(t *testing.T) {
	cfg := &config.Config{
		DBDriver: "sqlite",
		DBDSN:    ":memory:",
		Env:      "testing",
	}

	db, err := database.Connect(cfg)
	assert.NoError(t, err)
	assert.NoError(t, database.AutoMigrate(db))

	user := entity.User{
		ID:        uuid.New(),
		Email:     "seeker@example.com",
		IsActive:  true,
		CreatedAt: time.Now().Add(-60 * 24 * time.Hour),
	}
	assert.NoError(t, db.Create(&user).Error)

	profile := entity.Profile{
		ID:        uuid.New(),
		UserID:    user.ID,
		FirstName: "Jane",
		LastName:  "Doe",
	}
	assert.NoError(t, db.Create(&profile).Error)

	// Old stale request (> 30 days)
	staleReq := entity.ServiceRequest{
		ID:          uuid.New(),
		UserID:      user.ID,
		Title:       "Old Garden Maintenance",
		Description: "Need lawn mowing",
		Status:      "OPEN",
		CreatedAt:   time.Now().Add(-40 * 24 * time.Hour),
	}
	assert.NoError(t, db.Create(&staleReq).Error)

	// Fresh request (< 30 days)
	freshReq := entity.ServiceRequest{
		ID:          uuid.New(),
		UserID:      user.ID,
		Title:       "Fresh Plumbing Fix",
		Description: "Leak under sink",
		Status:      "OPEN",
		CreatedAt:   time.Now().Add(-5 * 24 * time.Hour),
	}
	assert.NoError(t, db.Create(&freshReq).Error)

	worker := NewStaleRequestWorker(db, cfg, nil, nil)
	assert.NotNil(t, worker)

	// Execute processing
	worker.processStaleRequests()

	// Verify stale request was marked EXPIRED
	var updatedStale entity.ServiceRequest
	assert.NoError(t, db.First(&updatedStale, "id = ?", staleReq.ID).Error)
	assert.Equal(t, "EXPIRED", updatedStale.Status)

	// Verify fresh request remains OPEN
	var updatedFresh entity.ServiceRequest
	assert.NoError(t, db.First(&updatedFresh, "id = ?", freshReq.ID).Error)
	assert.Equal(t, "OPEN", updatedFresh.Status)

	// Verify an in-app notification was created for the seeker
	var notif entity.Notification
	assert.NoError(t, db.First(&notif, "user_id = ? AND notification_type = ?", user.ID, "REQUEST_EXPIRED").Error)
	assert.Contains(t, notif.Title, "Service Request Expired")
	assert.Contains(t, notif.Message, "Old Garden Maintenance")
}
