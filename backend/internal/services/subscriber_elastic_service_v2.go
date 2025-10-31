package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"bome-backend/internal/database"
)

// SubscriberElasticServiceV2 provides unified access to subscriber data using v2 tables
// This is the NEW implementation that uses the v2 Stripe schema with proper foreign keys
type SubscriberElasticServiceV2 struct {
	db *database.DB
}

// NewSubscriberElasticServiceV2 creates a new v2 elastic service instance
func NewSubscriberElasticServiceV2(db *database.DB) *SubscriberElasticServiceV2 {
	return &SubscriberElasticServiceV2{db: db}
}

// UnifiedSubscriberV2 represents the unified subscriber data model for v2
// This matches the v1 structure but pulls from v2 tables
type UnifiedSubscriberV2 struct {
	// User Fields
	ID            int     `json:"id"`
	Email         string  `json:"email"`
	FirstName     *string `json:"first_name"`
	LastName      *string `json:"last_name"`
	FullName      string  `json:"full_name"`
	Role          string  `json:"role"`
	EmailVerified bool    `json:"email_verified"`
	IsActive      bool    `json:"is_active"`
	CreatedAt     string  `json:"created_at"`
	LastLogin     *string `json:"last_login"`

	// Stripe Customer Fields
	StripeCustomerID  *string  `json:"stripe_customer_id"`
	StripeCustomerIDs []string `json:"stripe_customer_ids"`
	PrimaryCustomerID *string  `json:"primary_customer_id"` // NEW: From user_stripe_customers_v2

	// Subscription Fields
	SubscriptionID     *string `json:"subscription_id"`
	PlanName           *string `json:"plan_name"`
	PlanType           string  `json:"plan_type"`
	PlanStatus         string  `json:"plan_status"`
	PlanPrice          *int    `json:"plan_price"`
	PlanCurrency       string  `json:"plan_currency"`
	PlanInterval       string  `json:"plan_interval"`
	PlanLegacyStatus   string  `json:"plan_legacy_status"`
	PlanStartDate      *string `json:"plan_start_date"`
	BillingPeriodStart *string `json:"billing_period_start"`
	BillingPeriodEnd   *string `json:"billing_period_end"`
	DaysUntilExpiry    *int    `json:"days_until_expiry"`

	// Calculated Fields
	HasActivePlan       bool    `json:"has_active_plan"`
	HasVideoAccess      bool    `json:"has_video_access"`
	ManualAccessGranted bool    `json:"manual_access_granted"`
	IsExpiringSoon      bool    `json:"is_expiring_soon"`
	MRRContribution     float64 `json:"mrr_contribution"`
	ARRContribution     float64 `json:"arr_contribution"`
	LTVEstimate         float64 `json:"ltv_estimate"`
	AccountAgeDays      int     `json:"account_age_days"`
}

