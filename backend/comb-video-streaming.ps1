# Comb All Video Streaming Strands

Write-Host ""
Write-Host "VIDEO STREAMING BRAID - COMPLETE COMBING" -ForegroundColor Magenta
Write-Host "==========================================" -ForegroundColor Magenta
Write-Host ""

# STRAND 1: Video Upload & Processing
Write-Host "STRAND 1: Video Upload and Processing" -ForegroundColor Yellow
Write-Host "---------------------------------------" -ForegroundColor Yellow

$hasUploadHandler = Select-String -Path "video-streaming/handlers/*.go" -Pattern "func.*Upload|func.*ProcessVideo" -Quiet
if ($hasUploadHandler) { 
    Write-Host "  [OK] Upload handlers found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Upload handlers missing" -ForegroundColor Red 
}

Write-Host ""

# STRAND 2: Video Streaming Delivery
Write-Host "STRAND 2: Video Streaming Delivery" -ForegroundColor Yellow
Write-Host "-----------------------------------" -ForegroundColor Yellow

$hasStreamHandler = Select-String -Path "video-streaming/handlers/*.go" -Pattern "func.*Stream|func.*GetVideo" -Quiet
if ($hasStreamHandler) { 
    Write-Host "  [OK] Streaming handlers found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Streaming handlers missing" -ForegroundColor Red 
}

Write-Host ""

# STRAND 3: Bunny.net CDN Integration
Write-Host "STRAND 3: Bunny.net CDN Integration" -ForegroundColor Yellow
Write-Host "------------------------------------" -ForegroundColor Yellow

if (Test-Path "video-streaming/services/bunny.go") {
    Write-Host "  [OK] bunny.go service exists" -ForegroundColor Green
    
    $hasBunnyFuncs = Select-String -Path "video-streaming/services/bunny.go" -Pattern "func.*" | Measure-Object
    Write-Host "  [INFO] Found $($hasBunnyFuncs.Count) functions in bunny.go" -ForegroundColor Cyan
} else {
    Write-Host "  [SPLIT-END] bunny.go service missing" -ForegroundColor Red
}

if (Test-Path "video-streaming/services/bunny_optimized.go") {
    Write-Host "  [OK] bunny_optimized.go exists" -ForegroundColor Green
} else {
    Write-Host "  [SPLIT-END] bunny_optimized.go missing" -ForegroundColor Yellow
}

if (Test-Path "video-streaming/services/master_video_sync.go") {
    Write-Host "  [OK] master_video_sync.go exists" -ForegroundColor Green
} else {
    Write-Host "  [SPLIT-END] master_video_sync.go missing" -ForegroundColor Red
}

Write-Host ""

# STRAND 4: Video Access Control
Write-Host "STRAND 4: Video Access Control" -ForegroundColor Yellow
Write-Host "-------------------------------" -ForegroundColor Yellow

$hasAccessControl = Select-String -Path "video-streaming/models/*.go" -Pattern "func.*HasVideoAccess|func.*CheckAccess" -Quiet
if ($hasAccessControl) { 
    Write-Host "  [OK] Access control functions found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Access control missing" -ForegroundColor Red 
}

Write-Host ""

# STRAND 5: Video Analytics
Write-Host "STRAND 5: Video Analytics" -ForegroundColor Yellow
Write-Host "-------------------------" -ForegroundColor Yellow

$hasAnalytics = Select-String -Path "video-streaming/models/*.go" -Pattern "func.*Track|func.*Analytics|func.*View" -Quiet
if ($hasAnalytics) { 
    Write-Host "  [OK] Analytics tracking found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Analytics tracking missing" -ForegroundColor Yellow 
}

Write-Host ""

# Check models
Write-Host "MODEL FILES:" -ForegroundColor Cyan
if (Test-Path "video-streaming/models/video.go") {
    Write-Host "  [OK] video.go model exists" -ForegroundColor Green
} else {
    Write-Host "  [SPLIT-END] video.go missing" -ForegroundColor Red
}

if (Test-Path "video-streaming/models/master_video.go") {
    Write-Host "  [OK] master_video.go model exists" -ForegroundColor Green
} else {
    Write-Host "  [SPLIT-END] master_video.go missing" -ForegroundColor Red
}

Write-Host ""

# Check for compilation errors
Write-Host "CHECKING FOR KNOWN ISSUES:" -ForegroundColor Cyan
$compErrors = Select-String -Path "video-streaming/**/*.go" -Pattern "undefined:|declared and not used:" 2>$null
if ($compErrors) {
    Write-Host "  [ISSUE] Found potential compilation issues" -ForegroundColor Red
} else {
    Write-Host "  [OK] No obvious compilation issues detected" -ForegroundColor Green
}

Write-Host ""
Write-Host "ALL 5 STRANDS COMBED" -ForegroundColor Green
Write-Host ""

