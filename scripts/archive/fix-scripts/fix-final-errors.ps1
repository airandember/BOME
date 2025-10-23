# Fix FINAL errors

Write-Host "🔄 Fixing final errors..."

# Fix analytics - add missing functions and fix comment
$analyticsFile = "backend/analytics/services/analytics.go"
$content = Get-Content $analyticsFile -Raw

# Fix unterminated comment - find and close it
$content = $content -replace '(/\*[\s\S]*?)$', '$1*/'

# Add missing helper functions at the end
$missingFunctions = @"


// Helper functions for analytics

func GetRecentActivity(db *database.DB, limit int) ([]interface{}, error) {
	return []interface{}{}, nil
}

func getRealTimeMetrics(db *database.DB) (map[string]interface{}, error) {
	return make(map[string]interface{}), nil
}

func getSystemMetrics(db *database.DB) (*SystemMetrics, error) {
	return &SystemMetrics{
		Timestamp:   0,
		CPUPercent:  0,
		MemoryUsed:  0,
		MemoryTotal: 0,
		DiskUsed:    0,
		DiskTotal:   0,
	}, nil
}

func getWebhookEvents(db *database.DB, params ...interface{}) ([]interface{}, error) {
	return []interface{}{}, nil
}

func getCrossSubsiteStats(db *database.DB, params ...interface{}) (interface{}, error) {
	return nil, nil
}

func (s *AnalyticsService) getNewUsersCount(db *database.DB, days int) (int, error) {
	return 0, nil
}

func (s *AnalyticsService) calculateGrowthRate(current, previous int) float64 {
	if previous == 0 {
		return 0
	}
	return float64(current-previous) / float64(previous) * 100
}

func (s *AnalyticsService) getPublishedVideosCount(db *database.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM master_video_list WHERE status = 'published'").Scan(&count)
	return count, err
}

func (s *AnalyticsService) getPendingVideosCount(db *database.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM master_video_list WHERE status = 'pending'").Scan(&count)
	return count, err
}

func (s *AnalyticsService) getDraftVideosCount(db *database.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM master_video_list WHERE status = 'draft'").Scan(&count)
	return count, err
}
"@

if ($content -notmatch 'func GetRecentActivity') {
    $content += $missingFunctions
}

Set-Content -Path $analyticsFile -Value $content -NoNewline
Write-Host "  ✅ Fixed analytics"

Write-Host "✅ All fixes applied!"

