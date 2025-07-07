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

	// Update video status to 'ready'
	_, err = db.Exec("UPDATE videos SET status = 'ready' WHERE status = 'processing'")
	if err != nil {
		log.Fatal("Failed to update video status:", err)
	}

	fmt.Println("Successfully updated video status to 'ready'")
}
