package services

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"bome-backend/internal/database"
)

// SubscriberService handles business logic for subscribers (users with active subscriptions)
type SubscriberService struct {
	db *database.DB
}

// Subscriber represents a user with an active subscription
type Subscriber struct {
	ID                   int        `json:"id"`
	Email                string     `json:"email"`
	FirstName            string     `json:"first_name"`
	LastName             string     `json:"last_name"`
	Role                 string     `json:"role"`
	EmailVerified        bool       `json:"email_verified"`
	PlanID               *int       `json:"plan_id,omitempty"`
	PlanName             *string    `json:"plan_name,omitempty"`
	PlanPrice            *float64   `json:"plan_price,omitempty"`
	PlanCurrency         *string    `json:"plan_currency,omitempty"`
	PlanInterval         *string    `json:"plan_interval,omitempty"`
	PlanIntervalCount    *int       `json:"plan_interval_count,omitempty"`
	SubscriptionID       *int       `json:"subscription_id,omitempty"`
	SubID                *int       `json:"sub_id,omitempty"` // Alias for subscription_id
	SubscriptionStatus   *string    `json:"subscription_status,omitempty"`
	CurrentPeriodStart   *time.Time `json:"current_period_start,omitempty"`
	CurrentPeriodEnd     *time.Time `json:"current_period_end,omitempty"`
	StripeCustomerID     *string    `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID *string    `json:"stripe_subscription_id,omitempty"`
	LastLogin            *time.Time `json:"last_login,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// SubscriberStats represents subscriber statistics
type SubscriberStats struct {
	TotalSubscribers      int     `json:"total_subscribers"`
	ActiveSubscribers     int     `json:"active_subscribers"`
	TrialingSubscribers   int     `json:"trialing_subscribers"`
	PastDueSubscribers    int     `json:"past_due_subscribers"`
	CanceledSubscribers   int     `json:"canceled_subscribers"`
	MonthlyRevenue        float64 `json:"monthly_revenue"`
	AnnualRevenue         float64 `json:"annual_revenue"`
	AverageRevenuePerUser float64 `json:"average_revenue_per_user"`
	ChurnRate             float64 `json:"churn_rate"`
}

// SubscriberFilters represents filters for subscriber queries
type SubscriberFilters struct {
	PlanID        *int       `json:"plan_id"`
	Status        *string    `json:"status"`
	Search        string     `json:"search"`
	EmailVerified *bool      `json:"email_verified"`
	Role          *string    `json:"role"`
	LastLogin     *time.Time `json:"last_login"`
	CreatedDate   *time.Time `json:"created_date"`
	DateRange     *struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	} `json:"date_range"`
	HasSubscriptionHistory *bool `json:"has_subscription_history"` // true, false, or nil for all
}

// NewSubscriberService creates a new subscriber service
func NewSubscriberService(db *database.DB) *SubscriberService {
	return &SubscriberService{db: db}
}

