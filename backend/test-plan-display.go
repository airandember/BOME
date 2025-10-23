package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("🔍 TESTING PLAN DISPLAY FIX")
	fmt.Println(strings.Repeat("=", 50))

	// Connect to database
	db, err := sql.Open("postgres", "user=postgres password=postgres dbname=bome_db sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test the query that should return plan information
	query := `
		SELECT 
			sc.stripe_id,
			sc.name,
			sc.email,
			ss.stripe_id as subscription_id,
			ss.status,
			ss.product_name,
			ss.stripe_price_id,
			sp.video_approved
		FROM stripe_customers sc
		LEFT JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
		LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
		WHERE ss.status IN ('active', 'trialing')
		LIMIT 5
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Failed to execute query:", err)
	}
	defer rows.Close()

	fmt.Println("📊 CUSTOMERS WITH ACTIVE SUBSCRIPTIONS:")
	fmt.Println(strings.Repeat("-", 80))

	for rows.Next() {
		var stripeID, name, email, subscriptionID, status, productName, stripePriceID sql.NullString
		var videoApproved sql.NullBool

		err := rows.Scan(&stripeID, &name, &email, &subscriptionID, &status, &productName, &stripePriceID, &videoApproved)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		fmt.Printf("Customer: %s (%s)\n", name.String, email.String)
		fmt.Printf("  Stripe ID: %s\n", stripeID.String)
		fmt.Printf("  Subscription: %s\n", subscriptionID.String)
		fmt.Printf("  Status: %s\n", status.String)
		fmt.Printf("  Plan Name: %s\n", productName.String)
		fmt.Printf("  Price ID: %s\n", stripePriceID.String)
		fmt.Printf("  Video Approved: %t\n", videoApproved.Bool)
		fmt.Println(strings.Repeat("-", 40))
	}

	fmt.Println("✅ Test completed!")
}
