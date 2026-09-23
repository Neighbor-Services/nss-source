package usecase

import (
	"context"

	"backend-go/internal/domain/entity"
	"github.com/google/uuid"
)

type NotificationUseCase interface {
	GetNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]entity.Notification, error)
	MarkAsRead(ctx context.Context, notifID, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	RegisterDeviceToken(ctx context.Context, userID uuid.UUID, token, platform, deviceID string) error
	UnregisterDeviceToken(ctx context.Context, userID uuid.UUID, token string) error
	GetDeviceTokens(ctx context.Context, userID uuid.UUID) ([]entity.DeviceToken, error)
	SendPushNotification(ctx context.Context, userID uuid.UUID, title, body string, data map[string]string) error
	SendPushToTokens(ctx context.Context, tokens []string, title, body string, data map[string]string) error
}