// GetSubscribers retrieves all subscribers with optional filters
func (s *SubscriberService) GetSubscribers(limit, offset int, filters *SubscriberFilters) ([]*Subscriber, error) {
	log.Printf("🔍 GetSubscribers called with limit=%d, offset=%d", limit, offset)

	// UNIFIED PLANS APPROACH: Treat subscription_plans and stripe_products as one dataset
	// This eliminates complex COALESCE logic and provides cleaner plan management
	query := `
		WITH unified_plans AS (
			-- Legacy subscription plans
			SELECT 
				'legacy' as plan_source,
				sp.id::text as plan_id,
				sp.name as plan_name,
				sp.price as plan_price,
				sp.currency as plan_currency,
				sp.interval as plan_interval,
				sp.interval_count as plan_interval_count,
				sp.is_active as is_active,
				sp.created_at as created_at,
				sp.updated_at as updated_at,
				NULL as stripe_product_id,
				NULL as stripe_price_id
			FROM subscription_plans sp
			WHERE sp.is_active = true 
			  AND sp.deleted_at IS NULL
			
			UNION ALL
			
			-- Stripe products as plans (with pricing from stripe_prices)
			SELECT 
				'stripe' as plan_source,
				stripe_prod.stripe_id as plan_id,
				stripe_prod.name as plan_name,
				CASE WHEN stripe_price.unit_amount IS NOT NULL 
					THEN stripe_price.unit_amount::float / 100.0 
					ELSE 0.0 
				END as plan_price,
				COALESCE(stripe_price.currency, 'USD') as plan_currency,
				COALESCE(stripe_price.recurring_interval, 'month') as plan_interval,
				COALESCE(stripe_price.recurring_interval_count, 1) as plan_interval_count,
				stripe_prod.active as is_active,
				stripe_prod.created_at as created_at,
				stripe_prod.updated_at as updated_at,
				stripe_prod.stripe_id as stripe_product_id,
				stripe_price.stripe_id as stripe_price_id
			FROM stripe_products stripe_prod
			LEFT JOIN stripe_prices stripe_price ON stripe_prod.stripe_id = stripe_price.product_id
			WHERE stripe_prod.active = true
		),
		
		user_plans AS (
			SELECT 
				u.id as user_id,
				u.email,
				u.first_name,
				u.last_name,
				u.role,
				u.email_verified,
				u.stripe_customer_id,
				u.last_login,
				u.created_at as user_created_at,
				u.updated_at as user_updated_at,
				
				-- Plan information from unified plans
				up.plan_source,
				up.plan_id,
				up.plan_name,
				up.plan_price,
				up.plan_currency,
				up.plan_interval,
				up.plan_interval_count,
				
				-- Subscription status information
				CASE 
					WHEN up.plan_source = 'legacy' THEN 'active'
					ELSE COALESCE(ss.status, 'active')
				END as subscription_status,
				
				-- Subscription periods (only for Stripe)
				ss.current_period_start,
				ss.current_period_end,
				ss.stripe_id as stripe_subscription_id,
				
				-- Priority for DISTINCT ON ordering (legacy plans first)
				CASE 
					WHEN up.plan_source = 'legacy' THEN 1  -- Legacy plans highest priority
					WHEN up.plan_source = 'stripe' AND ss.stripe_price_id IS NOT NULL THEN 2
					WHEN up.plan_source = 'stripe' AND up.plan_name IS NOT NULL THEN 3
					ELSE 4
				END as plan_priority
				
			FROM users u
			
			-- Join with unified plans via legacy subscription
			LEFT JOIN unified_plans up ON (
				(up.plan_source = 'legacy' AND u.sub_id::text = up.plan_id) OR
				(up.plan_source = 'stripe' AND EXISTS (
					SELECT 1 FROM stripe_customers sc2 
					JOIN stripe_subscriptions ss2 ON sc2.id = ss2.customer_id
					WHERE (u.stripe_customer_id = sc2.stripe_id OR sc2.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}')))
					  AND ss2.stripe_product_id = up.stripe_product_id
					  AND ss2.status IN ('active', 'trialing')
					  AND (ss2.current_period_end IS NULL OR ss2.current_period_end > NOW())
				))
			)
			
			-- Join Stripe subscription details for Stripe plans
			LEFT JOIN stripe_customers sc ON (
				up.plan_source = 'stripe' AND 
				(u.stripe_customer_id = sc.stripe_id OR sc.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}')))
			)
			LEFT JOIN stripe_subscriptions ss ON (
				up.plan_source = 'stripe' AND 
				sc.id = ss.customer_id AND 
				ss.stripe_product_id = up.stripe_product_id AND
				ss.status IN ('active', 'trialing') AND
				(ss.current_period_end IS NULL OR ss.current_period_end > NOW())
			)
			
			WHERE up.plan_id IS NOT NULL  -- Only users with plans
			  AND u.is_active = true
		)
		
		-- Final query with unified plan data
		SELECT DISTINCT ON (user_id)
			user_id as id,
			email,
			first_name,
			last_name,
			role,
			email_verified,
			stripe_customer_id,
			last_login,
			user_created_at as created_at,
			user_updated_at as updated_at,
			
			-- Unified plan information
			plan_id as subscription_id,
			plan_id,
			plan_name,
			plan_price,
			plan_currency,
			plan_interval,
			plan_interval_count,
			subscription_status,
			current_period_start,
			current_period_end,
			stripe_subscription_id
			
		FROM user_plans
		WHERE 1=1  -- Placeholder for dynamic filters
		ORDER BY user_id, plan_priority ASC  -- Priority ensures legacy plans come first
	`

	args := []interface{}{}
	argCount := 0

	// Add filters
	if filters != nil {
		// Note: PlanID filter removed since we don't have stripe_prices table join

		if filters.Status != nil {
			argCount++
			query += fmt.Sprintf(" AND ss.status = $%d", argCount)
			args = append(args, *filters.Status)
		}

		if filters.Search != "" {
			argCount++
			query += fmt.Sprintf(" AND (u.email ILIKE $%d OR u.first_name ILIKE $%d OR u.last_name ILIKE $%d)",
				argCount, argCount, argCount)
			searchTerm := "%" + filters.Search + "%"
			args = append(args, searchTerm)
		}

		if filters.EmailVerified != nil {
			argCount++
			query += fmt.Sprintf(" AND u.email_verified = $%d", argCount)
			args = append(args, *filters.EmailVerified)
		}

		if filters.Role != nil {
			fmt.Printf("DEBUG: Processing role filter: %s\n", *filters.Role)
			argCount++
			query += fmt.Sprintf(" AND u.role = $%d", argCount)
			args = append(args, *filters.Role)
			fmt.Printf("DEBUG: Role filter added to query\n")
		}

		if filters.LastLogin != nil {
			fmt.Printf("DEBUG: Processing last login filter: %v\n", *filters.LastLogin)
			if filters.LastLogin.IsZero() {
				// Special case for "never" - users who have never logged in
				query += " AND u.last_login IS NULL"
				fmt.Printf("DEBUG: Last login filter added (IS NULL)\n")
			} else {
				argCount++
				query += fmt.Sprintf(" AND u.last_login >= $%d", argCount)
				args = append(args, *filters.LastLogin)
				fmt.Printf("DEBUG: Last login filter added (>= %v)\n", *filters.LastLogin)
			}
		}

		if filters.CreatedDate != nil {
			fmt.Printf("DEBUG: Processing created date filter: %v\n", *filters.CreatedDate)
			argCount++
			query += fmt.Sprintf(" AND u.created_at >= $%d", argCount)
			args = append(args, *filters.CreatedDate)
			fmt.Printf("DEBUG: Created date filter added (>= %v)\n", *filters.CreatedDate)
		}

		if filters.DateRange != nil {
			argCount++
			query += fmt.Sprintf(" AND u.created_at BETWEEN $%d AND $%d", argCount, argCount+1)
			args = append(args, filters.DateRange.Start, filters.DateRange.End)
			argCount++
		}

		if filters.HasSubscriptionHistory != nil {
			fmt.Printf("DEBUG: Processing subscription history filter: %v\n", *filters.HasSubscriptionHistory)
			if *filters.HasSubscriptionHistory {
				// Users who have subscription history (has_subbed = true)
				query += " AND COALESCE(u.has_subbed, false) = true"
			} else {
				// Users who have no subscription history (has_subbed = false or NULL)
				query += " AND COALESCE(u.has_subbed, false) = false"
			}
			fmt.Printf("DEBUG: Subscription history filter added\n")
		}
	}

	// Note: ORDER BY is already set in the main query for DISTINCT ON to work properly
	// Do not add another ORDER BY here as it will override the DISTINCT ON ordering

	if limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
	}

	if offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
	}

	log.Printf("getSubscribers: Executing query with %d args", len(args))
	log.Printf("🔍 Final query: %s", query)
	log.Printf("🔍 Query args: %+v", args)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		log.Printf("❌ Database query failed: %v", err)
		return nil, fmt.Errorf("failed to get subscribers: %w", err)
	}
	defer rows.Close()

	var subscribers []*Subscriber
	for rows.Next() {
		subscriber := &Subscriber{}

		// Temporary variables for nullable fields from unified query
		var subscriptionID sql.NullString // plan_id from unified query (now string)
		var planID sql.NullString         // plan_id from unified query (now string)
		var planName sql.NullString
		var planPrice sql.NullFloat64
		var planCurrency sql.NullString
		var planInterval sql.NullString
		var planIntervalCount sql.NullInt64
		var subscriptionStatus sql.NullString
		var stripeCustomerID sql.NullString
		var stripeSubscriptionID sql.NullString

		err := rows.Scan(
			&subscriber.ID, &subscriber.Email, &subscriber.FirstName, &subscriber.LastName,
			&subscriber.Role, &subscriber.EmailVerified, &stripeCustomerID,
			&subscriber.LastLogin, &subscriber.CreatedAt, &subscriber.UpdatedAt,
			&subscriptionID, &planID, &planName,
			&planPrice, &planCurrency, &planInterval, &planIntervalCount,
			&subscriptionStatus, &subscriber.CurrentPeriodStart, &subscriber.CurrentPeriodEnd,
			&stripeSubscriptionID,
		)
		if err != nil {
			log.Printf("❌ Row scan failed: %v", err)
			return nil, fmt.Errorf("failed to scan subscriber: %w", err)
		}

		// Debug: Diagnostic logging removed - using unified plans now

		// Assign nullable fields to pointers (updated for unified query)
		if planID.Valid {
			// Convert string plan_id to int for backward compatibility
			if id, err := strconv.Atoi(planID.String); err == nil {
				subscriber.PlanID = &id
			}
		}
		if planName.Valid {
			subscriber.PlanName = &planName.String
		}
		if planPrice.Valid {
			subscriber.PlanPrice = &planPrice.Float64
		}
		if planCurrency.Valid {
			subscriber.PlanCurrency = &planCurrency.String
		}
		if subscriptionID.Valid {
			// Convert string subscription_id to int for backward compatibility
			if id, err := strconv.Atoi(subscriptionID.String); err == nil {
				subscriber.SubscriptionID = &id
			}
		}
		if subscriptionStatus.Valid {
			subscriber.SubscriptionStatus = &subscriptionStatus.String
		}
		if stripeCustomerID.Valid {
			subscriber.StripeCustomerID = &stripeCustomerID.String
		}
		if stripeSubscriptionID.Valid {
			subscriber.StripeSubscriptionID = &stripeSubscriptionID.String
		}

		// Set SubID to the same value as SubscriptionID
		subscriber.SubID = subscriber.SubscriptionID

		// Set plan interval data if available
		if planInterval.Valid && planInterval.String != "" {
			subscriber.PlanInterval = &planInterval.String
		}
		if planIntervalCount.Valid && planIntervalCount.Int64 > 0 {
			count := int(planIntervalCount.Int64)
			subscriber.PlanIntervalCount = &count
		}

		subscribers = append(subscribers, subscriber)
	}

	log.Printf("getSubscribers: Retrieved %d subscribers", len(subscribers))

	return subscribers, nil
}

