package usecase_test

import (
	"context"
	"testing"
	"time"

	"backend-go/internal/domain/entity"
	"backend-go/internal/usecase"
	"github.com/google/uuid"
)

type mockConversationRepo struct {
	convs map[uuid.UUID]*entity.Conversation
}

func (m *mockConversationRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Conversation, error) {
	var list []entity.Conversation
	for _, c := range m.convs {
		list = append(list, *c)
	}
	return list, nil
}

func (m *mockConversationRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
	if c, ok := m.convs[id]; ok {
		return c, nil
	}
	return nil, nil
}

func (m *mockConversationRepo) GetByParticipants(ctx context.Context, u1, u2 uuid.UUID) (*entity.Conversation, error) {
	return nil, nil
}

func (m *mockConversationRepo) Create(ctx context.Context, conv *entity.Conversation, pIDs []uuid.UUID) error {
	m.convs[conv.ID] = conv
	return nil
}

type mockMessageRepo struct {
	msgs []*entity.Message
}

func (m *mockMessageRepo) ListByConversation(ctx context.Context, convID uuid.UUID, limit, offset int) ([]entity.Message, error) {
	var list []entity.Message
	for _, msg := range m.msgs {
		if msg.ConversationID == convID {
			list = append(list, *msg)
		}
	}
	return list, nil
}

func (m *mockMessageRepo) Create(ctx context.Context, msg *entity.Message) error {
	m.msgs = append(m.msgs, msg)
	return nil
}

func (m *mockMessageRepo) MarkConversationRead(ctx context.Context, convID, readerID uuid.UUID) error {
	return nil
}

func (m *mockMessageRepo) MarkSeenBetweenUsers(ctx context.Context, senderID, receiverID uuid.UUID) error {
	return nil
}

type mockChatBlockRepo struct {
	blocks map[string]*entity.ChatBlock
}

func (m *mockChatBlockRepo) Block(ctx context.Context, blockerID, blockedID uuid.UUID, convID *uuid.UUID) error {
	key := blockerID.String() + ":" + blockedID.String()
	m.blocks[key] = &entity.ChatBlock{
		ID:             uuid.New(),
		BlockerID:      blockerID,
		BlockedID:      blockedID,
		ConversationID: convID,
		CreatedAt:      time.Now(),
	}
	return nil
}

func (m *mockChatBlockRepo) Unblock(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	key := blockerID.String() + ":" + blockedID.String()
	delete(m.blocks, key)
	return nil
}

func (m *mockChatBlockRepo) IsBlocked(ctx context.Context, u1, u2 uuid.UUID) (bool, error) {
	k1 := u1.String() + ":" + u2.String()
	k2 := u2.String() + ":" + u1.String()
	_, b1 := m.blocks[k1]
	_, b2 := m.blocks[k2]
	return b1 || b2, nil
}

func (m *mockChatBlockRepo) ListBlocksByBlocker(ctx context.Context, blockerID uuid.UUID) ([]entity.ChatBlock, error) {
	var list []entity.ChatBlock
	for _, b := range m.blocks {
		if b.BlockerID == blockerID {
			list = append(list, *b)
		}
	}
	return list, nil
}

func TestChatUseCase_BlockAndMessaging(t *testing.T) {
	user1 := uuid.New()
	user2 := uuid.New()
	convID := uuid.New()

	convRepo := &mockConversationRepo{convs: map[uuid.UUID]*entity.Conversation{
		convID: {ID: convID},
	}}
	msgRepo := &mockMessageRepo{}
	blockRepo := &mockChatBlockRepo{blocks: make(map[string]*entity.ChatBlock)}

	chatUC := usecase.NewChatUseCase(convRepo, msgRepo, blockRepo, nil, nil)
	ctx := context.Background()

	// 1. Send normal message
	msg, err := chatUC.SendMessage(ctx, convID, user1, "Hello!", "", "TEXT")
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}
	if msg.Content != "Hello!" && msg.Message != "Hello!" {
		t.Errorf("Expected 'Hello!', got %s", msg.Content)
	}

	// 2. Block user2
	err = chatUC.BlockChat(ctx, user1, user2, &convID)
	if err != nil {
		t.Fatalf("Failed to block user: %v", err)
	}

	// 3. Verify user is blocked
	blocks, err := chatUC.GetBlockedUsers(ctx, user1)
	if err != nil || len(blocks) != 1 {
		t.Fatalf("Expected 1 blocked user, got %d", len(blocks))
	}

	// 4. Unblock user2
	err = chatUC.UnblockChat(ctx, user1, user2)
	if err != nil {
		t.Fatalf("Failed to unblock user: %v", err)
	}

	blocksAfter, _ := chatUC.GetBlockedUsers(ctx, user1)
	if len(blocksAfter) != 0 {
		t.Errorf("Expected 0 blocked users after unblock, got %d", len(blocksAfter))
	}
}
