package gorm

import (
	"context"

	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) repository.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]entity.Notification, error) {
	var list []entity.Notification
	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&list).Error
	return list, err
}

func (r *notificationRepository) Create(ctx context.Context, notif *entity.Notification) error {
	return r.db.WithContext(ctx).Create(notif).Error
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("user_id = ?", userID).
		Update("is_read", true).Error
}

func (r *notificationRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

type deviceTokenRepository struct {
	db *gorm.DB
}

func NewDeviceTokenRepository(db *gorm.DB) repository.DeviceTokenRepository {
	return &deviceTokenRepository{db: db}
}

func (r *deviceTokenRepository) Upsert(ctx context.Context, token *entity.DeviceToken) error {
	var existing entity.DeviceToken
	err := r.db.WithContext(ctx).Where("token = ?", token.Token).First(&existing).Error
	if err == nil {
		token.ID = existing.ID
		return r.db.WithContext(ctx).Save(token).Error
	}
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *deviceTokenRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.DeviceToken, error) {
	var list []entity.DeviceToken
	err := r.db.WithContext(ctx).Where("user_id = ? AND is_active = ?", userID, true).Find(&list).Error
	return list, err
}

func (r *deviceTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&entity.DeviceToken{}).Error
}