// GetUnifiedSubscriberByIDV2 returns a single subscriber's unified data using v2 tables
// This uses the customer linking table to ensure we only use the PRIMARY customer's subscription
func (s *SubscriberElasticServiceV2) GetUnifiedSubscriberByIDV2(userID int) (*UnifiedSubscriberV2, error) {
	log.Printf("🔍 [SubscriberElasticServiceV2] Fetching unified data for user %d", userID)

	query := `
		WITH user_primary_customer AS (
			-- Get the user's PRIMARY Stripe customer from the linking table
			SELECT 
				usc.user_id,
				usc.stripe_customer_id,
				sc.stripe_id as stripe_customer_stripe_id
			FROM user_stripe_customers_v2 usc
			JOIN stripe_customers_v2 sc ON sc.id = usc.stripe_customer_id
			WHERE usc.user_id = $1 AND usc.is_primary = true
		),
		v2_subscriptions AS (
			-- Get the most recent subscription for the PRIMARY customer only
			SELECT DISTINCT ON (u.id)
				u.id as user_id,
				ss.stripe_id as subscription_id,
				ss.status as subscription_status,
				ss.current_period_start,
				ss.current_period_end,
				ss.stripe_created_at as subscription_created_at,
				sp.name as product_name,
				sp.video_approved,
				sp.stripe_id as product_id,
				spr.unit_amount as product_price,
				spr.currency as product_currency,
				spr.recurring_interval as product_interval,
				CASE 
					WHEN sp.name ILIKE '%premium%' THEN 'premium'
					WHEN sp.name ILIKE '%basic%' THEN 'basic'
					ELSE 'none'
				END as plan_type,
				CASE 
					WHEN ss.status IN ('active', 'trialing') THEN 'current'
					WHEN ss.status IN ('canceled', 'incomplete', 'incomplete_expired', 'past_due', 'unpaid') THEN 'legacy'
					ELSE 'unknown'
				END as legacy_status
		FROM users u
		JOIN user_primary_customer upc ON upc.user_id = u.id
		JOIN stripe_customers_v2 sc ON sc.id = upc.stripe_customer_id
		JOIN stripe_subscriptions_v2 ss ON ss.customer_id = sc.id
		LEFT JOIN stripe_prices_v2 spr ON ss.price_id = spr.id
		LEFT JOIN stripe_products_v2 sp ON spr.product_id = sp.id
			WHERE u.id = $1 AND ss.status IN ('active', 'trialing', 'canceled', 'incomplete', 'incomplete_expired', 'past_due', 'unpaid')
			ORDER BY u.id, ss.stripe_created_at DESC
		),
		user_access AS (
			-- Determine video access based on subscription and manual overrides
			SELECT 
				u.id as user_id,
				CASE 
					WHEN u.manual_video_access = true THEN true
					WHEN us.subscription_status IN ('active', 'trialing') AND us.video_approved = true THEN true
					ELSE false
				END as has_video_access,
				u.manual_video_access as manual_access_granted,
				CASE 
					WHEN us.subscription_status IN ('active', 'trialing') THEN true
					ELSE false
				END as has_active_plan
			FROM users u
			LEFT JOIN v2_subscriptions us ON u.id = us.user_id
		)
		SELECT 
			u.id, u.email, u.first_name, u.last_name, u.role, u.email_verified, u.is_active,
			u.created_at, u.last_login, upc.stripe_customer_stripe_id,
			upc.stripe_customer_stripe_id as primary_customer_id,
			us.subscription_id, us.product_name as plan_name,
			us.plan_type, us.subscription_status as plan_status,
			us.product_price as plan_price, us.product_currency as plan_currency,
			us.product_interval as plan_interval, us.current_period_start as plan_start_date,
			us.current_period_start as billing_period_start, us.current_period_end as billing_period_end,
			CASE 
				WHEN us.current_period_end IS NOT NULL 
				THEN EXTRACT(DAYS FROM us.current_period_end - NOW())::int
				ELSE NULL
			END as days_until_expiry,
			ua.has_active_plan,
			ua.has_video_access,
			ua.manual_access_granted,
			us.legacy_status,
			EXTRACT(DAYS FROM NOW() - u.created_at)::int as account_age_days
		FROM users u
		LEFT JOIN user_primary_customer upc ON upc.user_id = u.id
		LEFT JOIN v2_subscriptions us ON u.id = us.user_id
		LEFT JOIN user_access ua ON u.id = ua.user_id
		WHERE u.id = $1
	`

	var sub UnifiedSubscriberV2
	var firstName, lastName sql.NullString
	var lastLogin, stripeCustomerID, primaryCustomerID sql.NullString
	var subscriptionID, planName sql.NullString
	var planPrice sql.NullInt64
	var planCurrency, planInterval, planType, planStatus, planLegacyStatus sql.NullString
	var planStartDate, billingPeriodStart, billingPeriodEnd sql.NullTime
	var daysUntilExpiry sql.NullInt64
	var createdAt time.Time

	err := s.db.QueryRow(query, userID).Scan(
		&sub.ID, &sub.Email, &firstName, &lastName, &sub.Role, &sub.EmailVerified, &sub.IsActive,
		&createdAt, &lastLogin, &stripeCustomerID,
		&primaryCustomerID,
		&subscriptionID, &planName,
		&planType, &planStatus,
		&planPrice, &planCurrency,
		&planInterval, &planStartDate,
		&billingPeriodStart, &billingPeriodEnd,
		&daysUntilExpiry,
		&sub.HasActivePlan,
		&sub.HasVideoAccess,
		&sub.ManualAccessGranted,
		&planLegacyStatus,
		&sub.AccountAgeDays,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %d", userID)
		}
		return nil, fmt.Errorf("failed to fetch user data: %w", err)
	}

	// Handle NULL values
	sub.FirstName = nullStringToStringPtrV2(firstName)
	sub.LastName = nullStringToStringPtrV2(lastName)
	sub.LastLogin = nullStringToStringPtrV2(lastLogin)
	sub.StripeCustomerID = nullStringToStringPtrV2(stripeCustomerID)
	sub.PrimaryCustomerID = nullStringToStringPtrV2(primaryCustomerID)
	sub.SubscriptionID = nullStringToStringPtrV2(subscriptionID)
	sub.PlanName = nullStringToStringPtrV2(planName)
	sub.PlanType = nullStringToStringV2(planType, "none")
	sub.PlanStatus = nullStringToStringV2(planStatus, "none")
	sub.PlanCurrency = nullStringToStringV2(planCurrency, "USD")
	sub.PlanInterval = nullStringToStringV2(planInterval, "monthly")
	sub.PlanLegacyStatus = nullStringToStringV2(planLegacyStatus, "unknown")

	if planPrice.Valid {
		price := int(planPrice.Int64)
		sub.PlanPrice = &price
	}

	if daysUntilExpiry.Valid {
		days := int(daysUntilExpiry.Int64)
		sub.DaysUntilExpiry = &days
		sub.IsExpiringSoon = days > 0 && days <= 30
	}

	// Format timestamps
	sub.CreatedAt = createdAt.Format(time.RFC3339)
	sub.PlanStartDate = nullTimeToStringPtrV2(planStartDate)
	sub.BillingPeriodStart = nullTimeToStringPtrV2(billingPeriodStart)
	sub.BillingPeriodEnd = nullTimeToStringPtrV2(billingPeriodEnd)

	// Build full name
	sub.FullName = buildFullNameV2(sub.FirstName, sub.LastName)

	// Calculate financial metrics
	if sub.PlanPrice != nil {
		priceInDollars := float64(*sub.PlanPrice) / 100.0

		switch sub.PlanInterval {
		case "month":
			sub.MRRContribution = priceInDollars
			sub.ARRContribution = priceInDollars * 12
		case "year":
			sub.MRRContribution = priceInDollars / 12
			sub.ARRContribution = priceInDollars
		case "quarter":
			sub.MRRContribution = priceInDollars / 3
			sub.ARRContribution = priceInDollars * 4
		default:
			sub.MRRContribution = priceInDollars
			sub.ARRContribution = priceInDollars * 12
		}

		// LTV estimate: ARR * 2 years (simple estimate)
		sub.LTVEstimate = sub.ARRContribution * 2
	}

	// Get all customer IDs (from legacy users.stripe_customer_ids array)
	if stripeCustomerID.Valid {
		// Query for all linked customers
		rows, err := s.db.Query(`
			SELECT sc.stripe_id
			FROM user_stripe_customers_v2 usc
			JOIN stripe_customers_v2 sc ON sc.id = usc.stripe_customer_id
			WHERE usc.user_id = $1
			ORDER BY usc.is_primary DESC, sc.created_at DESC
		`, userID)
		if err == nil {
			defer rows.Close()
			var customerIDs []string
			for rows.Next() {
				var cid string
				if err := rows.Scan(&cid); err == nil {
					customerIDs = append(customerIDs, cid)
				}
			}
			sub.StripeCustomerIDs = customerIDs
		}
	}

	log.Printf("✅ [SubscriberElasticServiceV2] Successfully fetched user %d (primary customer: %v)",
		userID, sub.PrimaryCustomerID)

	return &sub, nil
}

