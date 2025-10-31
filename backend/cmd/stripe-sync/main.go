package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"bome-backend/internal/config"
	"bome-backend/internal/database"
	"bome-backend/internal/services"

	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v74"
)

// Helper functions for min/max
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️  No .env file found, using system environment variables")
	}

	// Initialize database first (needed to get Stripe key)
	cfg := config.New()
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize crypto service (needed to decrypt Stripe key)
	cryptoSvc, err := services.NewCryptoServiceFromEnv()
	if err != nil {
		log.Fatalf("❌ Failed to initialize crypto service: %v", err)
	}
	services.SetGlobalCryptoService(cryptoSvc)
	log.Printf("✅ Crypto service initialized")

	// Get encrypted Stripe key from secure_settings table
	var encryptedKey string
	err = db.QueryRow(`
		SELECT value FROM secure_settings 
		WHERE key = 'stripe_secret_key'
	`).Scan(&encryptedKey)

	var stripeKey string
	if err != nil {
		log.Printf("⚠️  Failed to get Stripe key from database: %v", err)
		log.Printf("⚠️  Falling back to environment variable...")
		stripeKey = os.Getenv("STRIPE_SECRET_KEY")
		if stripeKey == "" {
			log.Fatal("❌ STRIPE_SECRET_KEY not found in database or environment")
		}
	} else {
		// Decrypt the key using the crypto service
		crypto := services.GetGlobalCryptoService()
		if crypto == nil {
			log.Fatal("❌ Crypto service not initialized - cannot decrypt Stripe key")
		}

		decryptedKey, err := crypto.DecryptString(encryptedKey)
		if err != nil {
			log.Printf("❌ Failed to decrypt Stripe key: %v", err)
			log.Printf("⚠️  Falling back to environment variable...")
			stripeKey = os.Getenv("STRIPE_SECRET_KEY")
			if stripeKey == "" {
				log.Fatal("❌ Could not decrypt database key and no environment variable set")
			}
		} else {
			stripeKey = decryptedKey
			log.Printf("✅ Stripe key decrypted successfully from database")
		}
	}

	stripe.Key = stripeKey

	log.Println("=" + "================================================================")
	log.Println("🚀 STRIPE V2 SYNC - Syncing all data from Stripe API")
	log.Println("=" + "================================================================")
	log.Println("")

	// Create sync service
	syncService := services.NewStripeSyncV2Service(db)

	// Run full sync
	ctx := context.Background()
	progress, err := syncService.SyncAll(ctx)
	if err != nil {
		log.Fatalf("❌ Sync failed: %v", err)
	}

	// Print summary
	log.Println("")
	log.Println("=" + "================================================================")
	log.Println("📊 SYNC SUMMARY")
	log.Println("=" + "================================================================")
	log.Println("")
	printSummary(progress)

	if len(progress.Errors) > 0 {
		log.Println("")
		log.Println("⚠️  ERRORS ENCOUNTERED:")
		for i, err := range progress.Errors {
			log.Printf("   %d. %s", i+1, err)
		}
		os.Exit(1)
	}

	log.Println("")
	log.Println("✅ Sync completed successfully!")
}

func printSummary(progress *services.SyncProgress) {
	duration := time.Since(progress.StartedAt)

	fmt.Printf("⏱️  Duration: %v\n", duration.Round(time.Millisecond))
	fmt.Println("")

	// Products
	productSuccess := float64(progress.ProductsSynced) / float64(progress.ProductsTotal) * 100
	fmt.Printf("📦 Products:       %d/%d (%.1f%%)\n", progress.ProductsSynced, progress.ProductsTotal, productSuccess)

	// Prices
	priceSuccess := float64(progress.PricesSynced) / float64(progress.PricesTotal) * 100
	fmt.Printf("💰 Prices:         %d/%d (%.1f%%)\n", progress.PricesSynced, progress.PricesTotal, priceSuccess)

	// Customers
	customerSuccess := float64(progress.CustomersSynced) / float64(progress.CustomersTotal) * 100
	fmt.Printf("👥 Customers:      %d/%d (%.1f%%)\n", progress.CustomersSynced, progress.CustomersTotal, customerSuccess)

	// Subscriptions
	subSuccess := float64(progress.SubscriptionsSynced) / float64(progress.SubscriptionsTotal) * 100
	fmt.Printf("📋 Subscriptions:  %d/%d (%.1f%%)\n", progress.SubscriptionsSynced, progress.SubscriptionsTotal, subSuccess)

	// Errors
	if len(progress.Errors) > 0 {
		fmt.Printf("❌ Errors:         %d\n", len(progress.Errors))
	} else {
		fmt.Printf("✅ Errors:         0\n")
	}
}