// GetSubscriptionHistory retrieves expired/inactive subscriptions for history view
func (s *SubscriberService) GetSubscriptionHistory(limit, offset int, filters *SubscriberFilters) ([]*Subscriber, error) {
	log.Printf("🔍 GetSubscriptionHistory called with limit=%d, offset=%d", limit, offset)

	// Get EXPIRED/INACTIVE subscriptions (legacy + Stripe)
	query := `
		SELECT DISTINCT ON (u.id, COALESCE(sp.id, ss.id))
			u.id, u.email, u.first_name, u.last_name, u.role, u.email_verified,
			u.stripe_customer_id, u.last_login, u.created_at, u.updated_at,
			COALESCE(u.sub_id, ss.id) as subscription_id,
			COALESCE(sp.id, 0) as plan_id, 
			COALESCE(
				sp.name,                    -- Legacy subscription plan name
				stripe_prod.name,           -- Stripe product name  
				ss.product_name,            -- Fallback from stripe_subscriptions
				CASE 
					WHEN ss.status = 'canceled' THEN 'Canceled Subscription'
					WHEN ss.status = 'incomplete_expired' THEN 'Expired Subscription'
					WHEN ss.status = 'past_due' THEN 'Past Due Subscription'
					ELSE 'Inactive Subscription'
				END
			) as plan_name,
			COALESCE(
				sp.price,                                           -- Legacy plan price
				CASE WHEN stripe_price.unit_amount IS NOT NULL 
					THEN stripe_price.unit_amount::float / 100.0 
					ELSE NULL END,                                  -- Stripe price (cents to dollars)
				CASE WHEN ss.unit_amount IS NOT NULL 
					THEN ss.unit_amount::float / 100.0 
					ELSE NULL END,                                  -- Fallback from subscription
				0.0
			) as plan_price, 
			COALESCE(
				sp.currency,                -- Legacy plan currency
				stripe_price.currency,      -- Stripe price currency
				ss.currency,                -- Fallback currency
				'USD'
			) as plan_currency,
			COALESCE(
				sp.interval,                        -- Legacy plan interval
				stripe_price.recurring_interval,    -- Stripe price interval
				'month'                             -- Default interval
			) as interval, 
			COALESCE(sp.interval_count, 1) as interval_count,
			COALESCE(ss.status, 'expired') as subscription_status,
			ss.current_period_start, 
			ss.current_period_end,
			COALESCE(ss.stripe_id, u.stripe_customer_id) as stripe_subscription_id
		FROM users u
		-- Join legacy subscription plans (including inactive ones)
		LEFT JOIN subscription_plans sp ON u.sub_id = sp.id 
			AND (sp.is_active = false OR sp.deleted_at IS NOT NULL)
		-- Join Stripe customers
		LEFT JOIN stripe_customers sc ON (
			u.stripe_customer_id = sc.stripe_id OR 
			sc.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}'))
		)
		-- Join EXPIRED/INACTIVE Stripe subscriptions
		LEFT JOIN stripe_subscriptions ss ON sc.id = ss.customer_id 
			AND (
				ss.status IN ('canceled', 'incomplete_expired', 'past_due', 'unpaid') 
				OR (ss.current_period_end IS NOT NULL AND ss.current_period_end <= NOW())
			)
		-- Join Stripe products (to get product names)
		LEFT JOIN stripe_products stripe_prod ON ss.stripe_product_id = stripe_prod.stripe_id
		-- Join Stripe prices (to get pricing info)
		LEFT JOIN stripe_prices stripe_price ON ss.stripe_price_id = stripe_price.stripe_id
		WHERE (
			-- Has inactive legacy subscription
			(u.sub_id IS NOT NULL AND (sp.is_active = false OR sp.deleted_at IS NOT NULL)) 
			OR 
			-- Has expired/inactive Stripe subscription
			(ss.id IS NOT NULL AND (
				ss.status IN ('canceled', 'incomplete_expired', 'past_due', 'unpaid') 
				OR (ss.current_period_end IS NOT NULL AND ss.current_period_end <= NOW())
			))
		)
		AND u.is_active = true
		ORDER BY u.id, COALESCE(sp.id, ss.id), COALESCE(ss.current_period_end, sp.updated_at) DESC NULLS LAST
	`

	args := []interface{}{}
	argCount := 0

	// Add filters (similar to main subscribers)
	if filters != nil {
		log.Printf("🔍 Applying history filters: %+v", filters)

		if filters.Status != nil {
			argCount++
			query += fmt.Sprintf(" AND ss.status = $%d", argCount)
			args = append(args, *filters.Status)
		}

		if filters.Search != "" {
			argCount++
			query += fmt.Sprintf(" AND (u.email ILIKE $%d OR u.first_name ILIKE $%d OR u.last_name ILIKE $%d)",
				argCount, argCount, argCount)
			searchTerm := "%" + filters.Search + "%"
			args = append(args, searchTerm)
		}

		if filters.EmailVerified != nil {
			argCount++
			query += fmt.Sprintf(" AND u.email_verified = $%d", argCount)
			args = append(args, *filters.EmailVerified)
		}

		if filters.Role != nil {
			argCount++
			query += fmt.Sprintf(" AND u.role = $%d", argCount)
			args = append(args, *filters.Role)
		}
	}

	// Add limit and offset
	if limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
	}

	if offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
	}

	log.Printf("🔍 Executing subscription history query with %d args", len(args))
	log.Printf("🔍 History query: %s", query)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		log.Printf("❌ Subscription history query failed: %v", err)
		return nil, fmt.Errorf("failed to get subscription history: %w", err)
	}
	defer rows.Close()

	var subscribers []*Subscriber
	for rows.Next() {
		subscriber := &Subscriber{}
		var interval sql.NullString
		var intervalCount sql.NullInt64

		// Temporary variables for nullable fields
		var planID sql.NullInt64
		var planName sql.NullString
		var planPrice sql.NullFloat64
		var planCurrency sql.NullString
		var subscriptionID sql.NullInt64
		var subscriptionStatus sql.NullString
		var stripeCustomerID sql.NullString
		var stripeSubscriptionID sql.NullString

		err := rows.Scan(
			&subscriber.ID, &subscriber.Email, &subscriber.FirstName, &subscriber.LastName,
			&subscriber.Role, &subscriber.EmailVerified, &stripeCustomerID,
			&subscriber.LastLogin, &subscriber.CreatedAt, &subscriber.UpdatedAt,
			&subscriptionID, &planID, &planName,
			&planPrice, &planCurrency, &interval, &intervalCount,
			&subscriptionStatus, &subscriber.CurrentPeriodStart, &subscriber.CurrentPeriodEnd,
			&stripeSubscriptionID,
		)
		if err != nil {
			log.Printf("❌ History row scan failed: %v", err)
			return nil, fmt.Errorf("failed to scan subscription history: %w", err)
		}

		// Assign nullable fields to pointers
		if planID.Valid {
			id := int(planID.Int64)
			subscriber.PlanID = &id
		}
		if planName.Valid {
			subscriber.PlanName = &planName.String
		}
		if planPrice.Valid {
			subscriber.PlanPrice = &planPrice.Float64
		}
		if planCurrency.Valid {
			subscriber.PlanCurrency = &planCurrency.String
		}
		if subscriptionID.Valid {
			id := int(subscriptionID.Int64)
			subscriber.SubscriptionID = &id
		}
		if subscriptionStatus.Valid {
			subscriber.SubscriptionStatus = &subscriptionStatus.String
		}
		if stripeCustomerID.Valid {
			subscriber.StripeCustomerID = &stripeCustomerID.String
		}
		if stripeSubscriptionID.Valid {
			subscriber.StripeSubscriptionID = &stripeSubscriptionID.String
		}

		// Set SubID to the same value as SubscriptionID
		subscriber.SubID = subscriber.SubscriptionID

		// Set plan interval data if available
		if interval.Valid && interval.String != "" {
			subscriber.PlanInterval = &interval.String
		}
		if intervalCount.Valid && intervalCount.Int64 > 0 {
			count := int(intervalCount.Int64)
			subscriber.PlanIntervalCount = &count
		}

		subscribers = append(subscribers, subscriber)
	}

	log.Printf("getSubscriptionHistory: Retrieved %d expired/inactive subscriptions", len(subscribers))

	return subscribers, nil
}

