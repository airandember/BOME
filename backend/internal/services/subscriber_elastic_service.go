package services

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"bome-backend/internal/database"
)

// SubscriberElasticService provides unified access to subscriber data across all tables
type SubscriberElasticService struct {
	db *database.DB
}

// NewSubscriberElasticService creates a new elastic service instance
func NewSubscriberElasticService(db *database.DB) *SubscriberElasticService {
	return &SubscriberElasticService{db: db}
}

// GetUnifiedSubscriberByID returns a single subscriber's unified data
// This is optimized for middleware auth checks - uses the same CTE logic as GetAllUnifiedSubscribers
// but with a WHERE clause to fetch only the requested user
func (s *SubscriberElasticService) GetUnifiedSubscriberByID(userID int) (*UnifiedSubscriber, error) {
	log.Printf("🔍 [SubscriberElasticService] Fetching unified data for user %d", userID)

	// Use the same optimized CTE query but filtered for single user
	query := `
		WITH user_subscriptions AS (
			-- Get the most recent active subscription for each user
			SELECT DISTINCT ON (u.id)
				u.id as user_id,
				ss.stripe_id as subscription_id,
				ss.status as subscription_status,
				ss.current_period_start,
				ss.current_period_end,
				ss.created_at as subscription_created_at,
				sp.name as product_name,
				sp.video_approved,
				sp.stripe_id as product_id,
				COALESCE(spr.unit_amount, 0) as product_price,
				COALESCE(spr.currency, 'USD') as product_currency,
				COALESCE(spr.recurring_interval, 'monthly') as product_interval,
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
			LEFT JOIN stripe_customers sc ON (
				u.stripe_customer_id = sc.stripe_id OR 
				sc.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}'))
			)
			LEFT JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
			LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
			LEFT JOIN stripe_prices spr ON sp.id = spr.product_id
			WHERE u.id = $1 AND ss.status IN ('active', 'trialing', 'canceled', 'incomplete', 'incomplete_expired', 'past_due', 'unpaid')
			ORDER BY u.id, ss.created_at DESC
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
			LEFT JOIN user_subscriptions us ON u.id = us.user_id
			WHERE u.id = $1
		)
		SELECT 
			u.id, u.email, u.first_name, u.last_name, u.role, u.email_verified, u.is_active,
			u.created_at, u.last_login, u.stripe_customer_id,
			ARRAY_TO_STRING(u.stripe_customer_ids, ',') as stripe_customer_ids_str,
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
			-- Calculate MRR/ARR contributions (convert from cents to dollars)
			CASE 
				WHEN us.product_interval = 'monthly' AND us.subscription_status IN ('active', 'trialing')
				THEN COALESCE(us.product_price, 0) / 100.0
				WHEN us.product_interval = 'yearly' AND us.subscription_status IN ('active', 'trialing')
				THEN COALESCE(us.product_price, 0) / 100.0 / 12
				WHEN us.product_interval = 'year' AND us.subscription_status IN ('active', 'trialing')
				THEN COALESCE(us.product_price, 0) / 100.0 / 12
				ELSE 0
			END as mrr_contribution,
			CASE 
				WHEN us.product_interval = 'monthly' AND us.subscription_status IN ('active', 'trialing')
				THEN COALESCE(us.product_price, 0) / 100.0 * 12
				WHEN us.product_interval = 'yearly' AND us.subscription_status IN ('active', 'trialing')
				THEN COALESCE(us.product_price, 0) / 100.0
				WHEN us.product_interval = 'year' AND us.subscription_status IN ('active', 'trialing')
				THEN COALESCE(us.product_price, 0) / 100.0
				ELSE 0
			END as arr_contribution,
			-- LTV estimate (simplified: ARR * 2 years, in dollars)
			CASE 
				WHEN us.product_interval = 'monthly' AND us.subscription_status IN ('active', 'trialing')
				THEN COALESCE(us.product_price, 0) / 100.0 * 12 * 2
				WHEN us.product_interval = 'yearly' AND us.subscription_status IN ('active', 'trialing')
				THEN COALESCE(us.product_price, 0) / 100.0 * 2
				WHEN us.product_interval = 'year' AND us.subscription_status IN ('active', 'trialing')
				THEN COALESCE(us.product_price, 0) / 100.0 * 2
				ELSE 0
			END as ltv_estimate,
			EXTRACT(DAYS FROM NOW() - u.created_at)::int as account_age_days,
			us.legacy_status as plan_legacy_status,
			-- Computed fields
			TRIM(CONCAT(u.first_name, ' ', u.last_name)) as full_name,
			CASE 
				WHEN us.current_period_end IS NOT NULL AND us.current_period_end <= NOW() + INTERVAL '30 days'
				THEN true
				ELSE false
			END as is_expiring_soon
		FROM users u
		LEFT JOIN user_subscriptions us ON u.id = us.user_id
		LEFT JOIN user_access ua ON u.id = ua.user_id
		WHERE u.id = $1
	`

	row := s.db.DB.QueryRow(query, userID)

	var sub UnifiedSubscriber
	var (
		stripeCustomerID, subscriptionID, planName, planType, planStatus, planCurrency, planInterval, planLegacyStatus sql.NullString
		planPrice                                                                                                      sql.NullFloat64
		planStartDate, billingPeriodStart, billingPeriodEnd, lastLogin                                                 sql.NullTime
		daysUntilExpiry, accountAgeDays                                                                                sql.NullInt64
		stripeCustomerIDsRaw                                                                                           sql.NullString
		mrrContribution, arrContribution, ltvEstimate                                                                  sql.NullFloat64
	)

	err := row.Scan(
		&sub.ID, &sub.Email, &sub.FirstName, &sub.LastName, &sub.Role, &sub.EmailVerified, &sub.IsActive,
		&sub.CreatedAt, &lastLogin, &stripeCustomerID, &stripeCustomerIDsRaw, &subscriptionID, &planName,
		&planType, &planStatus, &planPrice, &planCurrency, &planInterval, &planStartDate,
		&billingPeriodStart, &billingPeriodEnd, &daysUntilExpiry, &sub.HasActivePlan, &sub.HasVideoAccess,
		&sub.ManualAccessGranted, &mrrContribution, &arrContribution, &ltvEstimate, &accountAgeDays,
		&planLegacyStatus, &sub.FullName, &sub.IsExpiringSoon,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("❌ [SubscriberElasticService] User %d not found", userID)
			return nil, fmt.Errorf("user not found: %d", userID)
		}
		log.Printf("❌ [SubscriberElasticService] Error scanning subscriber %d: %v", userID, err)
		return nil, fmt.Errorf("failed to scan subscriber: %w", err)
	}

	// Assign scanned values to the struct, handling Null types
	sub.LastLogin = nullTimeToPtr(lastLogin)
	sub.StripeCustomerID = nullStringToStringPtr(stripeCustomerID)
	sub.SubscriptionID = nullStringToStringPtr(subscriptionID)
	sub.PlanName = nullStringToStringPtr(planName)
	sub.PlanType = nullStringToString(planType)
	sub.PlanStatus = nullStringToString(planStatus)
	sub.PlanPrice = nullFloat64ToFloat64(planPrice)
	sub.PlanCurrency = nullStringToString(planCurrency)
	sub.PlanInterval = nullStringToString(planInterval)
	sub.PlanStartDate = nullTimeToPtr(planStartDate)
	sub.BillingPeriodStart = nullTimeToPtr(billingPeriodStart)
	sub.BillingPeriodEnd = nullTimeToPtr(billingPeriodEnd)
	sub.DaysUntilExpiry = nullInt64ToIntPtr(daysUntilExpiry)
	sub.MRRContribution = nullFloat64ToFloat64(mrrContribution)
	sub.ARRContribution = nullFloat64ToFloat64(arrContribution)
	sub.LTVEstimate = nullFloat64ToFloat64(ltvEstimate)
	sub.AccountAgeDays = nullInt64ToInt(accountAgeDays)
	sub.PlanLegacyStatus = nullStringToString(planLegacyStatus)

	// Parse stripe_customer_ids from PostgreSQL array format
	if stripeCustomerIDsRaw.Valid && stripeCustomerIDsRaw.String != "" {
		idsStr := strings.Trim(stripeCustomerIDsRaw.String, "{}")
		if idsStr != "" {
			sub.StripeCustomerIDs = strings.Split(idsStr, ",")
			for i, id := range sub.StripeCustomerIDs {
				sub.StripeCustomerIDs[i] = strings.TrimSpace(id)
			}
		} else {
			sub.StripeCustomerIDs = []string{}
		}
	} else {
		sub.StripeCustomerIDs = []string{}
	}

	log.Printf("✅ [SubscriberElasticService] User %d fetched - HasActivePlan: %v, HasVideoAccess: %v, ManualAccess: %v",
		userID, sub.HasActivePlan, sub.HasVideoAccess, sub.ManualAccessGranted)

	return &sub, nil
}

