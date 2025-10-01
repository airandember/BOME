package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	log.Println("🔧 Starting subscription mapping fix...")

	// Connect to database using environment variable
	databaseURL := "your_database_url_here" // Replace with actual database URL

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✅ Connected to database")

	// Step 1: Build price mapping dynamically from subscription_plans and stripe_prices tables
	log.Println("🔍 Building price mappings from database tables...")
	
	// Query to get subscription plans and their corresponding stripe prices
	planQuery := `
		SELECT 
			sp.id as plan_id,
			sp.name as plan_name,
			sp.stripe_price_id,
			spr.unit_amount
		FROM subscription_plans sp
		LEFT JOIN stripe_prices spr ON sp.stripe_price_id = spr.stripe_id
		WHERE sp.is_active = true
		ORDER BY sp.id
	`
	
	rows, err := db.Query(planQuery)
	if err != nil {
		log.Fatalf("❌ Failed to query subscription plans: %v", err)
	}
	defer rows.Close()
	
	priceMapping := make(map[int64]int)
	planNames := make(map[int]string)
	
	log.Println("💰 Found subscription plans:")
	for rows.Next() {
		var planID int
		var planName string
		var stripePriceID sql.NullString
		var unitAmount sql.NullInt64
		
		err := rows.Scan(&planID, &planName, &stripePriceID, &unitAmount)
		if err != nil {
			log.Printf("❌ Error scanning row: %v", err)
			continue
		}
		
		planNames[planID] = planName
		
		if stripePriceID.Valid && unitAmount.Valid {
			priceMapping[unitAmount.Int64] = planID
			log.Printf("  Plan ID %d (%s): %s -> $%.2f (unit_amount: %d)", 
				planID, planName, stripePriceID.String, float64(unitAmount.Int64)/100, unitAmount.Int64)
		} else {
			log.Printf("  Plan ID %d (%s): No Stripe price mapping found", planID, planName)
		}
	}
	
	if len(priceMapping) == 0 {
		log.Println("⚠️  No price mappings found. Checking for additional price patterns...")
		
		// Fallback: Look for common unit_amounts in stripe_subscriptions and map to reasonable plans
		commonPricesQuery := `
			SELECT DISTINCT unit_amount, COUNT(*) as subscriber_count
			FROM stripe_subscriptions ss
			WHERE ss.status IN ('active', 'trialing')
			AND ss.unit_amount IS NOT NULL
			GROUP BY unit_amount
			ORDER BY subscriber_count DESC
		`
		
		priceRows, err := db.Query(commonPricesQuery)
		if err != nil {
			log.Printf("❌ Failed to query common prices: %v", err)
		} else {
			defer priceRows.Close()
			
			log.Println("🔍 Common subscription prices found:")
			for priceRows.Next() {
				var unitAmount int64
				var count int
				
				err := priceRows.Scan(&unitAmount, &count)
				if err != nil {
					continue
				}
				
				log.Printf("  $%.2f (unit_amount: %d) - %d subscribers", 
					float64(unitAmount)/100, unitAmount, count)
				
				// Smart mapping based on price ranges
				var planID int
				if unitAmount < 1500 { // Less than $15 = Monthly
					planID = 11
				} else { // $15+ = Annual
					planID = 12
				}
				
				priceMapping[unitAmount] = planID
				log.Printf("    → Mapped to Plan ID %d (%s)", planID, planNames[planID])
			}
		}
	}

	// Step 2: Update users.sub_id based on their active Stripe subscriptions
	updateQuery := `
		UPDATE users 
		SET sub_id = $1, updated_at = NOW()
		WHERE id IN (
			SELECT u.id
			FROM users u
			INNER JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
			INNER JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
			WHERE ss.status IN ('active', 'trialing')
			AND ss.unit_amount = $2
			AND u.sub_id IS NULL
		)
	`

	totalUpdated := 0

	for unitAmount, planID := range priceMapping {
		result, err := db.Exec(updateQuery, planID, unitAmount)
		if err != nil {
			log.Printf("❌ Error updating users for unit_amount %d: %v", unitAmount, err)
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		log.Printf("✅ Updated %d users with unit_amount %d to plan_id %d", rowsAffected, unitAmount, planID)
		totalUpdated += int(rowsAffected)
	}

	log.Printf("🎉 Total users updated: %d", totalUpdated)

	// Step 3: Update product_name in stripe_subscriptions based on unit_amount
	productNameMapping := map[int64]string{
		997:  "Basic Monthly",
		999:  "Premium Monthly",
		8982: "Premium Semi-Annual",
		9564: "Premium Annual",
		7200: "Best Value Annual",
	}

	log.Println("📝 Updating product names in stripe_subscriptions...")

	updateProductNameQuery := `
		UPDATE stripe_subscriptions 
		SET product_name = $1, updated_at = NOW()
		WHERE unit_amount = $2 
		AND (product_name IS NULL OR product_name = '')
	`

	for unitAmount, productName := range productNameMapping {
		result, err := db.Exec(updateProductNameQuery, productName, unitAmount)
		if err != nil {
			log.Printf("❌ Error updating product_name for unit_amount %d: %v", unitAmount, err)
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		log.Printf("✅ Updated %d subscriptions with unit_amount %d to product_name '%s'", rowsAffected, unitAmount, productName)
	}

	// Step 4: Verify the fix
	log.Println("🔍 Verifying the fix...")

	verifyQuery := `
		SELECT 
			u.id, u.email, u.sub_id, sp.name as plan_name, ss.unit_amount, ss.product_name
		FROM users u
		LEFT JOIN subscription_plans sp ON u.sub_id = sp.id
		LEFT JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
		LEFT JOIN stripe_subscriptions ss ON sc.id = ss.customer_id AND ss.status IN ('active', 'trialing')
		WHERE u.email = 'mycommonsensefinancial@yahoo.com'
	`

	var userID int
	var email string
	var subID sql.NullInt32
	var planName sql.NullString
	var unitAmount sql.NullInt64
	var productName sql.NullString

	err = db.QueryRow(verifyQuery).Scan(&userID, &email, &subID, &planName, &unitAmount, &productName)
	if err != nil {
		log.Printf("❌ Error verifying fix: %v", err)
	} else {
		log.Printf("🔍 Verification for %s (ID: %d):", email, userID)
		log.Printf("  sub_id: %v", subID)
		log.Printf("  plan_name: %v", planName)
		log.Printf("  unit_amount: %v", unitAmount)
		log.Printf("  product_name: %v", productName)
	}

	log.Println("🎉 Subscription mapping fix completed!")
}