// GetSubscriberByID retrieves a subscriber by user ID
func (s *SubscriberService) GetSubscriberByID(userID int) (*Subscriber, error) {
	query := `
		SELECT 
			u.id, u.email, u.first_name, u.last_name, u.role, u.email_verified,
			u.stripe_customer_id, u.last_login, u.created_at, u.updated_at,
			s.id as subscription_id, s.status as subscription_status,
			s.current_period_start, s.current_period_end, s.stripe_subscription_id,
			sp.id as plan_id, sp.name as plan_name, sp.price as plan_price, sp.currency as plan_currency
		FROM users u
		LEFT JOIN subscriptions s ON u.id = s.user_id AND s.deleted_at IS NULL
		LEFT JOIN subscription_plans sp ON s.plan_id = sp.id
		WHERE u.id = $1 AND u.id IN (
			SELECT DISTINCT user_id 
			FROM subscriptions 
			WHERE deleted_at IS NULL AND plan_id IS NOT NULL
		)
	`

	subscriber := &Subscriber{}
	err := s.db.QueryRow(query, userID).Scan(
		&subscriber.ID, &subscriber.Email, &subscriber.FirstName, &subscriber.LastName,
		&subscriber.Role, &subscriber.EmailVerified, &subscriber.StripeCustomerID,
		&subscriber.LastLogin, &subscriber.CreatedAt, &subscriber.UpdatedAt,
		&subscriber.SubscriptionID, &subscriber.SubscriptionStatus,
		&subscriber.CurrentPeriodStart, &subscriber.CurrentPeriodEnd,
		&subscriber.StripeSubscriptionID, &subscriber.PlanID, &subscriber.PlanName,
		&subscriber.PlanPrice, &subscriber.PlanCurrency,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber: %w", err)
	}

	return subscriber, nil
}

// GetUserAsSubscriber gets any user as a subscriber (doesn't require active subscription)
func (s *SubscriberService) GetUserAsSubscriber(userID int) (*Subscriber, error) {
	query := `
		SELECT 
			u.id, u.email, u.first_name, u.last_name, u.role, u.email_verified,
			u.stripe_customer_id, u.last_login, u.created_at, u.updated_at,
			COALESCE(s.id, 0) as subscription_id, 
			COALESCE(s.status, '') as subscription_status,
			s.current_period_start, s.current_period_end, s.stripe_subscription_id,
			COALESCE(sp.id, 0) as plan_id, 
			COALESCE(sp.name, '') as plan_name, 
			COALESCE(sp.price, 0) as plan_price, 
			COALESCE(sp.currency, '') as plan_currency
		FROM users u
		LEFT JOIN subscriptions s ON u.id = s.user_id AND s.deleted_at IS NULL
		LEFT JOIN subscription_plans sp ON s.plan_id = sp.id
		WHERE u.id = $1
	`

	subscriber := &Subscriber{}
	err := s.db.QueryRow(query, userID).Scan(
		&subscriber.ID, &subscriber.Email, &subscriber.FirstName, &subscriber.LastName,
		&subscriber.Role, &subscriber.EmailVerified, &subscriber.StripeCustomerID,
		&subscriber.LastLogin, &subscriber.CreatedAt, &subscriber.UpdatedAt,
		&subscriber.SubscriptionID, &subscriber.SubscriptionStatus,
		&subscriber.CurrentPeriodStart, &subscriber.CurrentPeriodEnd,
		&subscriber.StripeSubscriptionID, &subscriber.PlanID, &subscriber.PlanName,
		&subscriber.PlanPrice, &subscriber.PlanCurrency,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user as subscriber: %w", err)
	}

	return subscriber, nil
}

// GetSubscriberCount returns the total count of subscribers
func (s *SubscriberService) GetSubscriberCount(filters *SubscriberFilters) (int, error) {
	// ENHANCED: Count only CURRENTLY ACTIVE subscriptions (legacy + Stripe)
	// FIXED: Use DISTINCT to prevent counting duplicate users
	query := `
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		-- Join legacy subscription plans (for users with sub_id)
		LEFT JOIN subscription_plans sp ON u.sub_id = sp.id 
			AND sp.is_active = true 
			AND sp.deleted_at IS NULL
		-- Join Stripe customers (for users with stripe_customer_id)
		LEFT JOIN stripe_customers sc ON (
			u.stripe_customer_id = sc.stripe_id OR 
			sc.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}'))
		)
		-- Join current Stripe subscriptions (for ACTIVE Stripe subs only)
		LEFT JOIN stripe_subscriptions ss ON sc.id = ss.customer_id 
			AND ss.status IN ('active', 'trialing')
			AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
		WHERE (
			-- Has active legacy subscription
			(u.sub_id IS NOT NULL AND sp.id IS NOT NULL) 
			OR 
			-- Has active Stripe subscription
			(ss.id IS NOT NULL AND ss.status IN ('active', 'trialing') AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW()))
		)
		AND u.is_active = true
	`

	args := []interface{}{}
	argCount := 0

	// Add filters
	if filters != nil {
		// Note: PlanID filter removed since we don't have stripe_prices table join

		if filters.Status != nil {
			argCount++
			query += fmt.Sprintf(" AND ss.status = $%d", argCount)
			args = append(args, *filters.Status)
		}

		if filters.Search != "" {
			argCount++
			query += fmt.Sprintf(" AND (u.email ILIKE $%d OR u.first_name ILIKE $%d OR u.last_name ILIKE $%d)",
				argCount, argCount, argCount)
			searchTerm := "%" + filters.Search + "%"
			args = append(args, searchTerm)
		}

		if filters.EmailVerified != nil {
			argCount++
			query += fmt.Sprintf(" AND u.email_verified = $%d", argCount)
			args = append(args, *filters.EmailVerified)
		}

		if filters.Role != nil {
			argCount++
			query += fmt.Sprintf(" AND u.role = $%d", argCount)
			args = append(args, *filters.Role)
		}

		if filters.LastLogin != nil {
			argCount++
			query += fmt.Sprintf(" AND u.last_login = $%d", argCount)
			args = append(args, *filters.LastLogin)
		}

		if filters.CreatedDate != nil {
			argCount++
			query += fmt.Sprintf(" AND u.created_at = $%d", argCount)
			args = append(args, *filters.CreatedDate)
		}

		if filters.DateRange != nil {
			argCount++
			query += fmt.Sprintf(" AND u.created_at BETWEEN $%d AND $%d", argCount, argCount+1)
			args = append(args, filters.DateRange.Start, filters.DateRange.End)
			argCount++
		}

		if filters.HasSubscriptionHistory != nil {
			if *filters.HasSubscriptionHistory {
				// Users who have subscription history (has_subbed = true)
				query += " AND COALESCE(u.has_subbed, false) = true"
			} else {
				// Users who have no subscription history (has_subbed = false or NULL)
				query += " AND COALESCE(u.has_subbed, false) = false"
			}
		}
	}

	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get subscriber count: %w", err)
	}

	return count, nil
}

