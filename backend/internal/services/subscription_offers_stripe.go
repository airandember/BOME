package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"bome-backend/internal/database"
)

// SubscriptionOffersStripeService handles subscription offers with Stripe integration
type SubscriptionOffersStripeService struct {
	*SubscriptionOffersService // Embed existing service
	stripeService              *StripeService
}

// NewSubscriptionOffersStripeService creates a new service with Stripe integration
func NewSubscriptionOffersStripeService(db *database.DB, stripeService *StripeService) *SubscriptionOffersStripeService {
	return &SubscriptionOffersStripeService{
		SubscriptionOffersService: NewSubscriptionOffersService(db),
		stripeService:             stripeService,
	}
}

// CreateOfferWithStripe creates an offer and optionally creates Stripe coupon/promotion code
func (s *SubscriptionOffersStripeService) CreateOfferWithStripe(ctx context.Context, offer *database.SubscriptionOffer, autoCreateStripe bool) (*SubscriptionOfferResponse, error) {
	var stripeCouponID, stripePromotionCodeID string
	var err error

	// Auto-create Stripe coupon and promotion code if enabled and Stripe is available
	if autoCreateStripe && s.stripeService != nil && s.stripeService.IsEnabled() {
		// Enhanced duration mapping with business logic
		duration, durationInMonths, _ := mapOfferDurationToStripe(offer.OfferStartDate, offer.OffEndDate)

		// Validate coupon data
		if err := ValidateCouponData(offer.OffDiscountType, offer.OffDiscountValue, duration, durationInMonths, nil); err != nil {
			return nil, fmt.Errorf("invalid coupon configuration: %w", err)
		}

		// Enhanced metadata for better tracking and analytics
		metadata := map[string]string{
			"offer_id":       strconv.Itoa(offer.ID),
			"offer_type":     "subscription_offer",
			"plan_id":        strconv.Itoa(offer.PlanID),
			"discount_type":  offer.OffDiscountType,
			"discount_value": fmt.Sprintf("%.2f", offer.OffDiscountValue),
			"priority":       strconv.Itoa(offer.OffPriority),
			"auto_apply":     strconv.FormatBool(offer.OffAutoApply),
			"duration":       duration,
			"created_by":     "subscription_system",
			"auto_created":   "true",
			"business_unit":  "subscriptions",
			"campaign":       "auto_sync",
			"created_at":     time.Now().Format(time.RFC3339),
		}

		// Add date information if available
		if offer.OfferStartDate.Valid {
			metadata["active_from"] = offer.OfferStartDate.Time.Format(time.RFC3339)
		}
		if offer.OffEndDate.Valid {
			metadata["active_until"] = offer.OffEndDate.Time.Format(time.RFC3339)
		}

		// Create Stripe coupon
		var percentOff *float64
		var amountOff *int64
		var currency string

		if offer.OffDiscountType == "percentage" {
			percentOff = &offer.OffDiscountValue
		} else if offer.OffDiscountType == "amount" {
			amountOffValue := int64(offer.OffDiscountValue * 100) // Convert to cents
			amountOff = &amountOffValue
			currency = "usd" // Default currency, could be configurable
		}

		// Set max redemptions if specified
		var maxRedemptions *int64
		if offer.OffMaxUses.Valid && offer.OffMaxUses.Int32 > 0 {
			maxRedVal := int64(offer.OffMaxUses.Int32)
			maxRedemptions = &maxRedVal
		}

		// Calculate expiration timestamp
		expirationTimestamp := calculateExpirationTimestamp(offer.OffEndDate)

		stripeCoupon, err := s.stripeService.CreateCoupon(
			offer.OffName,
			percentOff,
			amountOff,
			currency,
			duration,
			durationInMonths,
			maxRedemptions,
			metadata,
			expirationTimestamp,
		)
		if err != nil {
			// Extract more specific error information for better frontend display
			var errorMsg string
			if strings.Contains(err.Error(), "redeem_by") {
				errorMsg = fmt.Sprintf("Invalid expiration date: %s. Please ensure the offer end date is in the future.", offer.OffEndDate)
			} else if strings.Contains(err.Error(), "amount_off") {
				errorMsg = "Invalid discount amount. Please check the discount value."
			} else if strings.Contains(err.Error(), "percent_off") {
				errorMsg = "Invalid discount percentage. Please check the discount value."
			} else {
				errorMsg = fmt.Sprintf("Stripe API error: %s", err.Error())
			}
			return nil, fmt.Errorf("failed to create Stripe coupon: %s", errorMsg)
		}
		stripeCouponID = stripeCoupon.ID
		log.Printf("Successfully created Stripe coupon: %s", stripeCouponID)

		// Create promotion code if offer has a code
		if offer.OffCode.Valid && offer.OffCode.String != "" {
			promotionMetadata := map[string]string{
				"offer_id":       strconv.Itoa(offer.ID),
				"offer_code":     offer.OffCode.String,
				"plan_id":        strconv.Itoa(offer.PlanID),
				"discount_type":  offer.OffDiscountType,
				"discount_value": fmt.Sprintf("%.2f", offer.OffDiscountValue),
				"priority":       strconv.Itoa(offer.OffPriority),
				"auto_apply":     strconv.FormatBool(offer.OffAutoApply),
				"created_by":     "subscription_system",
				"auto_created":   "true",
				"business_unit":  "subscriptions",
				"campaign":       "auto_sync",
				"created_at":     time.Now().Format(time.RFC3339),
			}

			// Add date information if available
			if offer.OfferStartDate.Valid {
				promotionMetadata["active_from"] = offer.OfferStartDate.Time.Format(time.RFC3339)
			}
			if offer.OffEndDate.Valid {
				promotionMetadata["active_until"] = offer.OffEndDate.Time.Format(time.RFC3339)
			}

			stripePromotionCode, err := s.stripeService.CreatePromotionCode(
				stripeCouponID,
				offer.OffCode.String,
				maxRedemptions,
				promotionMetadata,
			)
			if err != nil {
				log.Printf("Warning: Failed to create Stripe promotion code: %v", err)
			} else {
				stripePromotionCodeID = stripePromotionCode.ID
				log.Printf("Successfully created Stripe promotion code: %s", stripePromotionCodeID)
			}
		}
	}

	// Set Stripe IDs on the offer if they were created
	if stripeCouponID != "" {
		offer.StripeCouponID = sql.NullString{String: stripeCouponID, Valid: true}
	}
	if stripePromotionCodeID != "" {
		offer.StripePromotionCodeID = sql.NullString{String: stripePromotionCodeID, Valid: true}
	}

	// Create the subscription offer
	response, err := s.SubscriptionOffersService.CreateSubscriptionOffer(ctx, &CreateSubscriptionOfferRequest{
		PlanID:             offer.PlanID,
		ItemID:             getStringPointer(offer.ItemID),
		OffDiscountType:    offer.OffDiscountType,
		OffDiscountValue:   offer.OffDiscountValue,
		OfferStartDate:     getTimeStringPointer(offer.OfferStartDate),
		OffEndDate:         getTimeStringPointer(offer.OffEndDate),
		IsActive:           offer.IsActive,
		OffDescription:     getStringPointer(offer.OffDescription),
		OffName:            offer.OffName,
		OffCode:            getStringPointer(offer.OffCode),
		OffMaxUses:         getIntPointer(offer.OffMaxUses),
		OffCurrentUses:     offer.OffCurrentUses,
		OffTermsConditions: getStringPointer(offer.OffTermsConditions),
		OffTarget:          getStringPointer(offer.OffTarget),
		OffPriority:        offer.OffPriority,
		OffAutoApply:       offer.OffAutoApply,
	})
	if err != nil {
		// If offer creation fails but Stripe entities were created, we should ideally clean them up
		// For now, we'll log the issue
		if stripeCouponID != "" {
			log.Printf("Warning: Offer creation failed but Stripe entities were created. Coupon: %s, Promotion Code: %s", stripeCouponID, stripePromotionCodeID)
		}
		return nil, err
	}

	// Add Stripe IDs to response if they were created
	if stripeCouponID != "" {
		log.Printf("Offer created with Stripe integration: Coupon=%s, Promotion Code=%s", stripeCouponID, stripePromotionCodeID)
	}

	return response, nil
}

