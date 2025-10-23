package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	authModels "bome-backend/authentication/models"
	authServices "bome-backend/authentication/services"
	"bome-backend/infrastructure/config"
	"bome-backend/infrastructure/database"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	godotenv.Load()

	// Get user details from command line or use defaults
	email := "test.video@example.com"
	password := "VideoTest123!"
	firstName := "Video"
	lastName := "Tester"

	if len(os.Args) > 1 {
		email = os.Args[1]
	}
	if len(os.Args) > 2 {
		password = os.Args[2]
	}

	log.Printf("Creating test user: %s", email)

	// Initialize database
	cfg := config.New()
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Hash password
	hashedPassword, err := authServices.HashPassword(password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Check if user already exists
	existingUser, err := authModels.GetUserByEmail(db, email)
	if err == nil && existingUser != nil {
		log.Printf("User already exists with ID: %d", existingUser.ID)

		// Update to ensure verified and has password
		existingUser.EmailVerified = true
		existingUser.PasswordHash = hashedPassword
		existingUser.IsActive = sql.NullBool{Bool: true, Valid: true}

		// Use UpdateUserProfile for updates
		updates := map[string]interface{}{
			"email_verified": true,
			"password_hash":  hashedPassword,
			"is_active":      true,
		}
		if err := authModels.UpdateUserProfile(db, existingUser.ID, updates); err != nil {
			log.Fatalf("Failed to update existing user: %v", err)
		}

		log.Printf("✅ User updated and verified: %s (ID: %d)", email, existingUser.ID)
		fmt.Printf("EMAIL=%s\n", email)
		fmt.Printf("PASSWORD=%s\n", password)
		fmt.Printf("USER_ID=%d\n", existingUser.ID)
		return
	}

	// Create new user
	userID, err := authModels.CreateUser(db, email, firstName, lastName, hashedPassword, "user")
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	log.Printf("✅ Test user created successfully: %s (ID: %d)", email, userID)
	fmt.Printf("EMAIL=%s\n", email)
	fmt.Printf("PASSWORD=%s\n", password)
	fmt.Printf("USER_ID=%d\n", userID)
}
