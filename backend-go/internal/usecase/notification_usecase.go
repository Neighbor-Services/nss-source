package usecase

import (
	"context"
	"log"
	"strings"
	"time"

	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/internal/websocket"
	"backend-go/pkg/fcm"
	"github.com/google/uuid"
)

type notificationUseCase struct {
	notifRepo   repository.NotificationRepository
	tokenRepo   repository.DeviceTokenRepository
	profileRepo repository.ProfileRepository
	fcmClient   fcm.Client
}

func NewNotificationUseCase(
	notifRepo repository.NotificationRepository,
	tokenRepo repository.DeviceTokenRepository,
	profileRepo repository.ProfileRepository,
	fcmClient fcm.Client,
) domainUsecase.NotificationUseCase {
	return &notificationUseCase{
		notifRepo:   notifRepo,
		tokenRepo:   tokenRepo,
		profileRepo: profileRepo,
		fcmClient:   fcmClient,
	}
}

func (u *notificationUseCase) GetNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]entity.Notification, error) {
	list, err := u.notifRepo.ListByUser(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	// Never return provider-only job notifications to seekers
	isProvider := false
	if u.profileRepo != nil {
		if prof, pErr := u.profileRepo.GetByUserID(ctx, userID); pErr == nil && prof != nil {
			isProvider = strings.ToUpper(strings.TrimSpace(prof.UserType)) == "PROVIDER"
		}
	}
	if !isProvider {
		filtered := make([]entity.Notification, 0, len(list))
		for _, n := range list {
			nt := strings.ToUpper(strings.TrimSpace(n.NotificationType))
			if nt == "NEARBY_REQUEST" || nt == "BROADCAST_REQUEST" || nt == "DIRECT_REQUEST" || nt == "EMERGENCY_FLASH_DISPATCH" || nt == "REENGAGEMENT" || nt == "PROVIDER_SUBSCRIPTION_PROMO" {
				continue
			}
			filtered = append(filtered, n)
		}
		return filtered, nil
	}

	return list, nil
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
	// 1. Check if user turned off push notifications in their profile
	if u.profileRepo != nil {
		if prof, err := u.profileRepo.GetByUserID(ctx, userID); err == nil && prof != nil {
			if !prof.PushNotificationsEnabled {
				// User explicitly opted out of device push notifications
				return nil
			}
		}
	}

	// 2. Broadcast over WebSocket immediately to any active sessions (web, Flutter, admin)
	websocket.GlobalHub.SendToUser(userID.String(), map[string]interface{}{
 "type": "notification",
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

func (u *notificationUseCase) GetNotificationPreferences(ctx context.Context, userID uuid.UUID) (bool, bool, error) {
	if u.profileRepo == nil {
		return true, true, nil
	}
	prof, err := u.profileRepo.GetByUserID(ctx, userID)
	if err != nil || prof == nil {
		return true, true, nil
	}
	return prof.EmailNotificationsEnabled, prof.PushNotificationsEnabled, nil
}

func (u *notificationUseCase) UpdateNotificationPreferences(ctx context.Context, userID uuid.UUID, emailEnabled *bool, pushEnabled *bool) (bool, bool, error) {
	if u.profileRepo == nil {
		return true, true, nil
	}
	prof, err := u.profileRepo.GetByUserID(ctx, userID)
	if err != nil || prof == nil {
		prof = &entity.Profile{
			ID:                        uuid.New(),
			UserID:                    userID,
			UserType: "SEEKER",
			EmailNotificationsEnabled: true,
			PushNotificationsEnabled:  true,
		}
		if emailEnabled != nil {
			prof.EmailNotificationsEnabled = *emailEnabled
		}
		if pushEnabled != nil {
			prof.PushNotificationsEnabled = *pushEnabled
		}
		if createErr := u.profileRepo.Create(ctx, prof); createErr != nil {
			return false, false, createErr
		}
		return prof.EmailNotificationsEnabled, prof.PushNotificationsEnabled, nil
	}

	if emailEnabled != nil {
		prof.EmailNotificationsEnabled = *emailEnabled
	}
	if pushEnabled != nil {
		prof.PushNotificationsEnabled = *pushEnabled
	}
	prof.UpdatedAt = time.Now()
	if err := u.profileRepo.Update(ctx, prof); err != nil {
		return false, false, err
	}
	return prof.EmailNotificationsEnabled, prof.PushNotificationsEnabled, nil
}