// SyncOfferWithStripe syncs an existing offer with Stripe
func (s *SubscriptionOffersStripeService) SyncOfferWithStripe(ctx context.Context, offerID string) (*SubscriptionOfferResponse, error) {
	if s.stripeService == nil || !s.stripeService.IsEnabled() {
		return nil, fmt.Errorf("stripe service is not available")
	}

	// Get the existing offer
	offerIDInt, err := strconv.Atoi(offerID)
	if err != nil {
		return nil, fmt.Errorf("invalid offer ID: %w", err)
	}

	offer, err := s.GetSubscriptionOfferByID(ctx, offerIDInt)
	if err != nil {
		return nil, err
	}

	// Create Stripe coupon and promotion code if they don't exist
	if offer.StripeCouponID == nil || *offer.StripeCouponID == "" {
		// Validate offer data for Stripe
		var duration string
		var durationInMonths *int64

		// Map offer duration to Stripe duration
		duration = "once"

		// Validate coupon data
		if err := ValidateCouponData(offer.OffDiscountType, offer.OffDiscountValue, duration, durationInMonths, nil); err != nil {
			return nil, fmt.Errorf("invalid coupon configuration for offer '%s': %w", offer.OffName, err)
		}

		metadata := map[string]string{
			"offer_id":   offerID,
			"offer_type": "subscription_offer",
			"synced_at":  time.Now().Format(time.RFC3339),
		}

		// Create Stripe coupon
		var percentOff *float64
		var amountOff *int64
		var currency string

		if offer.OffDiscountType == "percentage" {
			percentOff = &offer.OffDiscountValue
		} else if offer.OffDiscountType == "amount" {
			amountOffValue := int64(offer.OffDiscountValue * 100) // Convert to cents
			amountOff = &amountOffValue
			currency = "usd" // Default currency, could be configurable
		}

		// Set max redemptions if specified
		var maxRedemptions *int64
		if offer.OffMaxUses != nil && *offer.OffMaxUses > 0 {
			maxRedVal := int64(*offer.OffMaxUses)
			maxRedemptions = &maxRedVal
		}

		// Calculate expiration timestamp from string date
		var expirationTimestamp *int64
		if offer.OffEndDate != nil && *offer.OffEndDate != "" {
			if parsedDate, err := time.Parse("2006-01-02", *offer.OffEndDate); err == nil {
				expirationTimestamp = calculateExpirationTimestamp(sql.NullTime{Time: parsedDate, Valid: true})
				if expirationTimestamp == nil {
					log.Printf("Warning: Offer end date '%s' is in the past or today, skipping expiration timestamp", *offer.OffEndDate)
				}
			}
		}

		stripeCoupon, err := s.stripeService.CreateCoupon(
			offer.OffName,
			percentOff,
			amountOff,
			currency,
			duration,
			durationInMonths,
			maxRedemptions,
			metadata,
			expirationTimestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create Stripe coupon: %w", err)
		}

		// Create promotion code if offer has a code
		var stripePromotionCodeID string
		if offer.OffCode != nil && *offer.OffCode != "" {
			promotionMetadata := map[string]string{
				"offer_id":       offerID,
				"offer_code":     *offer.OffCode,
				"plan_id":        strconv.Itoa(offer.PlanID),
				"discount_type":  offer.OffDiscountType,
				"discount_value": fmt.Sprintf("%.2f", offer.OffDiscountValue),
				"priority":       strconv.Itoa(offer.OffPriority),
				"auto_apply":     strconv.FormatBool(offer.OffAutoApply),
				"created_by":     "subscription_system",
				"synced_at":      time.Now().Format(time.RFC3339),
				"business_unit":  "subscriptions",
				"campaign":       "sync_operation",
			}

			// Add date information if available
			if offer.OffEndDate != nil && *offer.OffEndDate != "" {
				promotionMetadata["active_until"] = *offer.OffEndDate
			}

			stripePromotionCode, err := s.stripeService.CreatePromotionCode(
				stripeCoupon.ID,
				*offer.OffCode,
				maxRedemptions,
				promotionMetadata,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create Stripe promotion code: %w", err)
			}
			stripePromotionCodeID = stripePromotionCode.ID
		}

		// Update the offer with Stripe IDs
		updates := &UpdateSubscriptionOfferRequest{
			ID:             offerIDInt,
			StripeCouponID: &stripeCoupon.ID,
		}
		if stripePromotionCodeID != "" {
			updates.StripePromotionCodeID = &stripePromotionCodeID
		}

		return s.UpdateSubscriptionOffer(ctx, updates)
	}

	return offer, nil
}

