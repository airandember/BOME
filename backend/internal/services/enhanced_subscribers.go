package services

import (
	"fmt"
	"log"
	"time"

	"bome-backend/internal/database"
)

// EnhancedSubscriberService handles business logic for the unified subscriber dashboard
type EnhancedSubscriberService struct {
	db *database.DB
}

// EnhancedSubscriber represents a subscriber with business intelligence fields
type EnhancedSubscriber struct {
	// Core User Data
	ID            int    `json:"id"`
	Email         string `json:"email"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	FullName      string `json:"full_name"`
	EmailVerified bool   `json:"email_verified"`
	Role          string `json:"role"`

	// Subscription Status (Key Goals)
	HasActivePlan  bool `json:"has_active_plan"`
	HasVideoAccess bool `json:"has_video_access"`
	// VideoAccessSource removed - plans are now the only source

	// Plan Information
	PlanName           string     `json:"plan_name"`
	PlanLegacyStatus   string     `json:"plan_legacy_status"` // 'Legacy', 'Current', 'Unknown'
	PlanType           string     `json:"plan_type"`          // 'premium', 'basic', 'none'
	PlanStartDate      *time.Time `json:"plan_start_date"`
	BillingPeriodStart *time.Time `json:"billing_period_start"`
	BillingPeriodEnd   *time.Time `json:"billing_period_end"`
	PlanStatus         string     `json:"plan_status"` // 'active', 'expired', 'trial', 'cancelled', 'none'
	PlanPrice          float64    `json:"plan_price"`
	PlanCurrency       string     `json:"plan_currency"`

	// Business Intelligence
	MRRContribution          float64 `json:"mrr_contribution"`
	LTVEstimate              float64 `json:"ltv_estimate"`
	DaysUntilExpiry          *int    `json:"days_until_expiry"`
	SubscriptionDurationDays *int    `json:"subscription_duration_days"`

	// Admin Tracking
	LastLogin           *time.Time `json:"last_login"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	ManualAccessGranted bool       `json:"manual_access_granted"`
	StripeCustomerID    *string    `json:"stripe_customer_id"`

	// Computed Fields
	IsExpiringSoon bool `json:"is_expiring_soon"` // Within 7 days
	IsHighValue    bool `json:"is_high_value"`    // Above average MRR
	AccountAgeDays int  `json:"account_age_days"`
}

// EnhancedSubscriberFilters represents filters for enhanced subscriber queries
type EnhancedSubscriberFilters struct {
	Search         string  `json:"search"`
	PlanType       *string `json:"plan_type"` // 'premium', 'basic', 'none'
	HasActivePlan  *bool   `json:"has_active_plan"`
	HasVideoAccess *bool   `json:"has_video_access"`
	// VideoAccessSource removed - plans are now the only source
	IsExpiringSoon  *bool      `json:"is_expiring_soon"`
	EmailVerified   *bool      `json:"email_verified"`
	Role            *string    `json:"role"`
	CreatedDateFrom *time.Time `json:"created_date_from"`
	CreatedDateTo   *time.Time `json:"created_date_to"`
	LastLoginFrom   *time.Time `json:"last_login_from"`
	LastLoginTo     *time.Time `json:"last_login_to"`
	MinMRR          *float64   `json:"min_mrr"`
	MaxMRR          *float64   `json:"max_mrr"`
}

// SubscriberKPIs represents key performance indicators for subscribers
type SubscriberKPIs struct {
	TotalSubscribers  int     `json:"total_subscribers"`
	ActiveSubscribers int     `json:"active_subscribers"`
	VideoAccessUsers  int     `json:"video_access_users"`
	TotalMRR          float64 `json:"total_mrr"`
	AvgDaysToExpiry   float64 `json:"avg_days_to_expiry"`
	ChurnRiskCount    int     `json:"churn_risk_count"`
	PremiumUsers      int     `json:"premium_users"`
	BasicUsers        int     `json:"basic_users"`
	ManualAccessUsers int     `json:"manual_access_users"`
}

// SubscriberResponse represents the API response for enhanced subscribers
type SubscriberResponse struct {
	Subscribers []EnhancedSubscriber `json:"subscribers"`
	TotalCount  int                  `json:"total_count"`
	KPIs        SubscriberKPIs       `json:"kpis"`
	Pagination  struct {
		Page       int  `json:"page"`
		Limit      int  `json:"limit"`
		TotalPages int  `json:"total_pages"`
		HasNext    bool `json:"has_next"`
		HasPrev    bool `json:"has_prev"`
	} `json:"pagination"`
}

