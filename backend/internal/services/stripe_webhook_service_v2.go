package services

import (
	"context"
	"fmt"
	"log"

	"bome-backend/internal/database"

	"github.com/stripe/stripe-go/v74"
)

// StripeWebhookServiceV2 handles webhook events and routes them to v2 tables
type StripeWebhookServiceV2 struct {
	syncService         *StripeSyncV2Service
	linkingService      *CustomerLinkingService
	subscriptionManager *SubscriptionManagerService
	db                  *database.DB
}

// NewStripeWebhookServiceV2 creates a new webhook service for v2 tables
func NewStripeWebhookServiceV2(syncService *StripeSyncV2Service, linkingService *CustomerLinkingService, subscriptionManager *SubscriptionManagerService, db *database.DB) *StripeWebhookServiceV2 {
	return &StripeWebhookServiceV2{
		syncService:         syncService,
		linkingService:      linkingService,
		subscriptionManager: subscriptionManager,
		db:                  db,
	}
}

// ================================================================
// CUSTOMER WEBHOOK HANDLERS
// ================================================================

// HandleCustomerCreated processes customer.created webhook events
func (s *StripeWebhookServiceV2) HandleCustomerCreated(customer *stripe.Customer) error {
	log.Printf("👥 [Webhook v2] Customer created: %s (%s)", customer.ID, customer.Email)

	ctx := context.Background()

	// Step 1: Sync customer to stripe_customers_v2
	if err := s.syncService.SyncSingleCustomer(ctx, customer.ID); err != nil {
		return fmt.Errorf("failed to sync customer to v2: %w", err)
	}

	// Step 2: Auto-link to user by email (if email exists)
	if customer.Email != "" {
		log.Printf("🔗 [Webhook v2] Attempting to link customer %s to user with email: %s", customer.ID, customer.Email)

		// Find user by email
		user, err := s.db.GetUserByEmail(customer.Email)
		if err != nil {
			log.Printf("ℹ️  [Webhook v2] No user found for email %s - customer synced but not linked", customer.Email)
		} else {
			// Link the user to this customer
			if _, err := s.linkingService.LinkUserToCustomers(user.ID); err != nil {
				log.Printf("⚠️  [Webhook v2] Failed to auto-link customer %s to user %d: %v", customer.ID, user.ID, err)
			} else {
				log.Printf("✅ [Webhook v2] Customer %s successfully linked to user %d (%s)", customer.ID, user.ID, user.Email)
			}
		}
	} else {
		log.Printf("ℹ️  [Webhook v2] Customer %s has no email - skipping auto-link", customer.ID)
	}

	return nil
}

// HandleCustomerUpdated processes customer.updated webhook events
func (s *StripeWebhookServiceV2) HandleCustomerUpdated(customer *stripe.Customer) error {
	log.Printf("👥 [Webhook v2] Customer updated: %s (%s)", customer.ID, customer.Email)

	ctx := context.Background()

	// Sync customer to stripe_customers_v2
	if err := s.syncService.SyncSingleCustomer(ctx, customer.ID); err != nil {
		return fmt.Errorf("failed to sync customer to v2: %w", err)
	}

	// Re-link if email changed
	if customer.Email != "" {
		log.Printf("🔗 [Webhook v2] Re-linking customer %s (email may have changed)", customer.ID)
		user, err := s.db.GetUserByEmail(customer.Email)
		if err == nil {
			if _, err := s.linkingService.LinkUserToCustomers(user.ID); err != nil {
				log.Printf("⚠️  [Webhook v2] Failed to re-link customer %s: %v", customer.ID, err)
			}
		}
	}

	return nil
}

