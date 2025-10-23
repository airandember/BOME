package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	// Connect to database
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=bome_db sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("🔍 Checking product_name in stripe_subscriptions for active subscriptions...")

	// Query active subscriptions with their product names
	query := `
		SELECT 
			ss.stripe_id,
			ss.status,
			ss.product_name,
			ss.stripe_product_id,
			sc.stripe_id as customer_stripe_id,
			sc.email
		FROM stripe_subscriptions ss
		JOIN stripe_customers sc ON ss.customer_id = sc.id
		WHERE ss.status = 'active'
		ORDER BY ss.created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Failed to query subscriptions:", err)
	}
	defer rows.Close()

	fmt.Println("\n📊 Active Subscriptions with Product Names:")
	fmt.Println(strings.Repeat("=", 80))

	for rows.Next() {
		var stripeID, status, productName, stripeProductID, customerStripeID, email sql.NullString

		err := rows.Scan(&stripeID, &status, &productName, &stripeProductID, &customerStripeID, &email)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		fmt.Printf("Customer: %s (%s)\n", email.String, customerStripeID.String)
		fmt.Printf("  Subscription: %s\n", stripeID.String)
		fmt.Printf("  Status: %s\n", status.String)
		fmt.Printf("  Product Name: %s\n", productName.String)
		fmt.Printf("  Stripe Product ID: %s\n", stripeProductID.String)
		fmt.Println()
	}

	// Also check if there are any NULL product_name values
	fmt.Println("🔍 Checking for NULL product_name values...")

	var nullCount int
	err = db.QueryRow("SELECT COUNT(*) FROM stripe_subscriptions WHERE product_name IS NULL").Scan(&nullCount)
	if err != nil {
		log.Printf("Error counting NULL product_name: %v", err)
	} else {
		fmt.Printf("Subscriptions with NULL product_name: %d\n", nullCount)
	}

	var totalCount int
	err = db.QueryRow("SELECT COUNT(*) FROM stripe_subscriptions").Scan(&totalCount)
	if err != nil {
		log.Printf("Error counting total subscriptions: %v", err)
	} else {
		fmt.Printf("Total subscriptions: %d\n", totalCount)
	}
}
