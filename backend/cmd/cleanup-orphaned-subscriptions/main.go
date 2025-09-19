package main

import (
	"fmt"
	"log"
	"os"

	"bome-backend/internal/config"
	"bome-backend/internal/database"
)

// OrphanedSubscriptionStats holds statistics about orphaned subscriptions
type OrphanedSubscriptionStats struct {
	TotalActiveSubscriptions int `json:"total_active_subscriptions"`
	MissingProductNames      int `json:"missing_product_names"`
	OrphanedProductIDs       int `json:"orphaned_product_ids"`
	ValidProductReferences   int `json:"valid_product_references"`
	TotalProblematicSubs     int `json:"total_problematic_subs"`
}

// OrphanedSubscription represents a subscription with invalid product references
type OrphanedSubscription struct {
	ID               int     `json:"id"`
	StripeID         string  `json:"stripe_id"`
	Status           string  `json:"status"`
	StripeProductID  *string `json:"stripe_product_id"`
	ProductName      *string `json:"product_name"`
	CurrentPeriodEnd *string `json:"current_period_end"`
	CustomerStripeID *string `json:"customer_stripe_id"`
	UserEmail        *string `json:"user_email"`
}

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Check command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: cleanup-orphaned-subscriptions [analyze|cleanup|verify]")
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  analyze  - Show statistics and problematic subscriptions")
		fmt.Println("  cleanup  - Mark orphaned subscriptions as canceled")
		fmt.Println("  verify   - Verify cleanup results")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "analyze":
		analyzeOrphanedSubscriptions(db)
	case "cleanup":
		cleanupOrphanedSubscriptions(db)
	case "verify":
		verifyCleanup(db)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func analyzeOrphanedSubscriptions(db *database.DB) {
	fmt.Println("🔍 ANALYZING ORPHANED STRIPE SUBSCRIPTIONS")
	fmt.Println("==========================================")

	// Get statistics
	stats, err := getOrphanedSubscriptionStats(db)
	if err != nil {
		log.Fatalf("Failed to get statistics: %v", err)
	}

	// Display statistics
	fmt.Printf("📊 STATISTICS:\n")
	fmt.Printf("  Total Active Subscriptions: %d\n", stats.TotalActiveSubscriptions)
	fmt.Printf("  Missing Product Names: %d\n", stats.MissingProductNames)
	fmt.Printf("  Orphaned Product IDs: %d\n", stats.OrphanedProductIDs)
	fmt.Printf("  Valid Product References: %d\n", stats.ValidProductReferences)
	fmt.Printf("  Total Problematic: %d\n", stats.TotalProblematicSubs)
	fmt.Println()

	if stats.TotalProblematicSubs == 0 {
		fmt.Println("✅ No orphaned subscriptions found!")
		return
	}

	// Show problematic subscriptions
	fmt.Printf("⚠️  PROBLEMATIC SUBSCRIPTIONS (showing first 10):\n")
	orphaned, err := getOrphanedSubscriptions(db, 10)
	if err != nil {
		log.Fatalf("Failed to get orphaned subscriptions: %v", err)
	}

	for i, sub := range orphaned {
		fmt.Printf("  %d. ID: %d, Stripe: %s, Status: %s\n", i+1, sub.ID, sub.StripeID, sub.Status)
		if sub.StripeProductID != nil {
			fmt.Printf("     Product ID: %s\n", *sub.StripeProductID)
		}
		if sub.ProductName != nil {
			fmt.Printf("     Product Name: '%s'\n", *sub.ProductName)
		} else {
			fmt.Printf("     Product Name: <NULL>\n")
		}
		if sub.UserEmail != nil {
			fmt.Printf("     User: %s\n", *sub.UserEmail)
		}
		fmt.Println()
	}

	fmt.Printf("💡 RECOMMENDATION:\n")
	fmt.Printf("   Run 'cleanup-orphaned-subscriptions cleanup' to mark these as canceled\n")
	fmt.Printf("   This will preserve the data but exclude them from active plan logic\n")
}

