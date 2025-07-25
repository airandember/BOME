package database

import (
	"database/sql"
	"fmt" // Added for debugging

	// Added for logging
	"log"
	"strings" // Added for strings.Join
	"time"
)

// SubscriptionPlan represents a subscription plan in the database
type SubscriptionPlan struct {
	ID                 int
	Name               string
	Description        string
	Price              float64
	Currency           string
	Interval           string // 'month', 'year', 'week', 'day'
	IntervalCount      int
	StripePriceID      sql.NullString
	Features           sql.NullString // JSON array of features
	IsActive           bool
	IsPromoted         bool
	PromotionEndDate   sql.NullTime
	SortOrder          int
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          sql.NullTime
	ShortDesc          sql.NullString
	IsDeleted          sql.NullBool // Nullable boolean for soft delete status
	PromotionStartDate sql.NullTime
	PromotionHistory   sql.NullString // JSON array of promotion history
	SubType            int            // 100 = standard plan, 300 = promotional plan
}

// CreateSubscriptionPlan creates a new subscription plan
func (db *DB) CreateSubscriptionPlan(plan *SubscriptionPlan) (*SubscriptionPlan, error) {
	log.Printf("Creating subscription plan: %+v", plan)
	query := `
		INSERT INTO subscription_plans (
			name, description, price, currency, interval, interval_count, stripe_price_id, features, 
			is_active, is_promoted, promotion_end_date, sort_order, created_at, updated_at, deleted_at, 
			short_desc, is_deleted, promotion_start_date, promotion_history, sub_type
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		) RETURNING 
			id, name, description, price, currency, interval, interval_count, stripe_price_id, features, 
			is_active, is_promoted, promotion_end_date, sort_order, created_at, updated_at, deleted_at, 
			short_desc, is_deleted, promotion_start_date, promotion_history, sub_type
	`

	row := db.QueryRow(query,
		plan.Name, plan.Description, plan.Price, plan.Currency, plan.Interval, plan.IntervalCount,
		plan.StripePriceID, plan.Features, plan.IsActive, plan.IsPromoted, plan.PromotionEndDate,
		plan.SortOrder, plan.CreatedAt, plan.UpdatedAt, plan.DeletedAt, plan.ShortDesc, plan.IsDeleted,
		plan.PromotionStartDate, plan.PromotionHistory, plan.SubType,
	)

	createdPlan := &SubscriptionPlan{}
	err := row.Scan(
		&createdPlan.ID, &createdPlan.Name, &createdPlan.Description, &createdPlan.Price, &createdPlan.Currency,
		&createdPlan.Interval, &createdPlan.IntervalCount, &createdPlan.StripePriceID, &createdPlan.Features,
		&createdPlan.IsActive, &createdPlan.IsPromoted, &createdPlan.PromotionEndDate, &createdPlan.SortOrder,
		&createdPlan.CreatedAt, &createdPlan.UpdatedAt, &createdPlan.DeletedAt, &createdPlan.ShortDesc,
		&createdPlan.IsDeleted, &createdPlan.PromotionStartDate, &createdPlan.PromotionHistory, &createdPlan.SubType,
	)
	if err != nil {
		return nil, err
	}

	log.Printf("Created subscription plan: %+v", createdPlan)
	return createdPlan, nil
}

