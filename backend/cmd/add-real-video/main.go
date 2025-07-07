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

	// TODO: Replace with your actual Bunny.net video ID
	// You can find this in your Bunny.net dashboard under Stream Library
	actualBunnyVideoID := "YOUR_ACTUAL_BUNNY_VIDEO_ID_HERE"

	// Create a video record with your real Bunny.net video
	video, err := db.CreateVideo(
		"Book of Mormon Archaeological Evidence",                                        // Your actual title
		"Exploring archaeological evidence that supports the Book of Mormon narrative.", // Your description
		actualBunnyVideoID, // ← YOUR REAL BUNNY VIDEO ID
		"",                 // Let Bunny.net generate thumbnail
		"Archaeology",      // Category
		0,                  // Duration (will be updated from Bunny.net)
		0,                  // File size (will be updated)
		[]string{"archaeology", "book-of-mormon", "evidence", "history"}, // Your tags
		1, // Created by user ID 1
	)
	if err != nil {
		log.Fatal("Failed to create video:", err)
	}

	fmt.Printf("Successfully created real video: %s (ID: %d)\n", video.Title, video.ID)
	fmt.Printf("Bunny Video ID: %s\n", video.BunnyVideoID)
	fmt.Println("Real video added to database!")
}
