package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"bome-backend/internal/database"
)

// SubscriptionPlanService handles business logic for subscription plans
type SubscriptionPlanService struct {
	db *database.DB
}

// SubscriptionPlanRequest represents a request to create or update a subscription plan
type SubscriptionPlanRequest struct {
	Name             string     `json:"name" validate:"required,min=1,max=255"`
	Description      string     `json:"description"`
	Price            float64    `json:"price" validate:"required,min=0"`
	Currency         string     `json:"currency" validate:"required,oneof=USD EUR GBP CAD"`
	Interval         string     `json:"interval" validate:"required,oneof=monthly annual weekly daily"`
	IntervalCount    int        `json:"interval_count" validate:"required,min=1"`
	StripePriceID    string     `json:"stripe_price_id"`
	Features         []string   `json:"features"`
	IsActive         bool       `json:"is_active"`
	IsPromoted       bool       `json:"is_promoted"`
	PromotionEndDate *time.Time `json:"promotion_end_date"`
	SortOrder        int        `json:"sort_order"`
}

// SubscriptionPlanResponse represents a subscription plan response
type SubscriptionPlanResponse struct {
	ID               int        `json:"id"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	Price            float64    `json:"price"`
	Currency         string     `json:"currency"`
	Interval         string     `json:"interval"`
	IntervalCount    int        `json:"interval_count"`
	StripePriceID    *string    `json:"stripe_price_id,omitempty"`
	Features         []string   `json:"features"`
	IsActive         bool       `json:"is_active"`
	IsPromoted       bool       `json:"is_promoted"`
	PromotionEndDate *time.Time `json:"promotion_end_date,omitempty"`
	SortOrder        int        `json:"sort_order"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// NewSubscriptionPlanService creates a new subscription plan service
func NewSubscriptionPlanService(db *database.DB) *SubscriptionPlanService {
	return &SubscriptionPlanService{db: db}
}

// CreateSubscriptionPlan creates a new subscription plan
func (s *SubscriptionPlanService) CreateSubscriptionPlan(req *SubscriptionPlanRequest, userID int) (*SubscriptionPlanResponse, error) {
	// Validate request
	if err := s.validateSubscriptionPlanRequest(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Check if plan with same name already exists
	existingPlans, err := s.db.GetSubscriptionPlansWithFilters(1, 0, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing plans: %w", err)
	}

	for _, plan := range existingPlans {
		if plan.Name == req.Name {
			return nil, fmt.Errorf("subscription plan with name '%s' already exists", req.Name)
		}
	}

	// Create the subscription plan
	plan, err := s.db.CreateSubscriptionPlan(
		req.Name,
		req.Description,
		req.Price,
		req.Currency,
		req.Interval,
		req.IntervalCount,
		req.StripePriceID,
		req.Features,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription plan: %w", err)
	}

	// Log audit event
	s.logAuditEvent("subscription_plan_created", userID, plan.ID, map[string]interface{}{
		"plan_name": plan.Name,
		"price":     plan.Price,
		"currency":  plan.Currency,
		"interval":  plan.Interval,
	})

	return s.convertToResponse(plan), nil
}

// GetSubscriptionPlan retrieves a subscription plan by ID
func (s *SubscriptionPlanService) GetSubscriptionPlan(id int) (*SubscriptionPlanResponse, error) {
	plan, err := s.db.GetSubscriptionPlanByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription plan: %w", err)
	}

	return s.convertToResponse(plan), nil
}

// GetActiveSubscriptionPlans retrieves all active subscription plans
func (s *SubscriptionPlanService) GetActiveSubscriptionPlans() ([]*SubscriptionPlanResponse, error) {
	plans, err := s.db.GetActiveSubscriptionPlans()
	if err != nil {
		return nil, fmt.Errorf("failed to get active subscription plans: %w", err)
	}

	var responses []*SubscriptionPlanResponse
	for _, plan := range plans {
		responses = append(responses, s.convertToResponse(plan))
	}

	return responses, nil
}

// GetPromotedSubscriptionPlans retrieves currently promoted subscription plans
func (s *SubscriptionPlanService) GetPromotedSubscriptionPlans() ([]*SubscriptionPlanResponse, error) {
	plans, err := s.db.GetPromotedSubscriptionPlans()
	if err != nil {
		return nil, fmt.Errorf("failed to get promoted subscription plans: %w", err)
	}

	var responses []*SubscriptionPlanResponse
	for _, plan := range plans {
		responses = append(responses, s.convertToResponse(plan))
	}

	return responses, nil
}

// UpdateSubscriptionPlan updates a subscription plan
func (s *SubscriptionPlanService) UpdateSubscriptionPlan(id int, req *SubscriptionPlanRequest, userID int) (*SubscriptionPlanResponse, error) {
	// Validate request
	if err := s.validateSubscriptionPlanRequest(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Check if plan exists
	existingPlan, err := s.db.GetSubscriptionPlanByID(id)
	if err != nil {
		return nil, fmt.Errorf("subscription plan not found: %w", err)
	}

	// Check if name is being changed and if it conflicts with another plan
	if req.Name != existingPlan.Name {
		allPlans, err := s.db.GetSubscriptionPlansWithFilters(100, 0, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing plans: %w", err)
		}

		for _, plan := range allPlans {
			if plan.ID != id && plan.Name == req.Name {
				return nil, fmt.Errorf("subscription plan with name '%s' already exists", req.Name)
			}
		}
	}

	// Build updates map
	updates := map[string]interface{}{
		"name":           req.Name,
		"description":    req.Description,
		"price":          req.Price,
		"currency":       req.Currency,
		"interval":       req.Interval,
		"interval_count": req.IntervalCount,
		"is_active":      req.IsActive,
		"is_promoted":    req.IsPromoted,
		"sort_order":     req.SortOrder,
	}

	if req.StripePriceID != "" {
		updates["stripe_price_id"] = req.StripePriceID
	}

	if req.Features != nil {
		featuresJSON, err := json.Marshal(req.Features)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal features: %w", err)
		}
		updates["features"] = string(featuresJSON)
	}

	if req.PromotionEndDate != nil {
		updates["promotion_end_date"] = *req.PromotionEndDate
	} else {
		updates["promotion_end_date"] = nil
	}

	// Update the plan
	err = s.db.UpdateSubscriptionPlan(id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription plan: %w", err)
	}

	// Get updated plan
	updatedPlan, err := s.db.GetSubscriptionPlanByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated subscription plan: %w", err)
	}

	// Log audit event
	s.logAuditEvent("subscription_plan_updated", userID, id, map[string]interface{}{
		"plan_name": updatedPlan.Name,
		"changes":   updates,
	})

	return s.convertToResponse(updatedPlan), nil
}