// HandleCustomerDeleted processes customer.deleted webhook events
func (s *StripeWebhookServiceV2) HandleCustomerDeleted(customer *stripe.Customer) error {
	log.Printf("👥 [Webhook v2] Customer deleted: %s", customer.ID)

	// Mark as deleted in stripe_customers_v2 (soft delete via deleted_at timestamp)
	query := `
		UPDATE stripe_customers_v2 
		SET deleted_at = NOW(), 
		    last_synced_at = NOW() 
		WHERE stripe_id = $1
	`

	result, err := s.db.Exec(query, customer.ID)
	if err != nil {
		return fmt.Errorf("failed to mark customer as deleted: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("⚠️  [Webhook v2] Customer %s not found in v2 tables", customer.ID)
	}

	// Note: We don't delete from user_stripe_customers_v2 - we keep the historical link
	log.Printf("✅ [Webhook v2] Customer %s marked as deleted in v2 tables", customer.ID)

	return nil
}

// ================================================================
// SUBSCRIPTION WEBHOOK HANDLERS
// ================================================================

// HandleSubscriptionCreated processes customer.subscription.created webhook events
func (s *StripeWebhookServiceV2) HandleSubscriptionCreated(subscription *stripe.Subscription) error {
	log.Printf("📋 [Webhook v2] Subscription created: %s (Customer: %s, Status: %s)",
		subscription.ID, subscription.Customer.ID, subscription.Status)

	ctx := context.Background()

	// Step 1: Sync subscription to stripe_subscriptions_v2
	if err := s.syncService.SyncSingleSubscription(ctx, subscription.ID); err != nil {
		return fmt.Errorf("failed to sync subscription to v2: %w", err)
	}

	// Step 2: Check if customer is linked to a user
	user, err := s.linkingService.GetUserByStripeCustomerID(subscription.Customer.ID)
	if err != nil {
		log.Printf("ℹ️  [Webhook v2] No user found for customer %s - subscription synced but not linked", subscription.Customer.ID)
		return nil // Not an error - customer might not have a user account yet
	}

	log.Printf("✅ [Webhook v2] Subscription %s linked to user %d (%s)", subscription.ID, user.ID, user.Email)

	// Step 3: Phase 6 - Enforce single subscription rule
	log.Printf("🔒 [Webhook v2] Enforcing single subscription rule for user %d", user.ID)
	result, err := s.subscriptionManager.EnforceSingleSubscription(user.ID, subscription.ID)
	if err != nil {
		log.Printf("⚠️  [Webhook v2] Failed to enforce single subscription: %v", err)
		// Don't fail the webhook - subscription is still synced
	} else if len(result.CanceledSubscriptionIDs) > 0 {
		log.Printf("✅ [Webhook v2] Canceled %d old subscriptions for user %d", len(result.CanceledSubscriptionIDs), user.ID)
	}

	// Step 4: Grant video access if subscription is active
	if subscription.Status == stripe.SubscriptionStatusActive || subscription.Status == stripe.SubscriptionStatusTrialing {
		if err := s.subscriptionManager.GrantVideoAccess(user.ID, fmt.Sprintf("subscription %s is %s", subscription.ID, subscription.Status)); err != nil {
			log.Printf("⚠️  [Webhook v2] Failed to grant video access: %v", err)
			// Don't fail the webhook
		}
	}

	return nil
}

// HandleSubscriptionUpdated processes customer.subscription.updated webhook events
func (s *StripeWebhookServiceV2) HandleSubscriptionUpdated(subscription *stripe.Subscription) error {
	log.Printf("📋 [Webhook v2] Subscription updated: %s (Customer: %s, Status: %s)",
		subscription.ID, subscription.Customer.ID, subscription.Status)

	ctx := context.Background()

	// Sync subscription to stripe_subscriptions_v2
	if err := s.syncService.SyncSingleSubscription(ctx, subscription.ID); err != nil {
		return fmt.Errorf("failed to sync subscription to v2: %w", err)
	}

	// Phase 6: Update video access based on subscription status
	if err := s.subscriptionManager.UpdateVideoAccessForSubscription(subscription.ID); err != nil {
		log.Printf("⚠️  [Webhook v2] Failed to update video access for subscription %s: %v", subscription.ID, err)
		// Don't fail the webhook
	}

	return nil
}

// HandleSubscriptionDeleted processes customer.subscription.deleted webhook events
func (s *StripeWebhookServiceV2) HandleSubscriptionDeleted(subscription *stripe.Subscription) error {
	log.Printf("📋 [Webhook v2] Subscription deleted: %s (Customer: %s)",
		subscription.ID, subscription.Customer.ID)

	// Mark as deleted in stripe_subscriptions_v2
	query := `
		UPDATE stripe_subscriptions_v2 
		SET status = 'canceled',
		    canceled_at = NOW(),
		    last_synced_at = NOW() 
		WHERE stripe_id = $1
	`

	result, err := s.db.Exec(query, subscription.ID)
	if err != nil {
		return fmt.Errorf("failed to mark subscription as deleted: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("⚠️  [Webhook v2] Subscription %s not found in v2 tables", subscription.ID)
	}

	// Phase 6: Revoke video access if user has no other active subscriptions
	user, err := s.linkingService.GetUserByStripeCustomerID(subscription.Customer.ID)
	if err == nil {
		if err := s.subscriptionManager.RevokeVideoAccess(user.ID, fmt.Sprintf("subscription %s was deleted", subscription.ID)); err != nil {
			log.Printf("⚠️  [Webhook v2] Failed to revoke video access: %v", err)
			// Don't fail the webhook
		}
	}

	log.Printf("✅ [Webhook v2] Subscription %s marked as deleted in v2 tables", subscription.ID)

	return nil
}

// ================================================================
// PRODUCT WEBHOOK HANDLERS
// ================================================================

// HandleProductCreated processes product.created webhook events
func (s *StripeWebhookServiceV2) HandleProductCreated(product *stripe.Product) error {
	log.Printf("📦 [Webhook v2] Product created: %s (%s)", product.ID, product.Name)

	ctx := context.Background()

	if err := s.syncService.SyncSingleProduct(ctx, product.ID); err != nil {
		return fmt.Errorf("failed to sync product to v2: %w", err)
	}

	return nil
}

// HandleProductUpdated processes product.updated webhook events
func (s *StripeWebhookServiceV2) HandleProductUpdated(product *stripe.Product) error {
	log.Printf("📦 [Webhook v2] Product updated: %s (%s)", product.ID, product.Name)

	ctx := context.Background()

	if err := s.syncService.SyncSingleProduct(ctx, product.ID); err != nil {
		return fmt.Errorf("failed to sync product to v2: %w", err)
	}

	return nil
}

// HandleProductDeleted processes product.deleted webhook events
func (s *StripeWebhookServiceV2) HandleProductDeleted(product *stripe.Product) error {
	log.Printf("📦 [Webhook v2] Product deleted: %s", product.ID)

	// Mark as deleted in stripe_products_v2
	query := `
		UPDATE stripe_products_v2 
		SET active = false,
		    deleted_at = NOW(),
		    last_synced_at = NOW() 
		WHERE stripe_id = $1
	`

	result, err := s.db.Exec(query, product.ID)
	if err != nil {
		return fmt.Errorf("failed to mark product as deleted: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("⚠️  [Webhook v2] Product %s not found in v2 tables", product.ID)
	}

	log.Printf("✅ [Webhook v2] Product %s marked as deleted in v2 tables", product.ID)

	return nil
}

// ================================================================
// PRICE WEBHOOK HANDLERS
// ================================================================

// HandlePriceCreated processes price.created webhook events
func (s *StripeWebhookServiceV2) HandlePriceCreated(price *stripe.Price) error {
	log.Printf("💰 [Webhook v2] Price created: %s (Amount: %d)", price.ID, price.UnitAmount)

	ctx := context.Background()

	if err := s.syncService.SyncSinglePrice(ctx, price.ID); err != nil {
		return fmt.Errorf("failed to sync price to v2: %w", err)
	}

	return nil
}

// HandlePriceUpdated processes price.updated webhook events
func (s *StripeWebhookServiceV2) HandlePriceUpdated(price *stripe.Price) error {
	log.Printf("💰 [Webhook v2] Price updated: %s (Amount: %d)", price.ID, price.UnitAmount)

	ctx := context.Background()

	if err := s.syncService.SyncSinglePrice(ctx, price.ID); err != nil {
		return fmt.Errorf("failed to sync price to v2: %w", err)
	}

	return nil
}

// HandlePriceDeleted processes price.deleted webhook events
func (s *StripeWebhookServiceV2) HandlePriceDeleted(price *stripe.Price) error {
	log.Printf("💰 [Webhook v2] Price deleted: %s", price.ID)

	// Mark as deleted in stripe_prices_v2
	query := `
		UPDATE stripe_prices_v2 
		SET active = false,
		    deleted_at = NOW(),
		    last_synced_at = NOW() 
		WHERE stripe_id = $1
	`

	result, err := s.db.Exec(query, price.ID)
	if err != nil {
		return fmt.Errorf("failed to mark price as deleted: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("⚠️  [Webhook v2] Price %s not found in v2 tables", price.ID)
	}

	log.Printf("✅ [Webhook v2] Price %s marked as deleted in v2 tables", price.ID)

	return nil
}

// ================================================================
// INVOICE PAYMENT WEBHOOK HANDLERS (Phase 6)
// ================================================================

// HandleInvoicePaymentSucceeded processes invoice.payment_succeeded webhook events
func (s *StripeWebhookServiceV2) HandleInvoicePaymentSucceeded(invoice *stripe.Invoice) error {
	log.Printf("🧾 [Webhook v2] Invoice payment succeeded: %s (Amount: %d)", invoice.ID, invoice.AmountPaid)

	// Grant video access when payment succeeds
	if invoice.Customer != nil && invoice.Subscription != nil {
		log.Printf("🎥 [Webhook v2] Processing video access for customer %s (subscription: %s)",
			invoice.Customer.ID, invoice.Subscription.ID)

		// Get user linked to this customer
		user, err := s.linkingService.GetUserByStripeCustomerID(invoice.Customer.ID)
		if err != nil {
			log.Printf("ℹ️  [Webhook v2] No user found for customer %s", invoice.Customer.ID)
			return nil // Not an error - customer might not have a user account
		}

		// Grant video access
		reason := fmt.Sprintf("invoice %s paid (subscription: %s)", invoice.ID, invoice.Subscription.ID)
		if err := s.subscriptionManager.GrantVideoAccess(user.ID, reason); err != nil {
			log.Printf("⚠️  [Webhook v2] Failed to grant video access: %v", err)
			// Don't fail the webhook
		} else {
			log.Printf("✅ [Webhook v2] Video access granted to user %d", user.ID)
		}
	}

	return nil
}

// HandleInvoicePaymentFailed processes invoice.payment_failed webhook events
func (s *StripeWebhookServiceV2) HandleInvoicePaymentFailed(invoice *stripe.Invoice) error {
	log.Printf("🧾 [Webhook v2] Invoice payment failed: %s", invoice.ID)

	// Revoke video access when payment fails
	if invoice.Customer != nil && invoice.Subscription != nil {
		log.Printf("🚫 [Webhook v2] Payment failed for customer %s (subscription: %s)",
			invoice.Customer.ID, invoice.Subscription.ID)

		// Get user linked to this customer
		user, err := s.linkingService.GetUserByStripeCustomerID(invoice.Customer.ID)
		if err != nil {
			log.Printf("ℹ️  [Webhook v2] No user found for customer %s", invoice.Customer.ID)
			return nil // Not an error
		}

		// Revoke video access (but only if they have no other active subscriptions)
		reason := fmt.Sprintf("invoice %s payment failed (subscription: %s)", invoice.ID, invoice.Subscription.ID)
		if err := s.subscriptionManager.RevokeVideoAccess(user.ID, reason); err != nil {
			log.Printf("⚠️  [Webhook v2] Failed to revoke video access: %v", err)
			// Don't fail the webhook
		} else {
			log.Printf("✅ [Webhook v2] Video access revoked from user %d", user.ID)
		}
	}

	return nil
}
