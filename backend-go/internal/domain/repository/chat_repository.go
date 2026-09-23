package repository

import (
	"context"

	"backend-go/internal/domain/entity"
	"github.com/google/uuid"
)

type ConversationRepository interface {
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Conversation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error)
	GetByParticipants(ctx context.Context, user1ID, user2ID uuid.UUID) (*entity.Conversation, error)
	Create(ctx context.Context, conv *entity.Conversation, participantIDs []uuid.UUID) error
}

type MessageRepository interface {
	ListByConversation(ctx context.Context, convID uuid.UUID, limit, offset int) ([]entity.Message, error)
	Create(ctx context.Context, msg *entity.Message) error
	MarkConversationRead(ctx context.Context, convID, readerID uuid.UUID) error
	MarkSeenBetweenUsers(ctx context.Context, senderID, receiverID uuid.UUID) error
}

type ChatBlockRepository interface {
	Block(ctx context.Context, blockerID, blockedID uuid.UUID, convID *uuid.UUID) error
	Unblock(ctx context.Context, blockerID, blockedID uuid.UUID) error
	IsBlocked(ctx context.Context, user1ID, user2ID uuid.UUID) (bool, error)
	ListBlocksByBlocker(ctx context.Context, blockerID uuid.UUID) ([]entity.ChatBlock, error)
}
