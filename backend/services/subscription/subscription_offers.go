package subscription

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"bome-backend/infrastructure/database"
)

// SubscriptionOffersService handles business logic for subscription offers
type SubscriptionOffersService struct {
	db  *database.DB
	hub WebSocketHub
}

// SubscriptionOffer represents a promotional offer for subscription plans
type SubscriptionOffer struct {
	ID                    int        `json:"id"`
	PlanID                int        `json:"plan_id"`
	PlanName              string     `json:"plan_name,omitempty"`
	ItemID                *string    `json:"item_id,omitempty"`
	OffDiscountType       string     `json:"off_discount_type"`  // "percentage" or "fixed"
	OffDiscountValue      float64    `json:"off_discount_value"` // e.g., 20 for 20% or $20
	OfferStartDate        *time.Time `json:"offer_start_date,omitempty"`
	OffEndDate            *time.Time `json:"off_end_date,omitempty"`
	IsActive              bool       `json:"is_active"`
	OffDescription        *string    `json:"off_description,omitempty"`
	OffName               string     `json:"off_name"`
	OffCode               *string    `json:"off_code,omitempty"`
	OffMaxUses            *int       `json:"off_max_uses,omitempty"`
	OffCurrentUses        int        `json:"off_current_uses"`
	OffTermsConditions    *string    `json:"off_terms_conditions,omitempty"`
	OffTarget             *string    `json:"off_target,omitempty"` // "all", "new", "existing"
	OffPriority           int        `json:"off_priority"`         // Higher = shown first
	OffAutoApply          bool       `json:"off_auto_apply"`       // Auto-apply at checkout
	StripeCouponID        *string    `json:"stripe_coupon_id,omitempty"`
	StripePromotionCodeID *string    `json:"stripe_promotion_code_id,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// CreateSubscriptionOfferRequest represents a request to create an offer
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
	OffTermsConditions *string `json:"off_terms_conditions"`
	OffTarget          *string `json:"off_target"`
	OffPriority        int     `json:"off_priority"`
	OffAutoApply       bool    `json:"off_auto_apply"`
}

// UpdateSubscriptionOfferRequest represents a request to update an offer
type UpdateSubscriptionOfferRequest struct {
	PlanID             *int     `json:"plan_id"`
	ItemID             *string  `json:"item_id"`
	OffDiscountType    *string  `json:"off_discount_type"`
	OffDiscountValue   *float64 `json:"off_discount_value"`
	OfferStartDate     *string  `json:"offer_start_date"`
	OffEndDate         *string  `json:"off_end_date"`
	IsActive           *bool    `json:"is_active"`
	OffDescription     *string  `json:"off_description"`
	OffName            *string  `json:"off_name"`
	OffCode            *string  `json:"off_code"`
	OffMaxUses         *int     `json:"off_max_uses"`
	OffCurrentUses     *int     `json:"off_current_uses"`
	OffTermsConditions *string  `json:"off_terms_conditions"`
	OffTarget          *string  `json:"off_target"`
	OffPriority        *int     `json:"off_priority"`
	OffAutoApply       *bool    `json:"off_auto_apply"`
}

// NewSubscriptionOffersService creates a new subscription offers service
func NewSubscriptionOffersService(db *database.DB, hub WebSocketHub) *SubscriptionOffersService {
	return &SubscriptionOffersService{
		db:  db,
		hub: hub,
	}
}

// GetAllOffers retrieves all subscription offers
func (s *SubscriptionOffersService) GetAllOffers() ([]*SubscriptionOffer, error) {
	query := `
		SELECT 
			so.id, so.plan_id, so.item_id, so.off_discount_type, so.off_discount_value,
			so.offer_start_date, so.off_end_date, so.is_active, so.off_description,
			so.off_name, so.off_code, so.off_max_uses, so.off_current_uses,
			so.off_terms_conditions, so.off_target, so.off_priority, so.off_auto_apply,
			so.stripe_coupon_id, so.stripe_promotion_code_id, so.off_created_at, so.off_updated_at,
			sp.name as plan_name
		FROM subscription_offers so
		LEFT JOIN subscription_plans sp ON so.plan_id = sp.id
		ORDER BY so.off_priority DESC, so.off_created_at DESC
	`

	rows, err := s.db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offers []*SubscriptionOffer
	for rows.Next() {
		offer := &SubscriptionOffer{}
		var planName sql.NullString

		err := rows.Scan(
			&offer.ID, &offer.PlanID, &offer.ItemID, &offer.OffDiscountType, &offer.OffDiscountValue,
			&offer.OfferStartDate, &offer.OffEndDate, &offer.IsActive, &offer.OffDescription,
			&offer.OffName, &offer.OffCode, &offer.OffMaxUses, &offer.OffCurrentUses,
			&offer.OffTermsConditions, &offer.OffTarget, &offer.OffPriority, &offer.OffAutoApply,
			&offer.StripeCouponID, &offer.StripePromotionCodeID, &offer.CreatedAt, &offer.UpdatedAt,
			&planName,
		)
		if err != nil {
			log.Printf("Error scanning offer: %v", err)
			continue
		}

		if planName.Valid {
			offer.PlanName = planName.String
		}

		offers = append(offers, offer)
	}

	return offers, nil
}

// GetOfferByID retrieves a subscription offer by ID
func (s *SubscriptionOffersService) GetOfferByID(id int) (*SubscriptionOffer, error) {
	query := `
		SELECT 
			so.id, so.plan_id, so.item_id, so.off_discount_type, so.off_discount_value,
			so.offer_start_date, so.off_end_date, so.is_active, so.off_description,
			so.off_name, so.off_code, so.off_max_uses, so.off_current_uses,
			so.off_terms_conditions, so.off_target, so.off_priority, so.off_auto_apply,
			so.stripe_coupon_id, so.stripe_promotion_code_id, so.off_created_at, so.off_updated_at,
			sp.name as plan_name
		FROM subscription_offers so
		LEFT JOIN subscription_plans sp ON so.plan_id = sp.id
		WHERE so.id = $1
	`

	offer := &SubscriptionOffer{}
	var planName sql.NullString

	err := s.db.DB.QueryRow(query, id).Scan(
		&offer.ID, &offer.PlanID, &offer.ItemID, &offer.OffDiscountType, &offer.OffDiscountValue,
		&offer.OfferStartDate, &offer.OffEndDate, &offer.IsActive, &offer.OffDescription,
		&offer.OffName, &offer.OffCode, &offer.OffMaxUses, &offer.OffCurrentUses,
		&offer.OffTermsConditions, &offer.OffTarget, &offer.OffPriority, &offer.OffAutoApply,
		&offer.StripeCouponID, &offer.StripePromotionCodeID, &offer.CreatedAt, &offer.UpdatedAt,
		&planName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("offer not found")
		}
		return nil, err
	}

	if planName.Valid {
		offer.PlanName = planName.String
	}

	return offer, nil
}

// CreateOffer creates a new subscription offer
func (s *SubscriptionOffersService) CreateOffer(ctx context.Context, req *CreateSubscriptionOfferRequest) (*SubscriptionOffer, error) {
	// Parse dates if provided
	var offerStartDate, offEndDate *time.Time
	if req.OfferStartDate != nil {
		if date, err := time.Parse("2006-01-02", *req.OfferStartDate); err == nil {
			offerStartDate = &date
		}
	}
	if req.OffEndDate != nil {
		if date, err := time.Parse("2006-01-02", *req.OffEndDate); err == nil {
			offEndDate = &date
		}
	}

	query := `
		INSERT INTO subscription_offers 
		(plan_id, item_id, off_discount_type, off_discount_value, offer_start_date, off_end_date,
		 is_active, off_description, off_name, off_code, off_max_uses, off_current_uses,
		 off_terms_conditions, off_target, off_priority, off_auto_apply, off_created_at, off_updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW())
		RETURNING id, off_created_at, off_updated_at
	`

	var id int
	var createdAt, updatedAt time.Time
	err := s.db.DB.QueryRowContext(ctx, query,
		req.PlanID, req.ItemID, req.OffDiscountType, req.OffDiscountValue,
		offerStartDate, offEndDate, req.IsActive, req.OffDescription,
		req.OffName, req.OffCode, req.OffMaxUses, 0, // off_current_uses starts at 0
		req.OffTermsConditions, req.OffTarget, req.OffPriority, req.OffAutoApply,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create offer: %w", err)
	}

	// Get created offer
	offer, err := s.GetOfferByID(id)
	if err != nil {
		return nil, err
	}

	// Broadcast event via WebSocket
	if s.hub != nil {
		s.hub.BroadcastEvent("offer.created", map[string]interface{}{
			"offer": offer,
		}, fmt.Sprintf("New offer created: %s", offer.OffName))
		log.Printf("📡 Broadcasted offer creation: %s", offer.OffName)
	}

	return offer, nil
}

// UpdateOffer updates a subscription offer
func (s *SubscriptionOffersService) UpdateOffer(ctx context.Context, id int, req *UpdateSubscriptionOfferRequest) (*SubscriptionOffer, error) {
	// Build dynamic update query based on provided fields
	updates := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.PlanID != nil {
		updates = append(updates, fmt.Sprintf("plan_id = $%d", argIdx))
		args = append(args, *req.PlanID)
		argIdx++
	}
	if req.OffDiscountType != nil {
		updates = append(updates, fmt.Sprintf("off_discount_type = $%d", argIdx))
		args = append(args, *req.OffDiscountType)
		argIdx++
	}
	if req.OffDiscountValue != nil {
		updates = append(updates, fmt.Sprintf("off_discount_value = $%d", argIdx))
		args = append(args, *req.OffDiscountValue)
		argIdx++
	}
	if req.IsActive != nil {
		updates = append(updates, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *req.IsActive)
		argIdx++
	}
	if req.OffName != nil {
		updates = append(updates, fmt.Sprintf("off_name = $%d", argIdx))
		args = append(args, *req.OffName)
		argIdx++
	}
	if req.OffPriority != nil {
		updates = append(updates, fmt.Sprintf("off_priority = $%d", argIdx))
		args = append(args, *req.OffPriority)
		argIdx++
	}
	if req.OffAutoApply != nil {
		updates = append(updates, fmt.Sprintf("off_auto_apply = $%d", argIdx))
		args = append(args, *req.OffAutoApply)
		argIdx++
	}

	if len(updates) == 0 {
		return s.GetOfferByID(id)
	}

	updates = append(updates, "off_updated_at = NOW()")
	query := fmt.Sprintf("UPDATE subscription_offers SET %s WHERE id = $%d",
		fmt.Sprintf("%s", updates[0]), argIdx)
	for i := 1; i < len(updates); i++ {
		query = fmt.Sprintf("%s, %s", query, updates[i])
	}
	args = append(args, id)

	_, err := s.db.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update offer: %w", err)
	}

	// Get updated offer
	offer, err := s.GetOfferByID(id)
	if err != nil {
		return nil, err
	}

	// Broadcast event via WebSocket
	if s.hub != nil {
		s.hub.BroadcastEvent("offer.updated", map[string]interface{}{
			"offer": offer,
		}, fmt.Sprintf("Offer updated: %s", offer.OffName))
		log.Printf("📡 Broadcasted offer update: %s", offer.OffName)
	}

	return offer, nil
}

// DeleteOffer deletes a subscription offer
func (s *SubscriptionOffersService) DeleteOffer(ctx context.Context, id int) error {
	query := `DELETE FROM subscription_offers WHERE id = $1`

	result, err := s.db.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete offer: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("offer not found")
	}

	// Broadcast event via WebSocket
	if s.hub != nil {
		s.hub.BroadcastEvent("offer.deleted", map[string]interface{}{
			"offer_id": id,
		}, "Offer deleted")
		log.Printf("📡 Broadcasted offer deletion: ID %d", id)
	}

	return nil
}

// GetActiveOffersByPlan retrieves all active offers for a specific plan
func (s *SubscriptionOffersService) GetActiveOffersByPlan(planID int) ([]*SubscriptionOffer, error) {
	query := `
		SELECT 
			so.id, so.plan_id, so.item_id, so.off_discount_type, so.off_discount_value,
			so.offer_start_date, so.off_end_date, so.is_active, so.off_description,
			so.off_name, so.off_code, so.off_max_uses, so.off_current_uses,
			so.off_terms_conditions, so.off_target, so.off_priority, so.off_auto_apply,
			so.stripe_coupon_id, so.stripe_promotion_code_id, so.off_created_at, so.off_updated_at,
			sp.name as plan_name
		FROM subscription_offers so
		LEFT JOIN subscription_plans sp ON so.plan_id = sp.id
		WHERE so.plan_id = $1 AND so.is_active = true
		  AND (so.offer_start_date IS NULL OR so.offer_start_date <= NOW())
		  AND (so.off_end_date IS NULL OR so.off_end_date >= NOW())
		ORDER BY so.off_priority DESC, so.off_created_at DESC
	`

	rows, err := s.db.DB.Query(query, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offers []*SubscriptionOffer
	for rows.Next() {
		offer := &SubscriptionOffer{}
		var planName sql.NullString

		err := rows.Scan(
			&offer.ID, &offer.PlanID, &offer.ItemID, &offer.OffDiscountType, &offer.OffDiscountValue,
			&offer.OfferStartDate, &offer.OffEndDate, &offer.IsActive, &offer.OffDescription,
			&offer.OffName, &offer.OffCode, &offer.OffMaxUses, &offer.OffCurrentUses,
			&offer.OffTermsConditions, &offer.OffTarget, &offer.OffPriority, &offer.OffAutoApply,
			&offer.StripeCouponID, &offer.StripePromotionCodeID, &offer.CreatedAt, &offer.UpdatedAt,
			&planName,
		)
		if err != nil {
			log.Printf("Error scanning offer: %v", err)
			continue
		}

		if planName.Valid {
			offer.PlanName = planName.String
		}

		offers = append(offers, offer)
	}

	return offers, nil
}
