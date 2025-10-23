package subscription

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"strings"
	"time"

	"bome-backend/infrastructure/database"
)

// WebSocketHub interface to avoid circular imports
type WebSocketHub interface {
	BroadcastSubscriberCreated(subscriber interface{})
	BroadcastSubscriberUpdated(subscriber interface{})
	BroadcastKPIUpdate(kpis interface{})
	BroadcastEvent(eventType string, data map[string]interface{}, message string)
}

// SubscriberService handles business logic for subscribers (users with active subscriptions)
type SubscriberService struct {
	db  *database.DB
	hub WebSocketHub
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

// NonSubscriber represents a user without an active subscription
type NonSubscriber struct {
	ID                     int        `json:"id"`
	Email                  string     `json:"email"`
	FirstName              string     `json:"first_name"`
	LastName               string     `json:"last_name"`
	Role                   string     `json:"role"`
	EmailVerified          bool       `json:"email_verified"`
	LastLogin              *time.Time `json:"last_login,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
	HasSubscriptionHistory bool       `json:"has_subscription_history"`
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
	HasSubscriptionHistory *bool `json:"has_subscription_history"`
}

// NonSubscriberFilters represents filters for non-subscriber queries
type NonSubscriberFilters struct {
	Search        string     `json:"search"`
	EmailVerified *bool      `json:"email_verified"`
	Role          *string    `json:"role"`
	LastLogin     *time.Time `json:"last_login"`
	CreatedDate   *time.Time `json:"created_date"`
}

// SubscriberHistory represents a subscriber's subscription history
type SubscriberHistory struct {
	ID                 int        `json:"id"`
	UserID             int        `json:"user_id"`
	SubscriptionPlanID int        `json:"subscription_plan_id"`
	PlanName           string     `json:"plan_name"`
	StartDate          time.Time  `json:"start_date"`
	EndDate            *time.Time `json:"end_date"`
	Status             string     `json:"status"`
	CreatedAt          time.Time  `json:"created_at"`
}

// NewSubscriberService creates a new subscriber service
func NewSubscriberService(db *database.DB, hub WebSocketHub) *SubscriberService {
	return &SubscriberService{
		db:  db,
		hub: hub,
	}
}

// GetSubscribers retrieves all subscribers with optional filters and pagination
func (s *SubscriberService) GetSubscribers(page, limit int, filters *SubscriberFilters) ([]*Subscriber, int, error) {
	offset := (page - 1) * limit

	// Build query
	query := `
		SELECT 
			u.id,
			u.email,
			u.first_name,
			u.last_name,
			u.role,
			u.email_verified,
			u.sub_id,
			sp.name as plan_name,
			sp.price as plan_price,
			sp.currency as plan_currency,
			sp.interval as plan_interval,
			u.stripe_customer_id,
			u.last_login,
			u.created_at,
			u.updated_at
		FROM users u
		LEFT JOIN subscription_plans sp ON u.sub_id = sp.id
		WHERE u.is_active = true
			AND (u.sub_id IS NOT NULL OR u.stripe_customer_id IS NOT NULL)
	`

	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if filters != nil {
		if filters.Search != "" {
			query += fmt.Sprintf(" AND (LOWER(u.email) LIKE LOWER($%d) OR LOWER(u.first_name) LIKE LOWER($%d) OR LOWER(u.last_name) LIKE LOWER($%d))", argIndex, argIndex, argIndex)
			args = append(args, "%"+filters.Search+"%")
			argIndex++
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
	}

	// Get total count
	countQuery := strings.Replace(query, "SELECT u.id, u.email, u.first_name, u.last_name, u.role, u.email_verified, u.sub_id, sp.name as plan_name, sp.price as plan_price, sp.currency as plan_currency, sp.interval as plan_interval, u.stripe_customer_id, u.last_login, u.created_at, u.updated_at", "SELECT COUNT(*)", 1)
	var total int
	err := s.db.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Add ordering and pagination
	query += " ORDER BY u.created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := s.db.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var subscribers []*Subscriber
	for rows.Next() {
		var sub Subscriber
		err := rows.Scan(
			&sub.ID,
			&sub.Email,
			&sub.FirstName,
			&sub.LastName,
			&sub.Role,
			&sub.EmailVerified,
			&sub.SubID,
			&sub.PlanName,
			&sub.PlanPrice,
			&sub.PlanCurrency,
			&sub.PlanInterval,
			&sub.StripeCustomerID,
			&sub.LastLogin,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning subscriber: %v", err)
			continue
		}
		subscribers = append(subscribers, &sub)
	}

	return subscribers, total, nil
}

// GetSubscriberCount returns the total count of subscribers
func (s *SubscriberService) GetSubscriberCount() (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM users u
		WHERE u.is_active = true
			AND (u.sub_id IS NOT NULL OR u.stripe_customer_id IS NOT NULL)
	`

	var count int
	err := s.db.DB.QueryRow(query).Scan(&count)
	return count, err
}

// GetSubscriberStats returns subscriber statistics
func (s *SubscriberService) GetSubscriberStats() (*SubscriberStats, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN ss.status = 'active' THEN 1 END) as active,
			COUNT(CASE WHEN ss.status = 'trialing' THEN 1 END) as trialing,
			COUNT(CASE WHEN ss.status = 'past_due' THEN 1 END) as past_due,
			COUNT(CASE WHEN ss.status = 'canceled' THEN 1 END) as canceled,
			COALESCE(SUM(CASE WHEN sp.interval = 'month' THEN sp.price ELSE 0 END), 0) as monthly_revenue,
			COALESCE(SUM(CASE WHEN sp.interval = 'year' THEN sp.price ELSE 0 END), 0) as annual_revenue
		FROM users u
		LEFT JOIN subscription_plans sp ON u.sub_id = sp.id
		LEFT JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
		LEFT JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
		WHERE u.is_active = true
	`

	var stats SubscriberStats

	err := s.db.DB.QueryRow(query).Scan(
		&stats.TotalSubscribers,
		&stats.ActiveSubscribers,
		&stats.TrialingSubscribers,
		&stats.PastDueSubscribers,
		&stats.CanceledSubscribers,
		&stats.MonthlyRevenue,
		&stats.AnnualRevenue,
	)

	if err != nil {
		return nil, err
	}

	if stats.TotalSubscribers > 0 {
		stats.AverageRevenuePerUser = (stats.MonthlyRevenue + stats.AnnualRevenue/12) / float64(stats.TotalSubscribers)
	}

	return &stats, nil
}

// GetSubscribersByEmailVerification retrieves subscribers by email verification status
func (s *SubscriberService) GetSubscribersByEmailVerification(verified bool, page, limit int) ([]*Subscriber, int, error) {
	filters := &SubscriberFilters{
		EmailVerified: &verified,
	}
	return s.GetSubscribers(page, limit, filters)
}

// GetSubscriberCountByEmailVerification returns count by email verification status
func (s *SubscriberService) GetSubscriberCountByEmailVerification(verified bool) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM users u
		WHERE u.is_active = true
			AND u.email_verified = $1
			AND (u.sub_id IS NOT NULL OR u.stripe_customer_id IS NOT NULL)
	`

	var count int
	err := s.db.DB.QueryRow(query, verified).Scan(&count)
	return count, err
}

