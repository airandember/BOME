package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"bome-backend/internal/database"
)

// RevenueAttributionService handles video revenue attribution calculations
type RevenueAttributionService struct {
	db *database.DB
}

// NewRevenueAttributionService creates a new revenue attribution service
func NewRevenueAttributionService(db *database.DB) *RevenueAttributionService {
	return &RevenueAttributionService{db: db}
}

// AttributionFormula represents an attribution calculation formula
type AttributionFormula struct {
	ID                    int                    `json:"id"`
	Name                  string                 `json:"name"`
	Description           string                 `json:"description"`
	FormulaType           string                 `json:"formula_type"`
	FormulaConfig         map[string]interface{} `json:"formula_config"`
	AttributionWindowDays int                    `json:"attribution_window_days"`
	MinWatchPercentage    float64                `json:"min_watch_percentage"`
	IsActive              bool                   `json:"is_active"`
	IsDefault             bool                   `json:"is_default"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
	CreatedBy             *int                   `json:"created_by,omitempty"`
}

// VideoAttribution represents a single attribution record
type VideoAttribution struct {
	ID                       int       `json:"id"`
	VideoID                  int       `json:"video_id"`
	VideoTitle               string    `json:"video_title,omitempty"`
	UserID                   int       `json:"user_id"`
	SubscriptionID           string    `json:"subscription_id"`
	FormulaID                *int      `json:"formula_id,omitempty"`
	AttributionType          string    `json:"attribution_type"`
	AttributionWeight        float64   `json:"attribution_weight"`
	AttributedRevenue        float64   `json:"attributed_revenue"`
	SubscriptionValue        float64   `json:"subscription_value"`
	ViewsBeforeConversion    int       `json:"views_before_conversion"`
	TotalWatchTimeSeconds    int       `json:"total_watch_time_seconds"`
	LastViewBeforeConversion time.Time `json:"last_view_before_conversion"`
	ConversionTimestamp      time.Time `json:"conversion_timestamp"`
	CreatedAt                time.Time `json:"created_at"`
}

// VideoConversionMetrics represents aggregated conversion data for a video
type VideoConversionMetrics struct {
	VideoID                  int       `json:"video_id"`
	VideoTitle               string    `json:"video_title,omitempty"`
	TotalConversions         int       `json:"total_conversions"`
	AssistedConversions      int       `json:"assisted_conversions"`
	TotalAttributedRevenue   float64   `json:"total_attributed_revenue"`
	AvgRevenuePerConversion  float64   `json:"avg_revenue_per_conversion"`
	TotalQualifiedViews      int       `json:"total_qualified_views"`
	ConversionRate           float64   `json:"conversion_rate"`
	AvgTimeToConversionHours float64   `json:"avg_time_to_conversion_hours"`
	LastCalculated           time.Time `json:"last_calculated"`
}

// ConversionJourneyStep represents a video view in the conversion journey
type ConversionJourneyStep struct {
	VideoID           int       `json:"video_id"`
	VideoTitle        string    `json:"video_title"`
	ViewTimestamp     time.Time `json:"view_timestamp"`
	WatchDuration     int       `json:"watch_duration"`
	WatchPercentage   float64   `json:"watch_percentage"`
	HoursToConversion float64   `json:"hours_to_conversion"`
	Position          int       `json:"position"` // 1 = first, N = last
}

// CreateFormulaRequest represents a request to create a new formula
type CreateFormulaRequest struct {
	Name                  string                 `json:"name" binding:"required"`
	Description           string                 `json:"description"`
	FormulaType           string                 `json:"formula_type" binding:"required"`
	FormulaConfig         map[string]interface{} `json:"formula_config"`
	AttributionWindowDays int                    `json:"attribution_window_days"`
	MinWatchPercentage    float64                `json:"min_watch_percentage"`
}

// UpdateFormulaRequest represents a request to update a formula
type UpdateFormulaRequest struct {
	Name                  *string                 `json:"name,omitempty"`
	Description           *string                 `json:"description,omitempty"`
	FormulaConfig         *map[string]interface{} `json:"formula_config,omitempty"`
	AttributionWindowDays *int                    `json:"attribution_window_days,omitempty"`
	MinWatchPercentage    *float64                `json:"min_watch_percentage,omitempty"`
	IsActive              *bool                   `json:"is_active,omitempty"`
	IsDefault             *bool                   `json:"is_default,omitempty"`
}

// AttributionReport represents a comprehensive attribution report
type AttributionReport struct {
	FormulaName        string                   `json:"formula_name"`
	ReportPeriodDays   int                      `json:"report_period_days"`
	TotalRevenue       float64                  `json:"total_revenue"`
	TotalConversions   int                      `json:"total_conversions"`
	VideosWithImpact   int                      `json:"videos_with_impact"`
	TopVideos          []VideoConversionMetrics `json:"top_videos"`
	RecentAttributions []VideoAttribution       `json:"recent_attributions"`
	GeneratedAt        time.Time                `json:"generated_at"`
}

// ===========================
// FORMULA MANAGEMENT
// ===========================

// CreateFormula creates a new attribution formula
func (s *RevenueAttributionService) CreateFormula(req CreateFormulaRequest, userID int) (*AttributionFormula, error) {
	// Validate formula type
	validTypes := map[string]bool{
		"last_touch": true, "first_touch": true, "linear": true,
		"time_decay": true, "position_based": true, "custom": true,
	}
	if !validTypes[req.FormulaType] {
		return nil, fmt.Errorf("invalid formula type: %s", req.FormulaType)
	}

	// Set defaults
	if req.AttributionWindowDays == 0 {
		req.AttributionWindowDays = 7
	}
	if req.MinWatchPercentage == 0 {
		req.MinWatchPercentage = 25.0
	}

	// Convert formula config to JSON
	configJSON, err := json.Marshal(req.FormulaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal formula config: %v", err)
	}

	// Insert formula
	formula := &AttributionFormula{}
	err = s.db.DB.QueryRow(`
		INSERT INTO revenue_attribution_formulas 
		(name, description, formula_type, formula_config, attribution_window_days, 
		 min_watch_percentage, is_active, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, true, $7)
		RETURNING id, name, description, formula_type, formula_config, 
		          attribution_window_days, min_watch_percentage, is_active, 
		          is_default, created_at, updated_at, created_by
	`, req.Name, req.Description, req.FormulaType, configJSON, req.AttributionWindowDays,
		req.MinWatchPercentage, userID).Scan(
		&formula.ID, &formula.Name, &formula.Description, &formula.FormulaType,
		&configJSON, &formula.AttributionWindowDays, &formula.MinWatchPercentage,
		&formula.IsActive, &formula.IsDefault, &formula.CreatedAt, &formula.UpdatedAt,
		&formula.CreatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create formula: %v", err)
	}

	// Parse formula config back
	json.Unmarshal(configJSON, &formula.FormulaConfig)

	log.Printf("✅ Created attribution formula: %s (ID: %d)", formula.Name, formula.ID)
	return formula, nil
}

// GetFormula retrieves a formula by ID
func (s *RevenueAttributionService) GetFormula(formulaID int) (*AttributionFormula, error) {
	formula := &AttributionFormula{}
	var configJSON []byte

	err := s.db.DB.QueryRow(`
		SELECT id, name, description, formula_type, formula_config, 
		       attribution_window_days, min_watch_percentage, is_active, 
		       is_default, created_at, updated_at, created_by
		FROM revenue_attribution_formulas
		WHERE id = $1
	`, formulaID).Scan(
		&formula.ID, &formula.Name, &formula.Description, &formula.FormulaType,
		&configJSON, &formula.AttributionWindowDays, &formula.MinWatchPercentage,
		&formula.IsActive, &formula.IsDefault, &formula.CreatedAt, &formula.UpdatedAt,
		&formula.CreatedBy,
	)
	if err != nil {
		return nil, err
	}

	// Parse formula config
	if len(configJSON) > 0 {
		json.Unmarshal(configJSON, &formula.FormulaConfig)
	}

	return formula, nil
}

// GetAllFormulas retrieves all attribution formulas
func (s *RevenueAttributionService) GetAllFormulas(activeOnly bool) ([]AttributionFormula, error) {
	query := `
		SELECT id, name, description, formula_type, formula_config, 
		       attribution_window_days, min_watch_percentage, is_active, 
		       is_default, created_at, updated_at, created_by
		FROM revenue_attribution_formulas
	`
	if activeOnly {
		query += " WHERE is_active = true"
	}
	query += " ORDER BY is_default DESC, name ASC"

	rows, err := s.db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	formulas := []AttributionFormula{}
	for rows.Next() {
		formula := AttributionFormula{}
		var configJSON []byte

		err := rows.Scan(
			&formula.ID, &formula.Name, &formula.Description, &formula.FormulaType,
			&configJSON, &formula.AttributionWindowDays, &formula.MinWatchPercentage,
			&formula.IsActive, &formula.IsDefault, &formula.CreatedAt, &formula.UpdatedAt,
			&formula.CreatedBy,
		)
		if err != nil {
			log.Printf("Error scanning formula: %v", err)
			continue
		}

		// Parse formula config
		if len(configJSON) > 0 {
			json.Unmarshal(configJSON, &formula.FormulaConfig)
		}

		formulas = append(formulas, formula)
	}

	return formulas, nil
}

// UpdateFormula updates an existing formula
func (s *RevenueAttributionService) UpdateFormula(formulaID int, req UpdateFormulaRequest) error {
	// Build dynamic update query
	updates := []string{}
	args := []interface{}{}
	argPos := 1

	if req.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argPos))
		args = append(args, *req.Name)
		argPos++
	}
	if req.Description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argPos))
		args = append(args, *req.Description)
		argPos++
	}
	if req.FormulaConfig != nil {
		configJSON, _ := json.Marshal(*req.FormulaConfig)
		updates = append(updates, fmt.Sprintf("formula_config = $%d", argPos))
		args = append(args, configJSON)
		argPos++
	}
	if req.AttributionWindowDays != nil {
		updates = append(updates, fmt.Sprintf("attribution_window_days = $%d", argPos))
		args = append(args, *req.AttributionWindowDays)
		argPos++
	}
	if req.MinWatchPercentage != nil {
		updates = append(updates, fmt.Sprintf("min_watch_percentage = $%d", argPos))
		args = append(args, *req.MinWatchPercentage)
		argPos++
	}
	if req.IsActive != nil {
		updates = append(updates, fmt.Sprintf("is_active = $%d", argPos))
		args = append(args, *req.IsActive)
		argPos++
	}
	if req.IsDefault != nil {
		// If setting as default, unset all others first
		if *req.IsDefault {
			_, err := s.db.DB.Exec("UPDATE revenue_attribution_formulas SET is_default = false")
			if err != nil {
				return fmt.Errorf("failed to unset other defaults: %v", err)
			}
		}
		updates = append(updates, fmt.Sprintf("is_default = $%d", argPos))
		args = append(args, *req.IsDefault)
		argPos++
	}

	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Add updated_at
	updates = append(updates, "updated_at = CURRENT_TIMESTAMP")

	// Add formula ID as final parameter
	args = append(args, formulaID)

	query := fmt.Sprintf("UPDATE revenue_attribution_formulas SET %s WHERE id = $%d",
		joinStrings(updates, ", "), argPos)

	_, err := s.db.DB.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update formula: %v", err)
	}

	log.Printf("✅ Updated attribution formula ID: %d", formulaID)
	return nil
}

// DeleteFormula soft-deletes a formula (sets inactive)
func (s *RevenueAttributionService) DeleteFormula(formulaID int) error {
	_, err := s.db.DB.Exec(`
		UPDATE revenue_attribution_formulas 
		SET is_active = false, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, formulaID)
	if err != nil {
		return fmt.Errorf("failed to delete formula: %v", err)
	}

	log.Printf("🗑️ Deleted attribution formula ID: %d", formulaID)
	return nil
}

