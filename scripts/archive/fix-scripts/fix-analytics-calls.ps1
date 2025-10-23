# Fix analytics service to use cross-braid calls

Write-Host "🔄 Fixing analytics cross-braid calls..."

$file = "backend/analytics/services/analytics.go"
$content = Get-Content $file -Raw

# Fix function calls to use imported braid models
$content = $content -replace 's\.db\.GetUserCount\(\)', 'authModels.GetUserCount(s.db)'
$content = $content -replace 's\.db\.GetVideoCount\(\)', 'videoModels.GetVideoCount(s.db)'
$content = $content -replace 's\.db\.GetTotalViews\(\)', 'videoModels.GetTotalViews(s.db)'
$content = $content -replace 's\.db\.GetTotalLikes\(\)', 'videoModels.GetTotalLikes(s.db)'
$content = $content -replace 's\.db\.GetActiveSubscriptions\(\)', 'subModels.GetActiveSubscriptions(s.db)'
$content = $content -replace 's\.db\.GetRecentActivity\(', 'GetRecentActivity(s.db, '

# Add SystemHealth and SystemMetrics types
$typeDefs = @"

// SystemHealth represents system health metrics
type SystemHealth struct {
	Status      string ``json:"status"``
	Uptime      int64  ``json:"uptime"``
	CPUUsage    float64 ``json:"cpu_usage"``
	MemoryUsage float64 ``json:"memory_usage"``
	DiskUsage   float64 ``json:"disk_usage"``
}

// SystemMetrics represents detailed system metrics  
type SystemMetrics struct {
	Timestamp   int64   ``json:"timestamp"``
	CPUPercent  float64 ``json:"cpu_percent"``
	MemoryUsed  uint64  ``json:"memory_used"``
	MemoryTotal uint64  ``json:"memory_total"``
	DiskUsed    uint64  ``json:"disk_used"``
	DiskTotal   uint64  ``json:"disk_total"``
}

"@

# Insert types after package and imports
$content = $content -replace '(\nimport \([^)]+\))', "`$1$typeDefs"

Set-Content -Path $file -Value $content -NoNewline

Write-Host "✅ Analytics fixed!"