// GetSubscribersByPlan retrieves subscribers by plan ID
func (s *SubscriberService) GetSubscribersByPlan(planID, page, limit int) ([]*Subscriber, int, error) {
	filters := &SubscriberFilters{
		PlanID: &planID,
	}
	return s.GetSubscribers(page, limit, filters)
}

// GetSubscribersByStatus retrieves subscribers by status
func (s *SubscriberService) GetSubscribersByStatus(status string, page, limit int) ([]*Subscriber, int, error) {
	filters := &SubscriberFilters{
		Status: &status,
	}
	return s.GetSubscribers(page, limit, filters)
}

// SearchSubscribers searches subscribers by query
func (s *SubscriberService) SearchSubscribers(query string, page, limit int) ([]*Subscriber, int, error) {
	filters := &SubscriberFilters{
		Search: query,
	}
	return s.GetSubscribers(page, limit, filters)
}

// GetNonSubscribers retrieves users without active subscriptions
func (s *SubscriberService) GetNonSubscribers(page, limit int, filters *NonSubscriberFilters) ([]*NonSubscriber, int, error) {
	offset := (page - 1) * limit

	query := `
		SELECT 
			u.id,
			u.email,
			u.first_name,
			u.last_name,
			u.role,
			u.email_verified,
			u.last_login,
			u.created_at,
			u.updated_at,
			EXISTS(SELECT 1 FROM subscriber_history WHERE user_id = u.id) as has_history
		FROM users u
		WHERE u.is_active = true
			AND u.sub_id IS NULL
			AND u.stripe_customer_id IS NULL
	`

	args := []interface{}{}
	argIndex := 1

	if filters != nil {
		if filters.Search != "" {
			query += fmt.Sprintf(" AND (LOWER(u.email) LIKE LOWER($%d) OR LOWER(u.first_name) LIKE LOWER($%d) OR LOWER(u.last_name) LIKE LOWER($%d))", argIndex, argIndex, argIndex)
			args = append(args, "%"+filters.Search+"%")
			argIndex++
		}
	}

	// Get total count
	countQuery := strings.Replace(query, "SELECT u.id, u.email, u.first_name, u.last_name, u.role, u.email_verified, u.last_login, u.created_at, u.updated_at, EXISTS(SELECT 1 FROM subscriber_history WHERE user_id = u.id) as has_history", "SELECT COUNT(*)", 1)
	var total int
	err := s.db.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Add pagination
	query += " ORDER BY u.created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := s.db.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var nonSubscribers []*NonSubscriber
	for rows.Next() {
		var ns NonSubscriber
		err := rows.Scan(
			&ns.ID,
			&ns.Email,
			&ns.FirstName,
			&ns.LastName,
			&ns.Role,
			&ns.EmailVerified,
			&ns.LastLogin,
			&ns.CreatedAt,
			&ns.UpdatedAt,
			&ns.HasSubscriptionHistory,
		)
		if err != nil {
			continue
		}
		nonSubscribers = append(nonSubscribers, &ns)
	}

	return nonSubscribers, total, nil
}