// GetSubscribersByEmailVerification retrieves subscribers filtered by email verification status
func (s *SubscriberService) GetSubscribersByEmailVerification(emailVerified bool, limit, offset int, filters *SubscriberFilters) ([]*Subscriber, error) {
	// Create a new filter set that includes email verification status
	emailFilters := &SubscriberFilters{
		EmailVerified: &emailVerified,
	}

	// Copy other filters if provided
	if filters != nil {
		emailFilters.Status = filters.Status
		emailFilters.Search = filters.Search
		emailFilters.Role = filters.Role
		emailFilters.LastLogin = filters.LastLogin
		emailFilters.CreatedDate = filters.CreatedDate
		emailFilters.DateRange = filters.DateRange
		emailFilters.HasSubscriptionHistory = filters.HasSubscriptionHistory
	}

	return s.GetSubscribers(limit, offset, emailFilters)
}

// GetSubscriberCountByEmailVerification returns count of subscribers by email verification status
func (s *SubscriberService) GetSubscriberCountByEmailVerification(emailVerified bool, filters *SubscriberFilters) (int, error) {
	// Create a new filter set that includes email verification status
	emailFilters := &SubscriberFilters{
		EmailVerified: &emailVerified,
	}

	// Copy other filters if provided
	if filters != nil {
		emailFilters.Status = filters.Status
		emailFilters.Search = filters.Search
		emailFilters.Role = filters.Role
		emailFilters.LastLogin = filters.LastLogin
		emailFilters.CreatedDate = filters.CreatedDate
		emailFilters.DateRange = filters.DateRange
		emailFilters.HasSubscriptionHistory = filters.HasSubscriptionHistory
	}

	return s.GetSubscriberCount(emailFilters)
}

// GetSubscriberStats returns subscriber statistics
func (s *SubscriberService) GetSubscriberStats() (*SubscriberStats, error) {
	// Get total subscribers
	totalQuery := `
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		WHERE u.id IN (
			SELECT DISTINCT user_id 
			FROM subscriptions 
			WHERE deleted_at IS NULL AND plan_id IS NOT NULL
		)
	`
	var totalSubscribers int
	err := s.db.QueryRow(totalQuery).Scan(&totalSubscribers)
	if err != nil {
		return nil, fmt.Errorf("failed to get total subscribers: %w", err)
	}

	// Get subscribers by status
	statusQuery := `
		SELECT s.status, COUNT(DISTINCT u.id)
		FROM users u
		JOIN subscriptions s ON u.id = s.user_id AND s.deleted_at IS NULL
		WHERE s.plan_id IS NOT NULL
		GROUP BY s.status
	`
	rows, err := s.db.Query(statusQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscribers by status: %w", err)
	}
	defer rows.Close()

	stats := &SubscriberStats{TotalSubscribers: totalSubscribers}
	for rows.Next() {
		var status string
		var count int
		err := rows.Scan(&status, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan status count: %w", err)
		}

		switch status {
		case "active":
			stats.ActiveSubscribers = count
		case "trialing":
			stats.TrialingSubscribers = count
		case "past_due":
			stats.PastDueSubscribers = count
		case "canceled":
			stats.CanceledSubscribers = count
		}
	}

	// Get revenue statistics
	revenueQuery := `
		SELECT 
			SUM(sp.price) as total_revenue,
			AVG(sp.price) as avg_revenue_per_user
		FROM users u
		JOIN subscriptions s ON u.id = s.user_id AND s.deleted_at IS NULL
		JOIN subscription_plans sp ON s.plan_id = sp.id
		WHERE s.status IN ('active', 'trialing') AND s.plan_id IS NOT NULL
	`
	var totalRevenue, avgRevenuePerUser float64
	err = s.db.QueryRow(revenueQuery).Scan(&totalRevenue, &avgRevenuePerUser)
	if err != nil {
		return nil, fmt.Errorf("failed to get revenue stats: %w", err)
	}

	stats.MonthlyRevenue = totalRevenue
	stats.AnnualRevenue = totalRevenue * 12
	stats.AverageRevenuePerUser = avgRevenuePerUser

	// Calculate churn rate (simplified - would need more sophisticated logic in production)
	if totalSubscribers > 0 {
		stats.ChurnRate = float64(stats.CanceledSubscribers) / float64(totalSubscribers) * 100
	}

	return stats, nil
}

// GetSubscribersByPlan retrieves all subscribers for a specific plan
func (s *SubscriberService) GetSubscribersByPlan(planID int, limit, offset int) ([]*Subscriber, error) {
	filters := &SubscriberFilters{
		PlanID: &planID,
	}
	return s.GetSubscribers(limit, offset, filters)
}

// GetSubscribersByStatus retrieves all subscribers with a specific status
func (s *SubscriberService) GetSubscribersByStatus(status string, limit, offset int) ([]*Subscriber, error) {
	filters := &SubscriberFilters{
		Status: &status,
	}
	return s.GetSubscribers(limit, offset, filters)
}

// SearchSubscribers searches subscribers by email, first name, or last name
func (s *SubscriberService) SearchSubscribers(searchTerm string, limit, offset int) ([]*Subscriber, error) {
	filters := &SubscriberFilters{
		Search: searchTerm,
	}
	return s.GetSubscribers(limit, offset, filters)
}

