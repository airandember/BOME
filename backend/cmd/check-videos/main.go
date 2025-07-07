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

	// Check if videos table exists and has data
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM videos").Scan(&count)
	if err != nil {
		log.Fatal("Failed to count videos:", err)
	}

	fmt.Printf("Total videos in database: %d\n", count)

	// Get all videos regardless of status
	rows, err := db.Query("SELECT id, title, status, bunny_video_id FROM videos ORDER BY created_at DESC")
	if err != nil {
		log.Fatal("Failed to query videos:", err)
	}
	defer rows.Close()

	fmt.Println("\nAll videos in database:")
	for rows.Next() {
		var id int
		var title, status, bunnyVideoID string
		if err := rows.Scan(&id, &title, &status, &bunnyVideoID); err != nil {
			log.Fatal("Failed to scan video:", err)
		}
		fmt.Printf("ID: %d, Title: %s, Status: %s, Bunny ID: %s\n", id, title, status, bunnyVideoID)
	}

	// Test the GetVideos function directly
	fmt.Println("\nTesting GetVideos function:")
	videos, err := db.GetVideos(10, 0, "", "ready")
	if err != nil {
		fmt.Printf("Error calling GetVideos: %v\n", err)
	} else {
		fmt.Printf("GetVideos returned %d videos\n", len(videos))
		for _, video := range videos {
			fmt.Printf("- %s (ID: %d, Status: %s)\n", video.Title, video.ID, video.Status)
		}
	}
}
