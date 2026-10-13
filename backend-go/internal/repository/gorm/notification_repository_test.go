package gorm

import (
	"context"
	"testing"
	"time"

	"backend-go/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	gormPkg "gorm.io/gorm"
)

func TestDeviceTokenRepository_Upsert(t *testing.T) {
	db, err := gormPkg.Open(sqlite.Open("file::memory:?cache=shared"), &gormPkg.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&entity.DeviceToken{})
	assert.NoError(t, err)

	repo := NewDeviceTokenRepository(db)
	ctx := context.Background()

	userID1 := uuid.New()
	token := "fcm-device-token-12345"

	dt := entity.DeviceToken{
		ID:        uuid.New(),
		UserID:    userID1,
		Token:     token,
		Platform:  "ANDROID",
		DeviceID:  "dev-1",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 1. Initial insert
	err = repo.Upsert(ctx, &dt)
	assert.NoError(t, err)

	tokens, err := repo.ListByUser(ctx, userID1)
	assert.NoError(t, err)
	assert.Len(t, tokens, 1)
	assert.Equal(t, token, tokens[0].Token)
	assert.Equal(t, "dev-1", tokens[0].DeviceID)

	// 2. Upsert same token for user 2 (e.g. user re-login / switch accounts on same device)
	userID2 := uuid.New()
	dt2 := entity.DeviceToken{
		ID:        uuid.New(),
		UserID:    userID2,
		Token:     token,
		Platform:  "ANDROID",
		DeviceID:  "dev-2",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Upsert(ctx, &dt2)
	assert.NoError(t, err)

	tokensUser1, err := repo.ListByUser(ctx, userID1)
	assert.NoError(t, err)
	assert.Len(t, tokensUser1, 0)

	tokensUser2, err := repo.ListByUser(ctx, userID2)
	assert.NoError(t, err)
	assert.Len(t, tokensUser2, 1)
	assert.Equal(t, "dev-2", tokensUser2[0].DeviceID)
}
