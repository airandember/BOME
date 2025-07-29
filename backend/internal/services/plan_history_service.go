package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

// AddHistoryEvent adds a new event to the plan's change history
func (s *PlanHistoryService) AddHistoryEvent(ctx context.Context, planID int, event PlanHistoryEvent) error {
	// Get current history
	currentHistory, err := s.GetPlanHistory(ctx, planID)
	if err != nil {
		return fmt.Errorf("failed to get current history: %w", err)
	}

	// Add new event
	currentHistory = append(currentHistory, event)

	// Convert to JSON
	historyJSON, err := json.Marshal(currentHistory)
	if err != nil {
		return fmt.Errorf("failed to marshal history to JSON: %w", err)
	}

	// Update database
	updates := map[string]interface{}{
		"plan_change_history": string(historyJSON),
	}

	_, err = s.db.UpdateSubscriptionPlan(planID, updates)
	if err != nil {
		return fmt.Errorf("failed to update plan history: %w", err)
	}

	log.Printf("Added history event to plan %d: %s", planID, event.EventType)
	return nil
}

// GetPlanHistory retrieves the complete change history for a plan
func (s *PlanHistoryService) GetPlanHistory(ctx context.Context, planID int) ([]PlanHistoryEvent, error) {
	//plan, err := s.db.GetSubscriptionPlanByID(planID)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to get plan: %w", err)
	//}

	//if !plan.PlanChangeHistory.Valid || plan.PlanChangeHistory.String == "" {
	//return []PlanHistoryEvent{}, nil
	//}

	//var history []PlanHistoryEvent
	//err = json.Unmarshal([]byte(plan.PlanChangeHistory.String), &history)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to unmarshal history: %w", err)
	//}

	//return history, nil
	return nil, nil
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
	//plan, err := s.db.GetSubscriptionPlanByID(planID)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to get plan: %w", err)
	//}

	//if !plan.PromotionMetadata.Valid || plan.PromotionMetadata.String == "" {
	//	return &PromotionMetadata{}, nil
	//}

	//var metadata PromotionMetadata
	//err = json.Unmarshal([]byte(plan.PromotionMetadata.String), &metadata)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to unmarshal promotion metadata: %w", err)
	//}

	//return &metadata, nil
	return nil, nil
}

// CreatePlanCreatedEvent creates a history event for plan creation
func (s *PlanHistoryService) CreatePlanCreatedEvent(plan *database.SubscriptionPlan, userID string) PlanHistoryEvent {
	return PlanHistoryEvent{
		ID:          fmt.Sprintf("evt_%d", time.Now().Unix()),
		EventType:   "plan_created",
		Timestamp:   time.Now(),
		UserID:      userID,
		Description: fmt.Sprintf("Plan '%s' was created", plan.Name),
		NewValues: map[string]interface{}{
			"name":      plan.Name,
			"price":     plan.Price,
			"sub_type":  plan.SubType,
			"is_active": plan.IsActive,
		},
		Metadata: map[string]interface{}{
			"plan_name": plan.Name,
			"action":    "creation",
		},
	}
}

// CreatePlanUpdatedEvent creates a history event for plan updates
func (s *PlanHistoryService) CreatePlanUpdatedEvent(planID int, oldValues, newValues map[string]interface{}, userID string) PlanHistoryEvent {
	return PlanHistoryEvent{
		ID:          fmt.Sprintf("evt_%d", time.Now().Unix()),
		EventType:   "plan_updated",
		Timestamp:   time.Now(),
		UserID:      userID,
		Description: fmt.Sprintf("Plan was updated"),
		OldValues:   oldValues,
		NewValues:   newValues,
		Metadata: map[string]interface{}{
			"plan_id": planID,
			"action":  "update",
		},
	}
}

// CreatePromotionStartedEvent creates a history event for promotion start
func (s *PlanHistoryService) CreatePromotionStartedEvent(plan *database.SubscriptionPlan, userID string) PlanHistoryEvent {
	return PlanHistoryEvent{
		ID:        fmt.Sprintf("evt_%d", time.Now().Unix()),
		EventType: "promotion_started",
		Timestamp: time.Now(),
		UserID:    userID,
		Description: fmt.Sprintf("Promotion started for %s (until %s)",
			plan.Name,
			plan.PromotionEndDate.Time.Format("2006-01-02")),
		OldValues: map[string]interface{}{
			"sub_type": "stnd",
		},
		NewValues: map[string]interface{}{
			"sub_type":             "prmo",
			"promotion_start_date": plan.PromotionStartDate.Time,
			"promotion_end_date":   plan.PromotionEndDate.Time,
		},
		Metadata: map[string]interface{}{
			"plan_name": plan.Name,
			"action":    "promotion_start",
		},
	}
}

// CreatePromotionEndedEvent creates a history event for promotion end
func (s *PlanHistoryService) CreatePromotionEndedEvent(plan *database.SubscriptionPlan, userID string, reason string) PlanHistoryEvent {
	return PlanHistoryEvent{
		ID:          fmt.Sprintf("evt_%d", time.Now().Unix()),
		EventType:   "promotion_ended",
		Timestamp:   time.Now(),
		UserID:      userID,
		Description: fmt.Sprintf("Promotion ended for %s (%s)", plan.Name, reason),
		OldValues: map[string]interface{}{
			"sub_type": "prmo",
		},
		NewValues: map[string]interface{}{
			"sub_type":  "stnd",
			"is_active": false,
		},
		Metadata: map[string]interface{}{
			"plan_name": plan.Name,
			"action":    "promotion_end",
			"reason":    reason,
		},
	}
}

// CreateStatusToggleEvent creates a history event for status changes
func (s *PlanHistoryService) CreateStatusToggleEvent(plan *database.SubscriptionPlan, newStatus bool, userID string) PlanHistoryEvent {
	action := "activated"
	if !newStatus {
		action = "deactivated"
	}

	return PlanHistoryEvent{
		ID:          fmt.Sprintf("evt_%d", time.Now().Unix()),
		EventType:   "status_toggled",
		Timestamp:   time.Now(),
		UserID:      userID,
		Description: fmt.Sprintf("Plan '%s' was %s", plan.Name, action),
		OldValues: map[string]interface{}{
			"is_active": !newStatus,
		},
		NewValues: map[string]interface{}{
			"is_active": newStatus,
		},
		Metadata: map[string]interface{}{
			"plan_name": plan.Name,
			"action":    action,
		},
	}
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