// UnifiedSubscriber represents the complete subscriber data from all sources
type UnifiedSubscriber struct {
	// User data
	ID            int        `json:"id"`
	Email         string     `json:"email"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	Role          string     `json:"role"`
	EmailVerified bool       `json:"email_verified"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	LastLogin     *time.Time `json:"last_login"`

	// Stripe customer data
	StripeCustomerID  *string  `json:"stripe_customer_id"`
	StripeCustomerIDs []string `json:"stripe_customer_ids"`

	// Subscription data (from active subscription)
	SubscriptionID     *string    `json:"subscription_id"`
	PlanName           *string    `json:"plan_name"`
	PlanType           string     `json:"plan_type"`   // "premium", "basic", "none"
	PlanStatus         string     `json:"plan_status"` // "active", "trialing", "expired", etc.
	PlanPrice          float64    `json:"plan_price"`
	PlanCurrency       string     `json:"plan_currency"`
	PlanInterval       string     `json:"plan_interval"` // "monthly", "yearly"
	PlanStartDate      *time.Time `json:"plan_start_date"`
	BillingPeriodStart *time.Time `json:"billing_period_start"`
	BillingPeriodEnd   *time.Time `json:"billing_period_end"`
	DaysUntilExpiry    *int       `json:"days_until_expiry"`

	// Access control
	HasActivePlan       bool `json:"has_active_plan"`
	HasVideoAccess      bool `json:"has_video_access"`
	ManualAccessGranted bool `json:"manual_access_granted"`

	// Business intelligence
	MRRContribution float64 `json:"mrr_contribution"`
	ARRContribution float64 `json:"arr_contribution"`
	LTVEstimate     float64 `json:"ltv_estimate"`
	AccountAgeDays  int     `json:"account_age_days"`

	// Legacy status
	PlanLegacyStatus string `json:"plan_legacy_status"` // "legacy", "current", "unknown"

	// Computed fields
	FullName       string `json:"full_name"`
	IsExpiringSoon bool   `json:"is_expiring_soon"`
}