// GetNonSubscriberCount returns the total count of non-subscribers
func (s *SubscriberService) GetNonSubscriberCount() (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM users u
		WHERE u.is_active = true
			AND u.sub_id IS NULL
			AND u.stripe_customer_id IS NULL
	`

	var count int
	err := s.db.DB.QueryRow(query).Scan(&count)
	return count, err
}

// ExportSubscribersToCSV exports subscribers to CSV format
func (s *SubscriberService) ExportSubscribersToCSV() (string, error) {
	subscribers, _, err := s.GetSubscribers(1, 10000, nil) // Get all subscribers (max 10k)
	if err != nil {
		return "", err
	}

	var csvBuilder strings.Builder
	writer := csv.NewWriter(&csvBuilder)

	// Write header
	header := []string{"ID", "Email", "First Name", "Last Name", "Role", "Email Verified", "Plan Name", "Plan Price", "Created At"}
	if err := writer.Write(header); err != nil {
		return "", err
	}

	// Write rows
	for _, sub := range subscribers {
		planName := ""
		if sub.PlanName != nil {
			planName = *sub.PlanName
		}
		planPrice := ""
		if sub.PlanPrice != nil {
			planPrice = fmt.Sprintf("%.2f", *sub.PlanPrice)
		}

		row := []string{
			fmt.Sprintf("%d", sub.ID),
			sub.Email,
			sub.FirstName,
			sub.LastName,
			sub.Role,
			fmt.Sprintf("%t", sub.EmailVerified),
			planName,
			planPrice,
			sub.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(row); err != nil {
			return "", err
		}
	}

	writer.Flush()
	return csvBuilder.String(), writer.Error()
}

// ExportNonSubscribersToCSV exports non-subscribers to CSV format
func (s *SubscriberService) ExportNonSubscribersToCSV() (string, error) {
	nonSubscribers, _, err := s.GetNonSubscribers(1, 10000, nil)
	if err != nil {
		return "", err
	}

	var csvBuilder strings.Builder
	writer := csv.NewWriter(&csvBuilder)

	// Write header
	header := []string{"ID", "Email", "First Name", "Last Name", "Role", "Email Verified", "Has History", "Created At"}
	if err := writer.Write(header); err != nil {
		return "", err
	}

	// Write rows
	for _, ns := range nonSubscribers {
		row := []string{
			fmt.Sprintf("%d", ns.ID),
			ns.Email,
			ns.FirstName,
			ns.LastName,
			ns.Role,
			fmt.Sprintf("%t", ns.EmailVerified),
			fmt.Sprintf("%t", ns.HasSubscriptionHistory),
			ns.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(row); err != nil {
			return "", err
		}
	}

	writer.Flush()
	return csvBuilder.String(), writer.Error()
}

// GetSubscriberByID retrieves a subscriber by ID
func (s *SubscriberService) GetSubscriberByID(id int) (*Subscriber, error) {
	query := `
		SELECT 
			u.id,
			u.email,
			u.first_name,
			u.last_name,
			u.role,
			u.email_verified,
			u.sub_id,
			sp.name as plan_name,
			sp.price as plan_price,
			sp.currency as plan_currency,
			sp.interval as plan_interval,
			u.stripe_customer_id,
			u.last_login,
			u.created_at,
			u.updated_at
		FROM users u
		LEFT JOIN subscription_plans sp ON u.sub_id = sp.id
		WHERE u.id = $1 AND u.is_active = true
	`

	var sub Subscriber
	err := s.db.DB.QueryRow(query, id).Scan(
		&sub.ID,
		&sub.Email,
		&sub.FirstName,
		&sub.LastName,
		&sub.Role,
		&sub.EmailVerified,
		&sub.SubID,
		&sub.PlanName,
		&sub.PlanPrice,
		&sub.PlanCurrency,
		&sub.PlanInterval,
		&sub.StripeCustomerID,
		&sub.LastLogin,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &sub, nil
}

// UpdateSubscriber updates a subscriber's information
func (s *SubscriberService) UpdateSubscriber(id int, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Build SET clause
	setClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", key, argIndex))
		args = append(args, value)
		argIndex++
	}

	// Add updated_at
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add WHERE id
	args = append(args, id)

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIndex)

	_, err := s.db.DB.Exec(query, args...)

	// Broadcast update event via WebSocket
	if err == nil && s.hub != nil {
		// Get updated subscriber
		subscriber, getErr := s.GetSubscriberByID(id)
		if getErr == nil && subscriber != nil {
			s.hub.BroadcastSubscriberUpdated(subscriber)
			log.Printf("📡 Broadcasted subscriber update for ID %d", id)
		}
	}

	return err
}

// DeleteSubscriber soft deletes a subscriber
func (s *SubscriberService) DeleteSubscriber(id int) error {
	query := "UPDATE users SET is_active = false, updated_at = $1 WHERE id = $2"
	_, err := s.db.DB.Exec(query, time.Now(), id)
	return err
}

// GrantVideoAccess grants manual video access to a subscriber
func (s *SubscriberService) GrantVideoAccess(id int, expiryDate *time.Time, reason string) error {
	query := "UPDATE users SET manual_video_access = true, manual_access_expiry = $1, updated_at = $2 WHERE id = $3"
	_, err := s.db.DB.Exec(query, expiryDate, time.Now(), id)

	// Broadcast update event via WebSocket
	if err == nil && s.hub != nil {
		subscriber, getErr := s.GetSubscriberByID(id)
		if getErr == nil && subscriber != nil {
			s.hub.BroadcastSubscriberUpdated(subscriber)
			log.Printf("📡 Broadcasted video access grant for subscriber ID %d", id)
		}
	}

	return err
}

// GetSubscriberHistory retrieves a subscriber's subscription history
func (s *SubscriberService) GetSubscriberHistory(id int) ([]*SubscriberHistory, error) {
	query := `
		SELECT 
			h.id,
			h.user_id,
			h.subscription_plan_id,
			sp.name as plan_name,
			h.start_date,
			h.end_date,
			h.status,
			h.created_at
		FROM subscriber_history h
		LEFT JOIN subscription_plans sp ON h.subscription_plan_id = sp.id
		WHERE h.user_id = $1
		ORDER BY h.created_at DESC
	`

	rows, err := s.db.DB.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*SubscriberHistory
	for rows.Next() {
		var h SubscriberHistory
		err := rows.Scan(
			&h.ID,
			&h.UserID,
			&h.SubscriptionPlanID,
			&h.PlanName,
			&h.StartDate,
			&h.EndDate,
			&h.Status,
			&h.CreatedAt,
		)
		if err != nil {
			continue
		}
		history = append(history, &h)
	}

	return history, nil
}
