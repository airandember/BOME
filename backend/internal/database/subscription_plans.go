package database

import (
	"database/sql"
	"encoding/json"
	"time"
)

// SubscriptionPlan represents a subscription plan that users can subscribe to
type SubscriptionPlan struct {
	ID               int
	Name             string
	Description      string
	Price            float64
	Currency         string
	Interval         string // 'monthly', 'annual', 'weekly'
	IntervalCount    int
	StripePriceID    sql.NullString
	Features         sql.NullString // JSON array of features
	IsActive         bool
	IsPromoted       bool
	PromotionEndDate sql.NullTime
	SortOrder        int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        sql.NullTime
}

// CreateSubscriptionPlan inserts a new subscription plan
func (db *DB) CreateSubscriptionPlan(name, description string, price float64, currency, interval string, intervalCount int, stripePriceID string, features []string) (*SubscriptionPlan, error) {
	featuresJSON, err := json.Marshal(features)
	if err != nil {
		return nil, err
	}

	var id int
	err = db.QueryRow(
		`INSERT INTO subscription_plans (name, description, price, currency, interval, interval_count, stripe_price_id, features, created_at, updated_at) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()) RETURNING id`,
		name, description, price, currency, interval, intervalCount, stripePriceID, featuresJSON,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return db.GetSubscriptionPlanByID(id)
}

// GetSubscriptionPlanByID retrieves a subscription plan by ID
func (db *DB) GetSubscriptionPlanByID(id int) (*SubscriptionPlan, error) {
	plan := &SubscriptionPlan{}
	err := db.QueryRow(
		`SELECT id, name, description, price, currency, interval, interval_count, stripe_price_id, features, 
		        is_active, is_promoted, promotion_end_date, sort_order, created_at, updated_at, deleted_at 
		 FROM subscription_plans WHERE id = $1 AND deleted_at IS NULL`,
		id,
	).Scan(&plan.ID, &plan.Name, &plan.Description, &plan.Price, &plan.Currency, &plan.Interval, &plan.IntervalCount,
		&plan.StripePriceID, &plan.Features, &plan.IsActive, &plan.IsPromoted, &plan.PromotionEndDate,
		&plan.SortOrder, &plan.CreatedAt, &plan.UpdatedAt, &plan.DeletedAt)
	if err != nil {
		return nil, err
	}
	return plan, nil
}

// GetActiveSubscriptionPlans retrieves all active subscription plans
func (db *DB) GetActiveSubscriptionPlans() ([]*SubscriptionPlan, error) {
	rows, err := db.Query(
		`SELECT id, name, description, price, currency, interval, interval_count, stripe_price_id, features, 
		        is_active, is_promoted, promotion_end_date, sort_order, created_at, updated_at, deleted_at 
		 FROM subscription_plans 
		 WHERE is_active = true AND deleted_at IS NULL 
		 ORDER BY sort_order ASC, created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*SubscriptionPlan
	for rows.Next() {
		plan := &SubscriptionPlan{}
		err := rows.Scan(&plan.ID, &plan.Name, &plan.Description, &plan.Price, &plan.Currency, &plan.Interval,
			&plan.IntervalCount, &plan.StripePriceID, &plan.Features, &plan.IsActive, &plan.IsPromoted,
			&plan.PromotionEndDate, &plan.SortOrder, &plan.CreatedAt, &plan.UpdatedAt, &plan.DeletedAt)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	return plans, nil
}

// GetPromotedSubscriptionPlans retrieves currently promoted subscription plans
func (db *DB) GetPromotedSubscriptionPlans() ([]*SubscriptionPlan, error) {
	rows, err := db.Query(
		`SELECT id, name, description, price, currency, interval, interval_count, stripe_price_id, features, 
		        is_active, is_promoted, promotion_end_date, sort_order, created_at, updated_at, deleted_at 
		 FROM subscription_plans 
		 WHERE is_promoted = true AND is_active = true AND deleted_at IS NULL 
		 AND (promotion_end_date IS NULL OR promotion_end_date > NOW())
		 ORDER BY sort_order ASC, created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*SubscriptionPlan
	for rows.Next() {
		plan := &SubscriptionPlan{}
		err := rows.Scan(&plan.ID, &plan.Name, &plan.Description, &plan.Price, &plan.Currency, &plan.Interval,
			&plan.IntervalCount, &plan.StripePriceID, &plan.Features, &plan.IsActive, &plan.IsPromoted,
			&plan.PromotionEndDate, &plan.SortOrder, &plan.CreatedAt, &plan.UpdatedAt, &plan.DeletedAt)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	return plans, nil
}

// UpdateSubscriptionPlan updates a subscription plan
func (db *DB) UpdateSubscriptionPlan(id int, updates map[string]interface{}) error {
	// Build dynamic query based on provided updates
	query := "UPDATE subscription_plans SET updated_at = NOW()"
	args := []interface{}{}
	argCount := 1

	for field, value := range updates {
		query += ", " + field + " = $" + string(rune('0'+argCount))
		args = append(args, value)
		argCount++
	}

	query += " WHERE id = $" + string(rune('0'+argCount)) + " AND deleted_at IS NULL"
	args = append(args, id)

	_, err := db.Exec(query, args...)
	return err
}

// SoftDeleteSubscriptionPlan marks a subscription plan as deleted
func (db *DB) SoftDeleteSubscriptionPlan(id int) error {
	_, err := db.Exec(`UPDATE subscription_plans SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1`, id)
	return err
}

// GetSubscriptionPlanByStripePriceID retrieves a subscription plan by Stripe price ID
func (db *DB) GetSubscriptionPlanByStripePriceID(stripePriceID string) (*SubscriptionPlan, error) {
	plan := &SubscriptionPlan{}
	err := db.QueryRow(
		`SELECT id, name, description, price, currency, interval, interval_count, stripe_price_id, features, 
		        is_active, is_promoted, promotion_end_date, sort_order, created_at, updated_at, deleted_at 
		 FROM subscription_plans WHERE stripe_price_id = $1 AND deleted_at IS NULL`,
		stripePriceID,
	).Scan(&plan.ID, &plan.Name, &plan.Description, &plan.Price, &plan.Currency, &plan.Interval,
		&plan.IntervalCount, &plan.StripePriceID, &plan.Features, &plan.IsActive, &plan.IsPromoted,
		&plan.PromotionEndDate, &plan.SortOrder, &plan.CreatedAt, &plan.UpdatedAt, &plan.DeletedAt)
	if err != nil {
		return nil, err
	}
	return plan, nil
}

// GetSubscriptionPlansWithFilters retrieves subscription plans with optional filters
func (db *DB) GetSubscriptionPlansWithFilters(limit, offset int, isActive *bool, isPromoted *bool) ([]*SubscriptionPlan, error) {
	query := `SELECT id, name, description, price, currency, interval, interval_count, stripe_price_id, features, 
		        is_active, is_promoted, promotion_end_date, sort_order, created_at, updated_at, deleted_at 
		 FROM subscription_plans WHERE deleted_at IS NULL`
	args := []interface{}{}
	argCount := 1

	if isActive != nil {
		query += " AND is_active = $" + string(rune('0'+argCount))
		args = append(args, *isActive)
		argCount++
	}

	if isPromoted != nil {
		query += " AND is_promoted = $" + string(rune('0'+argCount))
		args = append(args, *isPromoted)
		argCount++
	}

	query += " ORDER BY sort_order ASC, created_at ASC LIMIT $" + string(rune('0'+argCount)) + " OFFSET $" + string(rune('0'+argCount+1))
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*SubscriptionPlan
	for rows.Next() {
		plan := &SubscriptionPlan{}
		err := rows.Scan(&plan.ID, &plan.Name, &plan.Description, &plan.Price, &plan.Currency, &plan.Interval,
			&plan.IntervalCount, &plan.StripePriceID, &plan.Features, &plan.IsActive, &plan.IsPromoted,
			&plan.PromotionEndDate, &plan.SortOrder, &plan.CreatedAt, &plan.UpdatedAt, &plan.DeletedAt)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	return plans, nil
}

// GetSubscriptionPlanCount returns the total count of subscription plans
func (db *DB) GetSubscriptionPlanCount(isActive *bool) (int, error) {
	query := "SELECT COUNT(*) FROM subscription_plans WHERE deleted_at IS NULL"
	args := []interface{}{}

	if isActive != nil {
		query += " AND is_active = $1"
		args = append(args, *isActive)
	}

	var count int
	err := db.QueryRow(query, args...).Scan(&count)
	return count, err
}