// NonSubscriber represents a user without an active subscription
type NonSubscriber struct {
	ID                     int        `json:"id"`
	Email                  string     `json:"email"`
	FirstName              string     `json:"first_name"`
	LastName               string     `json:"last_name"`
	Role                   string     `json:"role"`
	EmailVerified          bool       `json:"email_verified"`
	SubID                  *int       `json:"sub_id,omitempty"` // Will be null for non-subscribers
	HasSubscriptionHistory bool       `json:"has_subscription_history"`
	LastLogin              *time.Time `json:"last_login,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// NonSubscriberFilters represents filters for non-subscriber queries
type NonSubscriberFilters struct {
	Search        string     `json:"search"`
	EmailVerified *bool      `json:"email_verified"`
	Role          *string    `json:"role"`
	LastLogin     *time.Time `json:"last_login"`
	CreatedDate   *time.Time `json:"created_date"`
	DateRange     *struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	} `json:"date_range"`
	HasSubscriptionHistory *bool `json:"has_subscription_history"` // true, false, or nil for all
}

// GetNonSubscribers retrieves all users without active subscriptions
func (s *SubscriberService) GetNonSubscribers(limit, offset int, filters *NonSubscriberFilters) ([]*NonSubscriber, error) {
	query := `
		SELECT 
			u.id, u.email, u.first_name, u.last_name, u.role, u.email_verified,
			u.last_login, u.created_at, u.updated_at,
			COALESCE(u.has_subbed, false) as has_subscription_history
		FROM users u
		WHERE u.sub_id IS NULL
	`

	args := []interface{}{}
	argCount := 0

	// Add filters
	if filters != nil {
		if filters.Search != "" {
			argCount++
			query += fmt.Sprintf(" AND (u.email ILIKE $%d OR u.first_name ILIKE $%d OR u.last_name ILIKE $%d)",
				argCount, argCount, argCount)
			searchTerm := "%" + filters.Search + "%"
			args = append(args, searchTerm)
		}

		if filters.EmailVerified != nil {
			argCount++
			query += fmt.Sprintf(" AND u.email_verified = $%d", argCount)
			args = append(args, *filters.EmailVerified)
		}

		if filters.Role != nil {
			argCount++
			query += fmt.Sprintf(" AND u.role = $%d", argCount)
			args = append(args, *filters.Role)
		}

		if filters.LastLogin != nil {
			if filters.LastLogin.IsZero() {
				// Special case for "never" - users who have never logged in
				query += " AND u.last_login IS NULL"
			} else {
				argCount++
				query += fmt.Sprintf(" AND u.last_login >= $%d", argCount)
				args = append(args, *filters.LastLogin)
			}
		}

		if filters.CreatedDate != nil {
			argCount++
			query += fmt.Sprintf(" AND u.created_at >= $%d", argCount)
			args = append(args, *filters.CreatedDate)
		}

		if filters.DateRange != nil {
			argCount++
			query += fmt.Sprintf(" AND u.created_at BETWEEN $%d AND $%d", argCount, argCount+1)
			args = append(args, filters.DateRange.Start, filters.DateRange.End)
			argCount++
		}

		// Add subscription history filter using the new has_subbed column
		if filters.HasSubscriptionHistory != nil {
			argCount++
			query += fmt.Sprintf(" AND COALESCE(u.has_subbed, false) = $%d", argCount)
			args = append(args, *filters.HasSubscriptionHistory)
		}
	}

	query += " ORDER BY u.created_at DESC"

	if limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
	}

	if offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get non-subscribers: %w", err)
	}
	defer rows.Close()

	var nonSubscribers []*NonSubscriber
	for rows.Next() {
		nonSubscriber := &NonSubscriber{}
		err := rows.Scan(
			&nonSubscriber.ID, &nonSubscriber.Email, &nonSubscriber.FirstName, &nonSubscriber.LastName,
			&nonSubscriber.Role, &nonSubscriber.EmailVerified, &nonSubscriber.LastLogin,
			&nonSubscriber.CreatedAt, &nonSubscriber.UpdatedAt, &nonSubscriber.HasSubscriptionHistory,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan non-subscriber: %w", err)
		}
		nonSubscribers = append(nonSubscribers, nonSubscriber)
	}

	return nonSubscribers, nil
}

// GetNonSubscriberCount returns the total count of non-subscribers
func (s *SubscriberService) GetNonSubscriberCount(filters *NonSubscriberFilters) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM users u
		WHERE u.sub_id IS NULL
	`

	args := []interface{}{}
	argCount := 0

	// Add filters
	if filters != nil {
		if filters.Search != "" {
			argCount++
			query += fmt.Sprintf(" AND (u.email ILIKE $%d OR u.first_name ILIKE $%d OR u.last_name ILIKE $%d)",
				argCount, argCount, argCount)
			searchTerm := "%" + filters.Search + "%"
			args = append(args, searchTerm)
		}

		if filters.EmailVerified != nil {
			argCount++
			query += fmt.Sprintf(" AND u.email_verified = $%d", argCount)
			args = append(args, *filters.EmailVerified)
		}

		if filters.Role != nil {
			argCount++
			query += fmt.Sprintf(" AND u.role = $%d", argCount)
			args = append(args, *filters.Role)
		}

		if filters.LastLogin != nil {
			if filters.LastLogin.IsZero() {
				// Special case for "never" - users who have never logged in
				query += " AND u.last_login IS NULL"
			} else {
				argCount++
				query += fmt.Sprintf(" AND u.last_login >= $%d", argCount)
				args = append(args, *filters.LastLogin)
			}
		}

		if filters.CreatedDate != nil {
			argCount++
			query += fmt.Sprintf(" AND u.created_at >= $%d", argCount)
			args = append(args, *filters.CreatedDate)
		}

		if filters.DateRange != nil {
			argCount++
			query += fmt.Sprintf(" AND u.created_at BETWEEN $%d AND $%d", argCount, argCount+1)
			args = append(args, filters.DateRange.Start, filters.DateRange.End)
			argCount++
		}

		// Add subscription history filter using the new has_subbed column
		if filters.HasSubscriptionHistory != nil {
			argCount++
			query += fmt.Sprintf(" AND COALESCE(u.has_subbed, false) = $%d", argCount)
			args = append(args, *filters.HasSubscriptionHistory)
		}
	}

	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get non-subscriber count: %w", err)
	}

	return count, nil
}

// Debug method to check subscriptions table
func (s *SubscriberService) DebugSubscriptions() error {
	fmt.Println("DEBUG: Checking subscriptions table...")

	// Check all subscriptions
	rows, err := s.db.Query(`
		SELECT s.id, s.user_id, s.plan_id, s.status, s.deleted_at, u.email
		FROM subscriptions s
		LEFT JOIN users u ON s.user_id = u.id
		ORDER BY s.id
	`)
	if err != nil {
		return fmt.Errorf("failed to query subscriptions: %w", err)
	}
	defer rows.Close()

	fmt.Println("DEBUG: All subscriptions:")
	for rows.Next() {
		var id, userID int
		var planID sql.NullInt32
		var status string
		var deletedAt sql.NullTime
		var email string

		err := rows.Scan(&id, &userID, &planID, &status, &deletedAt, &email)
		if err != nil {
			return fmt.Errorf("failed to scan subscription: %w", err)
		}

		fmt.Printf("  ID: %d, UserID: %d, PlanID: %v, Status: %s, DeletedAt: %v, Email: %s\n",
			id, userID, planID, status, deletedAt, email)
	}

	// Check users with subscriptions
	rows2, err := s.db.Query(`
		SELECT u.id, u.email, s.id as sub_id, s.plan_id
		FROM users u
		LEFT JOIN subscriptions s ON u.sub_id = s.id AND s.deleted_at IS NULL
		WHERE s.id IS NOT NULL
		ORDER BY u.id
	`)
	if err != nil {
		return fmt.Errorf("failed to query users with subscriptions: %w", err)
	}
	defer rows2.Close()

	fmt.Println("DEBUG: Users with subscriptions:")
	for rows2.Next() {
		var userID int
		var email string
		var subID sql.NullInt32
		var planID sql.NullInt32

		err := rows2.Scan(&userID, &email, &subID, &planID)
		if err != nil {
			return fmt.Errorf("failed to scan user subscription: %w", err)
		}

		fmt.Printf("  UserID: %d, Email: %s, SubID: %v, PlanID: %v\n",
			userID, email, subID, planID)
	}

	// Check total users
	var totalUsers int
	err = s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}
	fmt.Printf("DEBUG: Total users in database: %d\n", totalUsers)

	// Check users with sub_id
	rows3, err := s.db.Query(`
		SELECT u.id, u.email, u.sub_id, sp.id as plan_id, sp.name as plan_name, sp.is_active
		FROM users u
		LEFT JOIN subscription_plans sp ON u.sub_id = sp.id
		WHERE u.sub_id IS NOT NULL
		ORDER BY u.id
	`)
	if err != nil {
		return fmt.Errorf("failed to query users with sub_id: %w", err)
	}
	defer rows3.Close()

	fmt.Println("DEBUG: Users with sub_id and plan data:")
	for rows3.Next() {
		var userID int
		var email string
		var subID sql.NullInt32
		var planID sql.NullInt32
		var planName sql.NullString
		var planActive sql.NullBool

		err := rows3.Scan(&userID, &email, &subID, &planID, &planName, &planActive)
		if err != nil {
			return fmt.Errorf("failed to scan user with plan: %w", err)
		}

		fmt.Printf("  UserID: %d, Email: %s, SubID: %v, PlanID: %v, PlanName: %v, PlanActive: %v\n",
			userID, email, subID, planID, planName, planActive)
	}

	return nil
}

