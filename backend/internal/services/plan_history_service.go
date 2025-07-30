package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"bome-backend/internal/database"
)

// PlanHistoryEvent represents a single event in the plan change history
type PlanHistoryEvent struct {
	ID          string                 `json:"id"`
	EventType   string                 `json:"event_type"`
	Timestamp   time.Time              `json:"timestamp"`
	UserID      string                 `json:"user_id,omitempty"`
	Description string                 `json:"description"`
	OldValues   map[string]interface{} `json:"old_values,omitempty"`
	NewValues   map[string]interface{} `json:"new_values,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PromotionMetadata represents promotion-specific analytics and settings
type PromotionMetadata struct {
	PromotionStats struct {
		TotalPromotions      int                   `json:"total_promotions"`
		CurrentPromotion     *CurrentPromotion     `json:"current_promotion,omitempty"`
		PerformanceMetrics   PerformanceMetrics    `json:"performance_metrics"`
		HistoricalPromotions []HistoricalPromotion `json:"historical_promotions,omitempty"`
	} `json:"promotion_stats"`
	PromotionSettings struct {
		AutoExtend             bool `json:"auto_extend"`
		NotifyBeforeExpiry     bool `json:"notify_before_expiry"`
		ExpiryNotificationDays int  `json:"expiry_notification_days"`
		AutoReactivate         bool `json:"auto_reactivate"`
		MaxPromotionDuration   int  `json:"max_promotion_duration"`
	} `json:"promotion_settings"`
}

type CurrentPromotion struct {
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	DurationDays     int       `json:"duration_days"`
	Status           string    `json:"status"` // upcoming, active, expired
	RevenueGenerated float64   `json:"revenue_generated"`
	ConversionRate   float64   `json:"conversion_rate"`
}

type PerformanceMetrics struct {
	TotalRevenueGenerated   float64 `json:"total_revenue_generated"`
	AverageConversionRate   float64 `json:"average_conversion_rate"`
	BestPerformingDuration  int     `json:"best_performing_duration"`
	TotalPromotionsRun      int     `json:"total_promotions_run"`
	AveragePromotionLength  int     `json:"average_promotion_length"`
	MostSuccessfulTimeframe string  `json:"most_successful_timeframe"`
}

type HistoricalPromotion struct {
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	Duration         int       `json:"duration"`
	RevenueGenerated float64   `json:"revenue_generated"`
	ConversionRate   float64   `json:"conversion_rate"`
	Status           string    `json:"status"`
}

type PlanHistoryService struct {
	db *database.DB
}

func NewPlanHistoryService(db *database.DB) *PlanHistoryService {
	return &PlanHistoryService{db: db}
}

// AddHistoryEvent adds a new event to the plan's change history using the separate table
func (s *PlanHistoryService) AddHistoryEvent(ctx context.Context, planID int, event PlanHistoryEvent) error {
	log.Printf("AddHistoryEvent: Starting to add event for plan %d", planID)
	log.Printf("AddHistoryEvent: Event details - Type=%s, Description=%s, UserID=%s", event.EventType, event.Description, event.UserID)

	// Use the new database method to add history event
	err := s.db.AddHistoryEvent(
		planID,
		event.EventType,
		event.UserID,
		event.Description,
		event.OldValues,
		event.NewValues,
		event.Metadata,
	)

	if err != nil {
		log.Printf("AddHistoryEvent: Failed to add history event: %v", err)
		return fmt.Errorf("failed to add history event: %w", err)
	}

	log.Printf("AddHistoryEvent: Successfully added history event for plan %d", planID)
	return nil
}

// GetPlanHistory retrieves all history events for a specific plan from the separate table
func (s *PlanHistoryService) GetPlanHistory(ctx context.Context, planID int) ([]PlanHistoryEvent, error) {
	log.Printf("GetPlanHistory: Getting history for plan %d", planID)

	// Use the new database method to get history events
	dbEvents, err := s.db.GetPlanHistory(planID)
	if err != nil {
		log.Printf("GetPlanHistory: Failed to get history from database: %v", err)
		return nil, fmt.Errorf("failed to get plan history: %w", err)
	}

	// Convert database events to service events
	var events []PlanHistoryEvent
	for _, dbEvent := range dbEvents {
		// Parse old values JSON
		var oldValues map[string]interface{}
		if dbEvent.OldValues.Valid {
			if err := json.Unmarshal([]byte(dbEvent.OldValues.String), &oldValues); err != nil {
				log.Printf("GetPlanHistory: Failed to parse old values for event %d: %v", dbEvent.ID, err)
				oldValues = make(map[string]interface{})
			}
		}

		// Parse new values JSON
		var newValues map[string]interface{}
		if dbEvent.NewValues.Valid {
			if err := json.Unmarshal([]byte(dbEvent.NewValues.String), &newValues); err != nil {
				log.Printf("GetPlanHistory: Failed to parse new values for event %d: %v", dbEvent.ID, err)
				newValues = make(map[string]interface{})
			}
		}

		// Parse metadata JSON
		var metadata map[string]interface{}
		if dbEvent.Metadata.Valid {
			if err := json.Unmarshal([]byte(dbEvent.Metadata.String), &metadata); err != nil {
				log.Printf("GetPlanHistory: Failed to parse metadata for event %d: %v", dbEvent.ID, err)
				metadata = make(map[string]interface{})
			}
		}

		// Convert user ID
		userID := ""
		if dbEvent.UserID.Valid {
			userID = dbEvent.UserID.String
		}

		// Convert description
		description := ""
		if dbEvent.Description.Valid {
			description = dbEvent.Description.String
		}

		event := PlanHistoryEvent{
			ID:          fmt.Sprintf("%d", dbEvent.ID),
			EventType:   dbEvent.EventType,
			Timestamp:   dbEvent.Timestamp,
			UserID:      userID,
			Description: description,
			OldValues:   oldValues,
			NewValues:   newValues,
			Metadata:    metadata,
		}

		events = append(events, event)
	}

	log.Printf("GetPlanHistory: Retrieved %d history events for plan %d", len(events), planID)
	return events, nil
}

// UpdatePromotionMetadata updates the promotion metadata for a plan
func (s *PlanHistoryService) UpdatePromotionMetadata(ctx context.Context, planID int, metadata PromotionMetadata) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal promotion metadata: %w", err)
	}

	updates := map[string]interface{}{
		"promotion_metadata": string(metadataJSON),
	}

	_, err = s.db.UpdateSubscriptionPlan(planID, updates)
	if err != nil {
		return fmt.Errorf("failed to update promotion metadata: %w", err)
	}

	log.Printf("Updated promotion metadata for plan %d", planID)
	return nil
}

// GetPromotionMetadata retrieves the promotion metadata for a plan
func (s *PlanHistoryService) GetPromotionMetadata(ctx context.Context, planID int) (*PromotionMetadata, error) {
	plan, err := s.db.GetSubscriptionPlanByID(planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	if !plan.PromotionMetadata.Valid || plan.PromotionMetadata.String == "" {
		return &PromotionMetadata{}, nil
	}

	var metadata PromotionMetadata
	err = json.Unmarshal([]byte(plan.PromotionMetadata.String), &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal promotion metadata: %w", err)
	}

	return &metadata, nil
}

// CreatePlanCreatedEvent creates a history event for plan creation
func (s *PlanHistoryService) CreatePlanCreatedEvent(plan *database.SubscriptionPlan, userDataJSON string) PlanHistoryEvent {
	eventID := fmt.Sprintf("evt_%d", time.Now().Unix())

	metadata := map[string]interface{}{
		"action":    "created",
		"plan_name": plan.Name,
	}

	event := PlanHistoryEvent{
		ID:          eventID,
		EventType:   "plan_created",
		Timestamp:   time.Now(),
		UserID:      userDataJSON, // Now contains JSONB user data
		Description: fmt.Sprintf("Plan '%s' was created", plan.Name),
		NewValues: map[string]interface{}{
			"id":              plan.ID,
			"name":            plan.Name,
			"price":           plan.Price,
			"currency":        plan.Currency,
			"features":        plan.Features.String,
			"interval":        plan.Interval,
			"sub_type":        plan.SubType,
			"is_active":       plan.IsActive,
			"short_desc":      plan.ShortDesc.String,
			"description":     plan.Description,
			"interval_count":  plan.IntervalCount,
			"stripe_price_id": plan.StripePriceID.String,
		},
		Metadata: metadata,
	}

	return event
}

// CreatePlanUpdatedEvent creates a history event for plan updates
func (s *PlanHistoryService) CreatePlanUpdatedEvent(planID int, oldValues, newValues map[string]interface{}, userDataJSON string) PlanHistoryEvent {
	eventID := fmt.Sprintf("evt_%d", time.Now().Unix())

	// Generate a more descriptive message based on what changed
	var changedFields []string
	for field := range newValues {
		changedFields = append(changedFields, field)
	}

	description := "Plan was updated"
	if len(changedFields) > 0 {
		description = fmt.Sprintf("Plan updated: %s", strings.Join(changedFields, ", "))
	}

	metadata := map[string]interface{}{
		"action":  "update",
		"plan_id": planID,
	}

	event := PlanHistoryEvent{
		ID:          eventID,
		EventType:   "plan_updated",
		Timestamp:   time.Now(),
		UserID:      userDataJSON, // Now contains JSONB user data
		Description: description,
		OldValues:   oldValues,
		NewValues:   newValues,
		Metadata:    metadata,
	}

	return event
}

// CreatePromotionStartedEvent creates a history event for promotion start
func (s *PlanHistoryService) CreatePromotionStartedEvent(plan *database.SubscriptionPlan, userDataJSON string) PlanHistoryEvent {
	eventID := fmt.Sprintf("evt_%d", time.Now().Unix())

	metadata := map[string]interface{}{
		"action":    "promotion_started",
		"plan_name": plan.Name,
	}

	event := PlanHistoryEvent{
		ID:          eventID,
		EventType:   "promotion_started",
		Timestamp:   time.Now(),
		UserID:      userDataJSON, // Now contains JSONB user data
		Description: fmt.Sprintf("Promotion started for plan '%s'", plan.Name),
		NewValues: map[string]interface{}{
			"sub_type":  "prmo",
			"is_active": true,
		},
		Metadata: metadata,
	}

	return event
}

// CreatePromotionEndedEvent creates a history event for promotion end
func (s *PlanHistoryService) CreatePromotionEndedEvent(plan *database.SubscriptionPlan, userDataJSON string, reason string) PlanHistoryEvent {
	eventID := fmt.Sprintf("evt_%d", time.Now().Unix())

	metadata := map[string]interface{}{
		"action":    "promotion_ended",
		"plan_name": plan.Name,
		"reason":    reason,
	}

	event := PlanHistoryEvent{
		ID:          eventID,
		EventType:   "promotion_ended",
		Timestamp:   time.Now(),
		UserID:      userDataJSON, // Now contains JSONB user data
		Description: fmt.Sprintf("Promotion ended for plan '%s' (%s)", plan.Name, reason),
		OldValues: map[string]interface{}{
			"sub_type": "prmo",
		},
		NewValues: map[string]interface{}{
			"sub_type":  "stnd",
			"is_active": false,
		},
		Metadata: metadata,
	}

	return event
}

// CreateStatusToggleEvent creates a history event for status changes
func (s *PlanHistoryService) CreateStatusToggleEvent(plan *database.SubscriptionPlan, newStatus bool, userDataJSON string) PlanHistoryEvent {
	eventID := fmt.Sprintf("evt_%d", time.Now().Unix())

	action := "activated"
	if !newStatus {
		action = "deactivated"
	}

	metadata := map[string]interface{}{
		"action":    action,
		"plan_name": plan.Name,
	}

	event := PlanHistoryEvent{
		ID:          eventID,
		EventType:   "status_toggled",
		Timestamp:   time.Now(),
		UserID:      userDataJSON, // Now contains JSONB user data
		Description: fmt.Sprintf("Plan '%s' was %s", plan.Name, action),
		OldValues: map[string]interface{}{
			"is_active": !newStatus,
		},
		NewValues: map[string]interface{}{
			"is_active": newStatus,
		},
		Metadata: metadata,
	}

	return event
}

// UpdatePromotionStats updates the promotion statistics in metadata
func (s *PlanHistoryService) UpdatePromotionStats(ctx context.Context, planID int, revenue float64, conversionRate float64) error {
	metadata, err := s.GetPromotionMetadata(ctx, planID)
	if err != nil {
		return fmt.Errorf("failed to get promotion metadata: %w", err)
	}

	// Update current promotion stats if exists
	if metadata.PromotionStats.CurrentPromotion != nil {
		metadata.PromotionStats.CurrentPromotion.RevenueGenerated = revenue
		metadata.PromotionStats.CurrentPromotion.ConversionRate = conversionRate
	}

	// Update overall stats
	metadata.PromotionStats.PerformanceMetrics.TotalRevenueGenerated += revenue
	metadata.PromotionStats.PerformanceMetrics.TotalPromotionsRun++

	// Calculate average conversion rate
	if metadata.PromotionStats.PerformanceMetrics.TotalPromotionsRun > 0 {
		metadata.PromotionStats.PerformanceMetrics.AverageConversionRate =
			metadata.PromotionStats.PerformanceMetrics.TotalRevenueGenerated / float64(metadata.PromotionStats.PerformanceMetrics.TotalPromotionsRun)
	}

	return s.UpdatePromotionMetadata(ctx, planID, *metadata)
}
