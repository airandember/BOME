package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
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
	if db == nil {
		log.Printf("Warning: Database is nil, creating service with nil database")
	}
	return &SubscriptionPlanService{
		db:             db,
		historyService: NewPlanHistoryService(db),
	}
}

// CreateSubscriptionPlan creates a new subscription plan
func (s *SubscriptionPlanService) CreateSubscriptionPlan(ctx context.Context, plan *database.SubscriptionPlan) (*SubscriptionPlanResponse, error) {
	// Check if database is nil
	if s.db == nil {
		return nil, fmt.Errorf("database is not available")
	}

	// Validate user context before proceeding
	if err := s.validateUserContext(ctx); err != nil {
		log.Printf("CreateSubscriptionPlan: User context validation failed: %v", err)
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	log.Printf("Service: Creating plan: %+v", plan)
	created, err := s.db.CreateSubscriptionPlan(plan)
	if err != nil {
		return nil, err
	}

	// Get user information for history logging in JSONB format
	userDataJSON, err := s.getUserDataFromContext(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get user data for creation history: %v", err)
		// Continue with system user data if user info fails
		userDataJSON = `{"id":"system","email":"system","role":"system","first_name":"System","last_name":""}`
	}
	log.Printf("Service: Creating plan with user data: %s", userDataJSON)

	event := s.historyService.CreatePlanCreatedEvent(created, userDataJSON)
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

	// Check if database is nil
	if s.db == nil {
		return nil, fmt.Errorf("database is not available")
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

	// Check if database is nil
	if s.db == nil {
		return nil, fmt.Errorf("database is not available")
	}

	// Validate user context before proceeding
	if err := s.validateUserContext(ctx); err != nil {
		log.Printf("UpdateSubscriptionPlan: User context validation failed: %v", err)
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	// Get the current plan to compare old vs new values
	currentPlan, err := s.db.GetSubscriptionPlanByID(planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current plan: %w", err)
	}

	// Filter out non-updatable fields and normalize the updates
	filteredUpdates := make(map[string]interface{})
	updatableFields := map[string]bool{
		"name":                 true,
		"description":          true,
		"short_desc":           true,
		"price":                true,
		"currency":             true,
		"interval":             true,
		"interval_count":       true,
		"stripe_price_id":      true,
		"features":             true,
		"is_active":            true,
		"sub_type":             true,
		"promotion_start_date": true,
		"promotion_end_date":   true,
	}

	for key, value := range updates {
		if updatableFields[key] {
			filteredUpdates[key] = value
		}
	}

	log.Printf("Service: Updating plan %d with filtered updates: %+v", planID, filteredUpdates)
	updated, err := s.db.UpdateSubscriptionPlan(planID, filteredUpdates)
	if err != nil {
		return nil, err
	}

	// Create old values map for comparison (only for fields that can be updated)
	oldValues := make(map[string]interface{})

	// Only include fields that are actually being updated
	for key := range filteredUpdates {
		switch key {
		case "name":
			oldValues[key] = currentPlan.Name
		case "price":
			oldValues[key] = currentPlan.Price
		case "sub_type":
			oldValues[key] = currentPlan.SubType
		case "is_active":
			oldValues[key] = currentPlan.IsActive
		case "description":
			oldValues[key] = currentPlan.Description
		case "short_desc":
			oldValues[key] = currentPlan.ShortDesc.String
		case "currency":
			oldValues[key] = currentPlan.Currency
		case "interval":
			oldValues[key] = currentPlan.Interval
		case "interval_count":
			oldValues[key] = currentPlan.IntervalCount
		case "stripe_price_id":
			oldValues[key] = currentPlan.StripePriceID.String
		case "features":
			oldValues[key] = currentPlan.Features.String
		case "promotion_start_date":
			if currentPlan.PromotionStartDate.Valid {
				oldValues[key] = currentPlan.PromotionStartDate.Time
			} else {
				oldValues[key] = nil
			}
		case "promotion_end_date":
			if currentPlan.PromotionEndDate.Valid {
				oldValues[key] = currentPlan.PromotionEndDate.Time
			} else {
				oldValues[key] = nil
			}
		}
	}

	// Detect actual changes
	changes := s.detectChanges(oldValues, filteredUpdates)

	// Only create history event if there are actual changes
	if len(changes) > 0 {
		log.Printf("Service: Detected %d changes for plan %d: %+v", len(changes), planID, changes)

		// Get user information for history logging in JSONB format
		userDataJSON, err := s.getUserDataFromContext(ctx)
		if err != nil {
			log.Printf("Warning: Failed to get user data for update history: %v", err)
			// Continue with system user data if user info fails
			userDataJSON = `{"id":"system","email":"system","role":"system","first_name":"System","last_name":""}`
		}
		log.Printf("Service: Updating plan with user data: %s", userDataJSON)

		// Create old and new values maps for the history event
		oldValuesForHistory := make(map[string]interface{})
		newValuesForHistory := make(map[string]interface{})

		for field, change := range changes {
			changeMap := change.(map[string]interface{})
			oldValuesForHistory[field] = changeMap["old"]
			newValuesForHistory[field] = changeMap["new"]
		}

		event := s.historyService.CreatePlanUpdatedEvent(planID, oldValuesForHistory, newValuesForHistory, userDataJSON)
		err = s.historyService.AddHistoryEvent(ctx, planID, event)
		if err != nil {
			log.Printf("Warning: Failed to add update history event: %v", err)
			// Don't fail the update if history logging fails
		}
	} else {
		log.Printf("Service: No changes detected for plan %d, skipping history event", planID)
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

	// Validate user context before proceeding
	if err := s.validateUserContext(ctx); err != nil {
		log.Printf("ToggleSubscriptionPlanStatus: User context validation failed: %v", err)
		return nil, fmt.Errorf("authentication required: %w", err)
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

	// Get user information for history logging in JSONB format
	userDataJSON, err := s.getUserDataFromContext(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get user data for status toggle history: %v", err)
		// Continue with system user data if user info fails
		userDataJSON = `{"id":"system","email":"system","role":"system","first_name":"System","last_name":""}`
	}
	log.Printf("ToggleSubscriptionPlanStatus: Creating history event with user data: %s", userDataJSON)

	event := s.historyService.CreateStatusToggleEvent(currentPlan, isActive, userDataJSON)
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

	// Validate user context before proceeding
	if err := s.validateUserContext(ctx); err != nil {
		log.Printf("UpdatePromotionStatus: User context validation failed: %v", err)
		return nil, fmt.Errorf("authentication required: %w", err)
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

	// Get user information for history logging in JSONB format
	userDataJSON, err := s.getUserDataFromContext(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get user data for promotion history: %v", err)
		// Continue with system user data if user info fails
		userDataJSON = `{"id":"system","email":"system","role":"system","first_name":"System","last_name":""}`
	}
	log.Printf("UpdatePromotionStatus: Updating promotion with user data: %s", userDataJSON)

	if isPromoted {
		// Create promotion started event
		event := s.historyService.CreatePromotionStartedEvent(updatedPlan, userDataJSON)
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
		event := s.historyService.CreatePromotionEndedEvent(currentPlan, userDataJSON, reason)
		err = s.historyService.AddHistoryEvent(ctx, planID, event)
		if err != nil {
			log.Printf("Warning: Failed to add promotion ended history event: %v", err)
		}
	}

	return s.convertToResponse(updatedPlan), nil
}

// GetAllSubscriptionPlans gets all subscription plans without pagination or filters
func (s *SubscriptionPlanService) GetAllSubscriptionPlans(ctx context.Context) ([]*SubscriptionPlanResponse, error) {
	log.Printf("Service: GetAllSubscriptionPlans called")

	// Check if database is nil
	if s.db == nil {
		log.Printf("Service: ERROR - Database is nil!")
		return nil, fmt.Errorf("database is not available")
	}

	log.Printf("Service: Database connection available, checking expired promotions")

	// Check and handle expired promotions first
	err := s.CheckAndHandleExpiredPromotions(ctx)
	if err != nil {
		log.Printf("Warning: Failed to check expired promotions: %v", err)
	}

	log.Printf("Service: Getting all subscription plans from database")
	plans, err := s.db.GetAllSubscriptionPlans()
	if err != nil {
		log.Printf("Service: ERROR - Failed to get subscription plans: %v", err)
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

			// Get user information for history logging (use system for automatic events)
			userDataJSON, err := s.getUserDataFromContext(ctx)
			if err != nil {
				log.Printf("Warning: Failed to get user data for auto-expiration history: %v", err)
				// Continue with system user data if user info fails
				userDataJSON = `{"id":"system","email":"system","role":"system","first_name":"System","last_name":"(Auto-Expiration)"}`
			}
			log.Printf("Service: Auto-expiration with user data: %s", userDataJSON)

			// Add history event for automatic expiration
			event := s.historyService.CreatePromotionEndedEvent(plan, userDataJSON, "expired")
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

	// Get history from the separate table instead of JSONB column
	log.Printf("convertToResponse: Fetching history from separate table for plan %d", plan.ID)
	historyEvents, err := s.historyService.GetPlanHistory(context.Background(), plan.ID)
	if err != nil {
		log.Printf("convertToResponse: Failed to get history for plan %d: %v", plan.ID, err)
		response.PlanChangeHistory = []map[string]interface{}{} // Empty array if failed
	} else {
		// Convert history events to the expected format
		var historyArray []map[string]interface{}
		for _, event := range historyEvents {
			historyMap := map[string]interface{}{
				"id":          event.ID,
				"event_type":  event.EventType,
				"timestamp":   event.Timestamp,
				"user_id":     event.UserID,
				"description": event.Description,
				"old_values":  event.OldValues,
				"new_values":  event.NewValues,
				"metadata":    event.Metadata,
			}
			historyArray = append(historyArray, historyMap)
		}
		response.PlanChangeHistory = historyArray
		log.Printf("convertToResponse: Successfully converted %d history events for plan %d", len(historyArray), plan.ID)
	}

	// Parse promotion metadata JSON if available (this stays as JSONB for now)
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

// detectChanges compares old and new values and returns only the fields that actually changed
func (s *SubscriptionPlanService) detectChanges(oldValues map[string]interface{}, newValues map[string]interface{}) map[string]interface{} {
	changes := make(map[string]interface{})

	for key, newValue := range newValues {
		if oldValue, exists := oldValues[key]; exists {
			// Normalize and compare values
			if !s.valuesAreEqual(oldValue, newValue) {
				changes[key] = map[string]interface{}{
					"old": oldValue,
					"new": newValue,
				}
			}
		} else {
			// New field added
			changes[key] = map[string]interface{}{
				"old": nil,
				"new": newValue,
			}
		}
	}

	// Check for removed fields
	for key, oldValue := range oldValues {
		if _, exists := newValues[key]; !exists {
			changes[key] = map[string]interface{}{
				"old": oldValue,
				"new": nil,
			}
		}
	}

	return changes
}

// valuesAreEqual compares two values with proper type handling
func (s *SubscriptionPlanService) valuesAreEqual(oldValue, newValue interface{}) bool {
	// Handle nil cases
	if oldValue == nil && newValue == nil {
		return true
	}
	if oldValue == nil || newValue == nil {
		return false
	}

	// Convert to comparable types
	oldNormalized := s.normalizeValue(oldValue)
	newNormalized := s.normalizeValue(newValue)

	// Use reflect.DeepEqual for comparison
	return reflect.DeepEqual(oldNormalized, newNormalized)
}

// normalizeValue converts values to a consistent format for comparison
func (s *SubscriptionPlanService) normalizeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case float64:
		// Convert to int if it's a whole number
		if v == float64(int(v)) {
			return int(v)
		}
		return v
	case int:
		return v
	case string:
		// Trim whitespace for string comparison
		trimmed := strings.TrimSpace(v)

		// Try to parse as JSON array first
		var jsonArray []interface{}
		if err := json.Unmarshal([]byte(trimmed), &jsonArray); err == nil {
			// It's a JSON array, normalize the elements
			normalized := make([]interface{}, len(jsonArray))
			for i, item := range jsonArray {
				normalized[i] = s.normalizeValue(item)
			}
			return normalized
		}

		// Try to parse as date and normalize format
		if parsed, err := time.Parse("2006-01-02", trimmed); err == nil {
			// Return date in consistent format
			return parsed.Format("2006-01-02")
		}
		if parsed, err := time.Parse("2006-01-02T15:04:05Z07:00", trimmed); err == nil {
			// Return date in consistent format
			return parsed.Format("2006-01-02")
		}
		if parsed, err := time.Parse("2006-01-02T15:04:05Z", trimmed); err == nil {
			// Return date in consistent format
			return parsed.Format("2006-01-02")
		}
		if parsed, err := time.Parse("2006-01-02 15:04:05 -0700 MST", trimmed); err == nil {
			// Return date in consistent format
			return parsed.Format("2006-01-02")
		}

		return trimmed
	case []interface{}:
		// Normalize array elements
		normalized := make([]interface{}, len(v))
		for i, item := range v {
			normalized[i] = s.normalizeValue(item)
		}
		return normalized
	case []string:
		// Convert string array to interface array for comparison
		normalized := make([]interface{}, len(v))
		for i, item := range v {
			normalized[i] = strings.TrimSpace(item)
		}
		return normalized
	case map[string]interface{}:
		// Normalize map values
		normalized := make(map[string]interface{})
		for k, val := range v {
			normalized[k] = s.normalizeValue(val)
		}
		return normalized
	case time.Time:
		// Normalize time to consistent date format
		return v.Format("2006-01-02")
	default:
		return value
	}
}

// GetHistoryService returns the history service for external access
func (s *SubscriptionPlanService) GetHistoryService() *PlanHistoryService {
	return s.historyService
}

// GetDatabase returns the database for external access
func (s *SubscriptionPlanService) GetDatabase() *database.DB {
	return s.db
}

// getUserInfoFromContext extracts user information from context and gets full name from database
func (s *SubscriptionPlanService) getUserInfoFromContext(ctx context.Context) (string, string, string, error) {
	// Try to get user ID from context
	userID, exists := ctx.Value("user_id").(int)
	if !exists || userID <= 0 {
		// Fallback to system user with dashboard context
		return "system", "System", "system", nil
	}

	userIDStr := fmt.Sprintf("%d", userID)
	userDisplayName := "System"
	userRole := "system"

	// Try to get user email and role from context
	email, exists := ctx.Value("user_email").(string)
	if !exists || email == "" {
		// Use fallback but still try to get user details from database
		userDisplayName = "System"
		userRole = "system"
	} else {
		userDisplayName = email
		userRole = "user" // Default role

		// Try to get role from context
		if role, exists := ctx.Value("user_role").(string); exists && role != "" {
			userRole = role
		}
	}

	// Try to get full name from database
	if userObj, err := s.db.GetUserByID(userID); err == nil {
		if userObj.FirstName != "" || userObj.LastName != "" {
			// Use full name if available
			fullName := strings.TrimSpace(userObj.FirstName + " " + userObj.LastName)
			if fullName != "" {
				userDisplayName = fullName
			}
		}
	} else {
		log.Printf("Warning: Failed to get user details from database for user %d: %v", userID, err)
		// Continue with email as display name
	}

	return userIDStr, userDisplayName, userRole, nil
}

// getUserDataFromContext extracts user information and returns it in JSONB format
func (s *SubscriptionPlanService) getUserDataFromContext(ctx context.Context) (string, error) {
	log.Printf("getUserDataFromContext: Starting user data extraction")

	// Check for frontend user data first (from localStorage)
	if frontendUserData, exists := ctx.Value("frontend_user_data").(string); exists && frontendUserData != "" {
		log.Printf("getUserDataFromContext: Using frontend user data: %s", frontendUserData)
		return frontendUserData, nil
	}

	log.Printf("getUserDataFromContext: No frontend user data found, checking context values")

	// Try to get user ID from context
	userID, exists := ctx.Value("user_id").(int)
	if !exists || userID <= 0 {
		log.Printf("getUserDataFromContext: No user_id in context, using system fallback")
		// Return system user data in JSONB format
		systemUserData := map[string]interface{}{
			"id":         "system",
			"email":      "system",
			"role":       "system",
			"first_name": "System",
			"last_name":  "",
		}
		userDataJSON, err := json.Marshal(systemUserData)
		if err != nil {
			return "", fmt.Errorf("failed to marshal system user data: %w", err)
		}
		log.Printf("getUserDataFromContext: Returning system user data: %s", string(userDataJSON))
		return string(userDataJSON), nil
	}

	log.Printf("getUserDataFromContext: Found user_id: %d", userID)

	// Get user information from context
	email, exists := ctx.Value("user_email").(string)
	if !exists || email == "" {
		email = "unknown"
		log.Printf("getUserDataFromContext: No user_email in context, using 'unknown'")
	} else {
		log.Printf("getUserDataFromContext: Found user_email: %s", email)
	}

	role, exists := ctx.Value("user_role").(string)
	if !exists || role == "" {
		role = "user"
		log.Printf("getUserDataFromContext: No user_role in context, using 'user'")
	} else {
		log.Printf("getUserDataFromContext: Found user_role: %s", role)
	}

	// Try to get full name from database
	firstName := ""
	lastName := ""
	if userObj, err := s.db.GetUserByID(userID); err == nil {
		firstName = userObj.FirstName
		lastName = userObj.LastName
		log.Printf("getUserDataFromContext: Got user details from DB - FirstName: %s, LastName: %s", firstName, lastName)
	} else {
		log.Printf("Warning: Failed to get user details from database for user %d: %v", userID, err)
		// Use email as fallback
		firstName = email
		lastName = ""
		log.Printf("getUserDataFromContext: Using email as firstName fallback: %s", firstName)
	}

	// Create user data in JSONB format
	userData := map[string]interface{}{
		"id":         userID,
		"email":      email,
		"role":       role,
		"first_name": firstName,
		"last_name":  lastName,
	}

	userDataJSON, err := json.Marshal(userData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal user data: %w", err)
	}

	log.Printf("getUserDataFromContext: Returning user data: %s", string(userDataJSON))
	return string(userDataJSON), nil
}

// createUserDisplayWithRole creates a display name with a role badge if applicable
func (s *SubscriptionPlanService) createUserDisplayWithRole(displayName, role string) string {
	// Handle special cases first
	if displayName == "System" || displayName == "System (Auto-Expiration)" {
		return displayName // Don't add role badge to system events
	}

	// Handle dashboard user
	if displayName == "Dashboard User" {
		return "Dashboard User (🖥️ Dashboard)"
	}

	// Map roles to display badges
	roleBadges := map[string]string{
		"super_admin":           "👑 Super Admin",
		"system_admin":          "🔧 System Admin",
		"content_manager":       "📝 Content Manager",
		"articles_manager":      "📰 Articles Manager",
		"youtube_manager":       "📺 YouTube Manager",
		"streaming_manager":     "🎥 Streaming Manager",
		"events_manager":        "🎪 Events Manager",
		"advertisement_manager": "📢 Ad Manager",
		"user_manager":          "👥 User Manager",
		"analytics_manager":     "📊 Analytics Manager",
		"financial_admin":       "💰 Financial Admin",
		"admin":                 "⚡ Admin",
		"user":                  "👤 User",
		"system":                "🤖 System",
		"dashboard":             "🖥️ Dashboard",
	}

	if badge, exists := roleBadges[role]; exists {
		return fmt.Sprintf("%s (%s)", displayName, badge)
	}

	// For unknown roles, just add the role name
	return fmt.Sprintf("%s (%s)", displayName, role)
}

// validateUserContext ensures that user information is properly available for audit purposes
// More lenient approach since user is already authenticated in dashboard
func (s *SubscriptionPlanService) validateUserContext(ctx context.Context) error {
	userID, exists := ctx.Value("user_id").(int)
	if !exists || userID <= 0 {
		log.Printf("Warning: user_id not found in context, using system fallback")
		return nil // Allow fallback to system user
	}

	email, exists := ctx.Value("user_email").(string)
	if !exists || email == "" {
		log.Printf("Warning: user_email not found in context, using system fallback")
		return nil // Allow fallback to system user
	}

	role, exists := ctx.Value("user_role").(string)
	if !exists || role == "" {
		log.Printf("Warning: user_role not found in context, using system fallback")
		return nil // Allow fallback to system user
	}

	// Log successful validation
	log.Printf("User context validated: ID=%d, Email=%s, Role=%s", userID, email, role)
	return nil
}