// UpdateSubscriber updates a subscriber's information
func (s *SubscriberService) UpdateSubscriber(userID int, updates map[string]interface{}) (*Subscriber, error) {
	// First check if user exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check if user exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user with ID %d does not exist", userID)
	}

	// Build update query dynamically
	query := "UPDATE users SET "
	args := []interface{}{}
	argCount := 0

	for field, value := range updates {
		if argCount > 0 {
			query += ", "
		}
		argCount++
		query += fmt.Sprintf("%s = $%d", field, argCount)
		args = append(args, value)
	}

	query += fmt.Sprintf(", updated_at = NOW() WHERE id = $%d", argCount+1)
	args = append(args, userID)

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscriber: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("no rows were updated for user ID %d", userID)
	}

	// Return the updated subscriber - use a simpler query that doesn't require subscriptions
	return s.GetUserAsSubscriber(userID)
}

// SuspendSubscriber suspends a subscriber's account
func (s *SubscriberService) SuspendSubscriber(userID int) (*Subscriber, error) {
	// Get current subscriber status
	subscriber, err := s.GetSubscriberByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber: %w", err)
	}

	// Update user status to suspended
	query := "UPDATE users SET status = 'suspended', updated_at = NOW() WHERE id = $1"
	_, err = s.db.Exec(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to suspend subscriber: %w", err)
	}

	// Add to subscriber history
	historyService := NewSubscriberHistoryService(s.db)
	previousStatus := "active"
	if subscriber.SubscriptionStatus != nil {
		previousStatus = *subscriber.SubscriptionStatus
	}

	err = historyService.AddSuspensionHistoryEntry(userID, "suspended", "Account suspended by admin", previousStatus, "suspended")
	if err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: Failed to add suspension history entry: %v\n", err)
	}

	// Return the updated subscriber
	return s.GetSubscriberByID(userID)
}

// ActivateSubscriber activates a subscriber's account
func (s *SubscriberService) ActivateSubscriber(userID int) (*Subscriber, error) {
	// Get current subscriber status
	subscriber, err := s.GetSubscriberByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber: %w", err)
	}

	// Update user status to active
	query := "UPDATE users SET status = 'active', updated_at = NOW() WHERE id = $1"
	_, err = s.db.Exec(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to activate subscriber: %w", err)
	}

	// Add to subscriber history
	historyService := NewSubscriberHistoryService(s.db)
	previousStatus := "suspended"
	if subscriber.SubscriptionStatus != nil {
		previousStatus = *subscriber.SubscriptionStatus
	}

	err = historyService.AddSuspensionHistoryEntry(userID, "activated", "Account activated by admin", previousStatus, "active")
	if err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: Failed to add activation history entry: %v\n", err)
	}

	// Return the updated subscriber
	return s.GetSubscriberByID(userID)
}

// GetSubscriberHistory retrieves the history of a subscriber
func (s *SubscriberService) GetSubscriberHistory(userID int) ([]map[string]interface{}, error) {
	// Create subscriber history service
	historyService := NewSubscriberHistoryService(s.db)

	// Get history from the new subscriber history service
	history, err := historyService.GetSubscriberHistory(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber history: %w", err)
	}

	// Convert the new format to the expected format for backward compatibility
	var result []map[string]interface{}

	// Add subscription history entries
	if subHistory, ok := history["subscription_history"].(map[string]interface{}); ok {
		if entries, ok := subHistory["entries"].([]interface{}); ok {
			for _, entry := range entries {
				if entryMap, ok := entry.(map[string]interface{}); ok {
					result = append(result, map[string]interface{}{
						"action":      entryMap["action"],
						"timestamp":   entryMap["timestamp"],
						"description": entryMap["description"],
						"metadata":    entryMap["metadata"],
					})
				}
			}
		}
	}

	// Add offer history entries
	if offHistory, ok := history["offer_history"].(map[string]interface{}); ok {
		if entries, ok := offHistory["entries"].([]interface{}); ok {
			for _, entry := range entries {
				if entryMap, ok := entry.(map[string]interface{}); ok {
					result = append(result, map[string]interface{}{
						"action":      entryMap["action"],
						"timestamp":   entryMap["timestamp"],
						"description": entryMap["description"],
						"metadata":    entryMap["metadata"],
					})
				}
			}
		}
	}

	// Add notes entries
	if notes, ok := history["notes"].(map[string]interface{}); ok {
		if noteEntries, ok := notes["notes"].([]interface{}); ok {
			for _, note := range noteEntries {
				if noteMap, ok := note.(map[string]interface{}); ok {
					result = append(result, map[string]interface{}{
						"action":      "note_added",
						"timestamp":   noteMap["timestamp"],
						"description": noteMap["note"],
						"metadata": map[string]interface{}{
							"category":   noteMap["category"],
							"visibility": noteMap["visibility"],
							"admin_name": noteMap["admin_name"],
						},
					})
				}
			}
		}
	}

	// Add updates entries
	if updates, ok := history["updates"].(map[string]interface{}); ok {
		if updateEntries, ok := updates["updates"].([]interface{}); ok {
			for _, update := range updateEntries {
				if updateMap, ok := update.(map[string]interface{}); ok {
					result = append(result, map[string]interface{}{
						"action":      updateMap["action"],
						"timestamp":   updateMap["timestamp"],
						"description": fmt.Sprintf("%s: %s", updateMap["action"], updateMap["category"]),
						"metadata":    updateMap["details"],
					})
				}
			}
		}
	}

	// Sort by timestamp (newest first)
	// Note: This is a simple sort - in production you might want to use a more sophisticated approach
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if timestamp1, ok1 := result[i]["timestamp"].(time.Time); ok1 {
				if timestamp2, ok2 := result[j]["timestamp"].(time.Time); ok2 {
					if timestamp1.Before(timestamp2) {
						result[i], result[j] = result[j], result[i]
					}
				}
			}
		}
	}

	return result, nil
}