// GetAllUnifiedSubscribers retrieves all subscribers with complete data
func (s *SubscriberElasticService) GetAllUnifiedSubscribers() ([]UnifiedSubscriber, error) {
	query := `
		WITH user_subscriptions AS (
			-- Get the most recent active subscription for each user
			SELECT DISTINCT ON (u.id)
				u.id as user_id,
				ss.stripe_id as subscription_id,
				ss.status as subscription_status,
				ss.current_period_start,
				ss.current_period_end,
				ss.created_at as subscription_created_at,
				sp.name as product_name,
				sp.video_approved,
				sp.stripe_id as product_id,
				COALESCE(spr.unit_amount, 0) as product_price,
				COALESCE(spr.currency, 'USD') as product_currency,
				COALESCE(spr.recurring_interval, 'monthly') as product_interval,
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
			LEFT JOIN stripe_customers sc ON (
				u.stripe_customer_id = sc.stripe_id OR 
				sc.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}'))
			)
			LEFT JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
			LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
			LEFT JOIN stripe_prices spr ON sp.id = spr.product_id
			WHERE ss.status IN ('active', 'trialing', 'canceled', 'incomplete', 'incomplete_expired', 'past_due', 'unpaid')
			ORDER BY u.id, ss.created_at DESC
		),
		user_access AS (
			-- Determine video access based on subscription and manual overrides
			SELECT 
				u.id as user_id,
				COALESCE(us.subscription_status IN ('active', 'trialing'), false) as has_active_plan,
				COALESCE(
					us.subscription_status IN ('active', 'trialing') AND us.video_approved = true,
					false
				) as has_video_access,
				COALESCE(u.manual_video_access, false) as manual_access_granted
			FROM users u
			LEFT JOIN user_subscriptions us ON u.id = us.user_id
		)
		SELECT 
			u.id,
			u.email,
			u.first_name,
			u.last_name,
			u.role,
			u.email_verified,
			COALESCE(u.is_active, true) as is_active,
			u.created_at,
			u.last_login,
			u.stripe_customer_id,
			COALESCE(u.stripe_customer_ids, '{}') as stripe_customer_ids,
			us.subscription_id,
			us.product_name as plan_name,
			us.plan_type,
			us.subscription_status as plan_status,
			COALESCE(us.product_price, 0) as plan_price,
			COALESCE(us.product_currency, 'USD') as plan_currency,
			us.product_interval as plan_interval,
			us.current_period_start as plan_start_date,
			us.current_period_start as billing_period_start,
			us.current_period_end as billing_period_end,
			CASE 
				WHEN us.current_period_end IS NOT NULL 
				THEN EXTRACT(DAYS FROM us.current_period_end - NOW())::int
				ELSE NULL
			END as days_until_expiry,
			ua.has_active_plan,
			ua.has_video_access,
			ua.manual_access_granted,
			-- Calculate MRR/ARR contributions (convert from cents to dollars)
			CASE 
				WHEN us.product_interval = 'monthly' AND us.subscription_status IN ('active', 'trialing')
				THEN COALESCE(us.product_price, 0) / 100.0
				WHEN us.product_interval = 'yearly' AND us.subscription_status IN ('active', 'trialing')
				THEN COALESCE(us.product_price, 0) / 100.0 / 12
				WHEN us.product_interval = 'year' AND us.subscription_status IN ('active', 'trialing')
				THEN COALESCE(us.product_price, 0) / 100.0 / 12
				ELSE 0
			END as mrr_contribution,
			CASE 
				WHEN us.product_interval = 'monthly' AND us.subscription_status IN ('active', 'trialing')
				THEN COALESCE(us.product_price, 0) / 100.0 * 12
				WHEN us.product_interval = 'yearly' AND us.subscription_status IN ('active', 'trialing')
				THEN COALESCE(us.product_price, 0) / 100.0
				WHEN us.product_interval = 'year' AND us.subscription_status IN ('active', 'trialing')
				THEN COALESCE(us.product_price, 0) / 100.0
				ELSE 0
			END as arr_contribution,
			-- LTV estimate (simplified: ARR * 2 years, in dollars)
			CASE 
				WHEN us.product_interval = 'monthly' AND us.subscription_status IN ('active', 'trialing')
				THEN COALESCE(us.product_price, 0) / 100.0 * 12 * 2
				WHEN us.product_interval = 'yearly' AND us.subscription_status IN ('active', 'trialing')
				THEN COALESCE(us.product_price, 0) / 100.0 * 2
				WHEN us.product_interval = 'year' AND us.subscription_status IN ('active', 'trialing')
				THEN COALESCE(us.product_price, 0) / 100.0 * 2
				ELSE 0
			END as ltv_estimate,
			EXTRACT(DAYS FROM NOW() - u.created_at)::int as account_age_days,
			us.legacy_status as plan_legacy_status,
			-- Computed fields
			TRIM(CONCAT(u.first_name, ' ', u.last_name)) as full_name,
			CASE 
				WHEN us.current_period_end IS NOT NULL AND us.current_period_end <= NOW() + INTERVAL '30 days'
				THEN true
				ELSE false
			END as is_expiring_soon
		FROM users u
		LEFT JOIN user_subscriptions us ON u.id = us.user_id
		LEFT JOIN user_access ua ON u.id = ua.user_id
		ORDER BY u.email
	`

	rows, err := s.db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query unified subscribers: %w", err)
	}
	defer rows.Close()

	var subscribers []UnifiedSubscriber
	for rows.Next() {
		var sub UnifiedSubscriber
		var stripeCustomerIDsRaw sql.NullString
		var lastLogin sql.NullTime
		var subscriptionID, planName, planType, planStatus, planCurrency, planInterval, planLegacyStatus sql.NullString
		var planStartDate, billingPeriodStart, billingPeriodEnd sql.NullTime
		var daysUntilExpiry sql.NullInt32

		err := rows.Scan(
			&sub.ID, &sub.Email, &sub.FirstName, &sub.LastName, &sub.Role,
			&sub.EmailVerified, &sub.IsActive, &sub.CreatedAt, &lastLogin,
			&sub.StripeCustomerID, &stripeCustomerIDsRaw, &subscriptionID,
			&planName, &planType, &planStatus, &sub.PlanPrice,
			&planCurrency, &planInterval, &planStartDate,
			&billingPeriodStart, &billingPeriodEnd, &daysUntilExpiry,
			&sub.HasActivePlan, &sub.HasVideoAccess, &sub.ManualAccessGranted,
			&sub.MRRContribution, &sub.ARRContribution, &sub.LTVEstimate,
			&sub.AccountAgeDays, &planLegacyStatus, &sub.FullName,
			&sub.IsExpiringSoon,
		)
		if err != nil {
			log.Printf("Error scanning subscriber row: %v", err)
			continue
		}

		// Handle nullable fields
		if lastLogin.Valid {
			sub.LastLogin = &lastLogin.Time
		}
		if subscriptionID.Valid {
			sub.SubscriptionID = &subscriptionID.String
		}
		if planName.Valid {
			sub.PlanName = &planName.String
		}
		if planType.Valid {
			sub.PlanType = planType.String
		} else {
			sub.PlanType = "none" // Default value for NULL plan_type
		}
		if planStatus.Valid {
			sub.PlanStatus = planStatus.String
		} else {
			sub.PlanStatus = "none" // Default value for NULL plan_status
		}
		if planCurrency.Valid {
			sub.PlanCurrency = planCurrency.String
		} else {
			sub.PlanCurrency = "USD" // Default value for NULL plan_currency
		}
		if planInterval.Valid {
			sub.PlanInterval = planInterval.String
		} else {
			sub.PlanInterval = "monthly" // Default value for NULL plan_interval
		}
		if planLegacyStatus.Valid {
			sub.PlanLegacyStatus = planLegacyStatus.String
		} else {
			sub.PlanLegacyStatus = "unknown" // Default value for NULL plan_legacy_status
		}
		if planStartDate.Valid {
			sub.PlanStartDate = &planStartDate.Time
		}
		if billingPeriodStart.Valid {
			sub.BillingPeriodStart = &billingPeriodStart.Time
		}
		if billingPeriodEnd.Valid {
			sub.BillingPeriodEnd = &billingPeriodEnd.Time
		}
		if daysUntilExpiry.Valid {
			days := int(daysUntilExpiry.Int32)
			sub.DaysUntilExpiry = &days
		}

		// Parse stripe_customer_ids from PostgreSQL array format
		if stripeCustomerIDsRaw.Valid && stripeCustomerIDsRaw.String != "" {
			// Remove curly braces and split by comma
			idsStr := strings.Trim(stripeCustomerIDsRaw.String, "{}")
			if idsStr != "" {
				sub.StripeCustomerIDs = strings.Split(idsStr, ",")
				// Clean up any whitespace
				for i, id := range sub.StripeCustomerIDs {
					sub.StripeCustomerIDs[i] = strings.TrimSpace(id)
				}
			} else {
				sub.StripeCustomerIDs = []string{}
			}
		} else {
			sub.StripeCustomerIDs = []string{}
		}

		subscribers = append(subscribers, sub)
	}

	log.Printf("✅ Retrieved %d unified subscribers", len(subscribers))
	return subscribers, nil
}

