package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"bome-backend/internal/config"
	"bome-backend/internal/database"
)

type DuplicateCustomer struct {
	Email         string   `json:"email"`
	UserID        int      `json:"user_id"`
	CustomerCount int      `json:"customer_count"`
	CustomerIDs   []string `json:"customer_ids"`
}

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("🔍 Duplicate Stripe Customers by Email Report")
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("")

	// Initialize database
	cfg := config.New()
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Find users with multiple Stripe customers
	query := `
		SELECT 
			u.id as user_id,
			u.email,
			COUNT(DISTINCT sc.id) as customer_count,
			ARRAY_AGG(DISTINCT sc.stripe_id ORDER BY sc.stripe_id) as customer_ids
		FROM users u
		JOIN user_stripe_customers_v2 usc ON u.id = usc.user_id
		JOIN stripe_customers_v2 sc ON usc.stripe_customer_id = sc.id
		GROUP BY u.id, u.email
		HAVING COUNT(DISTINCT sc.id) > 1
		ORDER BY customer_count DESC, u.email
		LIMIT 100
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("❌ Failed to query: %v", err)
	}
	defer rows.Close()

	var duplicates []DuplicateCustomer
	var totalDuplicateCustomers int

	for rows.Next() {
		var dup DuplicateCustomer
		var customerIDsArray sql.NullString

		err := rows.Scan(&dup.UserID, &dup.Email, &dup.CustomerCount, &customerIDsArray)
		if err != nil {
			log.Printf("⚠️  Failed to scan row: %v", err)
			continue
		}

		// Parse the PostgreSQL array (simple parsing)
		if customerIDsArray.Valid {
			// Remove braces and split
			ids := customerIDsArray.String
			if len(ids) > 2 {
				ids = ids[1 : len(ids)-1] // Remove { and }
			}
			// For now, store as single string (you could parse further)
			dup.CustomerIDs = []string{ids}
		}

		duplicates = append(duplicates, dup)
		totalDuplicateCustomers += (dup.CustomerCount - 1) // Subtract 1 because one is legitimate
	}

	// Display results
	fmt.Printf("📊 Found %d users with multiple Stripe customer IDs\n", len(duplicates))
	fmt.Printf("💰 Total duplicate customer records: %d\n", totalDuplicateCustomers)
	fmt.Println("")

	if len(duplicates) > 0 {
		fmt.Println("🔴 TOP 10 USERS WITH MOST CUSTOMER IDs:")
		fmt.Println("")

		for i, dup := range duplicates {
			if i >= 10 {
				break
			}
			fmt.Printf("%d. User %d - %s\n", i+1, dup.UserID, dup.Email)
			fmt.Printf("   Customer Count: %d\n", dup.CustomerCount)
			fmt.Printf("   Customer IDs: %v\n", dup.CustomerIDs)
			fmt.Println("")
		}
	}

	// Save to JSON
	jsonData, err := json.MarshalIndent(map[string]interface{}{
		"total_users_affected":    len(duplicates),
		"total_duplicate_records": totalDuplicateCustomers,
		"users":                   duplicates,
	}, "", "  ")
	if err != nil {
		log.Fatalf("❌ Failed to marshal JSON: %v", err)
	}

	err = os.WriteFile("duplicate-customers-report.json", jsonData, 0644)
	if err != nil {
		log.Fatalf("❌ Failed to write report: %v", err)
	}

	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("✅ Report Generated!")
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("")
	fmt.Printf("📄 Report saved to: duplicate-customers-report.json\n")
	fmt.Println("")

	if len(duplicates) > 0 {
		fmt.Println("⚠️  ISSUE IDENTIFIED:")
		fmt.Println("   Your Stripe checkout is creating new customer IDs")
		fmt.Println("   for each subscription attempt instead of reusing")
		fmt.Println("   existing customers by email.")
		fmt.Println("")
		fmt.Println("🔧 SOLUTION:")
		fmt.Println("   1. Update checkout flow to search for existing customer by email")
		fmt.Println("   2. Reuse existing customer ID when found")
		fmt.Println("   3. Only create new customer if email not found")
		fmt.Println("")
		fmt.Println("📋 CLEANUP:")
		fmt.Printf("   - %d users need customer consolidation\n", len(duplicates))
		fmt.Printf("   - %d duplicate customer records can be archived\n", totalDuplicateCustomers)
	} else {
		fmt.Println("✅ No duplicate customers found!")
	}
}
