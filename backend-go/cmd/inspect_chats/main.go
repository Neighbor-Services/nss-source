package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	gormRepo "backend-go/internal/repository/gorm"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	db, err := gorm.Open(postgres.Open(cfg.DBDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("DB connect err: %v", err)
	}

	convRepo := gormRepo.NewConversationRepository(db)
	msgRepo := gormRepo.NewMessageRepository(db)

	var users []entity.User
	db.Preload("Profile").Find(&users)
	fmt.Printf("Users found: %d\n", len(users))

	if len(users) >= 2 {
		u1 := users[2] // Afari Samuel
		u2 := users[3] // Kwame Samuel
		fmt.Printf("Testing conversation between u1=%s (%s) and u2=%s (%s)\n", u1.ID, u1.Email, u2.ID, u2.Email)

		existing, err := convRepo.GetByParticipants(context.Background(), u1.ID, u2.ID)
		if err != nil || existing == nil {
			fmt.Printf("No existing conversation, creating one...\n")
			newConv := entity.Conversation{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			err = convRepo.Create(context.Background(), &newConv, []uuid.UUID{u1.ID, u2.ID})
			if err != nil {
				log.Fatalf("Error creating conv: %v", err)
			}
			fmt.Printf("Created conv ID=%s\n", newConv.ID)
			existing = &newConv
		} else {
			fmt.Printf("Found existing conv ID=%s\n", existing.ID)
		}

		// Insert test message
		testMsg := entity.Message{
			ID:             uuid.New(),
			ConversationID: existing.ID,
			SenderID:       u1.ID,
			Content:        "Hello from Afari to Kwame!",
			Message:        "Hello from Afari to Kwame!",
			CreatedAt:      time.Now(),
			IsDelivered:    true,
		}
		err = msgRepo.Create(context.Background(), &testMsg)
		if err != nil {
			log.Fatalf("Error creating msg: %v", err)
		}
		fmt.Printf("Created test message ID=%s\n", testMsg.ID)

		// Test ListByUser
		list1, err := convRepo.ListByUser(context.Background(), u1.ID)
		fmt.Printf("User 1 (%s) conversations: %d (err=%v)\n", u1.Email, len(list1), err)
		for _, c := range list1 {
			fmt.Printf("  Conv %s: participants=%d, messages=%d\n", c.ID, len(c.Participants), len(c.Messages))
			for _, p := range c.Participants {
				pName := ""
				if p.Profile != nil {
					pName = p.Profile.FirstName + " " + p.Profile.LastName
				}
				fmt.Printf("    Participant: %s (%s)\n", p.Email, pName)
			}
		}

		list2, err := convRepo.ListByUser(context.Background(), u2.ID)
		fmt.Printf("User 2 (%s) conversations: %d (err=%v)\n", u2.Email, len(list2), err)
	}
}
