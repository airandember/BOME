package ports

import "time"

// AnalyticsPort defines the interface for analytics and subscription tracking
type AnalyticsPort interface {
	// Subscription Analytics
	GetActiveSubscriptionsCount(tx interface{}) (int, error)
	GetRevenueMetrics(startDate, endDate time.Time, tx interface{}) (map[string]interface{}, error)
	CalculateMRR(asOf time.Time, tx interface{}) (float64, error)
	CalculateChurnRate(startDate, endDate time.Time, tx interface{}) (map[string]interface{}, error)

	// Subscription Events
	TrackSubscriptionEvent(eventType string, data map[string]interface{})

	// Reports
	GenerateSubscriptionReport(startDate, endDate time.Time, tx interface{}) (map[string]interface{}, error)

	// Customer Analytics
	GetCustomerLifetimeValue(customerID string) (float64, error)
	GetCustomerRetentionRate(startDate, endDate time.Time) (float64, error)
}

// AnalyticsEvent represents an analytics event
type AnalyticsEvent struct {
	EventType string
	UserID    int
	Data      map[string]interface{}
	Timestamp time.Time
}

// SubscriptionMetrics represents subscription analytics metrics
type SubscriptionMetrics struct {
	ActiveCount    int
	NewCount       int
	CancelledCount int
	MRR            float64
	ARR            float64
	ChurnRate      float64
	GrowthRate     float64
}
