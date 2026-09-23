package usecase

import (
	"context"
	"log"
	"time"

	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/internal/websocket"
	"backend-go/pkg/fcm"
	"github.com/google/uuid"
)

type notificationUseCase struct {
	notifRepo repository.NotificationRepository
	tokenRepo repository.DeviceTokenRepository
	fcmClient fcm.Client
}

func NewNotificationUseCase(
	notifRepo repository.NotificationRepository,
	tokenRepo repository.DeviceTokenRepository,
	fcmClient fcm.Client,
) domainUsecase.NotificationUseCase {
	return &notificationUseCase{
		notifRepo: notifRepo,
		tokenRepo: tokenRepo,
		fcmClient: fcmClient,
	}
}

func (u *notificationUseCase) GetNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]entity.Notification, error) {
	return u.notifRepo.ListByUser(ctx, userID, limit)
}

func (u *notificationUseCase) MarkAsRead(ctx context.Context, notifID, userID uuid.UUID) error {
	return u.notifRepo.MarkAsRead(ctx, notifID, userID)
}

func (u *notificationUseCase) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return u.notifRepo.MarkAllAsRead(ctx, userID)
}

func (u *notificationUseCase) RegisterDeviceToken(ctx context.Context, userID uuid.UUID, token, platform, deviceID string) error {
	deviceToken := entity.DeviceToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     token,
		Platform:  platform,
		DeviceID:  deviceID,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return u.tokenRepo.Upsert(ctx, &deviceToken)
}

func (u *notificationUseCase) GetDeviceTokens(ctx context.Context, userID uuid.UUID) ([]entity.DeviceToken, error) {
	return u.tokenRepo.ListByUser(ctx, userID)
}

func (u *notificationUseCase) SendPushNotification(ctx context.Context, userID uuid.UUID, title, body string, data map[string]string) error {
	// 1. Broadcast over WebSocket immediately to any active sessions (web, Flutter, admin)
	websocket.GlobalHub.SendToUser(userID.String(), map[string]interface{}{
		"type":       "notification",
		"title":      title,
		"body":       body,
		"data":       data,
		"created_at": time.Now().UTC(),
	})

	if u.fcmClient == nil {
		return nil
	}

	tokens, err := u.tokenRepo.ListByUser(ctx, userID)
	if err != nil || len(tokens) == 0 {
		return err
	}

	tokenStrings := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if t.Token != "" && t.IsActive {
			tokenStrings = append(tokenStrings, t.Token)
		}
	}

	if len(tokenStrings) == 0 {
		return nil
	}

	go func() {
		invalidTokens, err := u.fcmClient.SendMulticast(context.Background(), tokenStrings, title, body, data)
		if err != nil {
			log.Printf("[PushNotification] Error sending multicast: %v", err)
		}
		// Clean up invalid or stale tokens
		for _, invToken := range invalidTokens {
			_ = u.tokenRepo.DeleteByToken(context.Background(), invToken)
		}
	}()

	return nil
}

func (u *notificationUseCase) SendPushToTokens(ctx context.Context, tokens []string, title, body string, data map[string]string) error {
	if u.fcmClient == nil || len(tokens) == 0 {
		return nil
	}

	go func() {
		invalidTokens, err := u.fcmClient.SendMulticast(context.Background(), tokens, title, body, data)
		if err != nil {
			log.Printf("[PushNotification] Error sending multicast to tokens: %v", err)
		}
		for _, invToken := range invalidTokens {
			_ = u.tokenRepo.DeleteByToken(context.Background(), invToken)
		}
	}()

	return nil
}

func (u *notificationUseCase) UnregisterDeviceToken(ctx context.Context, userID uuid.UUID, token string) error {
	if token == "" {
		return nil
	}
	return u.tokenRepo.DeleteByToken(ctx, token)
}