// GetStripeIntegrationStatus returns the Stripe integration status for an offer
func (s *SubscriptionOffersStripeService) GetStripeIntegrationStatus(ctx context.Context, offerID string) (map[string]interface{}, error) {
	offerIDInt, err := strconv.Atoi(offerID)
	if err != nil {
		return nil, fmt.Errorf("invalid offer ID: %w", err)
	}

	offer, err := s.GetSubscriptionOfferByID(ctx, offerIDInt)
	if err != nil {
		return nil, err
	}

	status := map[string]interface{}{
		"has_stripe_coupon":         false,
		"has_stripe_promotion_code": false,
		"sync_status":               "not_synced",
	}

	if offer.StripeCouponID != nil && *offer.StripeCouponID != "" {
		status["has_stripe_coupon"] = true
		status["stripe_coupon_id"] = *offer.StripeCouponID
		status["sync_status"] = "synced"

		// Check if we also have a promotion code
		if offer.StripePromotionCodeID != nil && *offer.StripePromotionCodeID != "" {
			status["has_stripe_promotion_code"] = true
			status["stripe_promotion_code_id"] = *offer.StripePromotionCodeID
		}

		// If we have Stripe service, we could verify the coupon exists
		if s.stripeService != nil && s.stripeService.IsEnabled() {
			// For now, assume if we have IDs, they exist
			// In a production system, you might want to verify with Stripe
		}
	}

	return status, nil
}

