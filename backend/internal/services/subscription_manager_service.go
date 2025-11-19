package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"bome-backend/internal/database"

	"github.com/lib/pq"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/subscription"
)

// SubscriptionManagerService handles subscription business logic
type SubscriptionManagerService struct {
	db             *database.DB
	linkingService *CustomerLinkingService
}

// NewSubscriptionManagerService creates a new subscription manager
func NewSubscriptionManagerService(db *database.DB, linkingService *CustomerLinkingService) *SubscriptionManagerService {
	return &SubscriptionManagerService{
		db:             db,
		linkingService: linkingService,
	}
}

// SubscriptionResult contains the result of subscription operations
type SubscriptionResult struct {
	UserID                  int      `json:"user_id"`
	NewSubscriptionID       string   `json:"new_subscription_id"`
	CanceledSubscriptionIDs []string `json:"canceled_subscription_ids"`
	VideoAccessGranted      bool     `json:"video_access_granted"`
	Error                   string   `json:"error,omitempty"`
}

// ================================================================
// SINGLE SUBSCRIPTION ENFORCEMENT
// ================================================================

// EnforceSingleSubscription ensures a user has only one active subscription
// When a new subscription is created, it cancels all other active subscriptions for that user
func (s *SubscriptionManagerService) EnforceSingleSubscription(userID int, newSubscriptionID string) (*SubscriptionResult, error) {
	result := &SubscriptionResult{
		UserID:                  userID,
		NewSubscriptionID:       newSubscriptionID,
		CanceledSubscriptionIDs: []string{},
		VideoAccessGranted:      false,
	}

	log.Printf("🔒 [Subscription Manager] Enforcing single subscription for user %d (new sub: %s)", userID, newSubscriptionID)

	// Step 1: Get all linked Stripe customers for this user
	linkedCustomers, err := s.linkingService.GetUserLinkedCustomers(userID)
	if err != nil {
		return result, fmt.Errorf("failed to get linked customers: %w", err)
	}

	if len(linkedCustomers) == 0 {
		log.Printf("ℹ️  [Subscription Manager] User %d has no linked customers", userID)
		return result, nil
	}

	log.Printf("📋 [Subscription Manager] User %d has %d linked customers", userID, len(linkedCustomers))

	// Step 2: Find all active subscriptions across all linked customers (excluding the new one)
	activeSubscriptions, err := s.findActiveSubscriptionsForUser(linkedCustomers, newSubscriptionID)
	if err != nil {
		return result, fmt.Errorf("failed to find active subscriptions: %w", err)
	}

	if len(activeSubscriptions) == 0 {
		log.Printf("✅ [Subscription Manager] No other active subscriptions found - user %d is good to go", userID)
		return result, nil
	}

	log.Printf("⚠️  [Subscription Manager] Found %d other active subscriptions for user %d", len(activeSubscriptions), userID)

	// Step 3: Cancel all other active subscriptions in Stripe
	for _, subID := range activeSubscriptions {
		log.Printf("❌ [Subscription Manager] Canceling old subscription: %s", subID)

		// Cancel in Stripe
		params := &stripe.SubscriptionParams{
			CancelAtPeriodEnd: stripe.Bool(true), // Don't cancel immediately - let them finish their period
		}

		_, err := subscription.Update(subID, params)
		if err != nil {
			log.Printf("⚠️  [Subscription Manager] Failed to cancel subscription %s in Stripe: %v", subID, err)
			// Continue canceling others even if one fails
			result.Error = fmt.Sprintf("Failed to cancel some subscriptions: %v", err)
		} else {
			result.CanceledSubscriptionIDs = append(result.CanceledSubscriptionIDs, subID)
			log.Printf("✅ [Subscription Manager] Subscription %s will cancel at period end", subID)
		}
	}

	log.Printf("🎯 [Subscription Manager] Single subscription enforcement complete for user %d", userID)
	log.Printf("   New subscription: %s", newSubscriptionID)
	log.Printf("   Canceled subscriptions: %d", len(result.CanceledSubscriptionIDs))

	return result, nil
}

