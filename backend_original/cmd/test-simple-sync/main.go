package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"bome-backend/internal/config"
	"bome-backend/internal/database"
	"bome-backend/internal/services"
)

func main() {
	fmt.Println("🧪 Testing Simple Stripe Sync...")

	// Load configuration
	cfg := config.New()
	if cfg == nil {
		log.Fatal("Failed to load configuration")
	}

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Stripe service
	stripeService := services.NewStripeService(db)

	// Create simple sync service
	simpleSync := services.NewSimpleStripeSyncService(db, stripeService)

	// Test the sync
	ctx := context.Background()
	fmt.Println("🚀 Starting simple Stripe sync test...")

	err = simpleSync.SyncAll(ctx)
	if err != nil {
		fmt.Printf("❌ Sync failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Simple Stripe sync completed successfully!")
}