// ===========================
// ATTRIBUTION CALCULATION
// ===========================

// CalculateAttribution calculates revenue attribution for a new subscription
func (s *RevenueAttributionService) CalculateAttribution(userID int, subscriptionID string, subscriptionValue float64, formulaID *int) error {
	// Get formula (use default if not specified)
	var formula *AttributionFormula
	var err error

	if formulaID != nil {
		formula, err = s.GetFormula(*formulaID)
	} else {
		formula, err = s.getDefaultFormula()
	}
	if err != nil {
		return fmt.Errorf("failed to get formula: %v", err)
	}

	// Get user's video journey within attribution window
	journey, err := s.getUserConversionJourney(userID, time.Now(), formula.AttributionWindowDays, formula.MinWatchPercentage)
	if err != nil {
		return fmt.Errorf("failed to get user journey: %v", err)
	}

	if len(journey) == 0 {
		log.Printf("ℹ️ No qualifying videos in attribution window for user %d", userID)
		return nil
	}

	// Calculate attribution weights based on formula
	weights, err := s.calculateAttributionWeights(journey, formula)
	if err != nil {
		return fmt.Errorf("failed to calculate weights: %v", err)
	}

	// Create attribution records
	for i, step := range journey {
		weight := weights[i]
		attributedRevenue := subscriptionValue * weight

		attributionType := s.getAttributionType(i, len(journey))

		_, err := s.db.DB.Exec(`
			INSERT INTO video_revenue_attribution 
			(video_id, user_id, subscription_id, formula_id, attribution_type, 
			 attribution_weight, attributed_revenue, subscription_value, 
			 views_before_conversion, total_watch_time_seconds, 
			 last_view_before_conversion, conversion_timestamp)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (video_id, user_id, subscription_id, formula_id) DO NOTHING
		`, step.VideoID, userID, subscriptionID, formula.ID, attributionType,
			weight, attributedRevenue, subscriptionValue, len(journey),
			step.WatchDuration, step.ViewTimestamp, time.Now())

		if err != nil {
			log.Printf("Error creating attribution record: %v", err)
			continue
		}
	}

	// Update conversion metrics
	for _, step := range journey {
		s.updateVideoConversionMetrics(step.VideoID, formula.ID)
	}

	log.Printf("✅ Calculated attribution for user %d, subscription %s: %d videos, $%.2f total",
		userID, subscriptionID, len(journey), subscriptionValue)
	return nil
}

