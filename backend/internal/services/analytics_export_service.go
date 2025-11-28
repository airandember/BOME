package services

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"bome-backend/internal/database"
)

// AnalyticsExportService handles exporting analytics data to various formats
type AnalyticsExportService struct {
	db *database.DB
}

// NewAnalyticsExportService creates a new analytics export service
func NewAnalyticsExportService(db *database.DB) *AnalyticsExportService {
	return &AnalyticsExportService{db: db}
}

// ExportRequest represents a request to export analytics data
type ExportRequest struct {
	ExportType string                 `json:"export_type"` // "video_analytics", "revenue_attribution", "user_stats"
	Format     string                 `json:"format"`      // "csv", "json"
	StartDate  time.Time              `json:"start_date"`
	EndDate    time.Time              `json:"end_date"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
}

// ExportResult represents the result of an export operation
type ExportResult struct {
	Content     string    `json:"content"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	RowCount    int       `json:"row_count"`
	ExportedAt  time.Time `json:"exported_at"`
}

// ===========================
// VIDEO ANALYTICS EXPORTS
// ===========================

// ExportVideoAnalytics exports video analytics data
func (s *AnalyticsExportService) ExportVideoAnalytics(req ExportRequest) (*ExportResult, error) {
	// Query video analytics data
	rows, err := s.db.DB.Query(`
		SELECT 
		v.id as video_id,
		v.title,
		COUNT(DISTINCT COALESCE(wh.user_id::text, wh.session_id)) as unique_viewers,
		SUM(wh.view_count) as total_views,
		AVG(wh.total_watch_time) as avg_watch_duration,
		SUM(wh.total_watch_time) as total_watch_time,
		AVG(CASE WHEN wh.completed THEN 100.0 ELSE 0.0 END) as completion_rate,
		COUNT(CASE WHEN wh.completed THEN 1 END) as completed_views,
		MIN(wh.first_watched_at) as first_view,
		MAX(wh.last_watched_at) as last_view
	FROM videos v
	LEFT JOIN watch_history wh ON wh.video_id = v.id
	WHERE wh.last_watched_at >= $1 AND wh.last_watched_at <= $2
	GROUP BY v.id, v.title
	ORDER BY total_views DESC
	`, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query analytics: %v", err)
	}
	defer rows.Close()

	// Prepare data for export
	var data [][]string
	headers := []string{
		"Video ID", "Title", "Unique Viewers", "Total Views",
		"Avg Watch Duration (seconds)", "Total Watch Time (seconds)",
		"Completion Rate (%)", "Completed Views", "First View", "Last View",
	}
	data = append(data, headers)

	rowCount := 0
	for rows.Next() {
		var videoID int
		var title string
		var uniqueViewers, totalViews, completedViews int
		var avgWatchDuration, totalWatchTime, completionRate float64
		var firstView, lastView time.Time

		err := rows.Scan(&videoID, &title, &uniqueViewers, &totalViews,
			&avgWatchDuration, &totalWatchTime, &completionRate, &completedViews,
			&firstView, &lastView)
		if err != nil {
			continue
		}

		row := []string{
			strconv.Itoa(videoID),
			title,
			strconv.Itoa(uniqueViewers),
			strconv.Itoa(totalViews),
			fmt.Sprintf("%.2f", avgWatchDuration),
			fmt.Sprintf("%.2f", totalWatchTime),
			fmt.Sprintf("%.2f", completionRate),
			strconv.Itoa(completedViews),
			firstView.Format("2006-01-02 15:04:05"),
			lastView.Format("2006-01-02 15:04:05"),
		}
		data = append(data, row)
		rowCount++
	}

	// Convert to requested format
	if req.Format == "csv" {
		content, err := s.convertToCSV(data)
		if err != nil {
			return nil, err
		}

		filename := fmt.Sprintf("video_analytics_%s_to_%s.csv",
			req.StartDate.Format("2006-01-02"),
			req.EndDate.Format("2006-01-02"))

		return &ExportResult{
			Content:     content,
			Filename:    filename,
			ContentType: "text/csv",
			RowCount:    rowCount,
			ExportedAt:  time.Now(),
		}, nil
	}

	return nil, fmt.Errorf("unsupported format: %s", req.Format)
}

