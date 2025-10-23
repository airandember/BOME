package services

import (
	"bome-backend/internal/database"
)

// BusinessIntelligenceService handles business intelligence and analytics
type BusinessIntelligenceService struct {
	db *database.DB
}

// NewBusinessIntelligenceService creates a new business intelligence service
func NewBusinessIntelligenceService(db *database.DB) *BusinessIntelligenceService {
	return &BusinessIntelligenceService{
		db: db,
	}
}

// ExecutiveSummaryData represents the executive summary analytics
type ExecutiveSummaryData struct {
	RevenueImpact struct {
		PromotionalRevenue float64 `json:"promotional_revenue"`
		StandardRevenue    float64 `json:"standard_revenue"`
		TotalMRR           float64 `json:"total_mrr"`
		GrowthRate         float64 `json:"growth_rate"`
	} `json:"revenue_impact"`
	CustomerImpact struct {
		NewCustomersPromos  int `json:"new_customers_promos"`
		StandardConversions int `json:"standard_conversions"`
		OverallGrowth       int `json:"overall_growth"`
	} `json:"customer_impact"`
	FunnelPerformance struct {
		PromoConversion    float64 `json:"promo_conversion"`
		StandardConversion float64 `json:"standard_conversion"`
		ConversionLift     float64 `json:"conversion_lift"`
	} `json:"funnel_performance"`
}

// FunnelAnalysisData represents funnel analysis data
type FunnelAnalysisData struct {
	Stages []struct {
		Name        string  `json:"name"`
		Standard    int     `json:"standard"`
		Promotional int     `json:"promotional"`
		Lift        float64 `json:"lift"`
	} `json:"stages"`
	ConversionRates struct {
		Standard    float64 `json:"standard"`
		Promotional float64 `json:"promotional"`
		Lift        float64 `json:"lift"`
	} `json:"conversion_rates"`
}

// RevenueImpactData represents revenue impact analysis
type RevenueImpactData struct {
	RevenueBreakdown struct {
		StandardPlans    float64 `json:"standard_plans"`
		PromotionalPlans float64 `json:"promotional_plans"`
		TotalRevenue     float64 `json:"total_revenue"`
	} `json:"revenue_breakdown"`
	PromotionalPerformance []struct {
		Name       string  `json:"name"`
		Revenue    float64 `json:"revenue"`
		Percentage float64 `json:"percentage"`
	} `json:"promotional_performance"`
	BaselineComparison struct {
		PrePromoMRR     float64 `json:"pre_promo_mrr"`
		CurrentMRR      float64 `json:"current_mrr"`
		PromotionalLift float64 `json:"promotional_lift"`
	} `json:"baseline_comparison"`
}

// CustomerJourneyData represents customer journey analysis
type CustomerJourneyData struct {
	JourneyMetrics []struct {
		Metric      string   `json:"metric"`
		Standard    float64  `json:"standard"`
		Promotional float64  `json:"promotional"`
		Improvement *float64 `json:"improvement,omitempty"`
		Difference  *float64 `json:"difference,omitempty"`
	} `json:"journey_metrics"`
	NetImpact string `json:"net_impact"`
}

// GetExecutiveSummary retrieves executive summary data
func (bis *BusinessIntelligenceService) GetExecutiveSummary(period string) (*ExecutiveSummaryData, error) {
	data := &ExecutiveSummaryData{}

	// Calculate revenue impact
	revenueImpact, err := bis.calculateRevenueImpact(period)
	if err != nil {
		return nil, err
	}
	data.RevenueImpact = revenueImpact

	// Calculate customer impact
	customerImpact, err := bis.calculateCustomerImpact(period)
	if err != nil {
		return nil, err
	}
	data.CustomerImpact = customerImpact

	// Calculate funnel performance
	funnelPerformance, err := bis.calculateFunnelPerformance(period)
	if err != nil {
		return nil, err
	}
	data.FunnelPerformance = funnelPerformance

	return data, nil
}

// GetFunnelAnalysis retrieves funnel analysis data
func (bis *BusinessIntelligenceService) GetFunnelAnalysis(period string) (*FunnelAnalysisData, error) {
	data := &FunnelAnalysisData{}

	// Calculate funnel stages
	stages, err := bis.calculateFunnelStages(period)
	if err != nil {
		return nil, err
	}
	data.Stages = stages

	// Calculate conversion rates
	conversionRates, err := bis.calculateConversionRates(period)
	if err != nil {
		return nil, err
	}
	data.ConversionRates = conversionRates

	return data, nil
}

// GetRevenueImpact retrieves revenue impact data
func (bis *BusinessIntelligenceService) GetRevenueImpact(period string) (*RevenueImpactData, error) {
	data := &RevenueImpactData{}

	// Calculate revenue breakdown
	revenueBreakdown, err := bis.calculateRevenueBreakdown(period)
	if err != nil {
		return nil, err
	}
	data.RevenueBreakdown = revenueBreakdown

	// Calculate promotional performance
	promotionalPerformance, err := bis.calculatePromotionalPerformance(period)
	if err != nil {
		return nil, err
	}
	data.PromotionalPerformance = promotionalPerformance

	// Calculate baseline comparison
	baselineComparison, err := bis.calculateBaselineComparison(period)
	if err != nil {
		return nil, err
	}
	data.BaselineComparison = baselineComparison

	return data, nil
}

