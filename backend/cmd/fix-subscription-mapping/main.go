package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"bome-backend/internal/config"
	"bome-backend/internal/database"

	_ "github.com/lib/pq"
)

func main() {
	log.Println("🔧 Starting subscription mapping fix...")

	// Load configuration
	cfg := config.Load()
	
	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✅ Connected to database")

	// Step 1: Create price mapping based on stripe_prices table and unit_amount
	// Based on the stripe_prices data provided
	priceMapping := map[int64]int{
		997:   11, // EMonth price -> Monthly plan (plan_id 11)
		999:   11, // premiummonthly price -> Monthly plan (plan_id 11)
		8982:  12, // PPlan price (Legacy Semi-Annual) -> Annual plan (plan_id 12)
		9564:  12, // YPremium price -> Annual plan (plan_id 12) 
		7200:  12, // bestvaluepromo price -> Annual plan (plan_id 12)
	}

	log.Println("💰 Price to Plan ID mapping:")
	for price, planID := range priceMapping {
		log.Printf("  $%.2f (unit_amount: %d) -> Plan ID: %d", float64(price)/100, price, planID)
	}

	// Step 2: Update users.sub_id based on their active Stripe subscriptions
	updateQuery := `
		UPDATE users 
		SET sub_id = $1, updated_at = NOW()
		WHERE id IN (
			SELECT u.id
			FROM users u
			INNER JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
			INNER JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
			WHERE ss.status IN ('active', 'trialing')
			AND ss.unit_amount = $2
			AND u.sub_id IS NULL
		)
	`

	totalUpdated := 0
	
	for unitAmount, planID := range priceMapping {
		result, err := db.Exec(updateQuery, planID, unitAmount)
		if err != nil {
			log.Printf("❌ Error updating users for unit_amount %d: %v", unitAmount, err)
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		log.Printf("✅ Updated %d users with unit_amount %d to plan_id %d", rowsAffected, unitAmount, planID)
		totalUpdated += int(rowsAffected)
	}

	log.Printf("🎉 Total users updated: %d", totalUpdated)

	// Step 3: Update product_name in stripe_subscriptions based on unit_amount
	productNameMapping := map[int64]string{
		997:   "Basic Monthly",
		999:   "Premium Monthly",
		8982:  "Premium Semi-Annual",
		9564:  "Premium Annual", 
		7200:  "Best Value Annual",
	}

	log.Println("📝 Updating product names in stripe_subscriptions...")

	updateProductNameQuery := `
		UPDATE stripe_subscriptions 
		SET product_name = $1, updated_at = NOW()
		WHERE unit_amount = $2 
		AND (product_name IS NULL OR product_name = '')
	`

	for unitAmount, productName := range productNameMapping {
		result, err := db.Exec(updateProductNameQuery, productName, unitAmount)
		if err != nil {
			log.Printf("❌ Error updating product_name for unit_amount %d: %v", unitAmount, err)
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		log.Printf("✅ Updated %d subscriptions with unit_amount %d to product_name '%s'", rowsAffected, unitAmount, productName)
	}

	// Step 4: Verify the fix
	log.Println("🔍 Verifying the fix...")
	
	verifyQuery := `
		SELECT 
			u.id, u.email, u.sub_id, sp.name as plan_name, ss.unit_amount, ss.product_name
		FROM users u
		LEFT JOIN subscription_plans sp ON u.sub_id = sp.id
		LEFT JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
		LEFT JOIN stripe_subscriptions ss ON sc.id = ss.customer_id AND ss.status IN ('active', 'trialing')
		WHERE u.email = 'mycommonsensefinancial@yahoo.com'
	`

	var userID int
	var email string
	var subID sql.NullInt32
	var planName sql.NullString
	var unitAmount sql.NullInt64
	var productName sql.NullString

	err = db.QueryRow(verifyQuery).Scan(&userID, &email, &subID, &planName, &unitAmount, &productName)
	if err != nil {
		log.Printf("❌ Error verifying fix: %v", err)
	} else {
		log.Printf("🔍 Verification for %s (ID: %d):", email, userID)
		log.Printf("  sub_id: %v", subID)
		log.Printf("  plan_name: %v", planName)
		log.Printf("  unit_amount: %v", unitAmount)
		log.Printf("  product_name: %v", productName)
	}

	log.Println("🎉 Subscription mapping fix completed!")
}
