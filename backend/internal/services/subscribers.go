package services

import (
	"fmt"
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
	SubscriptionID       *int       `json:"subscription_id,omitempty"`
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
	PlanID        *int    `json:"plan_id"`
	Status        *string `json:"status"`
	Search        string  `json:"search"`
	EmailVerified *bool   `json:"email_verified"`
	DateRange     *struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	} `json:"date_range"`
}

// NewSubscriberService creates a new subscriber service
func NewSubscriberService(db *database.DB) *SubscriberService {
	return &SubscriberService{db: db}
}

// GetSubscribers retrieves all subscribers with optional filters
func (s *SubscriberService) GetSubscribers(limit, offset int, filters *SubscriberFilters) ([]*Subscriber, error) {
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
		WHERE u.id IN (
			SELECT DISTINCT user_id 
			FROM subscriptions 
			WHERE deleted_at IS NULL AND plan_id IS NOT NULL
		)
	`

	args := []interface{}{}
	argCount := 0

	// Add filters
	if filters != nil {
		if filters.PlanID != nil {
			argCount++
			query += fmt.Sprintf(" AND sp.id = $%d", argCount)
			args = append(args, *filters.PlanID)
		}

		if filters.Status != nil {
			argCount++
			query += fmt.Sprintf(" AND s.status = $%d", argCount)
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

		if filters.DateRange != nil {
			argCount++
			query += fmt.Sprintf(" AND u.created_at BETWEEN $%d AND $%d", argCount, argCount+1)
			args = append(args, filters.DateRange.Start, filters.DateRange.End)
			argCount++
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
		return nil, fmt.Errorf("failed to get subscribers: %w", err)
	}
	defer rows.Close()

	var subscribers []*Subscriber
	for rows.Next() {
		subscriber := &Subscriber{}
		err := rows.Scan(
			&subscriber.ID, &subscriber.Email, &subscriber.FirstName, &subscriber.LastName,
			&subscriber.Role, &subscriber.EmailVerified, &subscriber.StripeCustomerID,
			&subscriber.LastLogin, &subscriber.CreatedAt, &subscriber.UpdatedAt,
			&subscriber.SubscriptionID, &subscriber.SubscriptionStatus,
			&subscriber.CurrentPeriodStart, &subscriber.CurrentPeriodEnd,
			&subscriber.StripeSubscriptionID, &subscriber.PlanID, &subscriber.PlanName,
			&subscriber.PlanPrice, &subscriber.PlanCurrency,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscriber: %w", err)
		}
		subscribers = append(subscribers, subscriber)
	}

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

// GetSubscriberCount returns the total count of subscribers
func (s *SubscriberService) GetSubscriberCount(filters *SubscriberFilters) (int, error) {
	query := `
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		LEFT JOIN subscriptions s ON u.id = s.user_id AND s.deleted_at IS NULL
		LEFT JOIN subscription_plans sp ON s.plan_id = sp.id
		WHERE u.id IN (
			SELECT DISTINCT user_id 
			FROM subscriptions 
			WHERE deleted_at IS NULL AND plan_id IS NOT NULL
		)
	`

	args := []interface{}{}
	argCount := 0

	// Add filters
	if filters != nil {
		if filters.PlanID != nil {
			argCount++
			query += fmt.Sprintf(" AND sp.id = $%d", argCount)
			args = append(args, *filters.PlanID)
		}

		if filters.Status != nil {
			argCount++
			query += fmt.Sprintf(" AND s.status = $%d", argCount)
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

		if filters.DateRange != nil {
			argCount++
			query += fmt.Sprintf(" AND u.created_at BETWEEN $%d AND $%d", argCount, argCount+1)
			args = append(args, filters.DateRange.Start, filters.DateRange.End)
			argCount++
		}
	}

	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get subscriber count: %w", err)
	}

	return count, nil
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