// NewEnhancedSubscriberService creates a new enhanced subscriber service
func NewEnhancedSubscriberService(db *database.DB) *EnhancedSubscriberService {
	return &EnhancedSubscriberService{db: db}
}

// GetEnhancedSubscribers retrieves subscribers with business intelligence data
func (s *EnhancedSubscriberService) GetEnhancedSubscribers(page, limit int, filters *EnhancedSubscriberFilters) (*SubscriberResponse, error) {
	offset := (page - 1) * limit

	// Build the main query with all business intelligence fields
	query := `
		SELECT 
			u.id,
			u.email,
			u.first_name,
			u.last_name,
			CONCAT(COALESCE(u.first_name, ''), ' ', COALESCE(u.last_name, '')) as full_name,
			u.email_verified,
			u.role,
			
			-- Subscription Status (Key Goals)
			CASE 
				WHEN (u.sub_id IS NOT NULL AND sp.is_active = true AND sp.deleted_at IS NULL) 
					OR (ss.status IN ('active', 'trialing') AND ss.current_period_end > NOW() AND spr.video_approved = true) 
				THEN true 
				ELSE false 
			END as has_active_plan,
			
			CASE 
				WHEN (u.sub_id IS NOT NULL AND sp.is_active = true AND sp.deleted_at IS NULL) 
					OR (ss.status IN ('active', 'trialing') AND ss.current_period_end > NOW() AND spr.video_approved = true)
					OR u.manual_video_access = true
				THEN true 
				ELSE false 
			END as has_video_access,
			
			-- Plan Information
			COALESCE(
				sp.name, 
				ss.product_name,
				CASE 
					WHEN ss.status = 'active' THEN 'Active Subscription'
					WHEN ss.status = 'trialing' THEN 'Trial Subscription'
					ELSE 'No Plan'
				END
			) as plan_name,
			
			CASE 
				WHEN spr.legacy_product = true THEN 'Legacy'
				WHEN spr.legacy_product = false THEN 'Current'
				ELSE 'Unknown'
			END as plan_legacy_status,
			
			CASE 
				WHEN LOWER(COALESCE(sp.name, ss.product_name, '')) LIKE '%premium%' 
					OR LOWER(COALESCE(sp.name, ss.product_name, '')) LIKE '%yearly%' 
				THEN 'premium'
				WHEN LOWER(COALESCE(sp.name, ss.product_name, '')) LIKE '%basic%' 
				THEN 'basic'
				ELSE 'none'
			END as plan_type,
			
			ss.created_at as plan_start_date,
			ss.current_period_start as billing_period_start,
			ss.current_period_end as billing_period_end,
			
			CASE 
				WHEN ss.status = 'active' AND ss.current_period_end > NOW() THEN 'active'
				WHEN ss.status = 'trialing' THEN 'trial'
				WHEN ss.status = 'canceled' THEN 'cancelled'
				WHEN ss.current_period_end <= NOW() THEN 'expired'
				ELSE 'none'
			END as plan_status,
			
			COALESCE(sp.price, ss.unit_amount::float / 100.0, 0.0) as plan_price,
			COALESCE(sp.currency, ss.currency, 'USD') as plan_currency,
			
			-- Business Intelligence
			CASE 
				WHEN COALESCE(sp.interval, 'month') = 'month' 
				THEN COALESCE(sp.price, ss.unit_amount::float / 100.0, 0.0)
				WHEN COALESCE(sp.interval, 'month') = 'year' 
				THEN COALESCE(sp.price, ss.unit_amount::float / 100.0, 0.0) / 12.0
				ELSE 0.0
			END as mrr_contribution,
			
			-- Simple LTV estimate: MRR * 24 months
			CASE 
				WHEN COALESCE(sp.interval, 'month') = 'month' 
				THEN COALESCE(sp.price, ss.unit_amount::float / 100.0, 0.0) * 24
				WHEN COALESCE(sp.interval, 'month') = 'year' 
				THEN COALESCE(sp.price, ss.unit_amount::float / 100.0, 0.0) * 2
				ELSE 0.0
			END as ltv_estimate,
			
			CASE 
				WHEN ss.current_period_end IS NOT NULL AND ss.current_period_end > NOW()
				THEN EXTRACT(DAY FROM (ss.current_period_end - NOW()))::int
				ELSE NULL
			END as days_until_expiry,
			
			CASE 
				WHEN ss.current_period_start IS NOT NULL AND ss.current_period_end IS NOT NULL
				THEN EXTRACT(DAY FROM (ss.current_period_end - ss.current_period_start))::int
				ELSE NULL
			END as subscription_duration_days,
			
			-- Admin Tracking
			u.last_login,
			u.created_at,
			u.updated_at,
			COALESCE(u.manual_video_access, false) as manual_access_granted,
			u.stripe_customer_id,
			
			-- Computed Fields
			CASE 
				WHEN ss.current_period_end IS NOT NULL 
					AND ss.current_period_end > NOW() 
					AND EXTRACT(DAY FROM (ss.current_period_end - NOW())) <= 7
				THEN true
				ELSE false
			END as is_expiring_soon,
			
			EXTRACT(DAY FROM (NOW() - u.created_at))::int as account_age_days
			
		FROM users u
		LEFT JOIN subscription_plans sp ON u.sub_id = sp.id AND sp.is_active = true AND sp.deleted_at IS NULL
		LEFT JOIN stripe_customers sc ON (
			u.stripe_customer_id = sc.stripe_id OR 
			sc.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}'))
		)
		LEFT JOIN stripe_subscriptions ss ON sc.id = ss.customer_id AND ss.status IN ('active', 'trialing') AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
		LEFT JOIN stripe_prices sp_price ON ss.price_id = sp_price.id
		LEFT JOIN stripe_products spr ON sp_price.product_id = spr.id
		WHERE u.is_active = true
	`

	// Apply filters
	args := []interface{}{}
	argIndex := 1

	if filters != nil {
		if filters.Search != "" {
			query += fmt.Sprintf(" AND (LOWER(u.email) LIKE LOWER($%d) OR LOWER(u.first_name) LIKE LOWER($%d) OR LOWER(u.last_name) LIKE LOWER($%d))", argIndex, argIndex, argIndex)
			args = append(args, "%"+filters.Search+"%")
			argIndex++
		}

		if filters.HasActivePlan != nil {
			if *filters.HasActivePlan {
				query += " AND ((u.sub_id IS NOT NULL AND sp.is_active = true AND sp.deleted_at IS NULL) OR (ss.status IN ('active', 'trialing') AND ss.current_period_end > NOW() AND spr.video_approved = true))"
			} else {
				query += " AND NOT ((u.sub_id IS NOT NULL AND sp.is_active = true AND sp.deleted_at IS NULL) OR (ss.status IN ('active', 'trialing') AND ss.current_period_end > NOW() AND spr.video_approved = true))"
			}
		}

		if filters.HasVideoAccess != nil {
			if *filters.HasVideoAccess {
				query += " AND ((u.sub_id IS NOT NULL AND sp.is_active = true AND sp.deleted_at IS NULL) OR (ss.status IN ('active', 'trialing') AND ss.current_period_end > NOW() AND spr.video_approved = true) OR u.manual_video_access = true)"
			} else {
				query += " AND NOT ((u.sub_id IS NOT NULL AND sp.is_active = true AND sp.deleted_at IS NULL) OR (ss.status IN ('active', 'trialing') AND ss.current_period_end > NOW() AND spr.video_approved = true) OR u.manual_video_access = true)"
			}
		}

		// VideoAccessSource filter removed - plans are now the only source

		if filters.PlanType != nil {
			switch *filters.PlanType {
			case "premium":
				query += " AND (LOWER(COALESCE(sp.name, ss.product_name, '')) LIKE '%premium%' OR LOWER(COALESCE(sp.name, ss.product_name, '')) LIKE '%yearly%')"
			case "basic":
				query += " AND LOWER(COALESCE(sp.name, ss.product_name, '')) LIKE '%basic%'"
			case "none":
				query += " AND (sp.name IS NULL AND ss.product_name IS NULL)"
			}
		}

		if filters.IsExpiringSoon != nil && *filters.IsExpiringSoon {
			query += " AND ss.current_period_end IS NOT NULL AND ss.current_period_end > NOW() AND EXTRACT(DAY FROM (ss.current_period_end - NOW())) <= 7"
		}

		if filters.EmailVerified != nil {
			query += fmt.Sprintf(" AND u.email_verified = $%d", argIndex)
			args = append(args, *filters.EmailVerified)
			argIndex++
		}

		if filters.Role != nil {
			query += fmt.Sprintf(" AND u.role = $%d", argIndex)
			args = append(args, *filters.Role)
			argIndex++
		}

		if filters.CreatedDateFrom != nil {
			query += fmt.Sprintf(" AND u.created_at >= $%d", argIndex)
			args = append(args, *filters.CreatedDateFrom)
			argIndex++
		}

		if filters.CreatedDateTo != nil {
			query += fmt.Sprintf(" AND u.created_at <= $%d", argIndex)
			args = append(args, *filters.CreatedDateTo)
			argIndex++
		}

		if filters.LastLoginFrom != nil {
			query += fmt.Sprintf(" AND u.last_login >= $%d", argIndex)
			args = append(args, *filters.LastLoginFrom)
			argIndex++
		}

		if filters.LastLoginTo != nil {
			query += fmt.Sprintf(" AND u.last_login <= $%d", argIndex)
			args = append(args, *filters.LastLoginTo)
			argIndex++
		}
	}

	// Get total count for pagination - use a simpler approach
	countQuery := `
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		LEFT JOIN subscription_plans sp ON u.sub_id = sp.id AND sp.is_active = true AND sp.deleted_at IS NULL
		LEFT JOIN stripe_customers sc ON (
			u.stripe_customer_id = sc.stripe_id OR 
			sc.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}'))
		)
		LEFT JOIN stripe_subscriptions ss ON sc.id = ss.customer_id AND ss.status IN ('active', 'trialing') AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
		LEFT JOIN stripe_prices sp_price ON ss.price_id = sp_price.id
		LEFT JOIN stripe_products spr ON sp_price.product_id = spr.id
		WHERE u.is_active = true
	`

	// Apply the same filters to the count query
	countArgs := []interface{}{}
	countArgIndex := 1

	if filters != nil {
		if filters.Search != "" {
			countQuery += fmt.Sprintf(" AND (LOWER(u.email) LIKE LOWER($%d) OR LOWER(u.first_name) LIKE LOWER($%d) OR LOWER(u.last_name) LIKE LOWER($%d))", countArgIndex, countArgIndex, countArgIndex)
			countArgs = append(countArgs, "%"+filters.Search+"%")
			countArgIndex++
		}

		if filters.HasActivePlan != nil {
			if *filters.HasActivePlan {
				countQuery += " AND ((u.sub_id IS NOT NULL AND sp.is_active = true AND sp.deleted_at IS NULL) OR (ss.status IN ('active', 'trialing') AND ss.current_period_end > NOW() AND spr.video_approved = true))"
			} else {
				countQuery += " AND NOT ((u.sub_id IS NOT NULL AND sp.is_active = true AND sp.deleted_at IS NULL) OR (ss.status IN ('active', 'trialing') AND ss.current_period_end > NOW() AND spr.video_approved = true))"
			}
		}

		if filters.HasVideoAccess != nil {
			if *filters.HasVideoAccess {
				countQuery += " AND ((u.sub_id IS NOT NULL AND sp.is_active = true AND sp.deleted_at IS NULL) OR (ss.status IN ('active', 'trialing') AND ss.current_period_end > NOW() AND spr.video_approved = true) OR u.manual_video_access = true)"
			} else {
				countQuery += " AND NOT ((u.sub_id IS NOT NULL AND sp.is_active = true AND sp.deleted_at IS NULL) OR (ss.status IN ('active', 'trialing') AND ss.current_period_end > NOW() AND spr.video_approved = true) OR u.manual_video_access = true)"
			}
		}

		//if filters.VideoAccessSource != nil {
		//	switch *filters.VideoAccessSource {
		//	case "manual":
		//		countQuery += " AND u.manual_video_access = true"
		//	case "plan":
		//		countQuery += " AND ((u.sub_id IS NOT NULL AND sp.is_active = true AND sp.deleted_at IS NULL) OR (ss.status IN ('active', 'trialing') AND ss.current_period_end > NOW() AND spr.video_approved = true)) AND COALESCE(u.manual_video_access, false) = false"
		//	case "none":
		//		countQuery += " AND NOT ((u.sub_id IS NOT NULL AND sp.is_active = true AND sp.deleted_at IS NULL) OR (ss.status IN ('active', 'trialing') AND ss.current_period_end > NOW() AND spr.video_approved = true) OR u.manual_video_access = true)"
		//	}
		//}

		if filters.PlanType != nil {
			switch *filters.PlanType {
			case "premium":
				countQuery += " AND (LOWER(COALESCE(sp.name, ss.product_name, '')) LIKE '%premium%' OR LOWER(COALESCE(sp.name, ss.product_name, '')) LIKE '%yearly%')"
			case "basic":
				countQuery += " AND LOWER(COALESCE(sp.name, ss.product_name, '')) LIKE '%basic%'"
			case "none":
				countQuery += " AND (sp.name IS NULL AND ss.product_name IS NULL)"
			}
		}

		if filters.IsExpiringSoon != nil && *filters.IsExpiringSoon {
			countQuery += " AND ss.current_period_end IS NOT NULL AND ss.current_period_end > NOW() AND EXTRACT(DAY FROM (ss.current_period_end - NOW())) <= 7"
		}

		if filters.EmailVerified != nil {
			countQuery += fmt.Sprintf(" AND u.email_verified = $%d", countArgIndex)
			countArgs = append(countArgs, *filters.EmailVerified)
			countArgIndex++
		}

		if filters.Role != nil {
			countQuery += fmt.Sprintf(" AND u.role = $%d", countArgIndex)
			countArgs = append(countArgs, *filters.Role)
			countArgIndex++
		}

		if filters.CreatedDateFrom != nil {
			countQuery += fmt.Sprintf(" AND u.created_at >= $%d", countArgIndex)
			countArgs = append(countArgs, *filters.CreatedDateFrom)
			countArgIndex++
		}

		if filters.CreatedDateTo != nil {
			countQuery += fmt.Sprintf(" AND u.created_at <= $%d", countArgIndex)
			countArgs = append(countArgs, *filters.CreatedDateTo)
			countArgIndex++
		}

		if filters.LastLoginFrom != nil {
			countQuery += fmt.Sprintf(" AND u.last_login >= $%d", countArgIndex)
			countArgs = append(countArgs, *filters.LastLoginFrom)
			countArgIndex++
		}

		if filters.LastLoginTo != nil {
			countQuery += fmt.Sprintf(" AND u.last_login <= $%d", countArgIndex)
			countArgs = append(countArgs, *filters.LastLoginTo)
			countArgIndex++
		}
	}

	var totalCount int
	err := s.db.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		log.Printf("Error getting subscriber count: %v", err)
		return nil, err
	}

	// Add ordering and pagination
	query += " ORDER BY u.first_name, u.last_name, u.email"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Execute main query
	rows, err := s.db.Query(query, args...)
	if err != nil {
		log.Printf("Error executing enhanced subscribers query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var subscribers []EnhancedSubscriber
	for rows.Next() {
		var sub EnhancedSubscriber
		err := rows.Scan(
			&sub.ID,
			&sub.Email,
			&sub.FirstName,
			&sub.LastName,
			&sub.FullName,
			&sub.EmailVerified,
			&sub.Role,
			&sub.HasActivePlan,
			&sub.HasVideoAccess,
			&sub.PlanName,
			&sub.PlanLegacyStatus,
			&sub.PlanType,
			&sub.PlanStartDate,
			&sub.BillingPeriodStart,
			&sub.BillingPeriodEnd,
			&sub.PlanStatus,
			&sub.PlanPrice,
			&sub.PlanCurrency,
			&sub.MRRContribution,
			&sub.LTVEstimate,
			&sub.DaysUntilExpiry,
			&sub.SubscriptionDurationDays,
			&sub.LastLogin,
			&sub.CreatedAt,
			&sub.UpdatedAt,
			&sub.ManualAccessGranted,
			&sub.StripeCustomerID,
			&sub.IsExpiringSoon,
			&sub.AccountAgeDays,
		)
		if err != nil {
			log.Printf("Error scanning subscriber row: %v", err)
			continue
		}

		// Calculate is_high_value based on average MRR (simple implementation)
		sub.IsHighValue = sub.MRRContribution > 20.0 // $20+ MRR considered high value

		subscribers = append(subscribers, sub)
	}

	// Get KPIs
	kpis, err := s.GetKPIs()
	if err != nil {
		log.Printf("Error getting KPIs: %v", err)
		// Continue without KPIs rather than failing
		kpis = &SubscriberKPIs{}
	}

	// Calculate pagination
	totalPages := (totalCount + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	response := &SubscriberResponse{
		Subscribers: subscribers,
		TotalCount:  totalCount,
		KPIs:        *kpis,
		Pagination: struct {
			Page       int  `json:"page"`
			Limit      int  `json:"limit"`
			TotalPages int  `json:"total_pages"`
			HasNext    bool `json:"has_next"`
			HasPrev    bool `json:"has_prev"`
		}{
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
		},
	}

	return response, nil
}

// GetKPIs calculates key performance indicators for subscribers
func (s *EnhancedSubscriberService) GetKPIs() (*SubscriberKPIs, error) {
	query := `
		SELECT 
			COUNT(DISTINCT u.id) as total_subscribers,
			COUNT(DISTINCT CASE 
				WHEN (u.sub_id IS NOT NULL AND sp.is_active = true AND sp.deleted_at IS NULL) 
					OR (ss.status IN ('active', 'trialing') AND ss.current_period_end > NOW() AND spr.video_approved = true) 
				THEN u.id 
			END) as active_subscribers,
			COUNT(DISTINCT CASE 
				WHEN (u.sub_id IS NOT NULL AND sp.is_active = true AND sp.deleted_at IS NULL) 
					OR (ss.status IN ('active', 'trialing') AND ss.current_period_end > NOW() AND spr.video_approved = true)
					OR u.manual_video_access = true
				THEN u.id 
			END) as video_access_users,
			COALESCE(SUM(
				CASE 
					WHEN COALESCE(sp.interval, 'month') = 'month' 
					THEN COALESCE(sp.price, ss.unit_amount::float / 100.0, 0.0)
					WHEN COALESCE(sp.interval, 'month') = 'year' 
					THEN COALESCE(sp.price, ss.unit_amount::float / 100.0, 0.0) / 12.0
					ELSE 0.0
				END
			), 0.0) as total_mrr,
			COALESCE(AVG(
				CASE 
					WHEN ss.current_period_end IS NOT NULL AND ss.current_period_end > NOW()
					THEN EXTRACT(DAY FROM (ss.current_period_end - NOW()))
					ELSE NULL
				END
			), 0.0) as avg_days_to_expiry,
			COUNT(DISTINCT CASE 
				WHEN ss.current_period_end IS NOT NULL 
					AND ss.current_period_end > NOW() 
					AND EXTRACT(DAY FROM (ss.current_period_end - NOW())) <= 7
				THEN u.id 
			END) as churn_risk_count,
			COUNT(DISTINCT CASE 
				WHEN LOWER(COALESCE(sp.name, ss.product_name, '')) LIKE '%premium%' 
					OR LOWER(COALESCE(sp.name, ss.product_name, '')) LIKE '%yearly%' 
				THEN u.id 
			END) as premium_users,
			COUNT(DISTINCT CASE 
				WHEN LOWER(COALESCE(sp.name, ss.product_name, '')) LIKE '%basic%' 
				THEN u.id 
			END) as basic_users,
			COUNT(DISTINCT CASE 
				WHEN u.manual_video_access = true 
				THEN u.id 
			END) as manual_access_users
		FROM users u
		LEFT JOIN subscription_plans sp ON u.sub_id = sp.id AND sp.is_active = true AND sp.deleted_at IS NULL
		LEFT JOIN stripe_customers sc ON (
			u.stripe_customer_id = sc.stripe_id OR 
			sc.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}'))
		)
		LEFT JOIN stripe_subscriptions ss ON sc.id = ss.customer_id AND ss.status IN ('active', 'trialing') AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
		LEFT JOIN stripe_prices sp_price ON ss.price_id = sp_price.id
		LEFT JOIN stripe_products spr ON sp_price.product_id = spr.id
		WHERE u.is_active = true
	`

	var kpis SubscriberKPIs
	err := s.db.QueryRow(query).Scan(
		&kpis.TotalSubscribers,
		&kpis.ActiveSubscribers,
		&kpis.VideoAccessUsers,
		&kpis.TotalMRR,
		&kpis.AvgDaysToExpiry,
		&kpis.ChurnRiskCount,
		&kpis.PremiumUsers,
		&kpis.BasicUsers,
		&kpis.ManualAccessUsers,
	)

	if err != nil {
		log.Printf("Error calculating KPIs: %v", err)
		return nil, err
	}

	return &kpis, nil
}