// BulkSuspendSubscribers suspends multiple subscribers
func (s *SubscriberService) BulkSuspendSubscribers(userIDs []int) ([]*Subscriber, error) {
	if len(userIDs) == 0 {
		return []*Subscriber{}, nil
	}

	// Build the query for bulk update
	query := "UPDATE users SET status = 'suspended', updated_at = NOW() WHERE id = ANY($1)"
	_, err := s.db.Exec(query, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk suspend subscribers: %w", err)
	}

	// Return the updated subscribers
	var subscribers []*Subscriber
	for _, userID := range userIDs {
		subscriber, err := s.GetSubscriberByID(userID)
		if err != nil {
			// Log error but continue with other subscribers
			fmt.Printf("Failed to get subscriber %d: %v\n", userID, err)
			continue
		}
		subscribers = append(subscribers, subscriber)
	}

	return subscribers, nil
}

// BulkActivateSubscribers activates multiple subscribers
func (s *SubscriberService) BulkActivateSubscribers(userIDs []int) ([]*Subscriber, error) {
	if len(userIDs) == 0 {
		return []*Subscriber{}, nil
	}

	// Build the query for bulk update
	query := "UPDATE users SET status = 'active', updated_at = NOW() WHERE id = ANY($1)"
	_, err := s.db.Exec(query, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk activate subscribers: %w", err)
	}

	// Return the updated subscribers
	var subscribers []*Subscriber
	for _, userID := range userIDs {
		subscriber, err := s.GetSubscriberByID(userID)
		if err != nil {
			// Log error but continue with other subscribers
			fmt.Printf("Failed to get subscriber %d: %v\n", userID, err)
			continue
		}
		subscribers = append(subscribers, subscriber)
	}

	return subscribers, nil
}

// BulkChangePlan changes the plan for multiple subscribers
func (s *SubscriberService) BulkChangePlan(planID int, userIDs []int) ([]*Subscriber, error) {
	if len(userIDs) == 0 {
		return []*Subscriber{}, nil
	}

	// Build the query for bulk update
	query := "UPDATE users SET sub_id = $1, updated_at = NOW() WHERE id = ANY($2)"
	_, err := s.db.Exec(query, planID, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk change plan: %w", err)
	}

	// Return the updated subscribers
	var subscribers []*Subscriber
	for _, userID := range userIDs {
		subscriber, err := s.GetSubscriberByID(userID)
		if err != nil {
			// Log error but continue with other subscribers
			fmt.Printf("Failed to get subscriber %d: %v\n", userID, err)
			continue
		}
		subscribers = append(subscribers, subscriber)
	}

	return subscribers, nil
}

// ExportSubscribers exports subscribers to CSV format
func (s *SubscriberService) ExportSubscribers() (string, error) {
	// Get all subscribers
	subscribers, err := s.GetSubscribers(10000, 0, nil) // Get all subscribers
	if err != nil {
		return "", fmt.Errorf("failed to get subscribers: %w", err)
	}

	// Generate CSV content
	var csvContent strings.Builder

	// Write CSV header
	csvContent.WriteString("ID,Email,First Name,Last Name,Role,Email Verified,Plan ID,Plan Name,Plan Price,Plan Currency,Subscription Status,Last Login,Created At,Updated At\n")

	// Write subscriber data
	for _, subscriber := range subscribers {
		lastLogin := ""
		if subscriber.LastLogin != nil {
			lastLogin = subscriber.LastLogin.Format("2006-01-02 15:04:05")
		}

		planID := ""
		if subscriber.PlanID != nil {
			planID = fmt.Sprintf("%d", *subscriber.PlanID)
		}

		planName := ""
		if subscriber.PlanName != nil {
			planName = *subscriber.PlanName
		}

		planPrice := ""
		if subscriber.PlanPrice != nil {
			planPrice = fmt.Sprintf("%.2f", *subscriber.PlanPrice)
		}

		planCurrency := ""
		if subscriber.PlanCurrency != nil {
			planCurrency = *subscriber.PlanCurrency
		}

		subscriptionStatus := ""
		if subscriber.SubscriptionStatus != nil {
			subscriptionStatus = *subscriber.SubscriptionStatus
		}

		// Escape CSV fields that contain commas or quotes
		email := strings.ReplaceAll(subscriber.Email, `"`, `""`)
		firstName := strings.ReplaceAll(subscriber.FirstName, `"`, `""`)
		lastName := strings.ReplaceAll(subscriber.LastName, `"`, `""`)
		role := strings.ReplaceAll(subscriber.Role, `"`, `""`)
		planName = strings.ReplaceAll(planName, `"`, `""`)
		planCurrency = strings.ReplaceAll(planCurrency, `"`, `""`)
		subscriptionStatus = strings.ReplaceAll(subscriptionStatus, `"`, `""`)

		// Write CSV row
		csvContent.WriteString(fmt.Sprintf(`%d,"%s","%s","%s","%s",%t,"%s","%s","%s","%s","%s","%s","%s","%s"`+"\n",
			subscriber.ID,
			email,
			firstName,
			lastName,
			role,
			subscriber.EmailVerified,
			planID,
			planName,
			planPrice,
			planCurrency,
			subscriptionStatus,
			lastLogin,
			subscriber.CreatedAt.Format("2006-01-02 15:04:05"),
			subscriber.UpdatedAt.Format("2006-01-02 15:04:05"),
		))
	}

	return csvContent.String(), nil
}

// ExportNonSubscribers exports non-subscribers to CSV format
func (s *SubscriberService) ExportNonSubscribers() (string, error) {
	fmt.Println("ExportNonSubscribers: Starting export")

	// Get all non-subscribers
	nonSubscribers, err := s.GetNonSubscribers(10000, 0, nil) // Get all non-subscribers
	if err != nil {
		fmt.Printf("ExportNonSubscribers: Error getting non-subscribers: %v\n", err)
		return "", fmt.Errorf("failed to get non-subscribers: %w", err)
	}

	fmt.Printf("ExportNonSubscribers: Retrieved %d non-subscribers\n", len(nonSubscribers))

	// Generate CSV content
	var csvContent strings.Builder

	// Write CSV header
	csvContent.WriteString("ID,Email,First Name,Last Name,Role,Email Verified,Last Login,Created At,Updated At\n")

	// Write non-subscriber data
	for _, nonSubscriber := range nonSubscribers {
		lastLogin := ""
		if nonSubscriber.LastLogin != nil {
			lastLogin = nonSubscriber.LastLogin.Format("2006-01-02 15:04:05")
		}

		// Escape CSV fields that contain commas or quotes
		email := strings.ReplaceAll(nonSubscriber.Email, `"`, `""`)
		firstName := strings.ReplaceAll(nonSubscriber.FirstName, `"`, `""`)
		lastName := strings.ReplaceAll(nonSubscriber.LastName, `"`, `""`)
		role := strings.ReplaceAll(nonSubscriber.Role, `"`, `""`)

		// Write CSV row
		csvContent.WriteString(fmt.Sprintf(`%d,"%s","%s","%s","%s",%t,"%s","%s","%s"`+"\n",
			nonSubscriber.ID,
			email,
			firstName,
			lastName,
			role,
			nonSubscriber.EmailVerified,
			lastLogin,
			nonSubscriber.CreatedAt.Format("2006-01-02 15:04:05"),
			nonSubscriber.UpdatedAt.Format("2006-01-02 15:04:05"),
		))
	}

	result := csvContent.String()
	fmt.Printf("ExportNonSubscribers: Generated CSV with %d characters\n", len(result))
	return result, nil
}
