package main

import (
	"database/sql"
	"fmt"
	"log"
	_ "github.com/lib/pq"
)

func main() {
	// Connect to database
	db, err := sql.Open("postgres", "user=postgres password=postgres dbname=bome_db sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test the video access query
	fmt.Println("🔍 Testing video access data...")
	
	// Check stripe_subscriptions with stripe_product_id
	var totalSubs, withProductID int
	err = db.QueryRow(`
		SELECT COUNT(*) as total_subs, COUNT(stripe_product_id) as with_product_id 
		FROM stripe_subscriptions 
		WHERE status IN ('active', 'trialing')
	`).Scan(&totalSubs, &withProductID)
	if err != nil {
		log.Fatal("Error querying subscriptions:", err)
	}
	
	fmt.Printf("📊 Active/Trialing Subscriptions: %d total, %d with stripe_product_id\n", totalSubs, withProductID)
	
	// Check if stripe_product_id is NULL/empty
	var nullProductID int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM stripe_subscriptions 
		WHERE status IN ('active', 'trialing') 
		AND (stripe_product_id IS NULL OR stripe_product_id = '')
	`).Scan(&nullProductID)
	if err != nil {
		log.Fatal("Error querying null product IDs:", err)
	}
	
	fmt.Printf("❌ Subscriptions with NULL/empty stripe_product_id: %d\n", nullProductID)
	
	// Test the actual video access query for a specific user
	fmt.Println("\n🔍 Testing video access query...")
	var hasAccess bool
	err = db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM users u
			INNER JOIN stripe_customers sc ON (
				u.stripe_customer_id = sc.stripe_id OR 
				sc.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}'))
			)
			INNER JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
			INNER JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
			WHERE u.id = 7342
			AND ss.status IN ('active', 'trialing')
			AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
			AND sp.video_approved = true
		)
	`).Scan(&hasAccess)
	if err != nil {
		log.Printf("❌ Video access query error: %v", err)
	} else {
		fmt.Printf("✅ User 7342 has video access: %t\n", hasAccess)
	}
}
