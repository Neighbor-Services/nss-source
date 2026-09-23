package gorm

import (
	"context"

	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type conversationRepository struct {
	db *gorm.DB
}

func NewConversationRepository(db *gorm.DB) repository.ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Conversation, error) {
	var list []entity.Conversation
	err := r.db.WithContext(ctx).
		Preload("Participants").
		Preload("Participants.Profile").
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(1)
		}).
		Joins("JOIN chat_conversation_participants ON chat_conversation_participants.conversation_id = chat_conversation.id").
		Where("chat_conversation_participants.user_id = ?", userID).
		Order("chat_conversation.updated_at DESC").
		Find(&list).Error
	return list, err
}

func (r *conversationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
	var conv entity.Conversation
	err := r.db.WithContext(ctx).
		Preload("Participants").
		Preload("Participants.Profile").
		First(&conv, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *conversationRepository) GetByParticipants(ctx context.Context, user1ID, user2ID uuid.UUID) (*entity.Conversation, error) {
	var conv entity.Conversation
	err := r.db.WithContext(ctx).
		Raw(`SELECT c.* FROM chat_conversation c 
			JOIN chat_conversation_participants p1 ON p1.conversation_id = c.id AND p1.user_id = ?
			JOIN chat_conversation_participants p2 ON p2.conversation_id = c.id AND p2.user_id = ?
			LIMIT 1`, user1ID, user2ID).
		Scan(&conv).Error

	if err != nil || conv.ID == uuid.Nil {
		return nil, gorm.ErrRecordNotFound
	}

	return r.GetByID(ctx, conv.ID)
}

func (r *conversationRepository) Create(ctx context.Context, conv *entity.Conversation, participantIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(conv).Error; err != nil {
			return err
		}

		type ConvParticipant struct {
			ConversationID uuid.UUID `gorm:"type:uuid"`
			UserID         uuid.UUID `gorm:"type:uuid"`
		}

		for _, pID := range participantIDs {
			participant := ConvParticipant{
				ConversationID: conv.ID,
				UserID:         pID,
			}
			if err := tx.Table("chat_conversation_participants").Create(&participant).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) repository.MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) ListByConversation(ctx context.Context, convID uuid.UUID, limit, offset int) ([]entity.Message, error) {
	var list []entity.Message
	query := r.db.WithContext(ctx).
		Preload("Sender").
		Where("conversation_id = ?", convID).
		Order("created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&list).Error
	return list, err
}

func (r *messageRepository) Create(ctx context.Context, msg *entity.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *messageRepository) MarkConversationRead(ctx context.Context, convID, readerID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entity.Message{}).
		Where("conversation_id = ? AND sender_id != ? AND is_seen = ?", convID, readerID, false).
		Update("is_seen", true).Error
}

func (r *messageRepository) MarkSeenBetweenUsers(ctx context.Context, senderID, receiverID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Exec(`UPDATE chat_message SET is_seen = true 
			WHERE is_seen = false AND sender_id = ? AND conversation_id IN (
				SELECT conversation_id FROM chat_conversation_participants WHERE user_id = ?
			)`, senderID, receiverID).Error
}

type chatBlockRepository struct {
	db *gorm.DB
}

func NewChatBlockRepository(db *gorm.DB) repository.ChatBlockRepository {
	return &chatBlockRepository{db: db}
}

func (r *chatBlockRepository) Block(ctx context.Context, blockerID, blockedID uuid.UUID, convID *uuid.UUID) error {
	block := entity.ChatBlock{
		ID:             uuid.New(),
		BlockerID:      blockerID,
		BlockedID:      blockedID,
		ConversationID: convID,
	}
	return r.db.WithContext(ctx).Create(&block).Error
}

func (r *chatBlockRepository) Unblock(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).
		Delete(&entity.ChatBlock{}).Error
}

func (r *chatBlockRepository) IsBlocked(ctx context.Context, user1ID, user2ID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.ChatBlock{}).
		Where("(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)", user1ID, user2ID, user2ID, user1ID).
		Count(&count).Error
	return count > 0, err
}

func (r *chatBlockRepository) ListBlocksByBlocker(ctx context.Context, blockerID uuid.UUID) ([]entity.ChatBlock, error) {
	var list []entity.ChatBlock
	err := r.db.WithContext(ctx).Where("blocker_id = ?", blockerID).Find(&list).Error
	return list, err
}