// GetUnifiedSubscriberByEmail retrieves a specific subscriber by email
func (s *SubscriberElasticService) GetUnifiedSubscriberByEmail(email string) (*UnifiedSubscriber, error) {
	subscribers, err := s.GetAllUnifiedSubscribers()
	if err != nil {
		return nil, err
	}

	for _, sub := range subscribers {
		if sub.Email == email {
			return &sub, nil
		}
	}

	return nil, fmt.Errorf("subscriber not found: %s", email)
}

// NOTE: GetUnifiedSubscriberByID is defined at line 26 with optimized WHERE clause

// GetSubscribersWithMultipleStripeCustomers finds users with multiple Stripe customer IDs
func (s *SubscriberElasticService) GetSubscribersWithMultipleStripeCustomers() ([]UnifiedSubscriber, error) {
	subscribers, err := s.GetAllUnifiedSubscribers()
	if err != nil {
		return nil, err
	}

	var multipleCustomers []UnifiedSubscriber
	for _, sub := range subscribers {
		if len(sub.StripeCustomerIDs) > 1 {
			multipleCustomers = append(multipleCustomers, sub)
		}
	}

	log.Printf("🔍 Found %d subscribers with multiple Stripe customer IDs", len(multipleCustomers))
	return multipleCustomers, nil
}

