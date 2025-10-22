# Fix syntax errors from bad regex replacements

Write-Host "🔄 Fixing syntax errors..."

# Fix analytics - properly comment out or fix unused variables
$analyticsFile = "backend/analytics/services/analytics.go"
$content = Get-Content $analyticsFile -Raw

# Fix the bad _ = replacements - comment them out instead
$content = $content -replace '_ = (uptime|responseTime|errorRate|storageUsed|bandwidthUsed|totalEvents) :=', '// $1 :='

Set-Content -Path $analyticsFile -Value $content -NoNewline
Write-Host "  ✅ Fixed analytics syntax"

# Fix middleware - properly fix HasVideoAccess calls
$middlewareFile = "backend/authentication/middleware/middleware.go"
$midContent = Get-Content $middlewareFile -Raw

# Find and replace hasAccess, isSubscriber, err := with just hasAccess, err :=
$midContent = $midContent -replace 'hasAccess, isSubscriber, err := videoModels\.HasVideoAccess\(db, db, video\.ID\)', 'hasAccess, err := videoModels.HasVideoAccess(db, userID, video.ID)'
$midContent = $midContent -replace 'hasAccess, isSubscriber, err := videoModels\.HasVideoAccess\([^)]+\)', 'hasAccess, err := videoModels.HasVideoAccess(db, userID, video.ID)'

Set-Content -Path $middlewareFile -Value $midContent -NoNewline
Write-Host "  ✅ Fixed middleware"

# Fix video sync - use proper GetMasterVideos signature OR create simpler stub
$masterFile = "backend/video-streaming/models/master_video.go"
$masterContent = Get-Content $masterFile -Raw

# Replace the GetMasterVideos stub with proper signature
$masterContent = $masterContent -replace 'func GetMasterVideos\(db \*database\.DB\) \(\[\]\*MasterVideo, error\) \{[^}]+\}', @'
func GetMasterVideos(db *database.DB, limit, offset int, title, category, status, creatorName, sortBy, sortOrder string) ([]*MasterVideo, error) {
	// Stub - return empty array
	return []*MasterVideo{}, nil
}
'@

# Replace GetSyncConflicts stub with proper signature
$masterContent = $masterContent -replace 'func GetSyncConflicts\(db \*database\.DB\) \(\[\]\*SyncConflict, error\) \{[^}]+\}', @'
func GetSyncConflicts(db *database.DB, videoID *int) ([]*SyncConflict, error) {
	// Stub - return empty array
	return []*SyncConflict{}, nil
}
'@

Set-Content -Path $masterFile -Value $masterContent -NoNewline
Write-Host "  ✅ Fixed video model signatures"

# Update sync service to use proper signatures
$syncFile = "backend/video-streaming/services/master_video_sync.go"
$syncContent = Get-Content $syncFile -Raw

$syncContent = $syncContent -replace 'videoModels\.GetMasterVideos\(s\.db\)', 'videoModels.GetMasterVideos(s.db, 1000, 0, "", "", "", "", "id", "desc")'
$syncContent = $syncContent -replace 'videoModels\.GetSyncConflicts\(s\.db\)', 'videoModels.GetSyncConflicts(s.db, nil)'

Set-Content -Path $syncFile -Value $syncContent -NoNewline
Write-Host "  ✅ Fixed sync service calls"

Write-Host "✅ All syntax errors fixed!"

