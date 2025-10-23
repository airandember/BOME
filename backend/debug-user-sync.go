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

	// Check user 7342 details
	fmt.Println("🔍 Checking user 7342 details...")
	
	var email, firstName, lastName string
	var stripeCustomerID sql.NullString
	err = db.QueryRow("SELECT email, first_name, last_name, stripe_customer_id FROM users WHERE id = 7342").Scan(&email, &firstName, &lastName, &stripeCustomerID)
	if err != nil {
		log.Fatal("Error getting user details:", err)
	}
	fmt.Printf("📊 User: %s %s (%s), stripe_customer_id: '%s'\n", firstName, lastName, email, stripeCustomerID.String)
	
	// Check if there's a Stripe customer with this email
	var stripeCustomerStripeID sql.NullString
	err = db.QueryRow("SELECT stripe_id FROM stripe_customers WHERE email = $1", email).Scan(&stripeCustomerStripeID)
	if err != nil {
		log.Printf("❌ No stripe_customers record found for email %s: %v", email, err)
	} else {
		fmt.Printf("✅ Found stripe_customers record: %s\n", stripeCustomerStripeID.String)
		
		// Check if this customer has active subscriptions
		var customerID int
		err = db.QueryRow("SELECT id FROM stripe_customers WHERE stripe_id = $1", stripeCustomerStripeID.String).Scan(&customerID)
		if err != nil {
			log.Fatal("Error getting customer ID:", err)
		}
		
		var subscriptionCount int
		err = db.QueryRow(`
			SELECT COUNT(*) 
			FROM stripe_subscriptions 
			WHERE customer_id = $1 AND status IN ('active', 'trialing')
		`, customerID).Scan(&subscriptionCount)
		if err != nil {
			log.Fatal("Error checking subscriptions:", err)
		}
		fmt.Printf("📊 Active/trialing subscriptions: %d\n", subscriptionCount)
		
		// Show subscription details
		if subscriptionCount > 0 {
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
	}
}
