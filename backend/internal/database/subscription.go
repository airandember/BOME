package database

import (
	"database/sql"
	"time"
)

// Subscription represents a user subscription
type Subscription struct {
	ID                   int
	UserID               int
	PlanID               sql.NullInt32 // New field
	StripeSubscriptionID string
	StripePriceID        string
	Status               string
	CurrentPeriodStart   *time.Time
	CurrentPeriodEnd     *time.Time
	CancelAtPeriodEnd    bool
	CancellationReason   sql.NullString  // New field
	RefundAmount         sql.NullFloat64 // New field
	RefundReason         sql.NullString  // New field
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeletedAt            sql.NullTime // New field
}

// CreateSubscription inserts a new subscription
func (db *DB) CreateSubscription(userID int, planID *int, stripeSubscriptionID, stripePriceID, status string, currentPeriodStart, currentPeriodEnd *time.Time) (*Subscription, error) {
	var id int
	var planIDValue interface{}
	if planID != nil {
		planIDValue = *planID
	} else {
		planIDValue = nil
	}

	err := db.QueryRow(
		`INSERT INTO subscriptions (user_id, plan_id, stripe_subscription_id, stripe_price_id, status, current_period_start, current_period_end, created_at, updated_at) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW()) RETURNING id`,
		userID, planIDValue, stripeSubscriptionID, stripePriceID, status, currentPeriodStart, currentPeriodEnd,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return db.GetSubscriptionByID(id)
}

// GetSubscriptionByID retrieves a subscription by ID
func (db *DB) GetSubscriptionByID(id int) (*Subscription, error) {
	subscription := &Subscription{}
	err := db.QueryRow(
		`SELECT id, user_id, plan_id, stripe_subscription_id, stripe_price_id, status, current_period_start, current_period_end, 
		        cancel_at_period_end, cancellation_reason, refund_amount, refund_reason, created_at, updated_at, deleted_at 
		 FROM subscriptions WHERE id = $1 AND deleted_at IS NULL`,
		id,
	).Scan(&subscription.ID, &subscription.UserID, &subscription.PlanID, &subscription.StripeSubscriptionID,
		&subscription.StripePriceID, &subscription.Status, &subscription.CurrentPeriodStart, &subscription.CurrentPeriodEnd,
		&subscription.CancelAtPeriodEnd, &subscription.CancellationReason, &subscription.RefundAmount,
		&subscription.RefundReason, &subscription.CreatedAt, &subscription.UpdatedAt, &subscription.DeletedAt)
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

// GetSubscriptionByUserID retrieves a user's active subscription
func (db *DB) GetSubscriptionByUserID(userID int) (*Subscription, error) {
	subscription := &Subscription{}
	err := db.QueryRow(
		`SELECT id, user_id, plan_id, stripe_subscription_id, stripe_price_id, status, current_period_start, current_period_end, 
		        cancel_at_period_end, cancellation_reason, refund_amount, refund_reason, created_at, updated_at, deleted_at 
		 FROM subscriptions WHERE user_id = $1 AND status IN ('active', 'trialing') AND deleted_at IS NULL 
		 ORDER BY created_at DESC LIMIT 1`,
		userID,
	).Scan(&subscription.ID, &subscription.UserID, &subscription.PlanID, &subscription.StripeSubscriptionID,
		&subscription.StripePriceID, &subscription.Status, &subscription.CurrentPeriodStart, &subscription.CurrentPeriodEnd,
		&subscription.CancelAtPeriodEnd, &subscription.CancellationReason, &subscription.RefundAmount,
		&subscription.RefundReason, &subscription.CreatedAt, &subscription.UpdatedAt, &subscription.DeletedAt)
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

// UpdateSubscriptionStatus updates a subscription's status
func (db *DB) UpdateSubscriptionStatus(subscriptionID int, status string) error {
	_, err := db.Exec(`UPDATE subscriptions SET status = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`, status, subscriptionID)
	return err
}

// CancelSubscription marks a subscription for cancellation with optional reason
func (db *DB) CancelSubscription(subscriptionID int, reason string) error {
	_, err := db.Exec(`UPDATE subscriptions SET cancel_at_period_end = TRUE, cancellation_reason = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`, reason, subscriptionID)
	return err
}

// ProcessRefund processes a refund for a subscription
func (db *DB) ProcessRefund(subscriptionID int, amount float64, reason string) error {
	_, err := db.Exec(`UPDATE subscriptions SET refund_amount = $1, refund_reason = $2, updated_at = NOW() WHERE id = $3 AND deleted_at IS NULL`, amount, reason, subscriptionID)
	return err
}

// SoftDeleteSubscription marks a subscription as deleted
func (db *DB) SoftDeleteSubscription(subscriptionID int) error {
	_, err := db.Exec(`UPDATE subscriptions SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1`, subscriptionID)
	return err
}

// GetUserSubscriptionStatus checks if a user has an active subscription
func (db *DB) GetUserSubscriptionStatus(userID int) (string, error) {
	var status sql.NullString
	err := db.QueryRow(
		`SELECT status FROM subscriptions WHERE user_id = $1 AND status IN ('active', 'trialing') AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 1`,
		userID,
	).Scan(&status)
	if err != nil {
		return "", err
	}
	if status.Valid {
		return status.String, nil
	}
	return "inactive", nil
}

// GetSubscriptionByStripeID retrieves a subscription by Stripe subscription ID
func (db *DB) GetSubscriptionByStripeID(stripeSubscriptionID string) (*Subscription, error) {
	subscription := &Subscription{}
	err := db.QueryRow(
		`SELECT id, user_id, plan_id, stripe_subscription_id, stripe_price_id, status, current_period_start, current_period_end, 
		        cancel_at_period_end, cancellation_reason, refund_amount, refund_reason, created_at, updated_at, deleted_at 
		 FROM subscriptions WHERE stripe_subscription_id = $1 AND deleted_at IS NULL`,
		stripeSubscriptionID,
	).Scan(&subscription.ID, &subscription.UserID, &subscription.PlanID, &subscription.StripeSubscriptionID,
		&subscription.StripePriceID, &subscription.Status, &subscription.CurrentPeriodStart, &subscription.CurrentPeriodEnd,
		&subscription.CancelAtPeriodEnd, &subscription.CancellationReason, &subscription.RefundAmount,
		&subscription.RefundReason, &subscription.CreatedAt, &subscription.UpdatedAt, &subscription.DeletedAt)
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

// GetUserSubscriptionHistory retrieves all subscriptions for a user
func (db *DB) GetUserSubscriptionHistory(userID int, limit, offset int) ([]*Subscription, error) {
	rows, err := db.Query(
		`SELECT id, user_id, plan_id, stripe_subscription_id, stripe_price_id, status, current_period_start, current_period_end, 
		        cancel_at_period_end, cancellation_reason, refund_amount, refund_reason, created_at, updated_at, deleted_at 
		 FROM subscriptions WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []*Subscription
	for rows.Next() {
		subscription := &Subscription{}
		err := rows.Scan(&subscription.ID, &subscription.UserID, &subscription.PlanID, &subscription.StripeSubscriptionID,
			&subscription.StripePriceID, &subscription.Status, &subscription.CurrentPeriodStart, &subscription.CurrentPeriodEnd,
			&subscription.CancelAtPeriodEnd, &subscription.CancellationReason, &subscription.RefundAmount,
			&subscription.RefundReason, &subscription.CreatedAt, &subscription.UpdatedAt, &subscription.DeletedAt)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, subscription)
	}
	return subscriptions, nil
}

// Add these methods at the end of the file

// GetActiveSubscriptionsCount returns the count of active subscriptions
func (db *DB) GetActiveSubscriptionsCount(planID *int) (int, error) {
	var count int
	var err error

	if planID != nil {
		err = db.QueryRow(
			`SELECT COUNT(*) FROM subscriptions 
			 WHERE status IN ('active', 'trialing') 
			 AND plan_id = $1 
			 AND deleted_at IS NULL`,
			*planID,
		).Scan(&count)
	} else {
		err = db.QueryRow(
			`SELECT COUNT(*) FROM subscriptions 
			 WHERE status IN ('active', 'trialing') 
			 AND deleted_at IS NULL`,
		).Scan(&count)
	}

	return count, err
}

// GetCancelledSubscriptionsCount returns the count of cancelled subscriptions in a date range
func (db *DB) GetCancelledSubscriptionsCount(startDate, endDate time.Time, planID *int) (int, error) {
	var count int
	var err error

	if planID != nil {
		err = db.QueryRow(
			`SELECT COUNT(*) FROM subscriptions 
			 WHERE updated_at BETWEEN $1 AND $2 
			 AND status = 'cancelled' 
			 AND plan_id = $3 
			 AND deleted_at IS NULL`,
			startDate, endDate, *planID,
		).Scan(&count)
	} else {
		err = db.QueryRow(
			`SELECT COUNT(*) FROM subscriptions 
			 WHERE updated_at BETWEEN $1 AND $2 
			 AND status = 'cancelled' 
			 AND deleted_at IS NULL`,
			startDate, endDate,
		).Scan(&count)
	}

	return count, err
}

// GetTotalRevenue calculates total revenue from active subscriptions
func (db *DB) GetTotalRevenue(planID *int) (float64, error) {
	var totalRevenue float64
	var err error

	if planID != nil {
		err = db.QueryRow(
			`SELECT COALESCE(SUM(sp.price), 0) 
			 FROM subscriptions s
			 JOIN subscription_plans sp ON s.plan_id = sp.id
			 WHERE s.status IN ('active', 'trialing') 
			 AND s.plan_id = $1 
			 AND s.deleted_at IS NULL 
			 AND sp.deleted_at IS NULL`,
			*planID,
		).Scan(&totalRevenue)
	} else {
		err = db.QueryRow(
			`SELECT COALESCE(SUM(sp.price), 0) 
			 FROM subscriptions s
			 JOIN subscription_plans sp ON s.plan_id = sp.id
			 WHERE s.status IN ('active', 'trialing') 
			 AND s.deleted_at IS NULL 
			 AND sp.deleted_at IS NULL`,
		).Scan(&totalRevenue)
	}

	return totalRevenue, err
}

// GetTotalCustomersCount returns the count of unique customers with subscriptions
func (db *DB) GetTotalCustomersCount() (int, error) {
	var count int
	err := db.QueryRow(
		`SELECT COUNT(DISTINCT user_id) 
		 FROM subscriptions 
		 WHERE deleted_at IS NULL`,
	).Scan(&count)
	return count, err
}
