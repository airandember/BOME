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

	// Check Stripe subscription with video_approved products
	var hasStripeAccess bool
	stripeQuery := `
		SELECT EXISTS(
			SELECT 1 FROM stripe_subscriptions ss
			INNER JOIN stripe_prices sp_price ON CAST(ss.price_id AS TEXT) = CAST(sp_price.id AS TEXT)
			INNER JOIN stripe_products sp ON CAST(sp_price.product_id AS TEXT) = CAST(sp.id AS TEXT)
			WHERE ss.customer_id = (
				SELECT stripe_customer_id FROM users WHERE id = $1
			)
			AND ss.status IN ('active', 'trialing')
			AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
			AND sp.video_approved = true
		)
	`
	log.Printf("🔍 [HasVideoAccess] Checking Stripe access for user %d", userID)
	err = db.QueryRow(stripeQuery, userID).Scan(&hasStripeAccess)
	if err != nil {
		log.Printf("❌ [HasVideoAccess] Stripe query error for user %d: %v", userID, err)
		return false, accessInfo, fmt.Errorf("failed to check Stripe video access: %w", err)
	}
	log.Printf("🔍 [HasVideoAccess] User %d Stripe access result: %v", userID, hasStripeAccess)

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
