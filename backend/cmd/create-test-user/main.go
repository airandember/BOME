package main

import (
	"fmt"
	"log"
	"os"

	"bome-backend/internal/database"
	authModels "bome-backend/internal/models/authentication"
	"bome-backend/internal/services/security/crypto"

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
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize crypto service
	cryptoService := crypto.GetGlobalCryptoService()
	if cryptoService == nil {
		log.Fatal("Crypto service not initialized")
	}

	// Hash password
	hashedPassword, err := cryptoService.HashPassword(password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create user
	user := &authModels.User{
		Email:         email,
		FirstName:     firstName,
		LastName:      lastName,
		Password:      hashedPassword,
		Role:          "user",
		EmailVerified: true, // Mark as verified for testing
		IsActive:      true,
	}

	// Check if user already exists
	existingUser, err := authModels.GetUserByEmail(db, email)
	if err == nil && existingUser != nil {
		log.Printf("User already exists with ID: %d", existingUser.ID)

		// Update to ensure verified and has password
		existingUser.EmailVerified = true
		existingUser.Password = hashedPassword
		existingUser.IsActive = true

		if err := authModels.UpdateUser(db, existingUser); err != nil {
			log.Fatalf("Failed to update existing user: %v", err)
		}

		log.Printf("✅ User updated and verified: %s (ID: %d)", email, existingUser.ID)
		fmt.Printf("EMAIL=%s\n", email)
		fmt.Printf("PASSWORD=%s\n", password)
		fmt.Printf("USER_ID=%d\n", existingUser.ID)
		return
	}

	// Create new user
	if err := authModels.CreateUser(db, user); err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	log.Printf("✅ Test user created successfully: %s (ID: %d)", email, user.ID)
	fmt.Printf("EMAIL=%s\n", email)
	fmt.Printf("PASSWORD=%s\n", password)
	fmt.Printf("USER_ID=%d\n", user.ID)
}
