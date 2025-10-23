package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func main() {
	// Try different database names that might exist
	dbNames := []string{"bome", "bome_db", "bome_streaming", "bome_production", "postgres"}

	for _, dbName := range dbNames {
		fmt.Printf("🔍 Trying database: %s\n", dbName)

		db, err := sql.Open("postgres", fmt.Sprintf("host=localhost port=5432 user=postgres password=postgres dbname=%s sslmode=disable", dbName))
		if err != nil {
			fmt.Printf("❌ Failed to connect to %s: %v\n", dbName, err)
			continue
		}

		// Test connection
		err = db.Ping()
		if err != nil {
			fmt.Printf("❌ Failed to ping %s: %v\n", dbName, err)
			db.Close()
			continue
		}

		// Check if stripe_subscriptions table exists
		var tableExists bool
		err = db.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables 
				WHERE table_name = 'stripe_subscriptions'
			)
		`).Scan(&tableExists)

		if err != nil {
			fmt.Printf("❌ Error checking tables in %s: %v\n", dbName, err)
			db.Close()
			continue
		}

		if tableExists {
			fmt.Printf("✅ Found stripe_subscriptions table in database: %s\n", dbName)

			// Get sample data
			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM stripe_subscriptions").Scan(&count)
			if err != nil {
				fmt.Printf("❌ Error counting subscriptions: %v\n", err)
			} else {
				fmt.Printf("📊 Found %d stripe_subscriptions\n", count)
			}

			db.Close()
			break
		} else {
			fmt.Printf("❌ No stripe_subscriptions table in %s\n", dbName)
			db.Close()
		}
	}
}
