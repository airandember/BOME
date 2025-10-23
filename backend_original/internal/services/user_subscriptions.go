package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"bome-backend/internal/database"
)

// UserSubscriptionService handles user subscription operations
type UserSubscriptionService struct {
	db           *database.DB
	emailService *EmailService
}

// UserSubscription represents a user's subscription
type UserSubscription struct {
	ID                   int                    `json:"id"`
	UserID               int                    `json:"user_id"`
	SubscriptionPlanID   int                    `json:"subscription_plan_id"`
	StripeSubscriptionID *string                `json:"stripe_subscription_id"`
	StripeCustomerID     *string                `json:"stripe_customer_id"`
	StripeSessionID      *string                `json:"stripe_session_id"`
	Status               string                 `json:"status"`
	CurrentPeriodStart   *time.Time             `json:"current_period_start"`
	CurrentPeriodEnd     *time.Time             `json:"current_period_end"`
	CancelAtPeriodEnd    bool                   `json:"cancel_at_period_end"`
	CancelledAt          *time.Time             `json:"cancelled_at"`
	TrialStart           *time.Time             `json:"trial_start"`
	TrialEnd             *time.Time             `json:"trial_end"`
	AmountPaid           *float64               `json:"amount_paid"`
	Currency             string                 `json:"currency"`
	PaymentMethod        *string                `json:"payment_method"`
	Metadata             map[string]interface{} `json:"metadata"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

// CreateSubscriptionRequest represents a request to create a subscription
type CreateSubscriptionRequest struct {
	UserID               int                    `json:"user_id"`
	SubscriptionPlanID   int                    `json:"subscription_plan_id"`
	StripeSubscriptionID *string                `json:"stripe_subscription_id"`
	StripeCustomerID     *string                `json:"stripe_customer_id"`
	StripeSessionID      *string                `json:"stripe_session_id"`
	Status               string                 `json:"status"`
	CurrentPeriodStart   *time.Time             `json:"current_period_start"`
	CurrentPeriodEnd     *time.Time             `json:"current_period_end"`
	AmountPaid           *float64               `json:"amount_paid"`
	Currency             string                 `json:"currency"`
	PaymentMethod        *string                `json:"payment_method"`
	Metadata             map[string]interface{} `json:"metadata"`
}

// NewUserSubscriptionService creates a new user subscription service
func NewUserSubscriptionService(db *database.DB, emailService *EmailService) *UserSubscriptionService {
	return &UserSubscriptionService{
		db:           db,
		emailService: emailService,
	}
}

// CreateSubscriptionFromStripeSession creates a subscription from a successful Stripe session
func (s *UserSubscriptionService) CreateSubscriptionFromStripeSession(sessionData map[string]interface{}, userID int) (*UserSubscription, error) {
	log.Printf("🔍 [SUBSCRIPTION] Creating subscription from Stripe session for user %d", userID)

	// Extract session data
	sessionID := sessionData["session_id"].(string)
	amountTotal := sessionData["amount_total"]
	currency := sessionData["currency"].(string)

	// Convert amount from cents to dollars
	var amountPaid *float64
	if amountTotal != nil {
		if amt, ok := amountTotal.(float64); ok {
			converted := amt / 100.0
			amountPaid = &converted
		}
	}

	// Get subscription ID if available
	var stripeSubscriptionID *string
	if subID, exists := sessionData["subscription_id"]; exists && subID != nil {
		if subIDStr, ok := subID.(string); ok {
			stripeSubscriptionID = &subIDStr
		}
	}

	// Get customer ID if available
	var stripeCustomerID *string
	if custID, exists := sessionData["customer_id"]; exists && custID != nil {
		if custIDStr, ok := custID.(string); ok {
			stripeCustomerID = &custIDStr
		}
	}

	// Get plan ID from metadata
	var planID int
	if metadata, exists := sessionData["metadata"]; exists {
		if metaMap, ok := metadata.(map[string]interface{}); ok {
			if planIDStr, exists := metaMap["plan_id"]; exists {
				if planIDInt, ok := planIDStr.(string); ok {
					// Convert string to int
					fmt.Sscanf(planIDInt, "%d", &planID)
				}
			}
		}
	}

	if planID == 0 {
		return nil, fmt.Errorf("plan ID not found in session metadata")
	}

	// Create subscription request
	request := CreateSubscriptionRequest{
		UserID:               userID,
		SubscriptionPlanID:   planID,
		StripeSubscriptionID: stripeSubscriptionID,
		StripeCustomerID:     stripeCustomerID,
		StripeSessionID:      &sessionID,
		Status:               "active",
		AmountPaid:           amountPaid,
		Currency:             currency,
		PaymentMethod:        stringPtr("stripe"),
	}

	// Create subscription
	subscription, err := s.CreateSubscription(request)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// Update user's subscription fields in the users table
	err = s.updateUserSubscriptionFields(userID, planID, true)
	if err != nil {
		log.Printf("⚠️ [SUBSCRIPTION] Failed to update user subscription fields: %v", err)
		// Don't fail the whole operation - subscription was created successfully
	}

	// Send confirmation email
	go s.sendSubscriptionConfirmationEmail(subscription)

	log.Printf("✅ [SUBSCRIPTION] Created subscription %d for user %d", subscription.ID, userID)
	return subscription, nil
}

// CreateSubscription creates a new user subscription
func (s *UserSubscriptionService) CreateSubscription(request CreateSubscriptionRequest) (*UserSubscription, error) {
	query := `
		INSERT INTO user_subscriptions (
			user_id, subscription_plan_id, stripe_subscription_id, stripe_customer_id, 
			stripe_session_id, status, current_period_start, current_period_end,
			amount_paid, currency, payment_method, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at`

	var subscription UserSubscription
	err := s.db.DB.QueryRow(
		query,
		request.UserID,
		request.SubscriptionPlanID,
		request.StripeSubscriptionID,
		request.StripeCustomerID,
		request.StripeSessionID,
		request.Status,
		request.CurrentPeriodStart,
		request.CurrentPeriodEnd,
		request.AmountPaid,
		request.Currency,
		request.PaymentMethod,
		request.Metadata,
	).Scan(&subscription.ID, &subscription.CreatedAt, &subscription.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// Fill in the rest of the fields
	subscription.UserID = request.UserID
	subscription.SubscriptionPlanID = request.SubscriptionPlanID
	subscription.StripeSubscriptionID = request.StripeSubscriptionID
	subscription.StripeCustomerID = request.StripeCustomerID
	subscription.StripeSessionID = request.StripeSessionID
	subscription.Status = request.Status
	subscription.CurrentPeriodStart = request.CurrentPeriodStart
	subscription.CurrentPeriodEnd = request.CurrentPeriodEnd
	subscription.AmountPaid = request.AmountPaid
	subscription.Currency = request.Currency
	subscription.PaymentMethod = request.PaymentMethod
	subscription.Metadata = request.Metadata

	return &subscription, nil
}

// GetUserActiveSubscription gets a user's active subscription
func (s *UserSubscriptionService) GetUserActiveSubscription(userID int) (*UserSubscription, error) {
	query := `
		SELECT id, user_id, subscription_plan_id, stripe_subscription_id, stripe_customer_id,
			   stripe_session_id, status, current_period_start, current_period_end,
			   cancel_at_period_end, cancelled_at, trial_start, trial_end,
			   amount_paid, currency, payment_method, metadata, created_at, updated_at
		FROM user_subscriptions 
		WHERE user_id = $1 AND status = 'active'
		ORDER BY created_at DESC
		LIMIT 1`

	var subscription UserSubscription
	var stripeSubscriptionID, stripeCustomerID, stripeSessionID, paymentMethod sql.NullString
	var currentPeriodStart, currentPeriodEnd, cancelledAt, trialStart, trialEnd sql.NullTime
	var amountPaid sql.NullFloat64

	err := s.db.DB.QueryRow(query, userID).Scan(
		&subscription.ID,
		&subscription.UserID,
		&subscription.SubscriptionPlanID,
		&stripeSubscriptionID,
		&stripeCustomerID,
		&stripeSessionID,
		&subscription.Status,
		&currentPeriodStart,
		&currentPeriodEnd,
		&subscription.CancelAtPeriodEnd,
		&cancelledAt,
		&trialStart,
		&trialEnd,
		&amountPaid,
		&subscription.Currency,
		&paymentMethod,
		&subscription.Metadata,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active subscription
		}
		return nil, fmt.Errorf("failed to get user subscription: %w", err)
	}

	// Handle nullable fields
	if stripeSubscriptionID.Valid {
		subscription.StripeSubscriptionID = &stripeSubscriptionID.String
	}
	if stripeCustomerID.Valid {
		subscription.StripeCustomerID = &stripeCustomerID.String
	}
	if stripeSessionID.Valid {
		subscription.StripeSessionID = &stripeSessionID.String
	}
	if paymentMethod.Valid {
		subscription.PaymentMethod = &paymentMethod.String
	}
	if currentPeriodStart.Valid {
		subscription.CurrentPeriodStart = &currentPeriodStart.Time
	}
	if currentPeriodEnd.Valid {
		subscription.CurrentPeriodEnd = &currentPeriodEnd.Time
	}
	if cancelledAt.Valid {
		subscription.CancelledAt = &cancelledAt.Time
	}
	if trialStart.Valid {
		subscription.TrialStart = &trialStart.Time
	}
	if trialEnd.Valid {
		subscription.TrialEnd = &trialEnd.Time
	}
	if amountPaid.Valid {
		subscription.AmountPaid = &amountPaid.Float64
	}

	return &subscription, nil
}

// sendSubscriptionConfirmationEmail sends a confirmation email for a new subscription
func (s *UserSubscriptionService) sendSubscriptionConfirmationEmail(subscription *UserSubscription) {
	if s.emailService == nil {
		log.Printf("⚠️ [SUBSCRIPTION] Email service not available, skipping confirmation email")
		return
	}

	// Get user details
	userQuery := `SELECT name, email FROM users WHERE id = $1`
	var userName, userEmail string
	err := s.db.DB.QueryRow(userQuery, subscription.UserID).Scan(&userName, &userEmail)
	if err != nil {
		log.Printf("❌ [SUBSCRIPTION] Failed to get user details: %v", err)
		return
	}

	// Get plan details
	planQuery := `SELECT name FROM subscription_plans WHERE id = $1`
	var planName string
	err = s.db.DB.QueryRow(planQuery, subscription.SubscriptionPlanID).Scan(&planName)
	if err != nil {
		log.Printf("❌ [SUBSCRIPTION] Failed to get plan details: %v", err)
		return
	}

	// Format amount and period end
	amount := 0.0
	if subscription.AmountPaid != nil {
		amount = *subscription.AmountPaid
	}

	periodEnd := "N/A"
	if subscription.CurrentPeriodEnd != nil {
		periodEnd = subscription.CurrentPeriodEnd.Format("January 2, 2006")
	}

	// Send confirmation email
	err = s.emailService.SendSubscriptionConfirmation(
		subscription.UserID,
		userEmail,
		userName,
		planName,
		amount,
		subscription.Currency,
		periodEnd,
	)

	if err != nil {
		log.Printf("❌ [SUBSCRIPTION] Failed to send confirmation email: %v", err)
	} else {
		log.Printf("✅ [SUBSCRIPTION] Confirmation email sent to %s", userEmail)
	}
}

// updateUserSubscriptionFields updates the legacy subscription fields in the users table
func (s *UserSubscriptionService) updateUserSubscriptionFields(userID, planID int, hasSubbed bool) error {
	query := `
		UPDATE users 
		SET sub_id = $1, has_subbed = $2, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $3`

	_, err := s.db.DB.Exec(query, planID, hasSubbed, userID)
	if err != nil {
		return fmt.Errorf("failed to update user subscription fields: %w", err)
	}

	log.Printf("✅ [SUBSCRIPTION] Updated user %d: sub_id=%d, has_subbed=%t", userID, planID, hasSubbed)
	return nil
}

// CancelUserSubscription cancels a user's subscription and updates legacy fields
func (s *UserSubscriptionService) CancelUserSubscription(userID int) error {
	// Update subscription status in user_subscriptions table
	updateSubQuery := `
		UPDATE user_subscriptions 
		SET status = 'cancelled', cancelled_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 AND status = 'active'`

	_, err := s.db.DB.Exec(updateSubQuery, userID)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	// Update legacy fields in users table (keep sub_id but set has_subbed to false)
	updateUserQuery := `
		UPDATE users 
		SET has_subbed = false, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1`

	_, err = s.db.DB.Exec(updateUserQuery, userID)
	if err != nil {
		return fmt.Errorf("failed to update user subscription status: %w", err)
	}

	log.Printf("✅ [SUBSCRIPTION] Cancelled subscription for user %d", userID)
	return nil
}

// GetUserSubscriptionStatus gets a user's subscription status including legacy fields
func (s *UserSubscriptionService) GetUserSubscriptionStatus(userID int) (map[string]interface{}, error) {
	// Get current subscription from new table
	subscription, err := s.GetUserActiveSubscription(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active subscription: %w", err)
	}

	// Get legacy fields from users table
	userQuery := `SELECT sub_id, has_subbed FROM users WHERE id = $1`
	var subID sql.NullInt64
	var hasSubbed sql.NullBool

	err = s.db.DB.QueryRow(userQuery, userID).Scan(&subID, &hasSubbed)
	if err != nil {
		return nil, fmt.Errorf("failed to get user subscription info: %w", err)
	}

	status := map[string]interface{}{
		"has_active_subscription": subscription != nil,
		"legacy_sub_id":           nil,
		"legacy_has_subbed":       false,
	}

	if subID.Valid {
		status["legacy_sub_id"] = subID.Int64
	}
	if hasSubbed.Valid {
		status["legacy_has_subbed"] = hasSubbed.Bool
	}

	if subscription != nil {
		status["subscription"] = subscription
	}

	return status, nil
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
