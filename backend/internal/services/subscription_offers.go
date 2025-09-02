package services

import (
	"context"
	"database/sql"
	"log"
	"time"

	"bome-backend/internal/database"

	"github.com/gin-gonic/gin"
)

// SubscriptionOfferResponse represents the response structure for subscription offers
type SubscriptionOfferResponse struct {
	ID                    int                      `json:"id"`
	PlanID                int                      `json:"plan_id"`
	ItemID                *string                  `json:"item_id,omitempty"`
	OffDiscountType       string                   `json:"off_discount_type"`
	OffDiscountValue      float64                  `json:"off_discount_value"`
	OfferStartDate        *string                  `json:"offer_start_date,omitempty"`
	OffEndDate            *string                  `json:"off_end_date,omitempty"`
	IsActive              bool                     `json:"is_active"`
	OffDescription        *string                  `json:"off_description,omitempty"`
	OffCreatedAt          string                   `json:"off_created_at"`
	OffUpdatedAt          string                   `json:"off_updated_at"`
	OffName               string                   `json:"off_name"`
	OffCode               *string                  `json:"off_code,omitempty"`
	OffMaxUses            *int                     `json:"off_max_uses,omitempty"`
	OffCurrentUses        int                      `json:"off_current_uses"`
	OffTermsConditions    *string                  `json:"off_terms_conditions,omitempty"`
	OffTarget             *string                  `json:"off_target,omitempty"`
	OffPriority           int                      `json:"off_priority"`
	OffAutoApply          bool                     `json:"off_auto_apply"`
	StripeCouponID        *string                  `json:"stripe_coupon_id,omitempty"`
	StripePromotionCodeID *string                  `json:"stripe_promotion_code_id,omitempty"`
	OfferHistory          []map[string]interface{} `json:"offer_history"`
}

// CreateSubscriptionOfferRequest represents the request structure for creating offers
type CreateSubscriptionOfferRequest struct {
	PlanID             int     `json:"plan_id" binding:"required"`
	ItemID             *string `json:"item_id"`
	OffDiscountType    string  `json:"off_discount_type" binding:"required"`
	OffDiscountValue   float64 `json:"off_discount_value" binding:"required"`
	OfferStartDate     *string `json:"offer_start_date"`
	OffEndDate         *string `json:"off_end_date"`
	IsActive           bool    `json:"is_active"`
	OffDescription     *string `json:"off_description"`
	OffName            string  `json:"off_name" binding:"required"`
	OffCode            *string `json:"off_code"`
	OffMaxUses         *int    `json:"off_max_uses"`
	OffCurrentUses     int     `json:"off_current_uses"`
	OffTermsConditions *string `json:"off_terms_conditions"`
	OffTarget          *string `json:"off_target"`
	OffPriority        int     `json:"off_priority"`
	OffAutoApply       bool    `json:"off_auto_apply"`
}

// UpdateSubscriptionOfferRequest represents the request structure for updating offers
type UpdateSubscriptionOfferRequest struct {
	ID                    int      `json:"id" binding:"required"`
	PlanID                *int     `json:"plan_id"`
	ItemID                *string  `json:"item_id"`
	OffDiscountType       *string  `json:"off_discount_type"`
	OffDiscountValue      *float64 `json:"off_discount_value"`
	OfferStartDate        *string  `json:"offer_start_date"`
	OffEndDate            *string  `json:"off_end_date"`
	IsActive              *bool    `json:"is_active"`
	OffDescription        *string  `json:"off_description"`
	OffName               *string  `json:"off_name"`
	OffCode               *string  `json:"off_code"`
	OffMaxUses            *int     `json:"off_max_uses"`
	OffCurrentUses        *int     `json:"off_current_uses"`
	OffTermsConditions    *string  `json:"off_terms_conditions"`
	OffTarget             *string  `json:"off_target"`
	OffPriority           *int     `json:"off_priority"`
	OffAutoApply          *bool    `json:"off_auto_apply"`
	StripeCouponID        *string  `json:"stripe_coupon_id"`
	StripePromotionCodeID *string  `json:"stripe_promotion_code_id"`
}

