package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"bome-backend/internal/database"
)

// SubscriptionPlanService handles business logic for subscription plans
type SubscriptionPlanService struct {
	db             *database.DB
	historyService *PlanHistoryService
}

// SubscriptionPlanResponse represents a subscription plan response
type SubscriptionPlanResponse struct {
	ID                 string                   `json:"id"`
	Name               string                   `json:"name"`
	Description        string                   `json:"description"`
	ShortDesc          string                   `json:"short_desc"`
	Price              float64                  `json:"price"`
	Currency           string                   `json:"currency"`
	Interval           string                   `json:"interval"`
	IntervalCount      int                      `json:"interval_count"`
	StripePriceID      *string                  `json:"stripe_price_id,omitempty"`
	Features           []string                 `json:"features"`
	IsActive           bool                     `json:"is_active"`
	PromotionEndDate   *time.Time               `json:"promotion_end_date,omitempty"`
	PromotionStartDate *time.Time               `json:"promotion_start_date,omitempty"`
	PlanChangeHistory  []map[string]interface{} `json:"plan_change_history"` // Always include, even if empty
	PromotionMetadata  map[string]interface{}   `json:"promotion_metadata"`  // Always include, even if empty
	SubType            string                   `json:"sub_type"`            // stnd = standard, prmo = promotional
	CreatedAt          time.Time                `json:"created_at"`
	UpdatedAt          time.Time                `json:"updated_at"`
}

// CreateSubscriptionPlanRequest represents a request to create a subscription plan
type CreateSubscriptionPlanRequest struct {
	Name               string   `json:"name" validate:"required,min=1,max=255"`
	Description        string   `json:"description"`
	ShortDesc          string   `json:"short_desc"`
	Price              float64  `json:"price" validate:"required,min=0"`
	Currency           string   `json:"currency" validate:"required,oneof=USD EUR GBP CAD"`
	Interval           string   `json:"interval" validate:"required,oneof=month year week day"`
	IntervalCount      int      `json:"interval_count" validate:"required,min=1"`
	StripePriceID      string   `json:"stripe_price_id"`
	Features           []string `json:"features"`
	IsActive           bool     `json:"is_active"`
	PromotionStartDate string   `json:"promotion_start_date"`
	PromotionEndDate   string   `json:"promotion_end_date"`
	SubType            string   `json:"sub_type"`
}

// UpdateSubscriptionPlanRequest represents a request to update a subscription plan
type UpdateSubscriptionPlanRequest struct {
	Name               string   `json:"name" validate:"required,min=1,max=255"`
	Description        string   `json:"description"`
	ShortDesc          string   `json:"short_desc"`
	Price              float64  `json:"price" validate:"required,min=0"`
	Currency           string   `json:"currency" validate:"required,oneof=USD EUR GBP CAD"`
	Interval           string   `json:"interval" validate:"required,oneof=month year week day"`
	IntervalCount      int      `json:"interval_count" validate:"required,min=1"`
	StripePriceID      string   `json:"stripe_price_id"`
	Features           []string `json:"features"`
	IsActive           bool     `json:"is_active"`
	PromotionStartDate string   `json:"promotion_start_date"`
	PromotionEndDate   string   `json:"promotion_end_date"`
	SubType            string   `json:"sub_type"`
}

// NewSubscriptionPlanService creates a new subscription plan service
func NewSubscriptionPlanService(db *database.DB) *SubscriptionPlanService {
	return &SubscriptionPlanService{
		db:             db,
		historyService: NewPlanHistoryService(db),
	}
}

