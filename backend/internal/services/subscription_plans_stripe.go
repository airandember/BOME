package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"bome-backend/internal/database"
)

// SubscriptionPlanStripeService handles subscription plans with Stripe integration
type SubscriptionPlanStripeService struct {
	*SubscriptionPlanService // Embed existing service
	stripeService            *StripeService
}

// NewSubscriptionPlanStripeService creates a new service with Stripe integration
func NewSubscriptionPlanStripeService(db *database.DB, stripeService *StripeService) *SubscriptionPlanStripeService {
	return &SubscriptionPlanStripeService{
		SubscriptionPlanService: NewSubscriptionPlanService(db),
		stripeService:           stripeService,
	}
}

// CreateSubscriptionPlanWithStripe creates a subscription plan and auto-creates Stripe product/price
func (s *SubscriptionPlanStripeService) CreateSubscriptionPlanWithStripe(ctx context.Context, plan *database.SubscriptionPlan, autoCreateStripe bool) (*SubscriptionPlanResponse, error) {
	var stripeProductID, stripePriceID string
	var err error

	// Auto-create Stripe product and price if enabled and Stripe is available
	if autoCreateStripe && s.stripeService != nil && s.stripeService.IsEnabled() {
		// Create Stripe product
		metadata := map[string]string{
			"plan_type":    plan.SubType,
			"created_by":   "subscription_system",
			"auto_created": "true",
		}

		description := plan.Description

		stripeProduct, err := s.stripeService.CreateProduct(plan.Name, description, metadata)
		if err != nil {
			log.Printf("Warning: Failed to create Stripe product: %v", err)
			// Continue without Stripe integration
		} else {
			stripeProductID = stripeProduct.ID

			// Create Stripe price
			unitAmount := int64(plan.Price * 100) // Convert to cents

			// Validate Stripe limitations
			if err := validateStripeInterval(plan.Interval, plan.IntervalCount); err != nil {
				return nil, fmt.Errorf("invalid interval configuration: %w", err)
			}

			stripePrice, err := s.stripeService.CreatePrice(
				stripeProductID,
				strings.ToLower(plan.Currency),
				unitAmount,
				plan.Interval,
				plan.IntervalCount,
				metadata,
			)
			if err != nil {
				log.Printf("Warning: Failed to create Stripe price: %v", err)
			} else {
				stripePriceID = stripePrice.ID
				// Update plan with Stripe IDs
				plan.StripePriceID = sql.NullString{String: stripePriceID, Valid: true}

				log.Printf("Successfully created Stripe product (%s) and price (%s) for plan", stripeProductID, stripePriceID)
			}
		}
	}

	// Create the subscription plan
	response, err := s.SubscriptionPlanService.CreateSubscriptionPlan(ctx, plan)
	if err != nil {
		// If plan creation fails but Stripe entities were created, we should ideally clean them up
		// For now, we'll log the issue
		if stripePriceID != "" {
			log.Printf("Warning: Plan creation failed but Stripe entities were created. Product: %s, Price: %s", stripeProductID, stripePriceID)
		}
		return nil, err
	}

	// Add Stripe IDs to response if they were created
	if stripeProductID != "" {
		log.Printf("Plan created with Stripe integration: Product=%s, Price=%s", stripeProductID, stripePriceID)
	}

	return response, nil
}

// SyncWithStripe syncs an existing plan with Stripe
func (s *SubscriptionPlanStripeService) SyncWithStripe(ctx context.Context, planID string) (*SubscriptionPlanResponse, error) {
	if s.stripeService == nil || !s.stripeService.IsEnabled() {
		return nil, fmt.Errorf("stripe service is not available")
	}

	// Get the existing plan
	plan, err := s.GetSubscriptionPlan(ctx, planID)
	if err != nil {
		return nil, err
	}

	// Create Stripe product and price if they don't exist
	if plan.StripePriceID == nil || *plan.StripePriceID == "" {
		metadata := map[string]string{
			"plan_id":   planID,
			"plan_type": plan.SubType,
			"synced_at": time.Now().Format(time.RFC3339),
		}

		// Create Stripe product
		stripeProduct, err := s.stripeService.CreateProduct(plan.Name, plan.Description, metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to create Stripe product: %w", err)
		}

		// Create Stripe price
		unitAmount := int64(plan.Price * 100)

		// Validate Stripe limitations
		if err := validateStripeInterval(plan.Interval, plan.IntervalCount); err != nil {
			return nil, fmt.Errorf("invalid interval configuration for plan '%s': %w", plan.Name, err)
		}

		stripePrice, err := s.stripeService.CreatePrice(
			stripeProduct.ID,
			strings.ToLower(plan.Currency),
			unitAmount,
			plan.Interval,
			plan.IntervalCount,
			metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create Stripe price: %w", err)
		}

		// Update the plan with Stripe IDs
		updates := map[string]interface{}{
			"stripe_price_id": stripePrice.ID,
		}

		return s.UpdateSubscriptionPlan(ctx, planID, updates)
	}

	return plan, nil
}

// GetStripeIntegrationStatus returns the Stripe integration status for a plan
func (s *SubscriptionPlanStripeService) GetStripeIntegrationStatus(ctx context.Context, planID string) (map[string]interface{}, error) {
	plan, err := s.GetSubscriptionPlan(ctx, planID)
	if err != nil {
		return nil, err
	}

	status := map[string]interface{}{
		"has_stripe_product": false,
		"has_stripe_price":   false,
		"sync_status":        "not_synced",
	}

	if plan.StripePriceID != nil && *plan.StripePriceID != "" {
		status["has_stripe_price"] = true
		status["stripe_price_id"] = *plan.StripePriceID
		status["sync_status"] = "synced"

		// If we have Stripe service, we could verify the price exists
		if s.stripeService != nil && s.stripeService.IsEnabled() {
			// For now, assume if we have a price ID, we have a product
			status["has_stripe_product"] = true
		}
	}

	return status, nil
}

// validateStripeInterval validates interval and interval_count against Stripe's limitations
func validateStripeInterval(interval string, intervalCount int) error {
	switch interval {
	case "year":
		if intervalCount > 3 {
			return fmt.Errorf("yearly subscriptions can have a maximum interval_count of 3, got %d", intervalCount)
		}
	case "month":
		if intervalCount > 12 {
			return fmt.Errorf("monthly subscriptions can have a maximum interval_count of 12, got %d", intervalCount)
		}
	case "week":
		if intervalCount > 52 {
			return fmt.Errorf("weekly subscriptions can have a maximum interval_count of 52, got %d", intervalCount)
		}
	case "day":
		if intervalCount > 365 {
			return fmt.Errorf("daily subscriptions can have a maximum interval_count of 365, got %d", intervalCount)
		}
	default:
		return fmt.Errorf("invalid interval '%s', must be one of: day, week, month, year", interval)
	}

	if intervalCount < 1 {
		return fmt.Errorf("interval_count must be at least 1, got %d", intervalCount)
	}

	return nil
}