func cleanupOrphanedSubscriptions(db *database.DB) {
	fmt.Println("🧹 CLEANING UP ORPHANED STRIPE SUBSCRIPTIONS")
	fmt.Println("=============================================")

	// Get count before cleanup
	stats, err := getOrphanedSubscriptionStats(db)
	if err != nil {
		log.Fatalf("Failed to get statistics: %v", err)
	}

	if stats.TotalProblematicSubs == 0 {
		fmt.Println("✅ No orphaned subscriptions found to clean up!")
		return
	}

	fmt.Printf("Found %d problematic subscriptions to clean up...\n", stats.TotalProblematicSubs)

	// Perform cleanup - mark as canceled
	query := `
		UPDATE stripe_subscriptions 
		SET 
			status = 'canceled',
			updated_at = NOW()
		WHERE id IN (
			SELECT ss.id
			FROM stripe_subscriptions ss
			LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
			WHERE ss.status IN ('active', 'trialing') 
				AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
				AND (
					-- Missing product name
					(ss.product_name IS NULL OR ss.product_name = '')
					OR 
					-- Orphaned product ID
					(ss.stripe_product_id IS NOT NULL AND sp.stripe_id IS NULL)
				)
		)
	`

	result, err := db.Exec(query)
	if err != nil {
		log.Fatalf("Failed to cleanup subscriptions: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("✅ Successfully marked %d subscriptions as canceled\n", rowsAffected)

	// Verify cleanup
	fmt.Println("\n🔍 Verifying cleanup...")
	verifyCleanup(db)
}

func verifyCleanup(db *database.DB) {
	fmt.Println("🔍 VERIFYING CLEANUP RESULTS")
	fmt.Println("============================")

	stats, err := getOrphanedSubscriptionStats(db)
	if err != nil {
		log.Fatalf("Failed to get statistics: %v", err)
	}

	fmt.Printf("📊 AFTER CLEANUP:\n")
	fmt.Printf("  Total Active Subscriptions: %d\n", stats.TotalActiveSubscriptions)
	fmt.Printf("  Problematic Subscriptions: %d\n", stats.TotalProblematicSubs)

	if stats.TotalProblematicSubs == 0 {
		fmt.Println("✅ Cleanup successful! No more orphaned subscriptions.")
	} else {
		fmt.Printf("⚠️  Still found %d problematic subscriptions\n", stats.TotalProblematicSubs)
	}
}

func getOrphanedSubscriptionStats(db *database.DB) (*OrphanedSubscriptionStats, error) {
	stats := &OrphanedSubscriptionStats{}

	// Total active subscriptions
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM stripe_subscriptions 
		WHERE status IN ('active', 'trialing') 
			AND (current_period_end IS NULL OR current_period_end > NOW())
	`).Scan(&stats.TotalActiveSubscriptions)
	if err != nil {
		return nil, err
	}

	// Missing product names
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM stripe_subscriptions 
		WHERE status IN ('active', 'trialing') 
			AND (current_period_end IS NULL OR current_period_end > NOW())
			AND (product_name IS NULL OR product_name = '')
	`).Scan(&stats.MissingProductNames)
	if err != nil {
		return nil, err
	}

	// Orphaned product IDs
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM stripe_subscriptions ss
		LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
		WHERE ss.status IN ('active', 'trialing') 
			AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
			AND ss.stripe_product_id IS NOT NULL
			AND sp.stripe_id IS NULL
	`).Scan(&stats.OrphanedProductIDs)
	if err != nil {
		return nil, err
	}

	// Valid product references
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM stripe_subscriptions ss
		INNER JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
		WHERE ss.status IN ('active', 'trialing') 
			AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
			AND ss.product_name IS NOT NULL 
			AND ss.product_name != ''
	`).Scan(&stats.ValidProductReferences)
	if err != nil {
		return nil, err
	}

	// Total problematic
	stats.TotalProblematicSubs = stats.MissingProductNames + stats.OrphanedProductIDs

	return stats, nil
}

func getOrphanedSubscriptions(db *database.DB, limit int) ([]OrphanedSubscription, error) {
	query := `
		SELECT 
			ss.id,
			ss.stripe_id,
			ss.status,
			ss.stripe_product_id,
			ss.product_name,
			ss.current_period_end,
			sc.stripe_id as customer_stripe_id,
			u.email as user_email
		FROM stripe_subscriptions ss
		LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
		LEFT JOIN stripe_customers sc ON ss.customer_id = sc.id
		LEFT JOIN users u ON (u.stripe_customer_id = sc.stripe_id OR sc.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}')))
		WHERE ss.status IN ('active', 'trialing') 
			AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
			AND (
				-- Missing product name
				(ss.product_name IS NULL OR ss.product_name = '')
				OR 
				-- Orphaned product ID
				(ss.stripe_product_id IS NOT NULL AND sp.stripe_id IS NULL)
			)
		ORDER BY ss.current_period_end DESC
		LIMIT $1
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []OrphanedSubscription
	for rows.Next() {
		var sub OrphanedSubscription
		err := rows.Scan(
			&sub.ID,
			&sub.StripeID,
			&sub.Status,
			&sub.StripeProductID,
			&sub.ProductName,
			&sub.CurrentPeriodEnd,
			&sub.CustomerStripeID,
			&sub.UserEmail,
		)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}
