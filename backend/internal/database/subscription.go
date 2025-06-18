package database

import (
	"database/sql"
	"time"
)

// Subscription represents a user subscription
type Subscription struct {
	ID                   int
	UserID               int
	StripeSubscriptionID string
	StripePriceID        string
	Status               string
	CurrentPeriodStart   *time.Time
	CurrentPeriodEnd     *time.Time
	CancelAtPeriodEnd    bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// CreateSubscription inserts a new subscription
func (db *DB) CreateSubscription(userID int, stripeSubscriptionID, stripePriceID, status string, currentPeriodStart, currentPeriodEnd *time.Time) (*Subscription, error) {
	var id int
	err := db.QueryRow(
		`INSERT INTO subscriptions (user_id, stripe_subscription_id, stripe_price_id, status, current_period_start, current_period_end, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) RETURNING id`,
		userID, stripeSubscriptionID, stripePriceID, status, currentPeriodStart, currentPeriodEnd,
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
		`SELECT id, user_id, stripe_subscription_id, stripe_price_id, status, current_period_start, current_period_end, cancel_at_period_end, created_at, updated_at FROM subscriptions WHERE id = $1`,
		id,
	).Scan(&subscription.ID, &subscription.UserID, &subscription.StripeSubscriptionID, &subscription.StripePriceID, &subscription.Status, &subscription.CurrentPeriodStart, &subscription.CurrentPeriodEnd, &subscription.CancelAtPeriodEnd, &subscription.CreatedAt, &subscription.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

// GetSubscriptionByUserID retrieves a user's active subscription
func (db *DB) GetSubscriptionByUserID(userID int) (*Subscription, error) {
	subscription := &Subscription{}
	err := db.QueryRow(
		`SELECT id, user_id, stripe_subscription_id, stripe_price_id, status, current_period_start, current_period_end, cancel_at_period_end, created_at, updated_at FROM subscriptions WHERE user_id = $1 AND status IN ('active', 'trialing') ORDER BY created_at DESC LIMIT 1`,
		userID,
	).Scan(&subscription.ID, &subscription.UserID, &subscription.StripeSubscriptionID, &subscription.StripePriceID, &subscription.Status, &subscription.CurrentPeriodStart, &subscription.CurrentPeriodEnd, &subscription.CancelAtPeriodEnd, &subscription.CreatedAt, &subscription.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

// UpdateSubscriptionStatus updates a subscription's status
func (db *DB) UpdateSubscriptionStatus(subscriptionID int, status string) error {
	_, err := db.Exec(`UPDATE subscriptions SET status = $1, updated_at = NOW() WHERE id = $2`, status, subscriptionID)
	return err
}

// CancelSubscription marks a subscription for cancellation
func (db *DB) CancelSubscription(subscriptionID int) error {
	_, err := db.Exec(`UPDATE subscriptions SET cancel_at_period_end = TRUE, updated_at = NOW() WHERE id = $1`, subscriptionID)
	return err
}

// GetUserSubscriptionStatus checks if a user has an active subscription
func (db *DB) GetUserSubscriptionStatus(userID int) (string, error) {
	var status sql.NullString
	err := db.QueryRow(
		`SELECT status FROM subscriptions WHERE user_id = $1 AND status IN ('active', 'trialing') ORDER BY created_at DESC LIMIT 1`,
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
		`SELECT id, user_id, stripe_subscription_id, stripe_price_id, status, current_period_start, current_period_end, cancel_at_period_end, created_at, updated_at FROM subscriptions WHERE stripe_subscription_id = $1`,
		stripeSubscriptionID,
	).Scan(&subscription.ID, &subscription.UserID, &subscription.StripeSubscriptionID, &subscription.StripePriceID, &subscription.Status, &subscription.CurrentPeriodStart, &subscription.CurrentPeriodEnd, &subscription.CancelAtPeriodEnd, &subscription.CreatedAt, &subscription.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return subscription, nil
}
