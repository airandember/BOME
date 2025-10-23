package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=bome_db sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("🔍 DIAGNOSING STRIPE SUBSCRIPTION PLAN ISSUES")
	fmt.Println(strings.Repeat("=", 50))

	// Check stripe_subscriptions table structure
	fmt.Println("\n📋 STRIPE_SUBSCRIPTIONS TABLE STRUCTURE:")
	rows, err := db.Query(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'stripe_subscriptions' 
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Printf("Error getting table structure: %v", err)
	} else {
		for rows.Next() {
			var colName, dataType, nullable string
			rows.Scan(&colName, &dataType, &nullable)
			fmt.Printf("  %-20s %-15s %s\n", colName, dataType, nullable)
		}
		rows.Close()
	}

	// Check sample stripe_subscriptions data
	fmt.Println("\n📊 SAMPLE STRIPE_SUBSCRIPTIONS DATA:")
	rows, err = db.Query(`
		SELECT 
			stripe_id, 
			customer_id, 
			status, 
			stripe_price_id, 
			stripe_product_id, 
			product_name,
			unit_amount,
			currency
		FROM stripe_subscriptions 
		WHERE status IN ('active', 'trialing')
		LIMIT 5
	`)
	if err != nil {
		log.Printf("Error getting subscription data: %v", err)
	} else {
		fmt.Printf("%-20s %-12s %-10s %-20s %-20s %-20s %-12s %-8s\n",
			"stripe_id", "customer_id", "status", "stripe_price_id", "stripe_product_id", "product_name", "unit_amount", "currency")
		fmt.Println(strings.Repeat("-", 120))

		for rows.Next() {
			var stripeID, customerID, status, priceID, productID, productName, currency sql.NullString
			var unitAmount sql.NullInt64

			rows.Scan(&stripeID, &customerID, &status, &priceID, &productID, &productName, &unitAmount, &currency)

			fmt.Printf("%-20s %-12s %-10s %-20s %-20s %-20s %-12d %-8s\n",
				getString(stripeID), getString(customerID), getString(status),
				getString(priceID), getString(productID), getString(productName),
				getInt64(unitAmount), getString(currency))
		}
		rows.Close()
	}

	// Check stripe_products table
	fmt.Println("\n🏷️ STRIPE_PRODUCTS TABLE:")
	rows, err = db.Query(`
		SELECT stripe_id, name, video_approved 
		FROM stripe_products 
		LIMIT 5
	`)
	if err != nil {
		log.Printf("Error getting products data: %v", err)
	} else {
		fmt.Printf("%-20s %-30s %-15s\n", "stripe_id", "name", "video_approved")
		fmt.Println(strings.Repeat("-", 65))

		for rows.Next() {
			var stripeID, name sql.NullString
			var videoApproved sql.NullBool

			rows.Scan(&stripeID, &name, &videoApproved)
			fmt.Printf("%-20s %-30s %-15t\n", getString(stripeID), getString(name), getBool(videoApproved))
		}
		rows.Close()
	}

	// Check subscription_plans table
	fmt.Println("\n📋 SUBSCRIPTION_PLANS TABLE:")
	rows, err = db.Query(`
		SELECT id, name, stripe_price_id, stripe_product_id, price, interval
		FROM subscription_plans 
		WHERE deleted_at IS NULL
		LIMIT 5
	`)
	if err != nil {
		log.Printf("Error getting subscription plans: %v", err)
	} else {
		fmt.Printf("%-5s %-30s %-20s %-20s %-10s %-10s\n", "id", "name", "stripe_price_id", "stripe_product_id", "price", "interval")
		fmt.Println(strings.Repeat("-", 95))

		for rows.Next() {
			var id int
			var name, priceID, productID, interval sql.NullString
			var price sql.NullFloat64

			rows.Scan(&id, &name, &priceID, &productID, &price, &interval)
			fmt.Printf("%-5d %-30s %-20s %-20s %-10.2f %-10s\n",
				id, getString(name), getString(priceID), getString(productID), getFloat64(price), getString(interval))
		}
		rows.Close()
	}

	// Check for missing product_name in stripe_subscriptions
	fmt.Println("\n❌ MISSING PRODUCT_NAMES IN STRIPE_SUBSCRIPTIONS:")
	var missingCount int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM stripe_subscriptions 
		WHERE product_name IS NULL OR product_name = ''
	`).Scan(&missingCount)
	if err != nil {
		log.Printf("Error counting missing product names: %v", err)
	} else {
		fmt.Printf("  %d subscriptions have missing product_name\n", missingCount)
	}

	// Check for missing stripe_product_id in stripe_subscriptions
	fmt.Println("\n❌ MISSING STRIPE_PRODUCT_ID IN STRIPE_SUBSCRIPTIONS:")
	var missingProductIDCount int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM stripe_subscriptions 
		WHERE stripe_product_id IS NULL OR stripe_product_id = ''
	`).Scan(&missingProductIDCount)
	if err != nil {
		log.Printf("Error counting missing product IDs: %v", err)
	} else {
		fmt.Printf("  %d subscriptions have missing stripe_product_id\n", missingProductIDCount)
	}

	fmt.Println("\n✅ DIAGNOSIS COMPLETE")
}

func getString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return "NULL"
}

func getInt64(i sql.NullInt64) int64 {
	if i.Valid {
		return i.Int64
	}
	return 0
}

func getFloat64(f sql.NullFloat64) float64 {
	if f.Valid {
		return f.Float64
	}
	return 0.0
}

func getBool(b sql.NullBool) bool {
	if b.Valid {
		return b.Bool
	}
	return false
}
