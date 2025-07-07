package main

import (
	"bome-backend/internal/config"
	"bome-backend/internal/database"
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.New()

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("Connected to database successfully")

	// Create a test video
	video, err := db.CreateVideo(
		"Test Video - Book of Mormon Evidence",
		"This is a test video to verify the backend-frontend integration is working properly.",
		"test-bunny-video-123",
		"https://example.com/test-thumbnail.jpg",
		"Archaeology",
		900,       // 15 minutes
		150000000, // 150MB
		[]string{"test", "archaeology", "evidence"},
		1, // created by user ID 1
	)
	if err != nil {
		log.Fatal("Failed to create video:", err)
	}

	fmt.Printf("Successfully created video: %s (ID: %d)\n", video.Title, video.ID)
	fmt.Println("Database seeding completed!")
}