// ExportTrendingVideos exports trending videos data
func (s *AnalyticsExportService) ExportTrendingVideos(limit int) (*ExportResult, error) {
	rows, err := s.db.DB.Query(`
		SELECT 
		v.id,
		v.title,
		SUM(CASE WHEN wh.last_watched_at >= NOW() - INTERVAL '24 hours' THEN wh.view_count END) as last_24h_views,
		SUM(CASE WHEN wh.last_watched_at >= NOW() - INTERVAL '7 days' THEN wh.view_count END) as last_7d_views,
		(SUM(CASE WHEN wh.last_watched_at >= NOW() - INTERVAL '24 hours' THEN wh.view_count END) * 2.0 +
		 SUM(CASE WHEN wh.last_watched_at >= NOW() - INTERVAL '7 days' THEN wh.view_count END) * 0.5) as trending_score
	FROM videos v
	LEFT JOIN watch_history wh ON wh.video_id = v.id
	WHERE wh.last_watched_at >= NOW() - INTERVAL '7 days'
	GROUP BY v.id, v.title
	HAVING SUM(wh.view_count) > 0
	ORDER BY trending_score DESC
	LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data [][]string
	headers := []string{"Video ID", "Title", "24h Views", "7d Views", "Trending Score"}
	data = append(data, headers)

	rowCount := 0
	for rows.Next() {
		var videoID, views24h, views7d int
		var title string
		var trendingScore float64

		err := rows.Scan(&videoID, &title, &views24h, &views7d, &trendingScore)
		if err != nil {
			continue
		}

		row := []string{
			strconv.Itoa(videoID),
			title,
			strconv.Itoa(views24h),
			strconv.Itoa(views7d),
			fmt.Sprintf("%.2f", trendingScore),
		}
		data = append(data, row)
		rowCount++
	}

	content, err := s.convertToCSV(data)
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("trending_videos_%s.csv", time.Now().Format("2006-01-02"))

	return &ExportResult{
		Content:     content,
		Filename:    filename,
		ContentType: "text/csv",
		RowCount:    rowCount,
		ExportedAt:  time.Now(),
	}, nil
}

// ===========================
// REVENUE ATTRIBUTION EXPORTS
// ===========================

// ExportRevenueAttribution exports revenue attribution data
func (s *AnalyticsExportService) ExportRevenueAttribution(formulaID int, periodDays int) (*ExportResult, error) {
	cutoff := time.Now().AddDate(0, 0, -periodDays)

	rows, err := s.db.DB.Query(`
		SELECT 
			vra.video_id,
			v.title,
			vra.user_id,
			vra.subscription_id,
			vra.attribution_type,
			vra.attribution_weight,
			vra.attributed_revenue,
			vra.subscription_value,
			vra.views_before_conversion,
			vra.total_watch_time_seconds,
			vra.conversion_timestamp,
			f.name as formula_name
		FROM video_revenue_attribution vra
		JOIN videos v ON v.id = vra.video_id
		JOIN revenue_attribution_formulas f ON f.id = vra.formula_id
		WHERE vra.formula_id = $1 
		  AND vra.conversion_timestamp >= $2
		ORDER BY vra.conversion_timestamp DESC
	`, formulaID, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data [][]string
	headers := []string{
		"Video ID", "Video Title", "User ID", "Subscription ID",
		"Attribution Type", "Attribution Weight", "Attributed Revenue",
		"Subscription Value", "Views Before Conversion", "Watch Time (seconds)",
		"Conversion Date", "Formula Name",
	}
	data = append(data, headers)

	rowCount := 0
	for rows.Next() {
		var videoID, userID, viewsBeforeConversion, watchTimeSeconds int
		var videoTitle, subscriptionID, attributionType, formulaName string
		var attributionWeight, attributedRevenue, subscriptionValue float64
		var conversionTimestamp time.Time

		err := rows.Scan(&videoID, &videoTitle, &userID, &subscriptionID,
			&attributionType, &attributionWeight, &attributedRevenue,
			&subscriptionValue, &viewsBeforeConversion, &watchTimeSeconds,
			&conversionTimestamp, &formulaName)
		if err != nil {
			continue
		}

		row := []string{
			strconv.Itoa(videoID),
			videoTitle,
			strconv.Itoa(userID),
			subscriptionID,
			attributionType,
			fmt.Sprintf("%.4f", attributionWeight),
			fmt.Sprintf("%.2f", attributedRevenue),
			fmt.Sprintf("%.2f", subscriptionValue),
			strconv.Itoa(viewsBeforeConversion),
			strconv.Itoa(watchTimeSeconds),
			conversionTimestamp.Format("2006-01-02 15:04:05"),
			formulaName,
		}
		data = append(data, row)
		rowCount++
	}

	content, err := s.convertToCSV(data)
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("revenue_attribution_%ddays_%s.csv",
		periodDays,
		time.Now().Format("2006-01-02"))

	return &ExportResult{
		Content:     content,
		Filename:    filename,
		ContentType: "text/csv",
		RowCount:    rowCount,
		ExportedAt:  time.Now(),
	}, nil
}

// ExportTopConvertingVideos exports top converting videos data
func (s *AnalyticsExportService) ExportTopConvertingVideos(formulaID int, limit int) (*ExportResult, error) {
	rows, err := s.db.DB.Query(`
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
			f.name as formula_name
		FROM video_conversion_metrics vcm
		JOIN videos v ON v.id = vcm.video_id
		JOIN revenue_attribution_formulas f ON f.id = vcm.formula_id
		WHERE vcm.formula_id = $1
		ORDER BY vcm.total_attributed_revenue DESC
		LIMIT $2
	`, formulaID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data [][]string
	headers := []string{
		"Video ID", "Title", "Total Conversions", "Assisted Conversions",
		"Total Attributed Revenue", "Avg Revenue Per Conversion",
		"Total Qualified Views", "Conversion Rate", "Avg Time to Conversion (hours)",
		"Formula Name",
	}
	data = append(data, headers)

	rowCount := 0
	for rows.Next() {
		var videoID, totalConversions, assistedConversions, totalQualifiedViews int
		var title, formulaName string
		var totalRevenue, avgRevenue, conversionRate, avgTimeToConversion float64

		err := rows.Scan(&videoID, &title, &totalConversions, &assistedConversions,
			&totalRevenue, &avgRevenue, &totalQualifiedViews, &conversionRate,
			&avgTimeToConversion, &formulaName)
		if err != nil {
			continue
		}

		row := []string{
			strconv.Itoa(videoID),
			title,
			strconv.Itoa(totalConversions),
			strconv.Itoa(assistedConversions),
			fmt.Sprintf("%.2f", totalRevenue),
			fmt.Sprintf("%.2f", avgRevenue),
			strconv.Itoa(totalQualifiedViews),
			fmt.Sprintf("%.4f", conversionRate),
			fmt.Sprintf("%.2f", avgTimeToConversion),
			formulaName,
		}
		data = append(data, row)
		rowCount++
	}

	content, err := s.convertToCSV(data)
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("top_converting_videos_%s.csv", time.Now().Format("2006-01-02"))

	return &ExportResult{
		Content:     content,
		Filename:    filename,
		ContentType: "text/csv",
		RowCount:    rowCount,
		ExportedAt:  time.Now(),
	}, nil
}

// ===========================
// USER STATISTICS EXPORTS
// ===========================

// ExportUserWatchStats exports aggregated user watch statistics
func (s *AnalyticsExportService) ExportUserWatchStats() (*ExportResult, error) {
	rows, err := s.db.DB.Query(`
		SELECT 
		u.id,
		u.email,
		u.username,
		COUNT(DISTINCT wh.video_id) as videos_watched,
		COUNT(CASE WHEN wh.completed THEN 1 END) as videos_completed,
		SUM(wh.total_watch_time) as total_watch_seconds,
		COUNT(DISTINCT DATE(wh.last_watched_at)) as days_active,
		MIN(wh.first_watched_at) as first_view,
		MAX(wh.last_watched_at) as last_view
	FROM users u
	LEFT JOIN watch_history wh ON wh.user_id = u.id
	GROUP BY u.id, u.email, u.username
	HAVING COUNT(wh.id) > 0
	ORDER BY total_watch_seconds DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data [][]string
	headers := []string{
		"User ID", "Email", "Username", "Videos Watched", "Videos Completed",
		"Total Watch Time (seconds)", "Days Active", "First View", "Last View",
	}
	data = append(data, headers)

	rowCount := 0
	for rows.Next() {
		var userID, videosWatched, videosCompleted, totalWatchSeconds, daysActive int
		var email, username string
		var firstView, lastView time.Time

		err := rows.Scan(&userID, &email, &username, &videosWatched, &videosCompleted,
			&totalWatchSeconds, &daysActive, &firstView, &lastView)
		if err != nil {
			continue
		}

		row := []string{
			strconv.Itoa(userID),
			email,
			username,
			strconv.Itoa(videosWatched),
			strconv.Itoa(videosCompleted),
			strconv.Itoa(totalWatchSeconds),
			strconv.Itoa(daysActive),
			firstView.Format("2006-01-02 15:04:05"),
			lastView.Format("2006-01-02 15:04:05"),
		}
		data = append(data, row)
		rowCount++
	}

	content, err := s.convertToCSV(data)
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("user_watch_stats_%s.csv", time.Now().Format("2006-01-02"))

	return &ExportResult{
		Content:     content,
		Filename:    filename,
		ContentType: "text/csv",
		RowCount:    rowCount,
		ExportedAt:  time.Now(),
	}, nil
}

// ===========================
// DAILY/WEEKLY REPORTS
// ===========================

// ExportDailyReport exports a comprehensive daily analytics report
func (s *AnalyticsExportService) ExportDailyReport(date time.Time) (*ExportResult, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	rows, err := s.db.DB.Query(`
		SELECT 
		v.id,
		v.title,
		COUNT(DISTINCT COALESCE(wh.user_id::text, wh.session_id)) as unique_viewers,
		SUM(wh.view_count) as total_views,
		AVG(wh.total_watch_time) as avg_watch_duration,
		SUM(wh.total_watch_time) as total_watch_time,
		COUNT(CASE WHEN wh.completed THEN 1 END) as completed_views
	FROM videos v
	LEFT JOIN watch_history wh ON wh.video_id = v.id
	WHERE wh.last_watched_at >= $1 AND wh.last_watched_at < $2
	GROUP BY v.id, v.title
	HAVING SUM(wh.view_count) > 0
	ORDER BY total_views DESC
	`, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data [][]string
	headers := []string{
		"Date", "Video ID", "Title", "Unique Viewers", "Total Views",
		"Avg Watch Duration (seconds)", "Total Watch Time (seconds)", "Completed Views",
	}
	data = append(data, headers)

	dateStr := date.Format("2006-01-02")
	rowCount := 0
	for rows.Next() {
		var videoID, uniqueViewers, totalViews, completedViews int
		var title string
		var avgWatchDuration, totalWatchTime float64

		err := rows.Scan(&videoID, &title, &uniqueViewers, &totalViews,
			&avgWatchDuration, &totalWatchTime, &completedViews)
		if err != nil {
			continue
		}

		row := []string{
			dateStr,
			strconv.Itoa(videoID),
			title,
			strconv.Itoa(uniqueViewers),
			strconv.Itoa(totalViews),
			fmt.Sprintf("%.2f", avgWatchDuration),
			fmt.Sprintf("%.2f", totalWatchTime),
			strconv.Itoa(completedViews),
		}
		data = append(data, row)
		rowCount++
	}

	content, err := s.convertToCSV(data)
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("daily_report_%s.csv", dateStr)

	return &ExportResult{
		Content:     content,
		Filename:    filename,
		ContentType: "text/csv",
		RowCount:    rowCount,
		ExportedAt:  time.Now(),
	}, nil
}

// ===========================
// HELPER FUNCTIONS
// ===========================

// convertToCSV converts 2D string array to CSV format
func (s *AnalyticsExportService) convertToCSV(data [][]string) (string, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	for _, row := range data {
		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("failed to write CSV row: %v", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("CSV writer error: %v", err)
	}

	return builder.String(), nil
}
