package usecase

import (
	"context"

	"backend-go/internal/domain/entity"
	"github.com/google/uuid"
)

type ChatUseCase interface {
	GetConversations(ctx context.Context, userID uuid.UUID) ([]entity.Conversation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error)
	GetByParticipants(ctx context.Context, user1ID, user2ID uuid.UUID) (*entity.Conversation, error)
	CreateConversation(ctx context.Context, senderID, recipientID uuid.UUID) (*entity.Conversation, error)
	GetMessages(ctx context.Context, convID, userID uuid.UUID, limit, offset int) ([]entity.Message, error)
	SendMessage(ctx context.Context, convID, senderID uuid.UUID, text, mediaURL, messageType string) (*entity.Message, error)
	SaveFullMessage(ctx context.Context, msg *entity.Message) (*entity.Message, error)
	SetSeen(ctx context.Context, currentUserID, receiverID uuid.UUID) error
	BlockChat(ctx context.Context, blockerID, blockedID uuid.UUID, convID *uuid.UUID) error
	UnblockChat(ctx context.Context, blockerID, blockedID uuid.UUID) error
	GetBlockedUsers(ctx context.Context, blockerID uuid.UUID) ([]entity.ChatBlock, error)
}