// Helper functions
func getStringPointer(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func getTimeStringPointer(nt sql.NullTime) *string {
	if nt.Valid {
		dateStr := nt.Time.Format("2006-01-02")
		return &dateStr
	}
	return nil
}

func getIntPointer(ni sql.NullInt32) *int {
	if ni.Valid {
		val := int(ni.Int32)
		return &val
	}
	return nil
}

// Helper function to map offer duration to Stripe duration with business logic
func mapOfferDurationToStripe(offerStartDate, offerEndDate sql.NullTime) (string, *int64, *int64) {
	// If no dates specified, default to single use
	if !offerStartDate.Valid || !offerEndDate.Valid {
		return "once", nil, nil
	}

	// Calculate duration between start and end dates
	duration := offerEndDate.Time.Sub(offerStartDate.Time)
	days := int(duration.Hours() / 24)

	// Business logic for duration mapping
	switch {
	case days <= 1:
		// Same day or 24-hour offers = single use
		return "once", nil, nil
	case days <= 7:
		// Week-long offers = single use (short-term campaigns)
		return "once", nil, nil
	case days <= 30:
		// Month-long offers = repeating for 1 month
		months := int64(1)
		return "repeating", &months, nil
	case days <= 90:
		// Quarter-long offers = repeating for 3 months
		months := int64(3)
		return "repeating", &months, nil
	case days <= 180:
		// Half-year offers = repeating for 6 months
		months := int64(6)
		return "repeating", &months, nil
	case days <= 365:
		// Year-long offers = repeating for 12 months
		months := int64(12)
		return "repeating", &months, nil
	default:
		// Long-term offers = forever (ongoing promotions)
		return "forever", nil, nil
	}
}

// Helper function to calculate expiration timestamp
func calculateExpirationTimestamp(offerEndDate sql.NullTime) *int64 {
	if !offerEndDate.Valid {
		return nil
	}

	// Set expiration to end of the end date (23:59:59)
	endOfDay := time.Date(
		offerEndDate.Time.Year(),
		offerEndDate.Time.Month(),
		offerEndDate.Time.Day(),
		23, 59, 59, 0,
		offerEndDate.Time.Location(),
	)

	timestamp := endOfDay.Unix()

	// Validate that the timestamp is in the future
	now := time.Now().Unix()
	if timestamp <= now {
		// Return nil if the date is in the past or today
		// This will prevent Stripe from receiving an invalid redeem_by timestamp
		return nil
	}

	return &timestamp
}