// GetAllSubscriptionPlans retrieves all subscription plans from the database
func (db *DB) GetAllSubscriptionPlans() ([]*SubscriptionPlan, error) {
	query := `
		SELECT id, name, description, price, currency, interval, interval_count, stripe_price_id, features, 
		       is_active, is_promoted, promotion_end_date, sort_order, created_at, updated_at, deleted_at, 
		       short_desc, is_deleted, promotion_start_date, promotion_history, sub_type
		FROM subscription_plans 
		WHERE deleted_at IS NULL
		ORDER BY sort_order ASC, created_at ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*SubscriptionPlan
	for rows.Next() {
		plan := &SubscriptionPlan{}
		err := rows.Scan(
			&plan.ID, &plan.Name, &plan.Description, &plan.Price, &plan.Currency, &plan.Interval,
			&plan.IntervalCount, &plan.StripePriceID, &plan.Features, &plan.IsActive, &plan.IsPromoted,
			&plan.PromotionEndDate, &plan.SortOrder, &plan.CreatedAt, &plan.UpdatedAt, &plan.DeletedAt,
			&plan.ShortDesc, &plan.IsDeleted, &plan.PromotionStartDate, &plan.PromotionHistory, &plan.SubType,
		)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	log.Printf("Retrieved %d subscription plans", len(plans))
	return plans, nil
}

// GetSubscriptionPlanByID retrieves a subscription plan by ID
func (db *DB) GetSubscriptionPlanByID(id int) (*SubscriptionPlan, error) {
	plan := &SubscriptionPlan{}
	err := db.QueryRow(`
		SELECT id, name, description, price, currency, interval, interval_count, stripe_price_id, features, 
		       is_active, is_promoted, promotion_end_date, sort_order, created_at, updated_at, deleted_at, 
		       short_desc, is_deleted, promotion_start_date, promotion_history, sub_type
		FROM subscription_plans WHERE id = $1
	`, id).Scan(
		&plan.ID, &plan.Name, &plan.Description, &plan.Price, &plan.Currency, &plan.Interval,
		&plan.IntervalCount, &plan.StripePriceID, &plan.Features, &plan.IsActive, &plan.IsPromoted,
		&plan.PromotionEndDate, &plan.SortOrder, &plan.CreatedAt, &plan.UpdatedAt, &plan.DeletedAt,
		&plan.ShortDesc, &plan.IsDeleted, &plan.PromotionStartDate, &plan.PromotionHistory, &plan.SubType,
	)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

// GetActiveSubscriptionPlans retrieves all active subscription plans
func (db *DB) GetActiveSubscriptionPlans() ([]*SubscriptionPlan, error) {
	query := `
		SELECT id, name, description, price, currency, interval, interval_count, stripe_price_id, features, 
		       is_active, is_promoted, promotion_end_date, sort_order, created_at, updated_at, deleted_at, 
		       short_desc, is_deleted, promotion_start_date, promotion_history, sub_type
		FROM subscription_plans 
		WHERE is_active = true AND deleted_at IS NULL
		ORDER BY sort_order ASC, created_at ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*SubscriptionPlan
	for rows.Next() {
		plan := &SubscriptionPlan{}
		err := rows.Scan(
			&plan.ID, &plan.Name, &plan.Description, &plan.Price, &plan.Currency, &plan.Interval,
			&plan.IntervalCount, &plan.StripePriceID, &plan.Features, &plan.IsActive, &plan.IsPromoted,
			&plan.PromotionEndDate, &plan.SortOrder, &plan.CreatedAt, &plan.UpdatedAt, &plan.DeletedAt,
			&plan.ShortDesc, &plan.IsDeleted, &plan.PromotionStartDate, &plan.PromotionHistory, &plan.SubType,
		)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	return plans, nil
}

// GetPromotedSubscriptionPlans retrieves all promoted subscription plans
func (db *DB) GetPromotedSubscriptionPlans() ([]*SubscriptionPlan, error) {
	query := `
		SELECT id, name, description, price, currency, interval, interval_count, stripe_price_id, features, 
		       is_active, is_promoted, promotion_end_date, sort_order, created_at, updated_at, deleted_at, 
		       short_desc, is_deleted, promotion_start_date, promotion_history, sub_type
		FROM subscription_plans 
		WHERE is_promoted = true AND deleted_at IS NULL
		ORDER BY sort_order ASC, created_at ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*SubscriptionPlan
	for rows.Next() {
		plan := &SubscriptionPlan{}
		err := rows.Scan(
			&plan.ID, &plan.Name, &plan.Description, &plan.Price, &plan.Currency, &plan.Interval,
			&plan.IntervalCount, &plan.StripePriceID, &plan.Features, &plan.IsActive, &plan.IsPromoted,
			&plan.PromotionEndDate, &plan.SortOrder, &plan.CreatedAt, &plan.UpdatedAt, &plan.DeletedAt,
			&plan.ShortDesc, &plan.IsDeleted, &plan.PromotionStartDate, &plan.PromotionHistory, &plan.SubType,
		)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	return plans, nil
}

// UpdateSubscriptionPlan updates a subscription plan in the database
func (db *DB) UpdateSubscriptionPlan(id int, updates map[string]interface{}) (*SubscriptionPlan, error) {
	log.Printf("Updating subscription plan %d with updates: %+v", id, updates)
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1
	for k, v := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", k, argIdx))
		args = append(args, v)
		argIdx++
	}
	setParts = append(setParts, fmt.Sprintf("updated_at = NOW()"))
	query := fmt.Sprintf(`
		UPDATE subscription_plans SET %s WHERE id = $%d 
		RETURNING id, name, description, price, currency, interval, interval_count, stripe_price_id, features, 
		          is_active, is_promoted, promotion_end_date, sort_order, created_at, updated_at, deleted_at, 
		          short_desc, is_deleted, promotion_start_date, promotion_history, sub_type
	`, strings.Join(setParts, ", "), argIdx)
	args = append(args, id)
	row := db.QueryRow(query, args...)
	plan := &SubscriptionPlan{}
	if err := row.Scan(
		&plan.ID, &plan.Name, &plan.Description, &plan.Price, &plan.Currency, &plan.Interval,
		&plan.IntervalCount, &plan.StripePriceID, &plan.Features, &plan.IsActive, &plan.IsPromoted,
		&plan.PromotionEndDate, &plan.SortOrder, &plan.CreatedAt, &plan.UpdatedAt, &plan.DeletedAt,
		&plan.ShortDesc, &plan.IsDeleted, &plan.PromotionStartDate, &plan.PromotionHistory, &plan.SubType,
	); err != nil {
		log.Printf("UpdateSubscriptionPlan error: %v", err)
		return nil, err
	}

	log.Printf("Updated plan: %+v", plan)
	return plan, nil
}

// SoftDeleteSubscriptionPlan marks a plan as deleted
func (db *DB) SoftDeleteSubscriptionPlan(id int) error {
	log.Printf("Soft deleting subscription plan %d", id)
	_, err := db.Exec(`UPDATE subscription_plans SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1`, id)
	return err
}

// GetSubscriptionPlanByStripePriceID retrieves a subscription plan by Stripe price ID
func (db *DB) GetSubscriptionPlanByStripePriceID(stripePriceID string) (*SubscriptionPlan, error) {
	plan := &SubscriptionPlan{}
	err := db.QueryRow(`
		SELECT id, name, description, price, currency, interval, interval_count, stripe_price_id, features, 
		       is_active, is_promoted, promotion_end_date, sort_order, created_at, updated_at, deleted_at, 
		       short_desc, is_deleted, promotion_start_date, promotion_history, sub_type
		FROM subscription_plans WHERE stripe_price_id = $1 AND deleted_at IS NULL
	`, stripePriceID).Scan(
		&plan.ID, &plan.Name, &plan.Description, &plan.Price, &plan.Currency, &plan.Interval,
		&plan.IntervalCount, &plan.StripePriceID, &plan.Features, &plan.IsActive, &plan.IsPromoted,
		&plan.PromotionEndDate, &plan.SortOrder, &plan.CreatedAt, &plan.UpdatedAt, &plan.DeletedAt,
		&plan.ShortDesc, &plan.IsDeleted, &plan.PromotionStartDate, &plan.PromotionHistory, &plan.SubType,
	)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

// GetSubscriptionPlansWithFilters retrieves subscription plans with optional filters
func (db *DB) GetSubscriptionPlansWithFilters(isActive *bool, isPromoted *bool, subType *int) ([]*SubscriptionPlan, error) {
	query := `
		SELECT id, name, description, price, currency, interval, interval_count, stripe_price_id, features, 
		       is_active, is_promoted, promotion_end_date, sort_order, created_at, updated_at, deleted_at, 
		       short_desc, is_deleted, promotion_start_date, promotion_history, sub_type
		FROM subscription_plans WHERE deleted_at IS NULL
	`

	var args []interface{}
	argCount := 1

	if isActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", argCount)
		args = append(args, *isActive)
		argCount++
	}

	if isPromoted != nil {
		query += fmt.Sprintf(" AND is_promoted = $%d", argCount)
		args = append(args, *isPromoted)
		argCount++
	}

	if subType != nil {
		query += fmt.Sprintf(" AND sub_type = $%d", argCount)
		args = append(args, *subType)
		argCount++
	}

	query += " ORDER BY sort_order ASC, created_at ASC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*SubscriptionPlan
	for rows.Next() {
		plan := &SubscriptionPlan{}
		err := rows.Scan(
			&plan.ID, &plan.Name, &plan.Description, &plan.Price, &plan.Currency, &plan.Interval,
			&plan.IntervalCount, &plan.StripePriceID, &plan.Features, &plan.IsActive, &plan.IsPromoted,
			&plan.PromotionEndDate, &plan.SortOrder, &plan.CreatedAt, &plan.UpdatedAt, &plan.DeletedAt,
			&plan.ShortDesc, &plan.IsDeleted, &plan.PromotionStartDate, &plan.PromotionHistory, &plan.SubType,
		)
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

// TestGetAllPlans is a debug function to check what's in the database
func (db *DB) TestGetAllPlans() error {
	fmt.Println("=== Testing database content ===")

	// Simple query to see all plans
	rows, err := db.Query("SELECT id, name, is_active, deleted_at FROM subscription_plans ORDER BY id")
	if err != nil {
		fmt.Printf("Test query error: %v\n", err)
		return err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var id int
		var name string
		var isActive bool
		var deletedAt sql.NullTime

		err := rows.Scan(&id, &name, &isActive, &deletedAt)
		if err != nil {
			fmt.Printf("Test scan error: %v\n", err)
			return err
		}

		count++
		fmt.Printf("Test row %d: ID=%d, Name=%s, IsActive=%v, IsDeleted=%v\n", count, id, name, isActive, deletedAt.Valid)
	}

	fmt.Printf("Total test rows: %d\n", count)
	fmt.Println("=== Test completed ===")
	return nil
}
