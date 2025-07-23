package services

import (
	"fmt"
	"log"
	"time"

	"bome-backend/internal/database"
)

// SubscriptionAnalyticsService handles subscription analytics and metrics
type SubscriptionAnalyticsService struct {
	db *database.DB
}

// SubscriptionMetrics represents aggregated subscription metrics
type SubscriptionMetrics struct {
	Date        time.Time `json:"date"`
	PlanID      *int      `json:"plan_id,omitempty"`
	MetricType  string    `json:"metric_type"`
	MetricValue float64   `json:"metric_value"`
	MetricCount int       `json:"metric_count"`
	Currency    string    `json:"currency"`
}

// SubscriptionEvent represents a subscription lifecycle event
type SubscriptionEvent struct {
	ID             int                    `json:"id"`
	SubscriptionID int                    `json:"subscription_id"`
	UserID         *int                   `json:"user_id,omitempty"`
	PlanID         *int                   `json:"plan_id,omitempty"`
	EventType      string                 `json:"event_type"`
	EventData      map[string]interface{} `json:"event_data,omitempty"`
	StripeEventID  *string                `json:"stripe_event_id,omitempty"`
	IPAddress      *string                `json:"ip_address,omitempty"`
	UserAgent      *string                `json:"user_agent,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// SubscriptionWebhook represents a Stripe webhook processing record
type SubscriptionWebhook struct {
	ID             int        `json:"id"`
	StripeEventID  string     `json:"stripe_event_id"`
	EventType      string     `json:"event_type"`
	SubscriptionID *int       `json:"subscription_id,omitempty"`
	UserID         *int       `json:"user_id,omitempty"`
	Status         string     `json:"status"`
	ProcessingTime *int       `json:"processing_time,omitempty"`
	ErrorMessage   *string    `json:"error_message,omitempty"`
	RetryCount     int        `json:"retry_count"`
	PayloadHash    *string    `json:"payload_hash,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	ProcessedAt    *time.Time `json:"processed_at,omitempty"`
}

// MRRData represents Monthly Recurring Revenue data
type MRRData struct {
	Date     time.Time `json:"date"`
	MRR      float64   `json:"mrr"`
	Currency string    `json:"currency"`
	PlanID   *int      `json:"plan_id,omitempty"`
}

// ChurnData represents churn rate data
type ChurnData struct {
	Date        time.Time `json:"date"`
	ChurnRate   float64   `json:"churn_rate"`
	ChurnedSubs int       `json:"churned_subscriptions"`
	TotalSubs   int       `json:"total_subscriptions"`
	PlanID      *int      `json:"plan_id,omitempty"`
}

// NewSubscriptionAnalyticsService creates a new subscription analytics service
func NewSubscriptionAnalyticsService(db *database.DB) *SubscriptionAnalyticsService {
	return &SubscriptionAnalyticsService{db: db}
}

// TrackSubscriptionEvent tracks a subscription lifecycle event
func (s *SubscriptionAnalyticsService) TrackSubscriptionEvent(subscriptionID, userID int, planID *int, eventType string, eventData map[string]interface{}, stripeEventID *string, ipAddress, userAgent string) error {
	// Insert event into database
	// TODO: Add method to database layer for inserting subscription events
	// err := s.db.CreateSubscriptionEvent(subscriptionID, userID, planID, eventType, eventDataJSON, stripeEventID, ipAddress, userAgent)
	// if err != nil {
	//     return fmt.Errorf("failed to track subscription event: %w", err)
	// }

	log.Printf("SUBSCRIPTION EVENT: %s - Subscription: %d, User: %d, Plan: %v", eventType, subscriptionID, userID, planID)

	// Update aggregated metrics
	go s.updateMetricsFromEvent(eventType, subscriptionID, planID, eventData)

	return nil
}

// GetSubscriptionMetrics retrieves subscription metrics for a date range
func (s *SubscriptionAnalyticsService) GetSubscriptionMetrics(startDate, endDate time.Time, planID *int, metricType string) ([]*SubscriptionMetrics, error) {
	// TODO: Add method to database layer for retrieving subscription metrics
	// metrics, err := s.db.GetSubscriptionMetrics(startDate, endDate, planID, metricType)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to get subscription metrics: %w", err)
	// }

	// For now, return empty slice
	return []*SubscriptionMetrics{}, nil
}

// CalculateMRR calculates Monthly Recurring Revenue for a given date
func (s *SubscriptionAnalyticsService) CalculateMRR(date time.Time, planID *int) (*MRRData, error) {
	// For now, calculate MRR as total revenue from active subscriptions
	// In a real implementation, you'd want to calculate this more accurately
	totalRevenue, err := s.db.GetTotalRevenue(planID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate MRR: %w", err)
	}

	mrr := &MRRData{
		Date:     date,
		MRR:      totalRevenue,
		Currency: "USD",
		PlanID:   planID,
	}

	return mrr, nil
}