// SubscriptionOffersService handles business logic for subscription offers
type SubscriptionOffersService struct {
	db             *database.DB
	historyService *OfferHistoryService
}

// NewSubscriptionOffersService creates a new subscription offers service
func NewSubscriptionOffersService(db *database.DB) *SubscriptionOffersService {
	return &SubscriptionOffersService{
		db:             db,
		historyService: NewOfferHistoryService(db),
	}
}

// CreateSubscriptionOffer creates a new subscription offer
func (s *SubscriptionOffersService) CreateSubscriptionOffer(ctx context.Context, req *CreateSubscriptionOfferRequest) (*SubscriptionOfferResponse, error) {
	log.Printf("Service: Creating subscription offer: %s", req.OffName)

	// Convert request to database model
	offer := &database.SubscriptionOffer{
		PlanID:           req.PlanID,
		OffDiscountType:  req.OffDiscountType,
		OffDiscountValue: req.OffDiscountValue,
		IsActive:         req.IsActive,
		OffName:          req.OffName,
		OffCurrentUses:   req.OffCurrentUses,
		OffPriority:      req.OffPriority,
		OffAutoApply:     req.OffAutoApply,
	}

	// Handle optional fields
	if req.ItemID != nil {
		offer.ItemID = sql.NullString{String: *req.ItemID, Valid: true}
	}
	if req.OfferStartDate != nil {
		if parsedDate, err := time.Parse("2006-01-02", *req.OfferStartDate); err == nil {
			offer.OfferStartDate = sql.NullTime{Time: parsedDate, Valid: true}
		}
	}
	if req.OffEndDate != nil {
		if parsedDate, err := time.Parse("2006-01-02", *req.OffEndDate); err == nil {
			offer.OffEndDate = sql.NullTime{Time: parsedDate, Valid: true}
		}
	}
	if req.OffDescription != nil {
		offer.OffDescription = sql.NullString{String: *req.OffDescription, Valid: true}
	}
	if req.OffCode != nil {
		offer.OffCode = sql.NullString{String: *req.OffCode, Valid: true}
	}
	if req.OffMaxUses != nil {
		offer.OffMaxUses = sql.NullInt32{Int32: int32(*req.OffMaxUses), Valid: true}
	}
	if req.OffTermsConditions != nil {
		offer.OffTermsConditions = sql.NullString{String: *req.OffTermsConditions, Valid: true}
	}
	if req.OffTarget != nil {
		offer.OffTarget = sql.NullString{String: *req.OffTarget, Valid: true}
	}

	// Create the offer in database
	err := s.db.CreateSubscriptionOffer(offer)
	if err != nil {
		log.Printf("Service: Error creating subscription offer: %v", err)
		return nil, err
	}

	// Add history event for offer creation
	userDataJSON := s.getUserDataFromContext(ctx)
	err = s.historyService.AddHistoryEvent(
		offer.ID,
		"offer_created",
		"Offer created",
		userDataJSON,
		nil, // No old offer for creation
		map[string]interface{}{
			"action":     "created",
			"offer_name": offer.OffName,
		},
	)
	if err != nil {
		log.Printf("Service: Error adding creation history event: %v", err)
		// Don't fail the creation if history logging fails
	}

	// Convert to response
	response := s.convertToResponse(offer)
	log.Printf("Service: Successfully created subscription offer with ID: %d", offer.ID)
	return response, nil
}

// GetAllSubscriptionOffers retrieves all subscription offers
func (s *SubscriptionOffersService) GetAllSubscriptionOffers(ctx context.Context) ([]*SubscriptionOfferResponse, error) {
	log.Printf("Service: Getting all subscription offers")

	offers, err := s.db.GetAllSubscriptionOffers()
	if err != nil {
		log.Printf("Service: Error getting subscription offers: %v", err)
		return []*SubscriptionOfferResponse{}, err // Return empty array instead of nil
	}

	var responses []*SubscriptionOfferResponse
	for _, offer := range offers {
		response := s.convertToResponse(offer)
		responses = append(responses, response)
	}

	// Ensure we never return nil, always return an array (even if empty)
	if responses == nil {
		responses = []*SubscriptionOfferResponse{}
	}

	log.Printf("Service: Retrieved %d subscription offers", len(responses))
	return responses, nil
}