// CreateSubscriptionPlan creates a new subscription plan
func (s *SubscriptionPlanService) CreateSubscriptionPlan(ctx context.Context, plan *database.SubscriptionPlan) (*SubscriptionPlanResponse, error) {
	log.Printf("Service: Creating plan: %+v", plan)
	created, err := s.db.CreateSubscriptionPlan(plan)
	if err != nil {
		return nil, err
	}

	// Add history event for plan creation
	// Extract user ID from context if available
	userID := "system" // Default user ID
	if user, exists := ctx.Value("user_id").(string); exists {
		userID = user
	}

	event := s.historyService.CreatePlanCreatedEvent(created, userID)
	err = s.historyService.AddHistoryEvent(ctx, created.ID, event)
	if err != nil {
		log.Printf("Warning: Failed to add creation history event: %v", err)
		// Don't fail the creation if history logging fails
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

	// Get the current plan to compare old vs new values
	currentPlan, err := s.db.GetSubscriptionPlanByID(planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current plan: %w", err)
	}

	log.Printf("Service: Updating plan %d with: %+v", planID, updates)
	updated, err := s.db.UpdateSubscriptionPlan(planID, updates)
	if err != nil {
		return nil, err
	}

	// Add history event for plan update
	// Extract user ID from context if available
	userID := "system" // Default user ID
	if user, exists := ctx.Value("user_id").(string); exists {
		userID = user
	}

	// Create old values map for comparison
	oldValues := map[string]interface{}{
		"name":      currentPlan.Name,
		"price":     currentPlan.Price,
		"sub_type":  currentPlan.SubType,
		"is_active": currentPlan.IsActive,
	}

	// Create new values map from updates
	newValues := make(map[string]interface{})
	for key, value := range updates {
		newValues[key] = value
	}

	event := s.historyService.CreatePlanUpdatedEvent(planID, oldValues, newValues, userID)
	err = s.historyService.AddHistoryEvent(ctx, planID, event)
	if err != nil {
		log.Printf("Warning: Failed to add update history event: %v", err)
		// Don't fail the update if history logging fails
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

	log.Printf("ToggleSubscriptionPlanStatus: Starting toggle for plan %d to isActive=%v", planID, isActive)

	// Check if plan exists
	currentPlan, err := s.db.GetSubscriptionPlanByID(planID)
	if err != nil {
		return nil, fmt.Errorf("subscription plan not found: %w", err)
	}

	log.Printf("ToggleSubscriptionPlanStatus: Current plan state - ID=%d, Name=%s, IsActive=%v, SubType=%s",
		currentPlan.ID, currentPlan.Name, currentPlan.IsActive, currentPlan.SubType)

	// Update status
	updates := map[string]interface{}{
		"is_active": isActive,
	}
	updatedPlan, err := s.db.UpdateSubscriptionPlan(planID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to toggle subscription plan status: %w", err)
	}

	log.Printf("ToggleSubscriptionPlanStatus: Plan updated successfully - ID=%d, IsActive=%v", updatedPlan.ID, updatedPlan.IsActive)

	// Add history event for status toggle
	// Extract user ID from context if available
	userID := "system" // Default user ID
	if user, exists := ctx.Value("user_id").(string); exists {
		userID = user
	}

	log.Printf("ToggleSubscriptionPlanStatus: Creating history event with userID=%s", userID)

	event := s.historyService.CreateStatusToggleEvent(currentPlan, isActive, userID)
	log.Printf("ToggleSubscriptionPlanStatus: Created event - Type=%s, Description=%s", event.EventType, event.Description)

	err = s.historyService.AddHistoryEvent(ctx, planID, event)
	if err != nil {
		log.Printf("Warning: Failed to add status toggle history event: %v", err)
		// Don't fail the toggle if history logging fails
	} else {
		log.Printf("ToggleSubscriptionPlanStatus: Successfully added history event for plan %d", planID)
	}

	return s.convertToResponse(updatedPlan), nil
}

// UpdatePromotionStatus updates the promotion status of a subscription plan
// If isPromoted is true, sets sub_type to prmo (promo), sets promotion_start_date to now, and optionally sets promotion_end_date.
// If isPromoted is false, sets sub_type to stnd (standard), sets is_active to false, and sets promotion_end_date to now.
func (s *SubscriptionPlanService) UpdatePromotionStatus(ctx context.Context, id string, isPromoted bool, promotionEndDate *time.Time) (*SubscriptionPlanResponse, error) {
	planID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid plan ID: %w", err)
	}
	log.Printf("UpdatePromotionStatus called for plan %d, isPromoted: %v", planID, isPromoted)

	currentPlan, err := s.db.GetSubscriptionPlanByID(planID)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if isPromoted {
		updates["sub_type"] = "prmo"
		updates["promotion_start_date"] = time.Now()
		if promotionEndDate != nil {
			updates["promotion_end_date"] = *promotionEndDate
		}
		updates["is_active"] = true
	} else {
		updates["sub_type"] = "stnd"
		updates["is_active"] = false
		updates["promotion_end_date"] = time.Now()
	}

	updatedPlan, err := s.db.UpdateSubscriptionPlan(planID, updates)
	if err != nil {
		return nil, err
	}

	// Add history event for promotion status change
	// Extract user ID from context if available
	userID := "system" // Default user ID
	if user, exists := ctx.Value("user_id").(string); exists {
		userID = user
	}

	if isPromoted {
		// Create promotion started event
		event := s.historyService.CreatePromotionStartedEvent(updatedPlan, userID)
		err = s.historyService.AddHistoryEvent(ctx, planID, event)
		if err != nil {
			log.Printf("Warning: Failed to add promotion started history event: %v", err)
		}
	} else {
		// Create promotion ended event
		reason := "manual"
		if currentPlan.SubType == "prmo" && currentPlan.PromotionEndDate.Valid && currentPlan.PromotionEndDate.Time.Before(time.Now()) {
			reason = "expired"
		}
		event := s.historyService.CreatePromotionEndedEvent(currentPlan, userID, reason)
		err = s.historyService.AddHistoryEvent(ctx, planID, event)
		if err != nil {
			log.Printf("Warning: Failed to add promotion ended history event: %v", err)
		}
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
		log.Printf("Service: Plan %d: ID=%d, Name=%s, IsActive=%v, SubType=%s, Interval=%s",
			i, plan.ID, plan.Name, plan.IsActive, plan.SubType, plan.Interval)
		log.Printf("Service: Plan %d PlanChangeHistory: Valid=%v, String='%s'",
			i, plan.PlanChangeHistory.Valid, plan.PlanChangeHistory.String)
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
		if plan.SubType == "prmo" && plan.PromotionEndDate.Valid {
			// If promotion end date has passed, mark for deactivation
			if plan.PromotionEndDate.Time.Before(now) {
				expiredPromotions = append(expiredPromotions, plan.ID)
			}
		}
	}

	// Deactivate expired promotions
	for _, planID := range expiredPromotions {
		// Get the plan before updating to create proper history event
		plan, err := s.db.GetSubscriptionPlanByID(planID)
		if err != nil {
			log.Printf("Service: Failed to get plan %d for history: %v", planID, err)
			continue
		}

		updates := map[string]interface{}{
			"sub_type":           "stnd",
			"is_active":          false,
			"promotion_end_date": now, // Set to current time when deactivated
			"updated_at":         now,
		}

		_, err = s.db.UpdateSubscriptionPlan(planID, updates)
		if err != nil {
			log.Printf("Service: Failed to deactivate expired promotion for plan %d: %v", planID, err)
		} else {
			log.Printf("Service: Deactivated expired promotion for plan %d", planID)

			// Add history event for automatic expiration
			event := s.historyService.CreatePromotionEndedEvent(plan, "system", "expired")
			err = s.historyService.AddHistoryEvent(ctx, planID, event)
			if err != nil {
				log.Printf("Warning: Failed to add expiration history event for plan %d: %v", planID, err)
			}
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
	log.Printf("convertToResponse: Starting conversion for plan %d", plan.ID)

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
		SubType:       plan.SubType,
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
	if plan.PlanChangeHistory.Valid {
		log.Printf("convertToResponse: Plan %d has valid PlanChangeHistory: %s", plan.ID, plan.PlanChangeHistory.String)
		var promotionHistory []map[string]interface{}
		if err := json.Unmarshal([]byte(plan.PlanChangeHistory.String), &promotionHistory); err == nil {
			log.Printf("convertToResponse: Successfully parsed %d history events for plan %d", len(promotionHistory), plan.ID)
			response.PlanChangeHistory = promotionHistory
		} else {
			log.Printf("convertToResponse: Failed to parse history for plan %d: %v", plan.ID, err)
			response.PlanChangeHistory = []map[string]interface{}{} // Empty array if parsing fails
		}
	} else {
		log.Printf("convertToResponse: Plan %d has no valid PlanChangeHistory", plan.ID)
		response.PlanChangeHistory = []map[string]interface{}{} // Empty array if no history
	}

	// Parse promotion metadata JSON if available
	if plan.PromotionMetadata.Valid {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(plan.PromotionMetadata.String), &metadata); err == nil {
			response.PromotionMetadata = metadata
		} else {
			response.PromotionMetadata = map[string]interface{}{} // Empty map if parsing fails
		}
	} else {
		response.PromotionMetadata = map[string]interface{}{} // Empty map if no metadata
	}

	log.Printf("convertToResponse: Final response for plan %d - PlanChangeHistory length: %d", plan.ID, len(response.PlanChangeHistory))
	return response
}

// logAuditEvent logs an audit event
func (s *SubscriptionPlanService) logAuditEvent(eventType string, userID, resourceID int, metadata map[string]interface{}) {
	// Audit logging logic here
	log.Printf("Service: Audit event: %s, User: %d, Resource: %d, Metadata: %v", eventType, userID, resourceID, metadata)
}

// ParseFlexibleDate handles multiple date formats
func ParseFlexibleDate(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}

	// List of supported date formats
	formats := []string{
		"2006-01-02T15:04:05Z07:00",   // RFC3339
		"2006-01-02T15:04:05",         // ISO without timezone
		"2006-01-02 15:04:05",         // Space separated
		"2006-01-02",                  // Date only (what frontend sends)
		"01/02/2006",                  // MM/DD/YYYY
		"02/01/2006",                  // DD/MM/YYYY
		"2006-01-02T15:04:05.000Z",    // ISO with milliseconds
		"2006-01-02T15:04:05.000000Z", // ISO with microseconds
	}

	for _, format := range formats {
		if parsed, err := time.Parse(format, dateStr); err == nil {
			return &parsed, nil
		}
	}

	return nil, fmt.Errorf("unable to parse date '%s' with any supported format", dateStr)
}

// FormatDateForDatabase ensures consistent date formatting for database storage
func FormatDateForDatabase(t *time.Time, isEndDate bool) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}

	var formattedTime time.Time
	if isEndDate {
		// Round up to end of day for end dates (23:59:59.999999999)
		formattedTime = time.Date(
			t.Year(), t.Month(), t.Day(),
			23, 59, 59, 999999999,
			t.Location(),
		)
	} else {
		// Round down to start of day for start dates (00:00:00)
		formattedTime = time.Date(
			t.Year(), t.Month(), t.Day(),
			0, 0, 0, 0,
			t.Location(),
		)
	}

	return sql.NullTime{Time: formattedTime, Valid: true}
}
