package analytics

// This file defines the ports (interfaces) for the analytics domain
// Implementations are in the subscription/ subdirectory

import "time"

// SubscriptionAnalyticsPort defines the interface for subscription analytics
// Implementation: services/analytics/subscription/subscription_analytics.go
type SubscriptionAnalyticsPort interface {
	GetActiveSubscriptionsCount(filters map[string]interface{}) (int, error)
	GetRevenueMetrics(startDate, endDate time.Time, filters map[string]interface{}) (map[string]interface{}, error)
	CalculateMRR(asOfDate time.Time, filters map[string]interface{}) (float64, error)
	CalculateChurnRate(startDate, endDate time.Time, filters map[string]interface{}) (map[string]interface{}, error)
	GenerateSubscriptionReport(startDate, endDate time.Time, filters map[string]interface{}) (map[string]interface{}, error)
	TrackSubscriptionEvent(eventType string, data map[string]interface{}) error
}

// BusinessIntelligencePort defines the interface for business intelligence analytics
type BusinessIntelligencePort interface {
	GetExecutiveSummary(period string) (map[string]interface{}, error)
	GetFunnelAnalysis(period string) (map[string]interface{}, error)
	GetRevenueImpact(period string) (map[string]interface{}, error)
	GetCustomerJourney(period string) (map[string]interface{}, error)
}

// GeneralAnalyticsPort defines the interface for general application analytics
type GeneralAnalyticsPort interface {
	TrackEvent(eventType string, userID *int, metadata map[string]interface{}) error
	GetDailyActiveUsers(startDate, endDate time.Time) (int, error)
	GetTotalUsers() (int, error)
	GetTopContent(limit int) ([]map[string]interface{}, error)
}