// GetCustomerJourney retrieves customer journey data
func (bis *BusinessIntelligenceService) GetCustomerJourney(period string) (*CustomerJourneyData, error) {
	data := &CustomerJourneyData{}

	// Calculate journey metrics
	journeyMetrics, err := bis.calculateJourneyMetrics(period)
	if err != nil {
		return nil, err
	}
	data.JourneyMetrics = journeyMetrics

	// Determine net impact
	data.NetImpact = bis.determineNetImpact(journeyMetrics)

	return data, nil
}

// Helper methods for calculations
func (bis *BusinessIntelligenceService) calculateRevenueImpact(period string) (struct {
	PromotionalRevenue float64 `json:"promotional_revenue"`
	StandardRevenue    float64 `json:"standard_revenue"`
	TotalMRR           float64 `json:"total_mrr"`
	GrowthRate         float64 `json:"growth_rate"`
}, error) {
	var result struct {
		PromotionalRevenue float64 `json:"promotional_revenue"`
		StandardRevenue    float64 `json:"standard_revenue"`
		TotalMRR           float64 `json:"total_mrr"`
		GrowthRate         float64 `json:"growth_rate"`
	}

	// Query for promotional revenue
	query := `
		SELECT COALESCE(SUM(price), 0) as promotional_revenue
		FROM subscription_plans 
		WHERE sub_type = 'prmo' AND is_active = true
	`
	err := bis.db.QueryRow(query).Scan(&result.PromotionalRevenue)
	if err != nil {
		return result, err
	}

	// Query for standard revenue
	query = `
		SELECT COALESCE(SUM(price), 0) as standard_revenue
		FROM subscription_plans 
		WHERE sub_type = 'stnd' AND is_active = true
	`
	err = bis.db.QueryRow(query).Scan(&result.StandardRevenue)
	if err != nil {
		return result, err
	}

	result.TotalMRR = result.PromotionalRevenue + result.StandardRevenue

	// Calculate growth rate (simplified calculation)
	if result.StandardRevenue > 0 {
		result.GrowthRate = (result.PromotionalRevenue / result.StandardRevenue) * 100
	}

	return result, nil
}

func (bis *BusinessIntelligenceService) calculateCustomerImpact(period string) (struct {
	NewCustomersPromos  int `json:"new_customers_promos"`
	StandardConversions int `json:"standard_conversions"`
	OverallGrowth       int `json:"overall_growth"`
}, error) {
	var result struct {
		NewCustomersPromos  int `json:"new_customers_promos"`
		StandardConversions int `json:"standard_conversions"`
		OverallGrowth       int `json:"overall_growth"`
	}

	// Mock data for now - in real implementation, this would query actual customer data
	result.NewCustomersPromos = 234
	result.StandardConversions = 156
	result.OverallGrowth = 18

	return result, nil
}

func (bis *BusinessIntelligenceService) calculateFunnelPerformance(period string) (struct {
	PromoConversion    float64 `json:"promo_conversion"`
	StandardConversion float64 `json:"standard_conversion"`
	ConversionLift     float64 `json:"conversion_lift"`
}, error) {
	var result struct {
		PromoConversion    float64 `json:"promo_conversion"`
		StandardConversion float64 `json:"standard_conversion"`
		ConversionLift     float64 `json:"conversion_lift"`
	}

	// Mock data for now
	result.PromoConversion = 2.9
	result.StandardConversion = 1.8
	result.ConversionLift = 61.1

	return result, nil
}

func (bis *BusinessIntelligenceService) calculateFunnelStages(period string) ([]struct {
	Name        string  `json:"name"`
	Standard    int     `json:"standard"`
	Promotional int     `json:"promotional"`
	Lift        float64 `json:"lift"`
}, error) {
	// Mock funnel stages data
	stages := []struct {
		Name        string  `json:"name"`
		Standard    int     `json:"standard"`
		Promotional int     `json:"promotional"`
		Lift        float64 `json:"lift"`
	}{
		{Name: "Awareness", Standard: 10000, Promotional: 15000, Lift: 50},
		{Name: "Interest", Standard: 2500, Promotional: 4500, Lift: 80},
		{Name: "Consideration", Standard: 1250, Promotional: 2700, Lift: 116},
		{Name: "Conversion", Standard: 180, Promotional: 432, Lift: 140},
		{Name: "Retention", Standard: 162, Promotional: 389, Lift: 140},
	}

	return stages, nil
}