// GetSubscribersWithActivePlansButNoAccess finds subscribers who should have access but don't
func (s *SubscriberElasticService) GetSubscribersWithActivePlansButNoAccess() ([]UnifiedSubscriber, error) {
	subscribers, err := s.GetAllUnifiedSubscribers()
	if err != nil {
		return nil, err
	}

	var noAccess []UnifiedSubscriber
	for _, sub := range subscribers {
		if sub.HasActivePlan && !sub.HasVideoAccess && !sub.ManualAccessGranted {
			noAccess = append(noAccess, sub)
		}
	}

	log.Printf("⚠️ Found %d subscribers with active plans but no video access", len(noAccess))
	return noAccess, nil
}

// GetSubscribersWithVideoAccessButNoPlan finds subscribers with video access but no active plan
func (s *SubscriberElasticService) GetSubscribersWithVideoAccessButNoPlan() ([]UnifiedSubscriber, error) {
	subscribers, err := s.GetAllUnifiedSubscribers()
	if err != nil {
		return nil, err
	}

	var manualAccess []UnifiedSubscriber
	for _, sub := range subscribers {
		if !sub.HasActivePlan && sub.HasVideoAccess {
			manualAccess = append(manualAccess, sub)
		}
	}

	log.Printf("🎬 Found %d subscribers with manual video access (no active plan)", len(manualAccess))
	return manualAccess, nil
}