// SoftDeleteSubscriptionPlan marks a subscription plan as deleted
func (s *SubscriptionPlanService) SoftDeleteSubscriptionPlan(id int, userID int) error {
	// Check if plan exists
	plan, err := s.db.GetSubscriptionPlanByID(id)
	if err != nil {
		return fmt.Errorf("subscription plan not found: %w", err)
	}

	// Check if plan has active subscriptions
	// TODO: Add method to check for active subscriptions
	// activeSubscriptions, err := s.db.GetActiveSubscriptionsByPlanID(id)
	// if err != nil {
	//     return fmt.Errorf("failed to check active subscriptions: %w", err)
	// }
	// if len(activeSubscriptions) > 0 {
	//     return fmt.Errorf("cannot delete plan with active subscriptions")
	// }

	// Soft delete the plan
	err = s.db.SoftDeleteSubscriptionPlan(id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription plan: %w", err)
	}

	// Log audit event
	s.logAuditEvent("subscription_plan_deleted", userID, id, map[string]interface{}{
		"plan_name": plan.Name,
	})

	return nil
}

// GetSubscriptionPlansWithFilters retrieves subscription plans with filters
func (s *SubscriptionPlanService) GetSubscriptionPlansWithFilters(limit, offset int, isActive, isPromoted *bool) ([]*SubscriptionPlanResponse, error) {
	plans, err := s.db.GetSubscriptionPlansWithFilters(limit, offset, isActive, isPromoted)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription plans: %w", err)
	}

	var responses []*SubscriptionPlanResponse
	for _, plan := range plans {
		responses = append(responses, s.convertToResponse(plan))
	}

	return responses, nil
}

// GetSubscriptionPlanCount returns the total count of subscription plans
func (s *SubscriptionPlanService) GetSubscriptionPlanCount(isActive *bool) (int, error) {
	count, err := s.db.GetSubscriptionPlanCount(isActive)
	if err != nil {
		return 0, fmt.Errorf("failed to get subscription plan count: %w", err)
	}
	return count, nil
}

// validateSubscriptionPlanRequest validates a subscription plan request
func (s *SubscriptionPlanService) validateSubscriptionPlanRequest(req *SubscriptionPlanRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}

	if req.Price < 0 {
		return fmt.Errorf("price must be non-negative")
	}

	if req.IntervalCount < 1 {
		return fmt.Errorf("interval count must be at least 1")
	}

	// Validate promotion end date
	if req.IsPromoted && req.PromotionEndDate != nil {
		if req.PromotionEndDate.Before(time.Now()) {
			return fmt.Errorf("promotion end date cannot be in the past")
		}
	}

	return nil
}

// convertToResponse converts a database subscription plan to a response
func (s *SubscriptionPlanService) convertToResponse(plan *database.SubscriptionPlan) *SubscriptionPlanResponse {
	response := &SubscriptionPlanResponse{
		ID:            plan.ID,
		Name:          plan.Name,
		Description:   plan.Description,
		Price:         plan.Price,
		Currency:      plan.Currency,
		Interval:      plan.Interval,
		IntervalCount: plan.IntervalCount,
		IsActive:      plan.IsActive,
		IsPromoted:    plan.IsPromoted,
		SortOrder:     plan.SortOrder,
		CreatedAt:     plan.CreatedAt,
		UpdatedAt:     plan.UpdatedAt,
	}

	if plan.StripePriceID.Valid {
		response.StripePriceID = &plan.StripePriceID.String
	}

	if plan.PromotionEndDate.Valid {
		response.PromotionEndDate = &plan.PromotionEndDate.Time
	}

	if plan.Features.Valid {
		var features []string
		if err := json.Unmarshal([]byte(plan.Features.String), &features); err == nil {
			response.Features = features
		}
	}

	return response
}

// logAuditEvent logs an audit event for subscription plan operations
func (s *SubscriptionPlanService) logAuditEvent(eventType string, userID, resourceID int, metadata map[string]interface{}) {
	// TODO: Integrate with existing audit system
	log.Printf("AUDIT: %s - User: %d, Resource: %d, Metadata: %+v", eventType, userID, resourceID, metadata)
}
