package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"bome-backend/internal/database"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/subscription"
)

// UserSubscriptionService handles user-facing subscription operations
type UserSubscriptionService struct {
	db             *database.DB
	linkingService *CustomerLinkingService
}

// NewUserSubscriptionService creates a new user subscription service
func NewUserSubscriptionService(db *database.DB, linkingService *CustomerLinkingService) *UserSubscriptionService {
	return &UserSubscriptionService{
		db:             db,
		linkingService: linkingService,
	}
}

// UserSubscription represents a user's subscription with all details
type UserSubscription struct {
	ID                 string    `json:"id"`
	StripeCustomerID   string    `json:"stripe_customer_id"`
	PlanName           string    `json:"plan_name"`
	Status             string    `json:"status"`
	Price              int64     `json:"price"` // in cents
	Currency           string    `json:"currency"`
	Interval           string    `json:"interval"` // month, year
	CurrentPeriodStart time.Time `json:"current_period_start"`
	CurrentPeriodEnd   time.Time `json:"current_period_end"`
	DaysUntilRenewal   int       `json:"days_until_renewal"`
	CancelAtPeriodEnd  bool      `json:"cancel_at_period_end"`
	CanceledAt         *string   `json:"canceled_at,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	IsLifetime         bool      `json:"is_lifetime"`
	IsPrimary          bool      `json:"is_primary"` // Is this the primary subscription?
}

// SubscriptionCounts holds counts of subscriptions by status
type SubscriptionCounts struct {
	Active   int `json:"active"`
	Trialing int `json:"trialing"`
	Canceled int `json:"canceled"`
	PastDue  int `json:"past_due"`
	Unpaid   int `json:"unpaid"`
	Total    int `json:"total"`
}

// UserSubscriptionsResponse is the response for getting all user subscriptions
type UserSubscriptionsResponse struct {
	ActiveSubscriptions   []UserSubscription `json:"active_subscriptions"`
	CanceledSubscriptions []UserSubscription `json:"canceled_subscriptions"`
	SubscriptionCounts    SubscriptionCounts `json:"subscription_count"`
	HasMultipleActive     bool               `json:"has_multiple_active"`
	VideoAccess           bool               `json:"video_access"`
}

// ================================================================
// GET USER SUBSCRIPTIONS
// ================================================================

// GetUserSubscriptions returns all subscriptions for a user
func (s *UserSubscriptionService) GetUserSubscriptions(userID int) (*UserSubscriptionsResponse, error) {
	log.Printf("📋 [User Subscriptions] Getting subscriptions for user %d", userID)

	// Get user's video access status
	var videoAccess bool
	err := s.db.QueryRow("SELECT manual_video_access FROM users WHERE id = $1", userID).Scan(&videoAccess)
	if err != nil {
		return nil, fmt.Errorf("failed to get user video access: %w", err)
	}

	// Get linked customers
	linkedCustomers, err := s.linkingService.GetUserLinkedCustomers(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get linked customers: %w", err)
	}

	if len(linkedCustomers) == 0 {
		log.Printf("ℹ️  [User Subscriptions] User %d has no linked customers", userID)
		return &UserSubscriptionsResponse{
			ActiveSubscriptions:   []UserSubscription{},
			CanceledSubscriptions: []UserSubscription{},
			SubscriptionCounts:    SubscriptionCounts{},
			HasMultipleActive:     false,
			VideoAccess:           videoAccess,
		}, nil
	}

	// Get primary customer ID
	primaryCustomerID, err := s.getPrimaryCustomerID(userID)
	if err != nil {
		log.Printf("⚠️  [User Subscriptions] Failed to get primary customer: %v", err)
		// Not fatal - continue without primary
		primaryCustomerID = ""
	}

	// Query all subscriptions for linked customers
	query := `
		SELECT 
			ss.stripe_id,
			sc.stripe_id as customer_id,
			COALESCE(sp.name, 'Unknown Plan') as plan_name,
			ss.status,
			COALESCE(spr.unit_amount, 0) as price,
			COALESCE(spr.currency, 'usd') as currency,
			COALESCE(spr.interval, 'month') as interval,
			ss.current_period_start,
			ss.current_period_end,
			ss.cancel_at_period_end,
			ss.canceled_at,
			ss.stripe_created_at
		FROM stripe_subscriptions_v2 ss
		JOIN stripe_customers_v2 sc ON ss.stripe_customer_id = sc.id
		LEFT JOIN stripe_prices_v2 spr ON ss.price_id = spr.id
		LEFT JOIN stripe_products_v2 sp ON spr.product_id = sp.id
		WHERE sc.stripe_id = ANY($1)
		ORDER BY ss.stripe_created_at DESC
	`

	rows, err := s.db.Query(query, linkedCustomers)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscriptions: %w", err)
	}
	defer rows.Close()

	var allSubscriptions []UserSubscription
	activeCount := 0
	trialingCount := 0
	canceledCount := 0
	pastDueCount := 0
	unpaidCount := 0

	for rows.Next() {
		var sub UserSubscription
		var canceledAt sql.NullString

		err := rows.Scan(
			&sub.ID,
			&sub.StripeCustomerID,
			&sub.PlanName,
			&sub.Status,
			&sub.Price,
			&sub.Currency,
			&sub.Interval,
			&sub.CurrentPeriodStart,
			&sub.CurrentPeriodEnd,
			&sub.CancelAtPeriodEnd,
			&canceledAt,
			&sub.CreatedAt,
		)
		if err != nil {
			log.Printf("⚠️  [User Subscriptions] Failed to scan subscription: %v", err)
			continue
		}

		// Handle nullable canceled_at
		if canceledAt.Valid {
			sub.CanceledAt = &canceledAt.String
		}

		// Calculate days until renewal
		daysUntilRenewal := int(time.Until(sub.CurrentPeriodEnd).Hours() / 24)
		sub.DaysUntilRenewal = daysUntilRenewal

		// Check if lifetime (no end date or very far in future)
		if sub.CurrentPeriodEnd.Year() > 2100 {
			sub.IsLifetime = true
		}

		// Check if this is the primary subscription
		if primaryCustomerID != "" && sub.StripeCustomerID == primaryCustomerID {
			sub.IsPrimary = true
		}

		allSubscriptions = append(allSubscriptions, sub)

		// Count by status
		switch sub.Status {
		case "active":
			if !sub.CancelAtPeriodEnd {
				activeCount++
			}
		case "trialing":
			trialingCount++
		case "canceled":
			canceledCount++
		case "past_due":
			pastDueCount++
		case "unpaid":
			unpaidCount++
		}
	}

	// Split into active and canceled
	var activeSubscriptions []UserSubscription
	var canceledSubscriptions []UserSubscription

	for _, sub := range allSubscriptions {
		if sub.Status == "active" || sub.Status == "trialing" {
			if !sub.CancelAtPeriodEnd {
				activeSubscriptions = append(activeSubscriptions, sub)
			} else {
				// Canceling at period end - show in canceled
				canceledSubscriptions = append(canceledSubscriptions, sub)
			}
		} else {
			canceledSubscriptions = append(canceledSubscriptions, sub)
		}
	}

	totalCount := len(allSubscriptions)
	hasMultipleActive := activeCount > 1

	log.Printf("✅ [User Subscriptions] User %d has %d active, %d canceled subscriptions", userID, activeCount, len(canceledSubscriptions))

	return &UserSubscriptionsResponse{
		ActiveSubscriptions:   activeSubscriptions,
		CanceledSubscriptions: canceledSubscriptions,
		SubscriptionCounts: SubscriptionCounts{
			Active:   activeCount,
			Trialing: trialingCount,
			Canceled: canceledCount,
			PastDue:  pastDueCount,
			Unpaid:   unpaidCount,
			Total:    totalCount,
		},
		HasMultipleActive: hasMultipleActive,
		VideoAccess:       videoAccess,
	}, nil
}

// getPrimaryCustomerID gets the primary Stripe customer ID for a user
func (s *UserSubscriptionService) getPrimaryCustomerID(userID int) (string, error) {
	query := `
		SELECT stripe_customer_id
		FROM user_stripe_customers_v2
		WHERE user_id = $1 AND is_primary = true
		LIMIT 1
	`

	var customerID string
	err := s.db.QueryRow(query, userID).Scan(&customerID)
	if err == sql.ErrNoRows {
		return "", nil // No primary customer yet
	}
	if err != nil {
		return "", err
	}

	return customerID, nil
}

// ================================================================
// CANCEL MULTIPLE SUBSCRIPTIONS
// ================================================================

// CancelMultipleRequest is the request body for canceling multiple subscriptions
type CancelMultipleRequest struct {
	SubscriptionIDs    []string `json:"subscription_ids"`
	KeepSubscriptionID string   `json:"keep_subscription_id,omitempty"`
	Reason             string   `json:"reason,omitempty"`
}

// CanceledSubscriptionInfo holds info about a canceled subscription
type CanceledSubscriptionInfo struct {
	ID                string `json:"id"`
	Status            string `json:"status"`
	CancelAtPeriodEnd bool   `json:"cancel_at_period_end"`
	CanceledAt        string `json:"canceled_at"`
	EndsOn            string `json:"ends_on"`
}

// CancelMultipleResponse is the response for canceling multiple subscriptions
type CancelMultipleResponse struct {
	Success               bool                       `json:"success"`
	Message               string                     `json:"message"`
	CanceledSubscriptions []CanceledSubscriptionInfo `json:"canceled_subscriptions"`
	KeptSubscription      *UserSubscription          `json:"kept_subscription,omitempty"`
}

// CancelMultipleSubscriptions cancels multiple subscriptions for a user
func (s *UserSubscriptionService) CancelMultipleSubscriptions(ctx context.Context, userID int, req CancelMultipleRequest) (*CancelMultipleResponse, error) {
	log.Printf("❌ [User Subscriptions] User %d requesting to cancel %d subscriptions", userID, len(req.SubscriptionIDs))

	if len(req.SubscriptionIDs) == 0 {
		return nil, fmt.Errorf("no subscription IDs provided")
	}

	// Step 1: Verify user owns ALL subscriptions
	if err := s.verifyUserOwnsSubscriptions(userID, req.SubscriptionIDs); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Step 2: Verify keep subscription is not in cancel list
	if req.KeepSubscriptionID != "" {
		for _, id := range req.SubscriptionIDs {
			if id == req.KeepSubscriptionID {
				return nil, fmt.Errorf("cannot cancel the subscription you want to keep")
			}
		}
	}

	// Step 3: Cancel each subscription in Stripe
	var canceledSubscriptions []CanceledSubscriptionInfo
	reason := req.Reason
	if reason == "" {
		reason = "User chose to consolidate subscriptions"
	}

	for _, subID := range req.SubscriptionIDs {
		log.Printf("❌ [User Subscriptions] Canceling subscription: %s", subID)

		// Cancel in Stripe (at period end)
		params := &stripe.SubscriptionParams{
			CancelAtPeriodEnd: stripe.Bool(true),
		}
		params.AddMetadata("cancellation_reason", reason)
		params.AddMetadata("canceled_by_user_id", fmt.Sprintf("%d", userID))

		stripeSub, err := subscription.Update(subID, params)
		if err != nil {
			log.Printf("⚠️  [User Subscriptions] Failed to cancel subscription %s in Stripe: %v", subID, err)
			// Continue with others
			continue
		}

		// Update in our database
		updateQuery := `
			UPDATE stripe_subscriptions_v2
			SET cancel_at_period_end = true,
			    canceled_at = NOW(),
			    last_synced_at = NOW()
			WHERE stripe_id = $1
		`
		_, err = s.db.Exec(updateQuery, subID)
		if err != nil {
			log.Printf("⚠️  [User Subscriptions] Failed to update subscription %s in database: %v", subID, err)
		}

		canceledSubscriptions = append(canceledSubscriptions, CanceledSubscriptionInfo{
			ID:                subID,
			Status:            string(stripeSub.Status),
			CancelAtPeriodEnd: stripeSub.CancelAtPeriodEnd,
			CanceledAt:        time.Now().Format(time.RFC3339),
			EndsOn:            time.Unix(stripeSub.CurrentPeriodEnd, 0).Format(time.RFC3339),
		})

		log.Printf("✅ [User Subscriptions] Subscription %s will cancel at period end (%s)", subID, time.Unix(stripeSub.CurrentPeriodEnd, 0).Format("2006-01-02"))
	}

	// Step 4: If keep subscription is provided, set as primary
	var keptSubscription *UserSubscription
	if req.KeepSubscriptionID != "" {
		// Get subscription details
		allSubs, err := s.GetUserSubscriptions(userID)
		if err == nil {
			for _, sub := range allSubs.ActiveSubscriptions {
				if sub.ID == req.KeepSubscriptionID {
					keptSubscription = &sub
					break
				}
			}
		}

		// Set as primary customer
		if keptSubscription != nil {
			setPrimaryQuery := `
				UPDATE user_stripe_customers_v2
				SET is_primary = (stripe_customer_id = (
					SELECT stripe_customer_id FROM stripe_subscriptions_v2 
					WHERE stripe_id = $1
				))
				WHERE user_id = $2
			`
			_, err := s.db.Exec(setPrimaryQuery, req.KeepSubscriptionID, userID)
			if err != nil {
				log.Printf("⚠️  [User Subscriptions] Failed to set primary subscription: %v", err)
			}
		}
	}

	message := fmt.Sprintf("%d subscriptions will be canceled at the end of their billing periods", len(canceledSubscriptions))
	log.Printf("✅ [User Subscriptions] User %d successfully canceled %d subscriptions", userID, len(canceledSubscriptions))

	return &CancelMultipleResponse{
		Success:               true,
		Message:               message,
		CanceledSubscriptions: canceledSubscriptions,
		KeptSubscription:      keptSubscription,
	}, nil
}

// verifyUserOwnsSubscriptions checks that user owns all provided subscription IDs
func (s *UserSubscriptionService) verifyUserOwnsSubscriptions(userID int, subscriptionIDs []string) error {
	// Get user's linked customers
	linkedCustomers, err := s.linkingService.GetUserLinkedCustomers(userID)
	if err != nil {
		return fmt.Errorf("failed to get linked customers: %w", err)
	}

	if len(linkedCustomers) == 0 {
		return fmt.Errorf("user has no linked customers")
	}

	// Verify each subscription belongs to one of the user's customers
	query := `
		SELECT ss.stripe_id
		FROM stripe_subscriptions_v2 ss
		JOIN stripe_customers_v2 sc ON ss.stripe_customer_id = sc.id
		WHERE sc.stripe_id = ANY($1)
		AND ss.stripe_id = ANY($2)
	`

	rows, err := s.db.Query(query, linkedCustomers, subscriptionIDs)
	if err != nil {
		return fmt.Errorf("failed to verify subscription ownership: %w", err)
	}
	defer rows.Close()

	ownedSubscriptions := make(map[string]bool)
	for rows.Next() {
		var subID string
		if err := rows.Scan(&subID); err != nil {
			continue
		}
		ownedSubscriptions[subID] = true
	}

	// Check all requested subscriptions are owned
	for _, subID := range subscriptionIDs {
		if !ownedSubscriptions[subID] {
			return fmt.Errorf("user does not own subscription %s", subID)
		}
	}

	return nil
}

// ================================================================
// CANCEL SINGLE SUBSCRIPTION
// ================================================================

// CancelSingleRequest is the request body for canceling a single subscription
type CancelSingleRequest struct {
	Reason string `json:"reason,omitempty"`
}

// CancelSingleResponse is the response for canceling a single subscription
type CancelSingleResponse struct {
	Success      bool                     `json:"success"`
	Message      string                   `json:"message"`
	Subscription CanceledSubscriptionInfo `json:"subscription"`
}

// CancelSingleSubscription cancels a single subscription
func (s *UserSubscriptionService) CancelSingleSubscription(ctx context.Context, userID int, subscriptionID string, req CancelSingleRequest) (*CancelSingleResponse, error) {
	log.Printf("❌ [User Subscriptions] User %d requesting to cancel subscription: %s", userID, subscriptionID)

	// Verify ownership
	if err := s.verifyUserOwnsSubscriptions(userID, []string{subscriptionID}); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Cancel in Stripe (at period end)
	reason := req.Reason
	if reason == "" {
		reason = "Canceled by user"
	}

	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}
	params.AddMetadata("cancellation_reason", reason)
	params.AddMetadata("canceled_by_user_id", fmt.Sprintf("%d", userID))

	stripeSub, err := subscription.Update(subscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel subscription in Stripe: %w", err)
	}

	// Update in database
	updateQuery := `
		UPDATE stripe_subscriptions_v2
		SET cancel_at_period_end = true,
		    canceled_at = NOW(),
		    last_synced_at = NOW()
		WHERE stripe_id = $1
	`
	_, err = s.db.Exec(updateQuery, subscriptionID)
	if err != nil {
		log.Printf("⚠️  [User Subscriptions] Failed to update subscription in database: %v", err)
	}

	log.Printf("✅ [User Subscriptions] Subscription %s will cancel at period end (%s)", subscriptionID, time.Unix(stripeSub.CurrentPeriodEnd, 0).Format("2006-01-02"))

	return &CancelSingleResponse{
		Success: true,
		Message: "Subscription will be canceled at the end of the billing period",
		Subscription: CanceledSubscriptionInfo{
			ID:                subscriptionID,
			Status:            string(stripeSub.Status),
			CancelAtPeriodEnd: stripeSub.CancelAtPeriodEnd,
			CanceledAt:        time.Now().Format(time.RFC3339),
			EndsOn:            time.Unix(stripeSub.CurrentPeriodEnd, 0).Format(time.RFC3339),
		},
	}, nil
}

// ================================================================
// SUBSCRIPTION HISTORY
// ================================================================

// GetSubscriptionHistory returns all subscriptions (active + historical) for a user
func (s *UserSubscriptionService) GetSubscriptionHistory(userID int) ([]UserSubscription, error) {
	log.Printf("📜 [User Subscriptions] Getting subscription history for user %d", userID)

	allSubs, err := s.GetUserSubscriptions(userID)
	if err != nil {
		return nil, err
	}

	// Combine active and canceled for history view
	history := append(allSubs.ActiveSubscriptions, allSubs.CanceledSubscriptions...)

	log.Printf("✅ [User Subscriptions] User %d has %d total subscriptions in history", userID, len(history))

	return history, nil
}

// ================================================================
// CHANGE SUBSCRIPTION PLAN (Phase 7.3)
// ================================================================

// ChangeSubscriptionPlanRequest is the request for changing a subscription plan
type ChangeSubscriptionPlanRequest struct {
	NewPriceID string `json:"new_price_id"` // Stripe price ID (e.g., price_1234567890)
}

// ChangeSubscriptionPlanResponse is the response for changing a subscription plan
type ChangeSubscriptionPlanResponse struct {
	Success             bool             `json:"success"`
	Message             string           `json:"message"`
	UpdatedSubscription UserSubscription `json:"updated_subscription"`
	ProrationAmount     int64            `json:"proration_amount"` // in cents (can be negative for credits)
}

// ChangeSubscriptionPlan updates an existing subscription to a new plan/price
// This prevents multiple active subscriptions by updating the existing one
func (s *UserSubscriptionService) ChangeSubscriptionPlan(ctx context.Context, userID int, req ChangeSubscriptionPlanRequest) (*ChangeSubscriptionPlanResponse, error) {
	log.Printf("🔄 [Change Plan] User %d requesting plan change to price: %s", userID, req.NewPriceID)

	// Step 1: Get user's current active subscription
	allSubs, err := s.GetUserSubscriptions(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user subscriptions: %w", err)
	}

	activeSubscriptions := allSubs.ActiveSubscriptions
	if len(activeSubscriptions) == 0 {
		return nil, fmt.Errorf("no active subscription found - user should create a new subscription instead")
	}

	if len(activeSubscriptions) > 1 {
		return nil, fmt.Errorf("user has multiple active subscriptions - please consolidate first using the admin tool")
	}

	currentSub := activeSubscriptions[0]
	log.Printf("📋 [Change Plan] Found current subscription: %s (plan: %s)", currentSub.ID, currentSub.PlanName)

	// Step 2: Get the subscription from Stripe
	stripeSub, err := subscription.Get(currentSub.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription from Stripe: %w", err)
	}

	if len(stripeSub.Items.Data) == 0 {
		return nil, fmt.Errorf("subscription has no items")
	}

	// Step 3: Update subscription in Stripe
	log.Printf("🔄 [Change Plan] Updating Stripe subscription %s to new price %s", currentSub.ID, req.NewPriceID)

	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(stripeSub.Items.Data[0].ID), // Keep same subscription item
				Price: stripe.String(req.NewPriceID),             // Change to new price
			},
		},
		ProrationBehavior: stripe.String("create_prorations"), // Auto pro-rate!
	}
	params.AddMetadata("changed_by_user_id", fmt.Sprintf("%d", userID))
	params.AddMetadata("old_price_id", stripeSub.Items.Data[0].Price.ID)
	params.AddMetadata("new_price_id", req.NewPriceID)

	updatedStripeSub, err := subscription.Update(currentSub.ID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription in Stripe: %w", err)
	}

	log.Printf("✅ [Change Plan] Subscription %s updated successfully in Stripe", updatedStripeSub.ID)

	// Step 4: Calculate proration amount (from latest invoice)
	prorationAmount := int64(0)
	if updatedStripeSub.LatestInvoice != nil {
		// The proration amount is in the invoice
		prorationAmount = updatedStripeSub.LatestInvoice.AmountDue
	}

	// Step 5: Sync to v2 tables
	syncService := NewStripeSyncV2Service(s.db)
	if err := syncService.SyncSingleSubscription(ctx, updatedStripeSub.ID); err != nil {
		log.Printf("⚠️  [Change Plan] Failed to sync updated subscription to v2: %v", err)
		// Don't fail - subscription was updated in Stripe successfully
	}

	// Step 6: Refresh subscription data from our database
	refreshedSubs, err := s.GetUserSubscriptions(userID)
	if err != nil {
		log.Printf("⚠️  [Change Plan] Failed to refresh subscription data: %v", err)
		// Return what we have from Stripe
		return &ChangeSubscriptionPlanResponse{
			Success: true,
			Message: fmt.Sprintf("Successfully changed to new plan! (Proration: %s)", s.formatProration(prorationAmount)),
			UpdatedSubscription: UserSubscription{
				ID:                 updatedStripeSub.ID,
				StripeCustomerID:   updatedStripeSub.Customer.ID,
				Status:             string(updatedStripeSub.Status),
				CurrentPeriodStart: time.Unix(updatedStripeSub.CurrentPeriodStart, 0),
				CurrentPeriodEnd:   time.Unix(updatedStripeSub.CurrentPeriodEnd, 0),
			},
			ProrationAmount: prorationAmount,
		}, nil
	}

	if len(refreshedSubs.ActiveSubscriptions) == 0 {
		return nil, fmt.Errorf("subscription updated but not found in database")
	}

	updatedSub := refreshedSubs.ActiveSubscriptions[0]

	log.Printf("✅ [Change Plan] User %d successfully changed plan (sub: %s, proration: %d cents)", userID, updatedSub.ID, prorationAmount)

	return &ChangeSubscriptionPlanResponse{
		Success:             true,
		Message:             fmt.Sprintf("Successfully changed to %s!", updatedSub.PlanName),
		UpdatedSubscription: updatedSub,
		ProrationAmount:     prorationAmount,
	}, nil
}

// formatProration formats the proration amount for display
func (s *UserSubscriptionService) formatProration(cents int64) string {
	if cents == 0 {
		return "No charge"
	} else if cents > 0 {
		dollars := float64(cents) / 100.0
		return fmt.Sprintf("Charged $%.2f", dollars)
	} else {
		dollars := float64(-cents) / 100.0
		return fmt.Sprintf("Credited $%.2f", dollars)
	}
}

// ================================================================
// CHECK IF USER CAN SUBSCRIBE (Phase 7.3)
// ================================================================

// CanUserSubscribe checks if a user can create a new subscription
// Returns false if they already have an active subscription (they should update instead)
func (s *UserSubscriptionService) CanUserSubscribe(userID int) (bool, string, error) {
	allSubs, err := s.GetUserSubscriptions(userID)
	if err != nil {
		return false, "", fmt.Errorf("failed to check user subscriptions: %w", err)
	}

	activeCount := len(allSubs.ActiveSubscriptions)

	if activeCount == 0 {
		return true, "User can subscribe", nil
	}

	if activeCount == 1 {
		activeSub := allSubs.ActiveSubscriptions[0]
		message := fmt.Sprintf("User already has an active subscription (%s). Please use the 'Change Plan' feature instead.", activeSub.PlanName)
		return false, message, nil
	}

	// Multiple active subscriptions - shouldn't happen but handle it
	message := fmt.Sprintf("User has %d active subscriptions. Please consolidate using admin tools before subscribing.", activeCount)
	return false, message, nil
}