// UpdateManualVideoAccess updates manual video access for a subscriber
func (s *SubscriberElasticService) UpdateManualVideoAccess(userID int, hasAccess bool) error {
	query := `
		UPDATE users 
		SET manual_video_access = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := s.db.DB.Exec(query, hasAccess, userID)
	if err != nil {
		return fmt.Errorf("failed to update manual video access: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %d", userID)
	}

	log.Printf("✅ Updated manual video access for user %d: %t", userID, hasAccess)
	return nil
}

// GetSubscriberStats returns comprehensive statistics about subscribers
func (s *SubscriberElasticService) GetSubscriberStats() (map[string]interface{}, error) {
	subscribers, err := s.GetAllUnifiedSubscribers()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_subscribers":         len(subscribers),
		"active_plans":              0,
		"video_access":              0,
		"manual_access":             0,
		"expiring_soon":             0,
		"multiple_stripe_customers": 0,
		"total_mrr":                 0.0,
		"total_arr":                 0.0,
		"plan_types": map[string]int{
			"premium": 0,
			"basic":   0,
			"none":    0,
		},
		"legacy_status": map[string]int{
			"current": 0,
			"legacy":  0,
			"unknown": 0,
		},
	}

	for _, sub := range subscribers {
		if sub.HasActivePlan {
			stats["active_plans"] = stats["active_plans"].(int) + 1
		}
		if sub.HasVideoAccess {
			stats["video_access"] = stats["video_access"].(int) + 1
		}
		if sub.ManualAccessGranted {
			stats["manual_access"] = stats["manual_access"].(int) + 1
		}
		if sub.IsExpiringSoon {
			stats["expiring_soon"] = stats["expiring_soon"].(int) + 1
		}
		if len(sub.StripeCustomerIDs) > 1 {
			stats["multiple_stripe_customers"] = stats["multiple_stripe_customers"].(int) + 1
		}

		stats["total_mrr"] = stats["total_mrr"].(float64) + sub.MRRContribution
		stats["total_arr"] = stats["total_arr"].(float64) + sub.ARRContribution

		// Count plan types
		planTypes := stats["plan_types"].(map[string]int)
		planTypes[sub.PlanType]++

		// Count legacy status
		legacyStatus := stats["legacy_status"].(map[string]int)
		legacyStatus[sub.PlanLegacyStatus]++
	}

	return stats, nil
}

// Helper functions for NULL handling in GetUnifiedSubscriberByID
func nullTimeToPtr(t sql.NullTime) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

func nullStringToStringPtr(s sql.NullString) *string {
	if s.Valid {
		return &s.String
	}
	return nil
}

func nullStringToString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func nullFloat64ToFloat64(f sql.NullFloat64) float64 {
	if f.Valid {
		return f.Float64
	}
	return 0.0
}

func nullInt64ToIntPtr(i sql.NullInt64) *int {
	if i.Valid {
		val := int(i.Int64)
		return &val
	}
	return nil
}

func nullInt64ToInt(i sql.NullInt64) int {
	if i.Valid {
		return int(i.Int64)
	}
	return 0
}
