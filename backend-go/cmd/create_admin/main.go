package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"backend-go/internal/config"
	"backend-go/internal/database"
	"backend-go/internal/domain/entity"
	"backend-go/pkg/auth"
	"github.com/google/uuid"
)

func main() {
	emailFlag := flag.String("email", "admin@neighborservice.com", "Admin user email address")
	passFlag := flag.String("password", "Admin123!", "Admin user password")
	firstFlag := flag.String("first-name", "Super", "Admin first name")
	lastFlag := flag.String("last-name", "Admin", "Admin last name")
	flag.Parse()

	email := strings.TrimSpace(*emailFlag)
	password := strings.TrimSpace(*passFlag)
	firstName := strings.TrimSpace(*firstFlag)
	lastName := strings.TrimSpace(*lastFlag)

	if email == "" || password == "" {
		log.Fatalf("Error: email and password must not be empty")
	}

	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL database: %v", err)
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	var existingUser entity.User
	result := db.Where("email = ?", email).First(&existingUser)

	if result.Error == nil {
		// User exists - update to superuser with new password
		existingUser.Password = hashedPassword
		existingUser.IsActive = true
		existingUser.IsStaff = true
		existingUser.IsSuperuser = true
		existingUser.IsVerified = true

		if err := db.Save(&existingUser).Error; err != nil {
			log.Fatalf("Failed to update existing admin user: %v", err)
		}

		// Ensure profile exists
		var prof entity.Profile
		if err := db.Where("user_id = ?", existingUser.ID).First(&prof).Error; err != nil {
			prof = entity.Profile{
				ID:        uuid.New(),
				UserID:    existingUser.ID,
				FirstName: firstName,
				LastName:  lastName,
				UserType:  "SEEKER",
			}
			_ = db.Create(&prof).Error
		} else {
			prof.FirstName = firstName
			prof.LastName = lastName
			_ = db.Save(&prof).Error
		}

		fmt.Println("\n========================================================")
		fmt.Println("  ✨ Existing user updated to Administrator successfully!")
		fmt.Println("========================================================")
		fmt.Printf("  Email:       %s\n", email)
		fmt.Printf("  Password:    %s\n", password)
		fmt.Printf("  Role:        Superuser / Admin Staff (is_staff=true)\n")
		fmt.Printf("  Portal URL:  http://localhost:4200/login\n")
		fmt.Println("========================================================")
		return
	}

	// Create new user
	newUserID := uuid.New()
	newUser := entity.User{
		ID:          newUserID,
		Email:       email,
		Password:    hashedPassword,
		IsActive:    true,
		IsStaff:     true,
		IsSuperuser: true,
		IsVerified:  true,
	}

	if err := db.Create(&newUser).Error; err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	newProfile := entity.Profile{
		ID:        uuid.New(),
		UserID:    newUserID,
		FirstName: firstName,
		LastName:  lastName,
		UserType:  "SEEKER",
	}
	if err := db.Create(&newProfile).Error; err != nil {
		log.Printf("Warning: Created user but profile creation gave: %v", err)
	}

	fmt.Println("\n========================================================")
	fmt.Println("  🚀 New Administrator account created successfully!")
	fmt.Println("========================================================")
	fmt.Printf("  Email:       %s\n", email)
	fmt.Printf("  Password:    %s\n", password)
	fmt.Printf("  Role:        Superuser / Admin Staff (is_staff=true)\n")
	fmt.Printf("  Portal URL:  http://localhost:4200/login\n")
	fmt.Println("========================================================")
	os.Exit(0)
}