// getUserConversionJourney gets all qualifying video views for a user
func (s *RevenueAttributionService) getUserConversionJourney(userID int, conversionTime time.Time, windowDays int, minWatchPercentage float64) ([]ConversionJourneyStep, error) {
	windowStart := conversionTime.AddDate(0, 0, -windowDays)

	rows, err := s.db.DB.Query(`
		SELECT 
			vv.video_id,
			v.title,
			vv.created_at,
			vv.watch_duration,
		CASE 
			WHEN v.duration > 0 THEN (wh.total_watch_time::DECIMAL / v.duration * 100)
			ELSE 0
		END as watch_percentage,
		EXTRACT(EPOCH FROM ($1 - wh.first_watched_at)) / 3600 as hours_to_conversion
	FROM watch_history wh
	JOIN videos v ON v.id = wh.video_id
	WHERE wh.user_id = $2
	  AND wh.first_watched_at >= $3
	  AND wh.first_watched_at <= $1
	  AND (v.duration = 0 OR (wh.total_watch_time::DECIMAL / v.duration * 100) >= $4)
	ORDER BY wh.first_watched_at ASC
	`, conversionTime, userID, windowStart, minWatchPercentage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	journey := []ConversionJourneyStep{}
	position := 1
	for rows.Next() {
		step := ConversionJourneyStep{}
		err := rows.Scan(&step.VideoID, &step.VideoTitle, &step.ViewTimestamp,
			&step.WatchDuration, &step.WatchPercentage, &step.HoursToConversion)
		if err != nil {
			log.Printf("Error scanning journey step: %v", err)
			continue
		}
		step.Position = position
		journey = append(journey, step)
		position++
	}

	return journey, nil
}

// calculateAttributionWeights calculates weights based on formula type
func (s *RevenueAttributionService) calculateAttributionWeights(journey []ConversionJourneyStep, formula *AttributionFormula) ([]float64, error) {
	n := len(journey)
	weights := make([]float64, n)

	switch formula.FormulaType {
	case "last_touch":
		// 100% to last video
		weights[n-1] = 1.0

	case "first_touch":
		// 100% to first video
		weights[0] = 1.0

	case "linear":
		// Equal distribution
		weight := 1.0 / float64(n)
		for i := range weights {
			weights[i] = weight
		}

	case "time_decay":
		// Exponential decay based on time to conversion
		decayRate := 0.5 // Default
		if rate, ok := formula.FormulaConfig["decay_rate"].(float64); ok {
			decayRate = rate
		}

		totalWeight := 0.0
		for i, step := range journey {
			// More recent = higher weight
			weights[i] = math.Exp(-decayRate * step.HoursToConversion / 24.0)
			totalWeight += weights[i]
		}
		// Normalize to sum to 1
		for i := range weights {
			weights[i] /= totalWeight
		}

	case "position_based":
		// 40% first, 40% last, 20% distributed to middle
		firstWeight := 0.4
		lastWeight := 0.4
		middleWeight := 0.2

		// Get weights from config if available
		if fw, ok := formula.FormulaConfig["first_weight"].(float64); ok {
			firstWeight = fw
		}
		if lw, ok := formula.FormulaConfig["last_weight"].(float64); ok {
			lastWeight = lw
		}
		if mw, ok := formula.FormulaConfig["middle_weight"].(float64); ok {
			middleWeight = mw
		}

		if n == 1 {
			weights[0] = 1.0
		} else if n == 2 {
			weights[0] = firstWeight
			weights[1] = lastWeight
		} else {
			weights[0] = firstWeight
			weights[n-1] = lastWeight
			middlePerVideo := middleWeight / float64(n-2)
			for i := 1; i < n-1; i++ {
				weights[i] = middlePerVideo
			}
		}

	case "custom":
		// Custom formula evaluation would go here
		// For now, fall back to linear
		weight := 1.0 / float64(n)
		for i := range weights {
			weights[i] = weight
		}
	}

	return weights, nil
}

// getAttributionType determines the attribution type based on position
func (s *RevenueAttributionService) getAttributionType(position, total int) string {
	if total == 1 {
		return "single_touch"
	}
	if position == 0 {
		return "first_touch"
	}
	if position == total-1 {
		return "last_touch"
	}
	return "assisted"
}

// getDefaultFormula gets the default attribution formula
func (s *RevenueAttributionService) getDefaultFormula() (*AttributionFormula, error) {
	formula := &AttributionFormula{}
	var configJSON []byte

	err := s.db.DB.QueryRow(`
		SELECT id, name, description, formula_type, formula_config, 
		       attribution_window_days, min_watch_percentage, is_active, 
		       is_default, created_at, updated_at, created_by
		FROM revenue_attribution_formulas
		WHERE is_default = true AND is_active = true
		LIMIT 1
	`).Scan(
		&formula.ID, &formula.Name, &formula.Description, &formula.FormulaType,
		&configJSON, &formula.AttributionWindowDays, &formula.MinWatchPercentage,
		&formula.IsActive, &formula.IsDefault, &formula.CreatedAt, &formula.UpdatedAt,
		&formula.CreatedBy,
	)
	if err != nil {
		return nil, err
	}

	if len(configJSON) > 0 {
		json.Unmarshal(configJSON, &formula.FormulaConfig)
	}

	return formula, nil
}

// updateVideoConversionMetrics updates aggregated conversion metrics
func (s *RevenueAttributionService) updateVideoConversionMetrics(videoID, formulaID int) error {
	_, err := s.db.DB.Exec(`
		INSERT INTO video_conversion_metrics (video_id, formula_id, total_conversions, assisted_conversions, 
		                                       total_attributed_revenue, avg_revenue_per_conversion, 
		                                       total_qualified_views, conversion_rate, 
		                                       avg_time_to_conversion_hours, last_calculated)
		SELECT 
			$1,
		$2,
		COUNT(DISTINCT CASE WHEN attribution_type IN ('last_touch', 'single_touch') THEN subscription_id END),
		COUNT(DISTINCT CASE WHEN attribution_type = 'assisted' THEN subscription_id END),
		SUM(attributed_revenue),
		AVG(attributed_revenue),
		(SELECT SUM(view_count) FROM watch_history WHERE video_id = $1),
		CASE 
			WHEN (SELECT SUM(view_count) FROM watch_history WHERE video_id = $1) > 0 
			THEN COUNT(DISTINCT subscription_id)::DECIMAL / (SELECT SUM(view_count) FROM watch_history WHERE video_id = $1)
			ELSE 0
		END,
		AVG(EXTRACT(EPOCH FROM (conversion_timestamp - last_view_before_conversion)) / 3600),
		CURRENT_TIMESTAMP
		FROM video_revenue_attribution
		WHERE video_id = $1 AND formula_id = $2
		ON CONFLICT (video_id, formula_id) DO UPDATE SET
			total_conversions = EXCLUDED.total_conversions,
			assisted_conversions = EXCLUDED.assisted_conversions,
			total_attributed_revenue = EXCLUDED.total_attributed_revenue,
			avg_revenue_per_conversion = EXCLUDED.avg_revenue_per_conversion,
			total_qualified_views = EXCLUDED.total_qualified_views,
			conversion_rate = EXCLUDED.conversion_rate,
			avg_time_to_conversion_hours = EXCLUDED.avg_time_to_conversion_hours,
			last_calculated = CURRENT_TIMESTAMP
	`, videoID, formulaID)

	return err
}

// ===========================
// REPORTING
// ===========================

// GetVideoConversionMetrics gets conversion metrics for a specific video
func (s *RevenueAttributionService) GetVideoConversionMetrics(videoID, formulaID int) (*VideoConversionMetrics, error) {
	metrics := &VideoConversionMetrics{}

	err := s.db.DB.QueryRow(`
		SELECT 
			vcm.video_id,
			v.title,
			vcm.total_conversions,
			vcm.assisted_conversions,
			vcm.total_attributed_revenue,
			vcm.avg_revenue_per_conversion,
			vcm.total_qualified_views,
			vcm.conversion_rate,
			vcm.avg_time_to_conversion_hours,
			vcm.last_calculated
		FROM video_conversion_metrics vcm
		JOIN videos v ON v.id = vcm.video_id
		WHERE vcm.video_id = $1 AND vcm.formula_id = $2
	`, videoID, formulaID).Scan(
		&metrics.VideoID, &metrics.VideoTitle, &metrics.TotalConversions,
		&metrics.AssistedConversions, &metrics.TotalAttributedRevenue,
		&metrics.AvgRevenuePerConversion, &metrics.TotalQualifiedViews,
		&metrics.ConversionRate, &metrics.AvgTimeToConversionHours,
		&metrics.LastCalculated,
	)
	if err == sql.ErrNoRows {
		// No metrics yet, trigger calculation
		s.updateVideoConversionMetrics(videoID, formulaID)
		return s.GetVideoConversionMetrics(videoID, formulaID)
	}
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

// GetTopConvertingVideos gets the top converting videos
func (s *RevenueAttributionService) GetTopConvertingVideos(formulaID int, limit int, sortBy string) ([]VideoConversionMetrics, error) {
	if limit <= 0 {
		limit = 10
	}

	validSortFields := map[string]string{
		"revenue":     "total_attributed_revenue DESC",
		"conversions": "total_conversions DESC",
		"rate":        "conversion_rate DESC",
		"avg_revenue": "avg_revenue_per_conversion DESC",
	}
	orderBy := validSortFields["revenue"] // default
	if sort, ok := validSortFields[sortBy]; ok {
		orderBy = sort
	}

	rows, err := s.db.DB.Query(fmt.Sprintf(`
		SELECT 
			vcm.video_id,
			v.title,
			vcm.total_conversions,
			vcm.assisted_conversions,
			vcm.total_attributed_revenue,
			vcm.avg_revenue_per_conversion,
			vcm.total_qualified_views,
			vcm.conversion_rate,
			vcm.avg_time_to_conversion_hours,
			vcm.last_calculated
		FROM video_conversion_metrics vcm
		JOIN videos v ON v.id = vcm.video_id
		WHERE vcm.formula_id = $1
		ORDER BY %s
		LIMIT $2
	`, orderBy), formulaID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	videos := []VideoConversionMetrics{}
	for rows.Next() {
		metrics := VideoConversionMetrics{}
		err := rows.Scan(
			&metrics.VideoID, &metrics.VideoTitle, &metrics.TotalConversions,
			&metrics.AssistedConversions, &metrics.TotalAttributedRevenue,
			&metrics.AvgRevenuePerConversion, &metrics.TotalQualifiedViews,
			&metrics.ConversionRate, &metrics.AvgTimeToConversionHours,
			&metrics.LastCalculated,
		)
		if err != nil {
			log.Printf("Error scanning video metrics: %v", err)
			continue
		}
		videos = append(videos, metrics)
	}

	return videos, nil
}

// GetAttributionReport generates a comprehensive attribution report
func (s *RevenueAttributionService) GetAttributionReport(formulaID int, periodDays int) (*AttributionReport, error) {
	formula, err := s.GetFormula(formulaID)
	if err != nil {
		return nil, err
	}

	// Get top converting videos
	topVideos, err := s.GetTopConvertingVideos(formulaID, 10, "revenue")
	if err != nil {
		return nil, err
	}

	// Get recent attributions
	cutoff := time.Now().AddDate(0, 0, -periodDays)
	rows, err := s.db.DB.Query(`
		SELECT 
			vra.id, vra.video_id, v.title, vra.user_id, vra.subscription_id,
			vra.formula_id, vra.attribution_type, vra.attribution_weight,
			vra.attributed_revenue, vra.subscription_value, vra.views_before_conversion,
			vra.total_watch_time_seconds, vra.last_view_before_conversion,
			vra.conversion_timestamp, vra.created_at
		FROM video_revenue_attribution vra
		JOIN videos v ON v.id = vra.video_id
		WHERE vra.formula_id = $1 AND vra.conversion_timestamp >= $2
		ORDER BY vra.conversion_timestamp DESC
		LIMIT 50
	`, formulaID, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recentAttributions := []VideoAttribution{}
	for rows.Next() {
		attr := VideoAttribution{}
		err := rows.Scan(
			&attr.ID, &attr.VideoID, &attr.VideoTitle, &attr.UserID, &attr.SubscriptionID,
			&attr.FormulaID, &attr.AttributionType, &attr.AttributionWeight,
			&attr.AttributedRevenue, &attr.SubscriptionValue, &attr.ViewsBeforeConversion,
			&attr.TotalWatchTimeSeconds, &attr.LastViewBeforeConversion,
			&attr.ConversionTimestamp, &attr.CreatedAt,
		)
		if err != nil {
			log.Printf("Error scanning attribution: %v", err)
			continue
		}
		recentAttributions = append(recentAttributions, attr)
	}

	// Calculate totals
	var totalRevenue float64
	var totalConversions int
	var videosWithImpact int
	for _, video := range topVideos {
		totalRevenue += video.TotalAttributedRevenue
		totalConversions += video.TotalConversions
		if video.TotalConversions > 0 {
			videosWithImpact++
		}
	}

	report := &AttributionReport{
		FormulaName:        formula.Name,
		ReportPeriodDays:   periodDays,
		TotalRevenue:       totalRevenue,
		TotalConversions:   totalConversions,
		VideosWithImpact:   videosWithImpact,
		TopVideos:          topVideos,
		RecentAttributions: recentAttributions,
		GeneratedAt:        time.Now(),
	}

	return report, nil
}

// ===========================
// UTILITY
// ===========================

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
