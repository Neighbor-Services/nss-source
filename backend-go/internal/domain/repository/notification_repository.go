package repository

import (
	"context"

	"backend-go/internal/domain/entity"
	"github.com/google/uuid"
)

type NotificationRepository interface {
	ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]entity.Notification, error)
	Create(ctx context.Context, notif *entity.Notification) error
	MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
}

type DeviceTokenRepository interface {
	Upsert(ctx context.Context, token *entity.DeviceToken) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.DeviceToken, error)
	DeleteByToken(ctx context.Context, token string) error
}
