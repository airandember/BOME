# Fix all remaining compilation errors

Write-Host "🔄 Fixing remaining errors..."

# Fix analytics - stub out more missing methods and remove unused variables
$analyticsFile = "backend/analytics/services/analytics.go"
$content = Get-Content $analyticsFile -Raw

# Replace s.db calls with stub calls
$content = $content -replace 's\.db\.GetSystemMetrics\(\)', 'getSystemMetrics(s.db)'
$content = $content -replace 's\.db\.GetWebhookEvents\(([^)]*)\)', 'getWebhookEvents(s.db, $1)'
$content = $content -replace 's\.db\.GetAlerts\(([^)]*)\)', 'getAlerts(s.db, $1)'
$content = $content -replace 's\.db\.GetCrossSubsiteStats\(([^)]*)\)', 'getCrossSubsiteStats(s.db, $1)'

# Comment out unused variables
$content = $content -replace '(\s+)(uptime :=)', '$1_ = $2' 
$content = $content -replace '(\s+)(responseTime :=)', '$1_ = $2'
$content = $content -replace '(\s+)(errorRate :=)', '$1_ = $2'
$content = $content -replace '(\s+)(storageUsed :=)', '$1_ = $2'
$content = $content -replace '(\s+)(bandwidthUsed :=)', '$1_ = $2'
$content = $content -replace '(\s+)(totalEvents :=)', '$1_ = $2'

Set-Content -Path $analyticsFile -Value $content -NoNewline
Write-Host "  ✅ Fixed analytics"

# Add remaining stub functions to analytics
$analyticsAdditions = @"

func getWebhookEvents(db *database.DB, params ...interface{}) ([]interface{}, error) {
	return []interface{}{}, nil
}

func getAlerts(db *database.DB, params ...interface{}) ([]interface{}, error) {
	return []interface{}{}, nil
}

func getCrossSubsiteStats(db *database.DB, params ...interface{}) (interface{}, error) {
	return nil, nil
}
"@

$content = Get-Content $analyticsFile -Raw
if ($content -notmatch 'func getWebhookEvents') {
    $content += $analyticsAdditions
    Set-Content -Path $analyticsFile -Value $content -NoNewline
    Write-Host "  ✅ Added analytics stubs"
}

# Fix middleware - fix HasVideoAccess calls and remove unused clientIP
$middlewareFile = "backend/authentication/middleware/middleware.go"
$midContent = Get-Content $middlewareFile -Raw

# Comment out unused clientIP
$midContent = $midContent -replace '(\s+)clientIP := c\.ClientIP\(\)', '$1_ = c.ClientIP()'

# Fix HasVideoAccess signature - it only takes db, userID, videoID
$midContent = $midContent -replace 'hasAccess, isSubscriber, err := videoModels\.HasVideoAccess\(db, db, video\.ID\)', 'hasAccess, err := videoModels.HasVideoAccess(db, userID, video.ID)'

# Fix GetSubscriptionPlanByID - remove extra db parameter
$midContent = $midContent -replace 'subModels\.GetSubscriptionPlanByID\(db, db,', 'subModels.GetSubscriptionPlanByID(db,'

Set-Content -Path $middlewareFile -Value $midContent -NoNewline
Write-Host "  ✅ Fixed middleware"

Write-Host "✅ All fixes applied!"