// GetAllUnifiedSubscribersV2 returns all subscribers' unified data using v2 tables
func (s *SubscriberElasticServiceV2) GetAllUnifiedSubscribersV2() ([]UnifiedSubscriberV2, error) {
	log.Printf("🔍 [SubscriberElasticServiceV2] Fetching all unified subscriber data")

	query := `
		WITH user_primary_customers AS (
			-- Get each user's PRIMARY Stripe customer
			SELECT 
				usc.user_id,
				usc.stripe_customer_id,
				sc.stripe_id as stripe_customer_stripe_id
			FROM user_stripe_customers_v2 usc
			JOIN stripe_customers_v2 sc ON sc.id = usc.stripe_customer_id
			WHERE usc.is_primary = true
		),
		v2_subscriptions AS (
			-- Get the most recent subscription for each user's PRIMARY customer
			SELECT DISTINCT ON (u.id)
				u.id as user_id,
				ss.stripe_id as subscription_id,
				ss.status as subscription_status,
				ss.current_period_start,
				ss.current_period_end,
				ss.stripe_created_at,
				sp.name as product_name,
				sp.video_approved,
				spr.unit_amount as product_price,
				spr.currency as product_currency,
				spr.recurring_interval as product_interval,
				CASE 
					WHEN sp.name ILIKE '%premium%' THEN 'premium'
					WHEN sp.name ILIKE '%basic%' THEN 'basic'
					ELSE 'none'
				END as plan_type,
				CASE 
					WHEN ss.status IN ('active', 'trialing') THEN 'current'
					WHEN ss.status IN ('canceled', 'incomplete', 'incomplete_expired', 'past_due', 'unpaid') THEN 'legacy'
					ELSE 'unknown'
				END as legacy_status
			FROM users u
			LEFT JOIN user_primary_customers upc ON upc.user_id = u.id
			LEFT JOIN stripe_customers_v2 sc ON sc.id = upc.stripe_customer_id
		LEFT JOIN stripe_subscriptions_v2 ss ON ss.customer_id = sc.id
		LEFT JOIN stripe_prices_v2 spr ON ss.price_id = spr.id
		LEFT JOIN stripe_products_v2 sp ON spr.product_id = sp.id
			WHERE ss.status IS NULL OR ss.status IN ('active', 'trialing', 'canceled', 'incomplete', 'incomplete_expired', 'past_due', 'unpaid')
			ORDER BY u.id, ss.stripe_created_at DESC
		),
		user_access AS (
			SELECT 
				u.id as user_id,
				CASE 
					WHEN u.manual_video_access = true THEN true
					WHEN us.subscription_status IN ('active', 'trialing') AND us.video_approved = true THEN true
					ELSE false
				END as has_video_access,
				u.manual_video_access as manual_access_granted,
				CASE 
					WHEN us.subscription_status IN ('active', 'trialing') THEN true
					ELSE false
				END as has_active_plan
			FROM users u
			LEFT JOIN v2_subscriptions us ON u.id = us.user_id
		)
		SELECT 
			u.id, u.email, u.first_name, u.last_name, u.role, u.email_verified, u.is_active,
			u.created_at, u.last_login, upc.stripe_customer_stripe_id,
			upc.stripe_customer_stripe_id as primary_customer_id,
			us.subscription_id, us.product_name,
			us.plan_type, us.subscription_status,
			us.product_price, us.product_currency,
			us.product_interval, us.current_period_start,
			us.current_period_start as billing_period_start, us.current_period_end as billing_period_end,
			CASE 
				WHEN us.current_period_end IS NOT NULL 
				THEN EXTRACT(DAYS FROM us.current_period_end - NOW())::int
				ELSE NULL
			END as days_until_expiry,
			ua.has_active_plan,
			ua.has_video_access,
			ua.manual_access_granted,
			us.legacy_status,
			EXTRACT(DAYS FROM NOW() - u.created_at)::int as account_age_days
		FROM users u
		LEFT JOIN user_primary_customers upc ON upc.user_id = u.id
		LEFT JOIN v2_subscriptions us ON u.id = us.user_id
		LEFT JOIN user_access ua ON u.id = ua.user_id
		ORDER BY u.id
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscribers: %w", err)
	}
	defer rows.Close()

	var subscribers []UnifiedSubscriberV2
	for rows.Next() {
		var sub UnifiedSubscriberV2
		var firstName, lastName sql.NullString
		var lastLogin, stripeCustomerID, primaryCustomerID sql.NullString
		var subscriptionID, planName sql.NullString
		var planPrice sql.NullInt64
		var planCurrency, planInterval, planType, planStatus, planLegacyStatus sql.NullString
		var planStartDate, billingPeriodStart, billingPeriodEnd sql.NullTime
		var daysUntilExpiry sql.NullInt64
		var createdAt time.Time

		err := rows.Scan(
			&sub.ID, &sub.Email, &firstName, &lastName, &sub.Role, &sub.EmailVerified, &sub.IsActive,
			&createdAt, &lastLogin, &stripeCustomerID,
			&primaryCustomerID,
			&subscriptionID, &planName,
			&planType, &planStatus,
			&planPrice, &planCurrency,
			&planInterval, &planStartDate,
			&billingPeriodStart, &billingPeriodEnd,
			&daysUntilExpiry,
			&sub.HasActivePlan,
			&sub.HasVideoAccess,
			&sub.ManualAccessGranted,
			&planLegacyStatus,
			&sub.AccountAgeDays,
		)

		if err != nil {
			log.Printf("⚠️  [SubscriberElasticServiceV2] Failed to scan row: %v", err)
			continue
		}

		// Handle NULL values
		sub.FirstName = nullStringToStringPtrV2(firstName)
		sub.LastName = nullStringToStringPtrV2(lastName)
		sub.LastLogin = nullStringToStringPtrV2(lastLogin)
		sub.StripeCustomerID = nullStringToStringPtrV2(stripeCustomerID)
		sub.PrimaryCustomerID = nullStringToStringPtrV2(primaryCustomerID)
		sub.SubscriptionID = nullStringToStringPtrV2(subscriptionID)
		sub.PlanName = nullStringToStringPtrV2(planName)
		sub.PlanType = nullStringToStringV2(planType, "none")
		sub.PlanStatus = nullStringToStringV2(planStatus, "none")
		sub.PlanCurrency = nullStringToStringV2(planCurrency, "USD")
		sub.PlanInterval = nullStringToStringV2(planInterval, "monthly")
		sub.PlanLegacyStatus = nullStringToStringV2(planLegacyStatus, "unknown")

		if planPrice.Valid {
			price := int(planPrice.Int64)
			sub.PlanPrice = &price
		}

		if daysUntilExpiry.Valid {
			days := int(daysUntilExpiry.Int64)
			sub.DaysUntilExpiry = &days
			sub.IsExpiringSoon = days > 0 && days <= 30
		}

		// Format timestamps
		sub.CreatedAt = createdAt.Format(time.RFC3339)
		sub.PlanStartDate = nullTimeToStringPtrV2(planStartDate)
		sub.BillingPeriodStart = nullTimeToStringPtrV2(billingPeriodStart)
		sub.BillingPeriodEnd = nullTimeToStringPtrV2(billingPeriodEnd)

		// Build full name
		sub.FullName = buildFullNameV2(sub.FirstName, sub.LastName)

		// Calculate financial metrics
		if sub.PlanPrice != nil {
			priceInDollars := float64(*sub.PlanPrice) / 100.0

			switch sub.PlanInterval {
			case "month":
				sub.MRRContribution = priceInDollars
				sub.ARRContribution = priceInDollars * 12
			case "year":
				sub.MRRContribution = priceInDollars / 12
				sub.ARRContribution = priceInDollars
			case "quarter":
				sub.MRRContribution = priceInDollars / 3
				sub.ARRContribution = priceInDollars * 4
			default:
				sub.MRRContribution = priceInDollars
				sub.ARRContribution = priceInDollars * 12
			}

			sub.LTVEstimate = sub.ARRContribution * 2
		}

		subscribers = append(subscribers, sub)
	}

	log.Printf("✅ [SubscriberElasticServiceV2] Retrieved %d subscribers", len(subscribers))
	return subscribers, nil
}

// GetSubscriberStatsV2 returns aggregated statistics using v2 tables
func (s *SubscriberElasticServiceV2) GetSubscriberStatsV2() (map[string]interface{}, error) {
	log.Printf("📊 [SubscriberElasticServiceV2] Calculating subscriber statistics")

	stats := make(map[string]interface{})

	// Total users
	var totalUsers int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}
	stats["total_users"] = totalUsers

	// Users with linked customers
	var usersWithCustomers int
	err = s.db.QueryRow(`
		SELECT COUNT(DISTINCT user_id) 
		FROM user_stripe_customers_v2
	`).Scan(&usersWithCustomers)
	if err != nil {
		return nil, fmt.Errorf("failed to count users with customers: %w", err)
	}
	stats["users_with_linked_customers"] = usersWithCustomers

	// Active subscriptions (from PRIMARY customers only)
	var activeSubscriptions int
	err = s.db.QueryRow(`
		SELECT COUNT(DISTINCT ss.id)
		FROM stripe_subscriptions_v2 ss
		JOIN stripe_customers_v2 sc ON sc.id = ss.stripe_customer_id
		JOIN user_stripe_customers_v2 usc ON usc.stripe_customer_id = sc.id
		WHERE usc.is_primary = true
		AND ss.status IN ('active', 'trialing')
	`).Scan(&activeSubscriptions)
	if err != nil {
		return nil, fmt.Errorf("failed to count active subscriptions: %w", err)
	}
	stats["active_subscriptions"] = activeSubscriptions

	// Manual access grants
	var manualAccessCount int
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM users WHERE manual_video_access = true
	`).Scan(&manualAccessCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count manual access: %w", err)
	}
	stats["manual_access_grants"] = manualAccessCount

	// Total MRR/ARR (from PRIMARY customers only)
	var totalMRR, totalARR float64
	err = s.db.QueryRow(`
		SELECT 
			SUM(CASE 
				WHEN spr.recurring_interval = 'month' THEN spr.unit_amount / 100.0
				WHEN spr.recurring_interval = 'year' THEN (spr.unit_amount / 100.0) / 12
				WHEN spr.recurring_interval = 'quarter' THEN (spr.unit_amount / 100.0) / 3
				ELSE spr.unit_amount / 100.0
			END) as total_mrr,
			SUM(CASE 
				WHEN spr.recurring_interval = 'month' THEN (spr.unit_amount / 100.0) * 12
				WHEN spr.recurring_interval = 'year' THEN spr.unit_amount / 100.0
				WHEN spr.recurring_interval = 'quarter' THEN (spr.unit_amount / 100.0) * 4
				ELSE (spr.unit_amount / 100.0) * 12
			END) as total_arr
	FROM stripe_subscriptions_v2 ss
	JOIN stripe_prices_v2 spr ON ss.price_id = spr.id
	JOIN stripe_customers_v2 sc ON sc.id = ss.customer_id
		JOIN user_stripe_customers_v2 usc ON usc.stripe_customer_id = sc.id
		WHERE usc.is_primary = true
		AND ss.status IN ('active', 'trialing')
	`).Scan(&totalMRR, &totalARR)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to calculate MRR/ARR: %w", err)
	}
	stats["total_mrr"] = totalMRR
	stats["total_arr"] = totalARR

	// Users by plan status
	var activeCount, trialingCount, canceledCount int
	s.db.QueryRow(`
		SELECT 
			COUNT(CASE WHEN ss.status = 'active' THEN 1 END) as active,
			COUNT(CASE WHEN ss.status = 'trialing' THEN 1 END) as trialing,
			COUNT(CASE WHEN ss.status = 'canceled' THEN 1 END) as canceled
		FROM stripe_subscriptions_v2 ss
		JOIN stripe_customers_v2 sc ON sc.id = ss.stripe_customer_id
		JOIN user_stripe_customers_v2 usc ON usc.stripe_customer_id = sc.id
		WHERE usc.is_primary = true
	`).Scan(&activeCount, &trialingCount, &canceledCount)

	stats["by_status"] = map[string]int{
		"active":   activeCount,
		"trialing": trialingCount,
		"canceled": canceledCount,
	}

	log.Printf("✅ [SubscriberElasticServiceV2] Statistics calculated successfully")
	return stats, nil
}

// Helper functions for v2 service (with V2 suffix to avoid conflicts with v1)
func nullTimeToStringPtrV2(t sql.NullTime) *string {
	if !t.Valid {
		return nil
	}
	s := t.Time.Format(time.RFC3339)
	return &s
}

func nullStringToStringPtrV2(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

func nullStringToStringV2(ns sql.NullString, defaultValue string) string {
	if !ns.Valid {
		return defaultValue
	}
	return ns.String
}

func buildFullNameV2(firstName, lastName *string) string {
	if firstName == nil && lastName == nil {
		return ""
	}
	if firstName == nil {
		return *lastName
	}
	if lastName == nil {
		return *firstName
	}
	return *firstName + " " + *lastName
}