// CalculateChurnRate calculates churn rate for a given period
func (s *SubscriptionAnalyticsService) CalculateChurnRate(startDate, endDate time.Time, planID *int) (*ChurnData, error) {
	// Get total active subscriptions at start of period
	totalSubs, err := s.db.GetActiveSubscriptionsCount(planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total subscriptions: %w", err)
	}

	// Get cancelled subscriptions during period
	cancelledSubs, err := s.db.GetCancelledSubscriptionsCount(startDate, endDate, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cancelled subscriptions: %w", err)
	}

	// Calculate churn rate
	var churnRate float64
	if totalSubs > 0 {
		churnRate = float64(cancelledSubs) / float64(totalSubs) * 100
	}

	churnData := &ChurnData{
		Date:        endDate,
		ChurnRate:   churnRate,
		ChurnedSubs: cancelledSubs,
		TotalSubs:   totalSubs,
		PlanID:      planID,
	}

	return churnData, nil
}

// GetSubscriptionEvents retrieves subscription events for a subscription
func (s *SubscriptionAnalyticsService) GetSubscriptionEvents(subscriptionID int, limit, offset int) ([]*SubscriptionEvent, error) {
	// TODO: Add method to database layer for retrieving subscription events
	// events, err := s.db.GetSubscriptionEvents(subscriptionID, limit, offset)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to get subscription events: %w", err)
	// }

	// For now, return empty slice
	return []*SubscriptionEvent{}, nil
}

// TrackWebhookEvent tracks a Stripe webhook processing event
func (s *SubscriptionAnalyticsService) TrackWebhookEvent(stripeEventID, eventType string, subscriptionID, userID *int, status string, processingTime *int, errorMessage *string, payloadHash *string) error {
	// TODO: Add method to database layer for inserting webhook events
	// err := s.db.CreateSubscriptionWebhook(stripeEventID, eventType, subscriptionID, userID, status, processingTime, errorMessage, payloadHash)
	// if err != nil {
	//     return fmt.Errorf("failed to track webhook event: %w", err)
	// }

	log.Printf("WEBHOOK EVENT: %s - Stripe Event: %s, Status: %s", eventType, stripeEventID, status)

	return nil
}

// GetWebhookEvents retrieves webhook events with filters
func (s *SubscriptionAnalyticsService) GetWebhookEvents(status, eventType string, limit, offset int) ([]*SubscriptionWebhook, error) {
	// TODO: Add method to database layer for retrieving webhook events
	// events, err := s.db.GetSubscriptionWebhooks(status, eventType, limit, offset)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to get webhook events: %w", err)
	// }

	// For now, return empty slice
	return []*SubscriptionWebhook{}, nil
}

// GenerateSubscriptionReport generates a comprehensive subscription report
func (s *SubscriptionAnalyticsService) GenerateSubscriptionReport(startDate, endDate time.Time, planID *int) (map[string]interface{}, error) {
	report := map[string]interface{}{
		"period": map[string]interface{}{
			"start_date": startDate,
			"end_date":   endDate,
		},
		"plan_id": planID,
		"metrics": map[string]interface{}{},
	}

	// Calculate MRR
	mrr, err := s.CalculateMRR(endDate, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate MRR: %w", err)
	}
	report["metrics"].(map[string]interface{})["mrr"] = mrr

	// Calculate churn rate
	churnData, err := s.CalculateChurnRate(startDate, endDate, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate churn rate: %w", err)
	}
	report["metrics"].(map[string]interface{})["churn"] = churnData

	// Get subscription metrics
	metrics, err := s.GetSubscriptionMetrics(startDate, endDate, planID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription metrics: %w", err)
	}
	report["metrics"].(map[string]interface{})["subscription_metrics"] = metrics

	return report, nil
}

// updateMetricsFromEvent updates aggregated metrics based on an event
func (s *SubscriptionAnalyticsService) updateMetricsFromEvent(eventType string, subscriptionID int, planID *int, eventData map[string]interface{}) {
	// TODO: Implement metric aggregation logic
	// This would involve:
	// 1. Determining which metrics to update based on event type
	// 2. Calculating the metric values
	// 3. Updating the subscription_metrics table

	log.Printf("UPDATING METRICS: Event: %s, Subscription: %d, Plan: %v", eventType, subscriptionID, planID)
}

// GetActiveSubscriptionsCount returns the count of active subscriptions
func (s *SubscriptionAnalyticsService) GetActiveSubscriptionsCount(planID *int) (int, error) {
	count, err := s.db.GetActiveSubscriptionsCount(planID)
	if err != nil {
		return 0, fmt.Errorf("failed to get active subscriptions count: %w", err)
	}
	return count, nil
}

// GetRevenueMetrics returns revenue-related metrics
func (s *SubscriptionAnalyticsService) GetRevenueMetrics(startDate, endDate time.Time, planID *int) (map[string]interface{}, error) {
	// Get total revenue
	totalRevenue, err := s.db.GetTotalRevenue(planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total revenue: %w", err)
	}

	// Get active subscriptions count for average calculation
	activeSubs, err := s.db.GetActiveSubscriptionsCount(planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active subscriptions: %w", err)
	}

	// Calculate average revenue per subscription
	var avgRevenuePerSub float64
	if activeSubs > 0 {
		avgRevenuePerSub = totalRevenue / float64(activeSubs)
	}
	// Get new subscriptions for growth calculation
	newSubs, err := s.db.GetNewSubscriptionsCount(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get new subscriptions: %w", err)
	}

	metrics := map[string]interface{}{
		"total":                totalRevenue,
		"avg_revenue_per_user": avgRevenuePerSub,
		"new_subscriptions":    newSubs,
		"currency":             "USD",
		"period": map[string]interface{}{
			"start_date": startDate,
			"end_date":   endDate,
		},
	}

	return metrics, nil
}
