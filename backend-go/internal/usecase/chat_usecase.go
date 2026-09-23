package usecase

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	domainUC "backend-go/internal/domain/usecase"
	"backend-go/pkg/fcm"
	"github.com/google/uuid"
)

type chatUseCase struct {
	convRepo      repository.ConversationRepository
	msgRepo       repository.MessageRepository
	blockRepo     repository.ChatBlockRepository
	tokenRepo     repository.DeviceTokenRepository
	fcmClient     fcm.Client
}

func NewChatUseCase(
	convRepo repository.ConversationRepository,
	msgRepo repository.MessageRepository,
	blockRepo repository.ChatBlockRepository,
	tokenRepo repository.DeviceTokenRepository,
	fcmClient fcm.Client,
) domainUC.ChatUseCase {
	return &chatUseCase{
		convRepo:  convRepo,
		msgRepo:   msgRepo,
		blockRepo: blockRepo,
		tokenRepo: tokenRepo,
		fcmClient: fcmClient,
	}
}

func (u *chatUseCase) GetConversations(ctx context.Context, userID uuid.UUID) ([]entity.Conversation, error) {
	return u.convRepo.ListByUser(ctx, userID)
}

func (u *chatUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
	return u.convRepo.GetByID(ctx, id)
}

func (u *chatUseCase) GetByParticipants(ctx context.Context, user1ID, user2ID uuid.UUID) (*entity.Conversation, error) {
	return u.convRepo.GetByParticipants(ctx, user1ID, user2ID)
}

func (u *chatUseCase) CreateConversation(ctx context.Context, senderID, recipientID uuid.UUID) (*entity.Conversation, error) {
	if senderID == recipientID {
		return nil, errors.New("cannot create conversation with oneself")
	}

	conv, err := u.convRepo.GetByParticipants(ctx, senderID, recipientID)
	if err == nil && conv != nil {
		return conv, nil
	}

	newConv := entity.Conversation{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	participantIDs := []uuid.UUID{senderID, recipientID}
	if err := u.convRepo.Create(ctx, &newConv, participantIDs); err != nil {
		return nil, err
	}

	return u.convRepo.GetByID(ctx, newConv.ID)
}

func (u *chatUseCase) GetMessages(ctx context.Context, convID, userID uuid.UUID, limit, offset int) ([]entity.Message, error) {
	_ = u.msgRepo.MarkConversationRead(ctx, convID, userID)
	return u.msgRepo.ListByConversation(ctx, convID, limit, offset)
}

func (u *chatUseCase) SendMessage(ctx context.Context, convID, senderID uuid.UUID, text, mediaURL, messageType string) (*entity.Message, error) {
	if text == "" && mediaURL == "" {
		return nil, errors.New("message content or media is required")
	}

	if messageType == "" {
		messageType = "TEXT"
	}

	fileName := ""
	if mediaURL != "" {
		fileName = filepath.Base(mediaURL)
	}

	isImage := mediaURL != "" && (strings.EqualFold(messageType, "IMAGE") || strings.HasSuffix(strings.ToLower(mediaURL), ".jpg") || strings.HasSuffix(strings.ToLower(mediaURL), ".jpeg") || strings.HasSuffix(strings.ToLower(mediaURL), ".png") || strings.HasSuffix(strings.ToLower(mediaURL), ".webp"))

	msg := entity.Message{
		ID:             uuid.New(),
		ConversationID: convID,
		SenderID:       senderID,
		Content:        text,
		Message:        text,
		MediaURL:       mediaURL,
		Image:          mediaURL,
		FileName:       fileName,
		WithImage:      isImage,
		IsSeen:         false,
		IsDelivered:    true,
		CreatedAt:      time.Now(),
	}

	return u.SaveFullMessage(ctx, &msg)
}

func (u *chatUseCase) SaveFullMessage(ctx context.Context, msg *entity.Message) (*entity.Message, error) {
	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	if msg.Message == "" && msg.Content != "" {
		msg.Message = msg.Content
	}
	if msg.Content == "" && msg.Message != "" {
		msg.Content = msg.Message
	}
	if msg.Image == "" && msg.MediaURL != "" {
		msg.Image = msg.MediaURL
	}
	if msg.MediaURL == "" && msg.Image != "" {
		msg.MediaURL = msg.Image
	}
	if msg.FileName == "" && msg.MediaURL != "" {
		msg.FileName = filepath.Base(msg.MediaURL)
	}

	if err := u.msgRepo.Create(ctx, msg); err != nil {
		return nil, err
	}

	// Asynchronously dispatch FCM push notification to recipient(s)
	if u.fcmClient != nil && u.tokenRepo != nil {
		go func() {
			conv, err := u.convRepo.GetByID(context.Background(), msg.ConversationID)
			if err != nil || conv == nil {
				return
			}

			senderName := "New Message"
			for _, p := range conv.Participants {
				if p.ID == msg.SenderID && p.Profile != nil && p.Profile.FirstName != "" {
					senderName = p.Profile.FirstName
					if p.Profile.LastName != "" {
						senderName = fmt.Sprintf("%s %s", p.Profile.FirstName, p.Profile.LastName)
					}
				}
			}

			body := msg.Message
			if body == "" && msg.MediaURL != "" {
				body = "Sent an attachment"
			}

			for _, p := range conv.Participants {
				if p.ID == msg.SenderID {
					continue
				}

				tokens, tErr := u.tokenRepo.ListByUser(context.Background(), p.ID)
				if tErr != nil || len(tokens) == 0 {
					continue
				}

				var tokenStrings []string
				for _, t := range tokens {
					if t.Token != "" && t.IsActive {
						tokenStrings = append(tokenStrings, t.Token)
					}
				}

				if len(tokenStrings) == 0 {
					continue
				}

				data := map[string]string{
					"notification_type": "message",
					"sender_id":         msg.SenderID.String(),
					"conversation_id":   msg.ConversationID.String(),
					"message_id":        msg.ID.String(),
				}

				invalidTokens, _ := u.fcmClient.SendMulticast(context.Background(), tokenStrings, senderName, body, data)
				for _, invToken := range invalidTokens {
					_ = u.tokenRepo.DeleteByToken(context.Background(), invToken)
				}
			}
		}()
	}

	return msg, nil
}

func (u *chatUseCase) SetSeen(ctx context.Context, currentUserID, receiverID uuid.UUID) error {
	return u.msgRepo.MarkSeenBetweenUsers(ctx, currentUserID, receiverID)
}

func (u *chatUseCase) BlockChat(ctx context.Context, blockerID, blockedID uuid.UUID, convID *uuid.UUID) error {
	return u.blockRepo.Block(ctx, blockerID, blockedID, convID)
}

func (u *chatUseCase) UnblockChat(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	return u.blockRepo.Unblock(ctx, blockerID, blockedID)
}

func (u *chatUseCase) GetBlockedUsers(ctx context.Context, blockerID uuid.UUID) ([]entity.ChatBlock, error) {
	return u.blockRepo.ListBlocksByBlocker(ctx, blockerID)
}
