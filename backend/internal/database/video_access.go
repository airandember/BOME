package database

import (
	"database/sql"
	"fmt"
	"log"
)

// VideoAccessInfo contains comprehensive video access information
type VideoAccessInfo struct {
	HasStripeAccess bool   `json:"has_stripe_access"`
	HasLegacyAccess bool   `json:"has_legacy_access"`
	HasManualAccess bool   `json:"has_manual_access"`
	AccessSource    string `json:"access_source"` // "stripe", "legacy", "manual", "none"
}

// HasVideoAccess checks if a user has video access from any source
// Returns: hasAccess (bool), accessInfo (*VideoAccessInfo), error
func (db *DB) HasVideoAccess(userID int) (bool, *VideoAccessInfo, error) {
	accessInfo := &VideoAccessInfo{}

	// Check manual video access first (highest priority)
	var manualAccess sql.NullBool
	err := db.QueryRow("SELECT manual_video_access FROM users WHERE id = $1", userID).Scan(&manualAccess)
	if err != nil {
		return false, accessInfo, fmt.Errorf("failed to check manual video access: %w", err)
	}

	accessInfo.HasManualAccess = manualAccess.Valid && manualAccess.Bool
	if accessInfo.HasManualAccess {
		accessInfo.AccessSource = "manual"
		return true, accessInfo, nil
	}

	// Check Stripe subscription with video_approved products (V2 tables)
	var hasStripeAccess bool
	stripeQuery := `
		SELECT EXISTS(
			SELECT 1 
			FROM user_stripe_customers_v2 usc
			INNER JOIN stripe_customers_v2 sc ON sc.id = usc.stripe_customer_id
			INNER JOIN stripe_subscriptions_v2 ss ON ss.customer_id = sc.id
			INNER JOIN stripe_prices_v2 sp ON sp.id = ss.price_id
			INNER JOIN stripe_products_v2 sprod ON sprod.id = sp.product_id
			WHERE usc.user_id = $1
			AND ss.status IN ('active', 'trialing')
			AND ss.current_period_end > NOW()
			AND sprod.video_approved = true
			AND sprod.active = true
		)
	`
	log.Printf("🔍 [HasVideoAccess] Checking Stripe V2 access for user %d", userID)
	err = db.QueryRow(stripeQuery, userID).Scan(&hasStripeAccess)
	if err != nil {
		log.Printf("❌ [HasVideoAccess] Stripe V2 query error for user %d: %v", userID, err)
		
		// Fallback: Try simpler query without video_approved check
		fallbackQuery := `
			SELECT EXISTS(
				SELECT 1 
				FROM user_stripe_customers_v2 usc
				INNER JOIN stripe_customers_v2 sc ON sc.id = usc.stripe_customer_id
				INNER JOIN stripe_subscriptions_v2 ss ON ss.customer_id = sc.id
				WHERE usc.user_id = $1
				AND ss.status IN ('active', 'trialing')
				AND ss.current_period_end > NOW()
			)
		`
		log.Printf("🔄 [HasVideoAccess] Trying V2 fallback query for user %d", userID)
		err = db.QueryRow(fallbackQuery, userID).Scan(&hasStripeAccess)
		if err != nil {
			log.Printf("❌ [HasVideoAccess] V2 fallback query also failed for user %d: %v", userID, err)
			return false, accessInfo, fmt.Errorf("failed to check Stripe video access: %w", err)
		}
		log.Printf("✅ [HasVideoAccess] V2 fallback query succeeded for user %d: %v", userID, hasStripeAccess)
	} else {
		log.Printf("✅ [HasVideoAccess] User %d Stripe V2 access result: %v", userID, hasStripeAccess)
	}

	accessInfo.HasStripeAccess = hasStripeAccess
	if hasStripeAccess {
		accessInfo.AccessSource = "stripe"
		return true, accessInfo, nil
	}

	// Check legacy subscription plans
	var hasLegacyAccess bool
	legacyQuery := `
		SELECT EXISTS(
			SELECT 1 FROM subscriptions s
			INNER JOIN subscription_plans sp ON s.plan_id = sp.id
			WHERE s.user_id = $1
			AND s.status = 'active'
			AND (s.current_period_end IS NULL OR s.current_period_end > NOW())
			AND s.deleted_at IS NULL
		)
	`
	err = db.QueryRow(legacyQuery, userID).Scan(&hasLegacyAccess)
	if err != nil {
		return false, accessInfo, fmt.Errorf("failed to check legacy video access: %w", err)
	}

	accessInfo.HasLegacyAccess = hasLegacyAccess
	if hasLegacyAccess {
		accessInfo.AccessSource = "legacy"
		return true, accessInfo, nil
	}

	// No access found
	accessInfo.AccessSource = "none"
	return false, accessInfo, nil
}
