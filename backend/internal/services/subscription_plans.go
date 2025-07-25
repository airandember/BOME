package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"bome-backend/internal/database"
)

// SubscriptionPlanService handles business logic for subscription plans
type SubscriptionPlanService struct {
	db *database.DB
}

// SubscriptionPlanResponse represents a subscription plan response
type SubscriptionPlanResponse struct {
	ID                 string     `json:"id"`
	Name               string     `json:"name"`
	Description        string     `json:"description"`
	ShortDesc          string     `json:"short_desc"`
	Price              float64    `json:"price"`
	Currency           string     `json:"currency"`
	Interval           string     `json:"interval"`
	IntervalCount      int        `json:"interval_count"`
	StripePriceID      *string    `json:"stripe_price_id,omitempty"`
	Features           []string   `json:"features"`
	IsActive           bool       `json:"is_active"`
	IsPromoted         bool       `json:"is_promoted"`
	PromotionEndDate   *time.Time `json:"promotion_end_date,omitempty"`
	PromotionStartDate *time.Time `json:"promotion_start_date,omitempty"`
	PromotionHistory   []string   `json:"promotion_history,omitempty"`
	SubType            int        `json:"sub_type"` // 100 = standard, 300 = promotional
	SortOrder          int        `json:"sort_order"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// CreateSubscriptionPlanRequest represents a request to create a subscription plan
type CreateSubscriptionPlanRequest struct {
	Name             string     `json:"name" validate:"required,min=1,max=255"`
	Description      string     `json:"description"`
	ShortDesc        string     `json:"short_desc"`
	Price            float64    `json:"price" validate:"required,min=0"`
	Currency         string     `json:"currency" validate:"required,oneof=USD EUR GBP CAD"`
	Interval         string     `json:"interval" validate:"required,oneof=month year week day"`
	IntervalCount    int        `json:"interval_count" validate:"required,min=1"`
	StripePriceID    string     `json:"stripe_price_id"`
	Features         []string   `json:"features"`
	IsActive         bool       `json:"is_active"`
	IsPromoted       bool       `json:"is_promoted"`
	PromotionEndDate *time.Time `json:"promotion_end_date"`
	SortOrder        int        `json:"sort_order"`
}

// UpdateSubscriptionPlanRequest represents a request to update a subscription plan
type UpdateSubscriptionPlanRequest struct {
	Name             string     `json:"name" validate:"required,min=1,max=255"`
	Description      string     `json:"description"`
	ShortDesc        string     `json:"short_desc"`
	Price            float64    `json:"price" validate:"required,min=0"`
	Currency         string     `json:"currency" validate:"required,oneof=USD EUR GBP CAD"`
	Interval         string     `json:"interval" validate:"required,oneof=month year week day"`
	IntervalCount    int        `json:"interval_count" validate:"required,min=1"`
	StripePriceID    string     `json:"stripe_price_id"`
	Features         []string   `json:"features"`
	IsActive         bool       `json:"is_active"`
	IsPromoted       bool       `json:"is_promoted"`
	PromotionEndDate *time.Time `json:"promotion_end_date"`
	SortOrder        int        `json:"sort_order"`
}

// NewSubscriptionPlanService creates a new subscription plan service
func NewSubscriptionPlanService(db *database.DB) *SubscriptionPlanService {
	return &SubscriptionPlanService{db: db}
}

// CreateSubscriptionPlan creates a new subscription plan
func (s *SubscriptionPlanService) CreateSubscriptionPlan(ctx context.Context, plan *database.SubscriptionPlan) (*SubscriptionPlanResponse, error) {
	log.Printf("Service: Creating plan: %+v", plan)
	created, err := s.db.CreateSubscriptionPlan(plan)
	if err != nil {
		return nil, err
	}
	return s.convertToResponse(created), nil
}

// GetSubscriptionPlan gets a subscription plan by ID
func (s *SubscriptionPlanService) GetSubscriptionPlan(ctx context.Context, id string) (*SubscriptionPlanResponse, error) {
	planID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid plan ID: %w", err)
	}

	plan, err := s.db.GetSubscriptionPlanByID(planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription plan: %w", err)
	}

	return s.convertToResponse(plan), nil
}

// UpdateSubscriptionPlan updates a subscription plan by ID
func (s *SubscriptionPlanService) UpdateSubscriptionPlan(ctx context.Context, id string, updates map[string]interface{}) (*SubscriptionPlanResponse, error) {
	planID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid plan ID: %w", err)
	}
	log.Printf("Service: Updating plan %d with: %+v", planID, updates)
	updated, err := s.db.UpdateSubscriptionPlan(planID, updates)
	if err != nil {
		return nil, err
	}
	return s.convertToResponse(updated), nil
}

// SoftDeleteSubscriptionPlan marks a plan as deleted
func (s *SubscriptionPlanService) SoftDeleteSubscriptionPlan(ctx context.Context, id string) error {
	planID, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid plan ID: %w", err)
	}
	log.Printf("Service: Soft deleting plan %d", planID)
	return s.db.SoftDeleteSubscriptionPlan(planID)
}

// ToggleSubscriptionPlanStatus toggles the active status of a subscription plan
func (s *SubscriptionPlanService) ToggleSubscriptionPlanStatus(ctx context.Context, id string, isActive bool) (*SubscriptionPlanResponse, error) {
	planID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid plan ID: %w", err)
	}

	// Check if plan exists
	_, err = s.db.GetSubscriptionPlanByID(planID)
	if err != nil {
		return nil, fmt.Errorf("subscription plan not found: %w", err)
	}

	// Update status
	updates := map[string]interface{}{
		"is_active": isActive,
	}
	_, err = s.db.UpdateSubscriptionPlan(planID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to toggle subscription plan status: %w", err)
	}

	// Get updated plan
	updatedPlan, err := s.db.GetSubscriptionPlanByID(planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated subscription plan: %w", err)
	}

	return s.convertToResponse(updatedPlan), nil
}

// UpdatePromotionStatus updates the promotion status of a subscription plan
// If isPromoted is true, sets sub_type to 300 (promo), sets promotion_start_date to now, and optionally sets promotion_end_date.
// If isPromoted is false, sets sub_type to 100 (standard), sets is_active to false, and sets promotion_end_date to now.
func (s *SubscriptionPlanService) UpdatePromotionStatus(ctx context.Context, id string, isPromoted bool, promotionEndDate *time.Time) (*SubscriptionPlanResponse, error) {
	planID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid plan ID: %w", err)
	}
	log.Printf("UpdatePromotionStatus called for plan %d, isPromoted: %v", planID, isPromoted)
	_, err = s.db.GetSubscriptionPlanByID(planID)
	if err != nil {
		return nil, err
	}
	updates := map[string]interface{}{}
	if isPromoted {
		updates["is_promoted"] = true
		updates["sub_type"] = 300
		updates["promotion_start_date"] = time.Now()
		if promotionEndDate != nil {
			updates["promotion_end_date"] = *promotionEndDate
		}
		updates["is_active"] = true
	} else {
		updates["is_promoted"] = false
		updates["sub_type"] = 100
		updates["is_active"] = false
		updates["promotion_end_date"] = time.Now()
	}
	updatedPlan, err := s.db.UpdateSubscriptionPlan(planID, updates)
	if err != nil {
		return nil, err
	}
	return s.convertToResponse(updatedPlan), nil
}

// GetAllSubscriptionPlans gets all subscription plans without pagination or filters
func (s *SubscriptionPlanService) GetAllSubscriptionPlans(ctx context.Context) ([]*SubscriptionPlanResponse, error) {
	// Check and handle expired promotions first
	err := s.CheckAndHandleExpiredPromotions(ctx)
	if err != nil {
		log.Printf("Warning: Failed to check expired promotions: %v", err)
	}

	plans, err := s.db.GetAllSubscriptionPlans()
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription plans: %w", err)
	}

	// Debug: Log the raw plans from database
	log.Printf("Service: Raw plans from database: %+v", plans)
	for i, plan := range plans {
		log.Printf("Service: Plan %d: ID=%d, Name=%s, IsActive=%v, IsPromoted=%v, Interval=%s",
			i, plan.ID, plan.Name, plan.IsActive, plan.IsPromoted, plan.Interval)
	}

	// Convert to response format
	responsePlans := make([]*SubscriptionPlanResponse, len(plans))
	for i, plan := range plans {
		responsePlans[i] = s.convertToResponse(plan)
	}

	// Debug: Log the response plans
	log.Printf("Service: Response plans: %+v", responsePlans)

	return responsePlans, nil
}

// CheckAndHandleExpiredPromotions checks for expired promotions and deactivates them
func (s *SubscriptionPlanService) CheckAndHandleExpiredPromotions(ctx context.Context) error {
	plans, err := s.db.GetAllSubscriptionPlans()
	if err != nil {
		return fmt.Errorf("failed to get subscription plans: %w", err)
	}

	now := time.Now()
	var expiredPromotions []int

	for _, plan := range plans {
		// Check if plan is promoted and has a promotion end date
		if plan.IsPromoted && plan.PromotionEndDate.Valid {
			// If promotion end date has passed, mark for deactivation
			if plan.PromotionEndDate.Time.Before(now) {
				expiredPromotions = append(expiredPromotions, plan.ID)
			}
		}
	}

	// Deactivate expired promotions
	for _, planID := range expiredPromotions {
		updates := map[string]interface{}{
			"is_promoted":        false,
			"is_active":          false,
			"promotion_end_date": now, // Set to current time when deactivated
			"updated_at":         now,
		}

		_, err := s.db.UpdateSubscriptionPlan(planID, updates)
		if err != nil {
			log.Printf("Service: Failed to deactivate expired promotion for plan %d: %v", planID, err)
		} else {
			log.Printf("Service: Deactivated expired promotion for plan %d", planID)
		}
	}

	return nil
}

// validateSubscriptionPlanRequest validates a subscription plan request
func (s *SubscriptionPlanService) validateSubscriptionPlanRequest(req interface{}) error {
	// Validation logic here
	return nil
}

// convertToResponse converts a database subscription plan to response format
func (s *SubscriptionPlanService) convertToResponse(plan *database.SubscriptionPlan) *SubscriptionPlanResponse {
	response := &SubscriptionPlanResponse{
		ID:            strconv.Itoa(plan.ID), // Convert int ID to string for response
		Name:          plan.Name,
		Description:   plan.Description,
		ShortDesc:     "", // Will be set below if valid
		Price:         plan.Price,
		Currency:      plan.Currency,
		Interval:      plan.Interval,
		IntervalCount: plan.IntervalCount,
		IsActive:      plan.IsActive,
		IsPromoted:    plan.IsPromoted,
		SubType:       plan.SubType,
		SortOrder:     plan.SortOrder,
		CreatedAt:     plan.CreatedAt,
		UpdatedAt:     plan.UpdatedAt,
	}

	// Handle nullable string fields
	if plan.ShortDesc.Valid {
		response.ShortDesc = plan.ShortDesc.String
	}

	if plan.StripePriceID.Valid {
		response.StripePriceID = &plan.StripePriceID.String
	}

	if plan.PromotionEndDate.Valid {
		response.PromotionEndDate = &plan.PromotionEndDate.Time
	}

	if plan.PromotionStartDate.Valid {
		response.PromotionStartDate = &plan.PromotionStartDate.Time
	}

	// Parse features JSON if available
	if plan.Features.Valid {
		var features []string
		if err := json.Unmarshal([]byte(plan.Features.String), &features); err == nil {
			response.Features = features
		} else {
			response.Features = []string{} // Empty array if parsing fails
		}
	} else {
		response.Features = []string{} // Empty array if no features
	}

	// Parse promotion history JSON if available
	if plan.PromotionHistory.Valid {
		var promotionHistory []string
		if err := json.Unmarshal([]byte(plan.PromotionHistory.String), &promotionHistory); err == nil {
			response.PromotionHistory = promotionHistory
		} else {
			response.PromotionHistory = []string{} // Empty array if parsing fails
		}
	} else {
		response.PromotionHistory = []string{} // Empty array if no history
	}

	return response
}

// logAuditEvent logs an audit event
func (s *SubscriptionPlanService) logAuditEvent(eventType string, userID, resourceID int, metadata map[string]interface{}) {
	// Audit logging logic here
	log.Printf("Service: Audit event: %s, User: %d, Resource: %d, Metadata: %v", eventType, userID, resourceID, metadata)
}
