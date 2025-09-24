package main

import (
	"fmt"
	"log"
	"time"

	"bome-backend/internal/config"
	"bome-backend/internal/database"
)

func main() {
	fmt.Println("🕰️ BOME Legacy Products Update")
	fmt.Println("===============================")

	// Load configuration
	cfg := config.New()

	// Connect to database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Update products older than 2 years to be legacy
	cutoffDate := time.Now().AddDate(-2, 0, 0)

	fmt.Printf("🔍 Marking products created before %s as legacy...\n", cutoffDate.Format("2006-01-02"))

	query := `
		UPDATE stripe_products 
		SET legacy_product = true 
		WHERE created_at < $1 AND legacy_product = false
	`

	result, err := db.DB.Exec(query, cutoffDate)
	if err != nil {
		log.Fatalf("❌ Failed to update legacy products: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("⚠️ Could not get rows affected: %v", err)
	} else {
		fmt.Printf("✅ Updated %d products to legacy status\n", rowsAffected)
	}

	// Show current legacy products
	fmt.Println("\n📋 Current Legacy Products:")
	showQuery := `
		SELECT id, stripe_id, name, created_at, legacy_product
		FROM stripe_products 
		WHERE legacy_product = true
		ORDER BY created_at ASC
	`

	rows, err := db.DB.Query(showQuery)
	if err != nil {
		log.Printf("❌ Failed to query legacy products: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id int
		var stripeID, name string
		var createdAt time.Time
		var legacyProduct bool

		err := rows.Scan(&id, &stripeID, &name, &createdAt, &legacyProduct)
		if err != nil {
			log.Printf("❌ Error scanning row: %v", err)
			continue
		}

		fmt.Printf("   %d. %s (%s) - Created: %s\n",
			id, name, stripeID, createdAt.Format("2006-01-02"))
		count++
	}

	fmt.Printf("\n🎉 Total legacy products: %d\n", count)
}
