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

	// Debug the video access query step by step
	fmt.Println("🔍 Debugging video access query for user 7342...")
	
	// Step 1: Check user's stripe_customer_id
	var stripeCustomerID sql.NullString
	err = db.QueryRow("SELECT stripe_customer_id FROM users WHERE id = 7342").Scan(&stripeCustomerID)
	if err != nil {
		log.Fatal("Error getting user's stripe_customer_id:", err)
	}
	fmt.Printf("📊 User 7342 stripe_customer_id: %s\n", stripeCustomerID.String)
	
	// Step 2: Check stripe_customers table
	var customerID int
	err = db.QueryRow("SELECT id FROM stripe_customers WHERE stripe_id = $1", stripeCustomerID.String).Scan(&customerID)
	if err != nil {
		log.Printf("❌ No stripe_customers record found for %s: %v", stripeCustomerID.String, err)
		return
	}
	fmt.Printf("📊 Found stripe_customers record with id: %d\n", customerID)
	
	// Step 3: Check stripe_subscriptions
	var subscriptionCount int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM stripe_subscriptions 
		WHERE customer_id = $1 AND status IN ('active', 'trialing')
	`, customerID).Scan(&subscriptionCount)
	if err != nil {
		log.Fatal("Error checking subscriptions:", err)
	}
	fmt.Printf("📊 Active/trialing subscriptions for customer: %d\n", subscriptionCount)
	
	// Step 4: Check stripe_products with video_approved
	var productCount int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM stripe_subscriptions ss
		INNER JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
		WHERE ss.customer_id = $1 
		AND ss.status IN ('active', 'trialing')
		AND sp.video_approved = true
	`, customerID).Scan(&productCount)
	if err != nil {
		log.Fatal("Error checking video-approved products:", err)
	}
	fmt.Printf("📊 Video-approved products for customer: %d\n", productCount)
	
	// Step 5: Show the actual subscription details
	fmt.Println("\n🔍 Subscription details:")
	rows, err := db.Query(`
		SELECT ss.stripe_id, ss.status, ss.current_period_end, 
		       sp.stripe_id as product_id, sp.name as product_name, sp.video_approved
		FROM stripe_subscriptions ss
		INNER JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
		WHERE ss.customer_id = $1 AND ss.status IN ('active', 'trialing')
	`, customerID)
	if err != nil {
		log.Fatal("Error getting subscription details:", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var subID, status, productID, productName string
		var currentPeriodEnd sql.NullTime
		var videoApproved bool
		
		err := rows.Scan(&subID, &status, &currentPeriodEnd, &productID, &productName, &videoApproved)
		if err != nil {
			log.Printf("Error scanning subscription: %v", err)
			continue
		}
		
		fmt.Printf("  📋 Subscription: %s, Status: %s, Product: %s (%s), Video Approved: %t\n", 
			subID, status, productName, productID, videoApproved)
	}
}