func (bis *BusinessIntelligenceService) calculateConversionRates(period string) (struct {
	Standard    float64 `json:"standard"`
	Promotional float64 `json:"promotional"`
	Lift        float64 `json:"lift"`
}, error) {
	var result struct {
		Standard    float64 `json:"standard"`
		Promotional float64 `json:"promotional"`
		Lift        float64 `json:"lift"`
	}

	result.Standard = 1.8
	result.Promotional = 2.9
	result.Lift = 61.1

	return result, nil
}

func (bis *BusinessIntelligenceService) calculateRevenueBreakdown(period string) (struct {
	StandardPlans    float64 `json:"standard_plans"`
	PromotionalPlans float64 `json:"promotional_plans"`
	TotalRevenue     float64 `json:"total_revenue"`
}, error) {
	var result struct {
		StandardPlans    float64 `json:"standard_plans"`
		PromotionalPlans float64 `json:"promotional_plans"`
		TotalRevenue     float64 `json:"total_revenue"`
	}

	// Query for standard plans revenue
	query := `
		SELECT COALESCE(SUM(price), 0) as standard_revenue
		FROM subscription_plans 
		WHERE sub_type = 'stnd' AND is_active = true
	`
	err := bis.db.QueryRow(query).Scan(&result.StandardPlans)
	if err != nil {
		return result, err
	}

	// Query for promotional plans revenue
	query = `
		SELECT COALESCE(SUM(price), 0) as promotional_revenue
		FROM subscription_plans 
		WHERE sub_type = 'prmo' AND is_active = true
	`
	err = bis.db.QueryRow(query).Scan(&result.PromotionalPlans)
	if err != nil {
		return result, err
	}

	result.TotalRevenue = result.StandardPlans + result.PromotionalPlans

	return result, nil
}

func (bis *BusinessIntelligenceService) calculatePromotionalPerformance(period string) ([]struct {
	Name       string  `json:"name"`
	Revenue    float64 `json:"revenue"`
	Percentage float64 `json:"percentage"`
}, error) {
	// Mock promotional performance data
	performance := []struct {
		Name       string  `json:"name"`
		Revenue    float64 `json:"revenue"`
		Percentage float64 `json:"percentage"`
	}{
		{Name: "Plan Share!", Revenue: 8200, Percentage: 66},
		{Name: "3 for 4", Revenue: 4250, Percentage: 34},
	}

	return performance, nil
}

func (bis *BusinessIntelligenceService) calculateBaselineComparison(period string) (struct {
	PrePromoMRR     float64 `json:"pre_promo_mrr"`
	CurrentMRR      float64 `json:"current_mrr"`
	PromotionalLift float64 `json:"promotional_lift"`
}, error) {
	var result struct {
		PrePromoMRR     float64 `json:"pre_promo_mrr"`
		CurrentMRR      float64 `json:"current_mrr"`
		PromotionalLift float64 `json:"promotional_lift"`
	}

	// Mock baseline comparison data
	result.PrePromoMRR = 42000
	result.CurrentMRR = 57650
	result.PromotionalLift = 37.3

	return result, nil
}

func (bis *BusinessIntelligenceService) calculateJourneyMetrics(period string) ([]struct {
	Metric      string   `json:"metric"`
	Standard    float64  `json:"standard"`
	Promotional float64  `json:"promotional"`
	Improvement *float64 `json:"improvement,omitempty"`
	Difference  *float64 `json:"difference,omitempty"`
}, error) {
	// Mock journey metrics data
	improvement50 := 50.0
	improvement7 := -7.0
	improvement33 := -33.0
	improvement16 := -16.0

	metrics := []struct {
		Metric      string   `json:"metric"`
		Standard    float64  `json:"standard"`
		Promotional float64  `json:"promotional"`
		Improvement *float64 `json:"improvement,omitempty"`
		Difference  *float64 `json:"difference,omitempty"`
	}{
		{Metric: "Time to Convert", Standard: 14, Promotional: 7, Improvement: &improvement50},
		{Metric: "Avg Order Value", Standard: 29.99, Promotional: 19.99, Difference: &improvement33},
		{Metric: "Retention Rate", Standard: 85, Promotional: 78, Difference: &improvement7},
		{Metric: "Upgrade Rate", Standard: 12, Promotional: 18, Improvement: &improvement50},
		{Metric: "LTV", Standard: 450, Promotional: 380, Difference: &improvement16},
	}

	return metrics, nil
}

func (bis *BusinessIntelligenceService) determineNetImpact(metrics []struct {
	Metric      string   `json:"metric"`
	Standard    float64  `json:"standard"`
	Promotional float64  `json:"promotional"`
	Improvement *float64 `json:"improvement,omitempty"`
	Difference  *float64 `json:"difference,omitempty"`
}) string {
	// Simple logic to determine net impact
	positiveCount := 0
	negativeCount := 0

	for _, metric := range metrics {
		if metric.Improvement != nil && *metric.Improvement > 0 {
			positiveCount++
		} else if metric.Difference != nil && *metric.Difference > 0 {
			positiveCount++
		} else {
			negativeCount++
		}
	}

	if positiveCount > negativeCount {
		return "positive"
	}
	return "negative"
}
