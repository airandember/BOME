# Test script for Bunny.net sync
Write-Host "Testing Bunny.net sync..." -ForegroundColor Green

# Test 1: Check current videos
Write-Host "`n1. Current videos in database:" -ForegroundColor Yellow
try {
    $videos = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/videos"
    Write-Host "Total videos: $($videos.pagination.total)"
    foreach ($video in $videos.videos) {
        Write-Host "  - $($video.title) (ID: $($video.id))"
    }
} catch {
    Write-Host "Error getting videos: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 2: Check Bunny.net configuration
Write-Host "`n2. Bunny.net configuration:" -ForegroundColor Yellow
try {
    $bunnyConfig = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/test/bunny"
    Write-Host "Configuration status: $($bunnyConfig.summary.all_configured)"
    Write-Host "Ready for testing: $($bunnyConfig.summary.ready_for_testing)"
} catch {
    Write-Host "Error checking Bunny.net config: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 3: Try to sync videos
Write-Host "`n3. Attempting to sync videos from Bunny.net..." -ForegroundColor Yellow
try {
    $syncResult = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/test/sync-bunny-videos" -Method POST
    Write-Host "Sync result: $($syncResult.message)" -ForegroundColor Green
    Write-Host "Total videos found: $($syncResult.total_videos)"
    Write-Host "Synced: $($syncResult.synced)"
    Write-Host "Skipped: $($syncResult.skipped)"
    Write-Host "Errors: $($syncResult.errors)"
} catch {
    Write-Host "Error syncing videos: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Trying alternative endpoint..." -ForegroundColor Yellow
    
    # Try the admin endpoint
    try {
        $syncResult = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/admin/sync-bunny-videos" -Method POST
        Write-Host "Admin sync result: $($syncResult.message)" -ForegroundColor Green
    } catch {
        Write-Host "Admin sync also failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 4: Check videos after sync
Write-Host "`n4. Videos after sync attempt:" -ForegroundColor Yellow
try {
    $videosAfter = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/videos"
    Write-Host "Total videos: $($videosAfter.pagination.total)"
    foreach ($video in $videosAfter.videos) {
        Write-Host "  - $($video.title) (ID: $($video.id), Status: $($video.status))"
    }
} catch {
    Write-Host "Error getting videos after sync: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`nTest completed!" -ForegroundColor Green 