// findActiveSubscriptionsForUser finds all active subscriptions for a user's linked customers
func (s *SubscriptionManagerService) findActiveSubscriptionsForUser(customerIDs []string, excludeSubscriptionID string) ([]string, error) {
	if len(customerIDs) == 0 {
		return []string{}, nil
	}

	// Build query to find all active subscriptions for these customers
	query := `
		SELECT ss.stripe_id
		FROM stripe_subscriptions_v2 ss
		JOIN stripe_customers_v2 sc ON ss.customer_id = sc.id
		WHERE sc.stripe_id = ANY($1)
		AND ss.status IN ('active', 'trialing')
		AND ss.stripe_id != $2
		AND ss.canceled_at IS NULL
	`

	rows, err := s.db.Query(query, pq.Array(customerIDs), excludeSubscriptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptionIDs []string
	for rows.Next() {
		var subID string
		if err := rows.Scan(&subID); err != nil {
			log.Printf("⚠️  [Subscription Manager] Failed to scan subscription ID: %v", err)
			continue
		}
		subscriptionIDs = append(subscriptionIDs, subID)
	}

	return subscriptionIDs, nil
}

// ================================================================
// VIDEO ACCESS MANAGEMENT
// ================================================================

// GrantVideoAccess grants video access to a user based on their subscription
// This function is idempotent - calling it multiple times won't cause errors
func (s *SubscriptionManagerService) GrantVideoAccess(userID int, reason string) error {
	// Check if user already has video access
	var hasAccess bool
	var currentSource string
	err := s.db.QueryRow(`
		SELECT COALESCE(has_video_access, false), COALESCE(video_access_source, '')
		FROM users 
		WHERE id = $1
	`, userID).Scan(&hasAccess, &currentSource)
	
	if err != nil {
		return fmt.Errorf("failed to check current access for user %d: %w", userID, err)
	}

	// If user already has access, update source to track all confirmation methods
	if hasAccess {
		// Append new source if not already present
		var updatedSource string
		if currentSource == "" {
			updatedSource = reason
		} else if !contains(currentSource, reason) {
			updatedSource = currentSource + "," + reason
		} else {
			// Source already tracked, no update needed
			log.Printf("ℹ️  [Subscription Manager] User %d already has video access from: %s", userID, currentSource)
			return nil
		}
		
		_, err = s.db.Exec(`
			UPDATE users 
			SET video_access_source = $1,
			    updated_at = NOW()
			WHERE id = $2
		`, updatedSource, userID)
		
		if err != nil {
			log.Printf("⚠️  [Subscription Manager] Failed to update access source for user %d: %v", userID, err)
			// Don't return error - access already granted
		} else {
			log.Printf("ℹ️  [Subscription Manager] User %d already has video access, updated source: %s", userID, updatedSource)
		}
		return nil
	}

	// Grant new video access
	log.Printf("🎥 [Subscription Manager] Granting video access to user %d (reason: %s)", userID, reason)
	
	query := `
		UPDATE users 
		SET has_video_access = true,
		    video_access_granted_at = NOW(),
		    video_access_source = $1,
		    manual_video_access = true,
		    updated_at = NOW()
		WHERE id = $2
	`

	result, err := s.db.Exec(query, reason, userID)
	if err != nil {
		return fmt.Errorf("failed to grant video access: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user %d not found", userID)
	}

	log.Printf("✅ [Subscription Manager] Video access granted to user %d", userID)
	return nil
}

// Helper function to check if a source string contains a specific value
func contains(source, value string) bool {
	if source == value {
		return true
	}
	// Check if value exists as a comma-separated item
	sources := splitSources(source)
	for _, s := range sources {
		if s == value {
			return true
		}
	}
	return false
}

// Helper function to split comma-separated sources
func splitSources(source string) []string {
	if source == "" {
		return []string{}
	}
	var result []string
	current := ""
	for _, char := range source {
		if char == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// RevokeVideoAccess revokes video access from a user
func (s *SubscriptionManagerService) RevokeVideoAccess(userID int, reason string) error {
	log.Printf("🚫 [Subscription Manager] Revoking video access from user %d (reason: %s)", userID, reason)

	// First, check if they have any other active subscriptions
	hasActiveSubscription, err := s.hasActiveSubscription(userID)
	if err != nil {
		return fmt.Errorf("failed to check for active subscriptions: %w", err)
	}

	if hasActiveSubscription {
		log.Printf("ℹ️  [Subscription Manager] User %d still has an active subscription - keeping video access", userID)
		return nil
	}

	query := `
		UPDATE users 
		SET manual_video_access = false,
		    updated_at = NOW()
		WHERE id = $1
	`

	result, err := s.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke video access: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user %d not found", userID)
	}

	log.Printf("✅ [Subscription Manager] Video access revoked from user %d", userID)
	return nil
}

// hasActiveSubscription checks if a user has any active subscriptions
func (s *SubscriptionManagerService) hasActiveSubscription(userID int) (bool, error) {
	// Get linked customers
	linkedCustomers, err := s.linkingService.GetUserLinkedCustomers(userID)
	if err != nil {
		return false, fmt.Errorf("failed to get linked customers: %w", err)
	}

	if len(linkedCustomers) == 0 {
		return false, nil
	}

	// Check for active subscriptions
	query := `
		SELECT COUNT(*)
		FROM stripe_subscriptions_v2 ss
		JOIN stripe_customers_v2 sc ON ss.customer_id = sc.id
		WHERE sc.stripe_id = ANY($1)
		AND ss.status IN ('active', 'trialing')
		AND ss.canceled_at IS NULL
		AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
	`

	var count int
	// Use pq.Array to properly convert Go slice to PostgreSQL array
	err = s.db.QueryRow(query, pq.Array(linkedCustomers)).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to count active subscriptions: %w", err)
	}

	return count > 0, nil
}

// UpdateVideoAccessForSubscription updates video access based on subscription status
func (s *SubscriptionManagerService) UpdateVideoAccessForSubscription(subscriptionID string) error {
	log.Printf("🔄 [Subscription Manager] Updating video access for subscription: %s", subscriptionID)

	// Step 1: Get subscription details
	query := `
		SELECT sc.stripe_id, ss.status
		FROM stripe_subscriptions_v2 ss
		JOIN stripe_customers_v2 sc ON ss.customer_id = sc.id
		WHERE ss.stripe_id = $1
	`

	var customerID, status string
	err := s.db.QueryRow(query, subscriptionID).Scan(&customerID, &status)
	if err != nil {
		return fmt.Errorf("failed to get subscription details: %w", err)
	}

	// Step 2: Get user linked to this customer
	user, err := s.linkingService.GetUserByStripeCustomerID(customerID)
	if err != nil {
		log.Printf("ℹ️  [Subscription Manager] No user linked to customer %s", customerID)
		return nil // Not an error - customer might not have a user account
	}

	// Step 3: Grant or revoke video access based on status
	if status == "active" || status == "trialing" {
		return s.GrantVideoAccess(user.ID, fmt.Sprintf("subscription %s is %s", subscriptionID, status))
	} else if status == "incomplete" {
		// For incomplete subscriptions, check if the latest invoice was paid
		// This handles cases where payment succeeds but the webhook for status update is delayed
		log.Printf("⚠️  [Subscription Manager] Subscription %s is incomplete - checking payment status", subscriptionID)
		
		// Query to check if the subscription has a paid invoice
		var hasPaidInvoice bool
		invoiceQuery := `
			SELECT EXISTS(
				SELECT 1 FROM stripe_invoices si
				WHERE si.subscription_id = $1
				AND si.status = 'paid'
				AND si.paid = true
			)
		`
		err := s.db.QueryRow(invoiceQuery, subscriptionID).Scan(&hasPaidInvoice)
		if err != nil {
			log.Printf("⚠️  [Subscription Manager] Failed to check invoice status for subscription %s: %v", subscriptionID, err)
			// Don't fail - just don't grant access yet
			return nil
		}
		
		if hasPaidInvoice {
			log.Printf("✅ [Subscription Manager] Subscription %s is incomplete but has paid invoice - granting access", subscriptionID)
			return s.GrantVideoAccess(user.ID, fmt.Sprintf("subscription %s has paid invoice", subscriptionID))
		}
		
		log.Printf("ℹ️  [Subscription Manager] Subscription %s is incomplete and payment not confirmed yet", subscriptionID)
		return nil
	} else if status == "canceled" || status == "past_due" || status == "unpaid" {
		return s.RevokeVideoAccess(user.ID, fmt.Sprintf("subscription %s is %s", subscriptionID, status))
	}

	log.Printf("ℹ️  [Subscription Manager] Subscription %s status is %s - no action taken", subscriptionID, status)
	return nil
}

// ================================================================
// SUBSCRIPTION DIAGNOSTICS
// ================================================================

// GetUserSubscriptionSummary returns a summary of a user's subscriptions
func (s *SubscriptionManagerService) GetUserSubscriptionSummary(userID int) (map[string]interface{}, error) {
	summary := make(map[string]interface{})
	summary["user_id"] = userID

	// Get linked customers
	linkedCustomers, err := s.linkingService.GetUserLinkedCustomers(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get linked customers: %w", err)
	}
	summary["linked_customers"] = linkedCustomers
	summary["linked_customer_count"] = len(linkedCustomers)

	if len(linkedCustomers) == 0 {
		summary["active_subscriptions"] = 0
		summary["has_video_access"] = false
		return summary, nil
	}

	// Count active subscriptions
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE ss.status IN ('active', 'trialing') AND ss.canceled_at IS NULL) as active_count,
			COUNT(*) FILTER (WHERE ss.status = 'canceled') as canceled_count,
			COUNT(*) FILTER (WHERE ss.status = 'past_due') as past_due_count,
			COUNT(*) as total_count
		FROM stripe_subscriptions_v2 ss
		JOIN stripe_customers_v2 sc ON ss.customer_id = sc.id
		WHERE sc.stripe_id = ANY($1)
	`

	var activeCount, canceledCount, pastDueCount, totalCount int
	err = s.db.QueryRow(query, pq.Array(linkedCustomers)).Scan(&activeCount, &canceledCount, &pastDueCount, &totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count subscriptions: %w", err)
	}

	summary["active_subscriptions"] = activeCount
	summary["canceled_subscriptions"] = canceledCount
	summary["past_due_subscriptions"] = pastDueCount
	summary["total_subscriptions"] = totalCount

	// Check video access
	var hasVideoAccess bool
	s.db.QueryRow("SELECT manual_video_access FROM users WHERE id = $1", userID).Scan(&hasVideoAccess)
	summary["has_video_access"] = hasVideoAccess

	// Recommendation
	if activeCount > 1 {
		summary["recommendation"] = fmt.Sprintf("User has %d active subscriptions - should be consolidated", activeCount)
		summary["action_needed"] = true
	} else {
		summary["recommendation"] = "All good!"
		summary["action_needed"] = false
	}

	return summary, nil
}

// FixMultipleSubscriptions finds and fixes users with multiple active subscriptions
func (s *SubscriptionManagerService) FixMultipleSubscriptions(ctx context.Context) ([]SubscriptionResult, error) {
	log.Printf("🔧 [Subscription Manager] Finding users with multiple active subscriptions...")

	// Query to find users with multiple active subscriptions
	query := `
		SELECT 
			usc.user_id,
			array_agg(DISTINCT ss.stripe_id ORDER BY ss.stripe_created_at DESC) as subscription_ids
		FROM user_stripe_customers_v2 usc
		JOIN stripe_customers_v2 sc ON usc.stripe_customer_id = sc.stripe_id
		JOIN stripe_subscriptions_v2 ss ON ss.customer_id = sc.id
		WHERE ss.status IN ('active', 'trialing')
		AND ss.canceled_at IS NULL
		GROUP BY usc.user_id
		HAVING COUNT(DISTINCT ss.stripe_id) > 1
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query multiple subscriptions: %w", err)
	}
	defer rows.Close()

	var results []SubscriptionResult
	for rows.Next() {
		var userID int
		var subscriptionIDs []string

		if err := rows.Scan(&userID, &subscriptionIDs); err != nil {
			log.Printf("⚠️  [Subscription Manager] Failed to scan row: %v", err)
			continue
		}

		log.Printf("⚠️  [Subscription Manager] User %d has %d active subscriptions", userID, len(subscriptionIDs))

		// Keep the newest subscription, cancel the rest
		if len(subscriptionIDs) > 1 {
			newestSubscription := subscriptionIDs[0] // Already sorted by created_at DESC
			result, err := s.EnforceSingleSubscription(userID, newestSubscription)
			if err != nil {
				log.Printf("❌ [Subscription Manager] Failed to enforce single subscription for user %d: %v", userID, err)
				results = append(results, SubscriptionResult{
					UserID: userID,
					Error:  err.Error(),
				})
			} else {
				results = append(results, *result)
			}

			// Add a small delay to avoid rate limiting
			time.Sleep(100 * time.Millisecond)
		}
	}

	log.Printf("✅ [Subscription Manager] Processed %d users with multiple subscriptions", len(results))
	return results, nil
}