// GetSubscriptionOfferByID retrieves a specific subscription offer by ID
func (s *SubscriptionOffersService) GetSubscriptionOfferByID(ctx context.Context, id int) (*SubscriptionOfferResponse, error) {
	log.Printf("Service: Getting subscription offer by ID: %d", id)

	offer, err := s.db.GetSubscriptionOfferByID(id)
	if err != nil {
		log.Printf("Service: Error getting subscription offer by ID: %v", err)
		return nil, err
	}

	response := s.convertToResponse(offer)
	log.Printf("Service: Successfully retrieved subscription offer: %s", offer.OffName)
	return response, nil
}

// UpdateSubscriptionOffer updates an existing subscription offer
func (s *SubscriptionOffersService) UpdateSubscriptionOffer(ctx context.Context, req *UpdateSubscriptionOfferRequest) (*SubscriptionOfferResponse, error) {
	log.Printf("Service: Updating subscription offer ID: %d", req.ID)

	// Get current offer for history comparison
	currentOffer, err := s.db.GetSubscriptionOfferByID(req.ID)
	if err != nil {
		log.Printf("Service: Error getting current offer: %v", err)
		return nil, err
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.PlanID != nil {
		updates["plan_id"] = *req.PlanID
	}
	if req.ItemID != nil {
		updates["item_id"] = *req.ItemID
	}
	if req.OffDiscountType != nil {
		updates["off_discount_type"] = *req.OffDiscountType
	}
	if req.OffDiscountValue != nil {
		updates["off_discount_value"] = *req.OffDiscountValue
	}
	if req.OfferStartDate != nil {
		if parsedDate, err := time.Parse("2006-01-02", *req.OfferStartDate); err == nil {
			updates["offer_start_date"] = parsedDate
		}
	}
	if req.OffEndDate != nil {
		if parsedDate, err := time.Parse("2006-01-02", *req.OffEndDate); err == nil {
			updates["off_end_date"] = parsedDate
		}
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.OffDescription != nil {
		updates["off_description"] = *req.OffDescription
	}
	if req.OffName != nil {
		updates["off_name"] = *req.OffName
	}
	if req.OffCode != nil {
		updates["off_code"] = *req.OffCode
	}
	if req.OffMaxUses != nil {
		updates["off_max_uses"] = *req.OffMaxUses
	}
	if req.OffCurrentUses != nil {
		updates["off_current_uses"] = *req.OffCurrentUses
	}
	if req.OffTermsConditions != nil {
		updates["off_terms_conditions"] = *req.OffTermsConditions
	}
	if req.OffTarget != nil {
		updates["off_target"] = *req.OffTarget
	}
	if req.OffPriority != nil {
		updates["off_priority"] = *req.OffPriority
	}
	if req.OffAutoApply != nil {
		updates["off_auto_apply"] = *req.OffAutoApply
	}
	if req.StripeCouponID != nil {
		updates["stripe_coupon_id"] = *req.StripeCouponID
	}
	if req.StripePromotionCodeID != nil {
		updates["stripe_promotion_code_id"] = *req.StripePromotionCodeID
	}

	// Update the offer in database
	updatedOffer, err := s.db.UpdateSubscriptionOffer(req.ID, updates)
	if err != nil {
		log.Printf("Service: Error updating subscription offer: %v", err)
		return nil, err
	}

	// Add history event for offer update
	userDataJSON := s.getUserDataFromContext(ctx)
	err = s.historyService.AddHistoryEvent(
		req.ID,
		"offer_updated",
		"Offer updated",
		userDataJSON,
		currentOffer,
		map[string]interface{}{
			"action":     "updated",
			"offer_name": updatedOffer.OffName,
		},
	)
	if err != nil {
		log.Printf("Service: Error adding update history event: %v", err)
		// Don't fail the update if history logging fails
	}

	// Convert to response
	response := s.convertToResponse(updatedOffer)
	log.Printf("Service: Successfully updated subscription offer with ID: %d", req.ID)
	return response, nil
}

// DeleteSubscriptionOffer deletes a subscription offer
func (s *SubscriptionOffersService) DeleteSubscriptionOffer(ctx context.Context, id int) error {
	log.Printf("Service: Deleting subscription offer ID: %d", id)

	// Get current offer for history
	currentOffer, err := s.db.GetSubscriptionOfferByID(id)
	if err != nil {
		log.Printf("Service: Error getting current offer: %v", err)
		return err
	}

	// Delete the offer from database
	err = s.db.DeleteSubscriptionOffer(id)
	if err != nil {
		log.Printf("Service: Error deleting subscription offer: %v", err)
		return err
	}

	// Add history event for offer deletion
	userDataJSON := s.getUserDataFromContext(ctx)
	err = s.historyService.AddHistoryEvent(
		id,
		"offer_deleted",
		"Offer deleted",
		userDataJSON,
		currentOffer,
		map[string]interface{}{
			"action":     "deleted",
			"offer_name": currentOffer.OffName,
		},
	)
	if err != nil {
		log.Printf("Service: Error adding deletion history event: %v", err)
		// Don't fail the deletion if history logging fails
	}

	log.Printf("Service: Successfully deleted subscription offer with ID: %d", id)
	return nil
}

// GetOfferHistory retrieves history events for a specific offer
func (s *SubscriptionOffersService) GetOfferHistory(ctx context.Context, offerID int) ([]map[string]interface{}, error) {
	log.Printf("Service: Getting history for offer %d", offerID)

	events, err := s.historyService.GetOfferHistory(offerID)
	if err != nil {
		log.Printf("Service: Error getting offer history: %v", err)
		return nil, err
	}

	// Convert to response format
	response := s.convertHistoryEvents(events)
	log.Printf("Service: Retrieved %d history events for offer %d", len(response), offerID)
	return response, nil
}

// convertToResponse converts a database offer to a response
func (s *SubscriptionOffersService) convertToResponse(offer *database.SubscriptionOffer) *SubscriptionOfferResponse {
	log.Printf("Service: Converting offer %d to response", offer.ID)

	response := &SubscriptionOfferResponse{
		ID:               offer.ID,
		PlanID:           offer.PlanID,
		OffDiscountType:  offer.OffDiscountType,
		OffDiscountValue: offer.OffDiscountValue,
		IsActive:         offer.IsActive,
		OffCreatedAt:     offer.OffCreatedAt.Format(time.RFC3339),
		OffUpdatedAt:     offer.OffUpdatedAt.Format(time.RFC3339),
		OffName:          offer.OffName,
		OffCurrentUses:   offer.OffCurrentUses,
		OffPriority:      offer.OffPriority,
		OffAutoApply:     offer.OffAutoApply,
	}

	// Handle nullable fields
	if offer.ItemID.Valid {
		response.ItemID = &offer.ItemID.String
	}
	if offer.OfferStartDate.Valid {
		startDate := offer.OfferStartDate.Time.Format("2006-01-02")
		response.OfferStartDate = &startDate
	}
	if offer.OffEndDate.Valid {
		endDate := offer.OffEndDate.Time.Format("2006-01-02")
		response.OffEndDate = &endDate
	}
	if offer.OffDescription.Valid {
		response.OffDescription = &offer.OffDescription.String
	}
	if offer.OffCode.Valid {
		response.OffCode = &offer.OffCode.String
	}
	if offer.OffMaxUses.Valid {
		maxUses := int(offer.OffMaxUses.Int32)
		response.OffMaxUses = &maxUses
	}
	if offer.OffTermsConditions.Valid {
		response.OffTermsConditions = &offer.OffTermsConditions.String
	}
	if offer.OffTarget.Valid {
		response.OffTarget = &offer.OffTarget.String
	}

	// Handle Stripe fields
	if offer.StripeCouponID.Valid {
		response.StripeCouponID = &offer.StripeCouponID.String
	}
	if offer.StripePromotionCodeID.Valid {
		response.StripePromotionCodeID = &offer.StripePromotionCodeID.String
	}

	// Get history events
	log.Printf("Service: Fetching history from separate table for offer %d", offer.ID)
	historyEvents, err := s.historyService.GetOfferHistory(offer.ID)
	if err != nil {
		log.Printf("Service: Error getting offer history: %v", err)
		// Continue without history if there's an error
		response.OfferHistory = []map[string]interface{}{}
	} else {
		response.OfferHistory = s.convertHistoryEvents(historyEvents)
	}
	log.Printf("Service: Successfully converted %d history events for offer %d", len(response.OfferHistory), offer.ID)

	log.Printf("Service: Final response for offer %d - OfferHistory length: %d", offer.ID, len(response.OfferHistory))
	return response
}

// convertHistoryEvents converts database history events to response format
func (s *SubscriptionOffersService) convertHistoryEvents(events []*database.SubscriptionOfferHistory) []map[string]interface{} {
	var converted []map[string]interface{}

	for _, event := range events {
		convertedEvent := map[string]interface{}{
			"id":         event.ID,
			"offer_id":   event.OfferID,
			"accepted":   event.Accepted,
			"event_type": event.EventType,
			"created_at": event.CreatedAt.Format(time.RFC3339),
		}

		// Handle nullable fields
		if event.UserID.Valid {
			convertedEvent["user_id"] = event.UserID.String
		}
		if event.SubPlanID.Valid {
			convertedEvent["sub_plan_id"] = event.SubPlanID.Int32
		}
		if event.OffUserIP.Valid {
			convertedEvent["off_user_ip"] = event.OffUserIP.String
		}
		if event.DeviceInfo.Valid {
			convertedEvent["device_info"] = event.DeviceInfo.String
		}
		if event.DiscountAmount.Valid {
			convertedEvent["discount_amount"] = event.DiscountAmount.Float64
		}
		if event.OriginalPrice.Valid {
			convertedEvent["original_price"] = event.OriginalPrice.Float64
		}
		if event.FinalPrice.Valid {
			convertedEvent["final_price"] = event.FinalPrice.Float64
		}
		if event.SessionID.Valid {
			convertedEvent["session_id"] = event.SessionID.String
		}
		if event.ReferrerURL.Valid {
			convertedEvent["referrer_url"] = event.ReferrerURL.String
		}
		if event.UserAgent.Valid {
			convertedEvent["user_agent"] = event.UserAgent.String
		}

		converted = append(converted, convertedEvent)
	}

	return converted
}

// getUserDataFromContext extracts user data from context
func (s *SubscriptionOffersService) getUserDataFromContext(ctx context.Context) string {
	log.Printf("Service: Starting user data extraction")

	// Try to get frontend user data first
	if frontendUserData, exists := ctx.Value("frontend_user_data").(string); exists && frontendUserData != "" {
		log.Printf("Service: Using frontend user data: %s", frontendUserData)
		return frontendUserData
	}

	// Try to get user data from gin context if available
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if userData := ginCtx.GetHeader("X-User-Data"); userData != "" {
			log.Printf("Service: Using X-User-Data header: %s", userData)
			return userData
		}
	}

	// Try to get from context directly
	if userData, exists := ctx.Value("X-User-Data").(string); exists && userData != "" {
		log.Printf("Service: Using X-User-Data from context: %s", userData)
		return userData
	}

	// Fallback to system user
	systemUserData := `{"id":"system","email":"🖥️System","role":"system","first_name":"🖥️","last_name":"System"}`
	log.Printf("Service: Using system user data: %s", systemUserData)
	return systemUserData
}
