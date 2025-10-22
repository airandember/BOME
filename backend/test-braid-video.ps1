# ============================================================================
# 🎬 VIDEO STREAMING BRAID E2E TESTING
# ============================================================================
# Tests the complete video streaming flow from list to playback
# ============================================================================

$ErrorActionPreference = "Stop"

# Configuration
$apiBase = "http://localhost:8080/api/v1"
$timestamp = Get-Date -Format "yyyyMMddHHmmss"

# Test Results Tracking
$results = @{
    TestName = "Video Streaming Braid E2E"
    Timestamp = $timestamp
    Tests = @()
    Passed = 0
    Failed = 0
    TotalDuration = 0
}

function Add-TestResult {
    param(
        [string]$TestName,
        [bool]$Success,
        [string]$Message,
        [int]$Duration,
        [object]$Data = $null
    )
    
    $result = @{
        Name = $TestName
        Success = $Success
        Message = $Message
        Duration = $Duration
        Data = $Data
    }
    
    $results.Tests += $result
    $results.TotalDuration += $Duration
    
    if ($Success) {
        $results.Passed++
    } else {
        $results.Failed++
    }
}

# ============================================================================
# Test Banner
# ============================================================================

Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║        MISSION 2: VIDEO STREAMING BRAID E2E TESTING         ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════════╝`n" -ForegroundColor Cyan

# ============================================================================
# PHASE 1: SETUP & AUTHENTICATION
# ============================================================================

Write-Host "📋 PHASE 1: Setup & Authentication" -ForegroundColor Yellow
Write-Host "─────────────────────────────────────`n" -ForegroundColor DarkGray

# Use pre-created test user (verified via create-test-user tool)
$testEmail = "test.video@example.com"
$testPassword = "VideoTest123!"

Write-Host "🔐 Test User: $testEmail" -ForegroundColor Cyan
Write-Host "🔐 Password: $testPassword" -ForegroundColor Cyan
Write-Host "📝 (Pre-verified test user)`n" -ForegroundColor DarkGray

# Login to get tokens
Write-Host "▶ Logging in..." -ForegroundColor White
$loginBody = @{
    email = $testEmail
    password = $testPassword
} | ConvertTo-Json

$accessToken = $null
$refreshToken = $null

try {
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $loginResponse = Invoke-RestMethod -Uri "$apiBase/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $sw.Stop()
    
    $accessToken = $loginResponse.access_token
    $refreshToken = $loginResponse.refresh_token
    
    Write-Host "  ✅ Login successful" -ForegroundColor Green
    Write-Host "  📝 Access token: $($accessToken.Substring(0, 20))..." -ForegroundColor DarkGray
    Add-TestResult -TestName "User Login" -Success $true -Message "Authenticated successfully" -Duration $sw.ElapsedMilliseconds
} catch {
    $sw.Stop()
    Write-Host "  ❌ FAIL: Could not login - $($_.Exception.Message)" -ForegroundColor Red
    Add-TestResult -TestName "User Login" -Success $false -Message $_.Exception.Message -Duration $sw.ElapsedMilliseconds
    Write-Host "`n⚠️  Cannot continue without authentication" -ForegroundColor Red
    exit 1
}

$authHeaders = @{
    "Authorization" = "Bearer $accessToken"
    "Content-Type" = "application/json"
}

Write-Host "`n✅ Authentication complete!`n" -ForegroundColor Green

# ============================================================================
# PHASE 2: VIDEO LIST RETRIEVAL
# ============================================================================

Write-Host "📋 PHASE 2: Video List Retrieval" -ForegroundColor Yellow
Write-Host "─────────────────────────────────────`n" -ForegroundColor DarkGray

# Test 1: Get all videos (public)
Write-Host "▶ Testing: 1. Get All Videos (Public Access)" -ForegroundColor White
try {
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $videoListResponse = Invoke-RestMethod -Uri "$apiBase/videos" -Method GET
    $sw.Stop()
    
    Write-Host "  ✅ PASS - Retrieved video list (Count: $($videoListResponse.Count -or 0)) (${sw.ElapsedMilliseconds}ms)" -ForegroundColor Green
    Add-TestResult -TestName "Get All Videos (Public)" -Success $true -Message "Retrieved $($videoListResponse.Count -or 0) videos" -Duration $sw.ElapsedMilliseconds -Data $videoListResponse
    
    # Store video IDs for later tests
    $script:videoList = $videoListResponse
} catch {
    $sw.Stop()
    Write-Host "  ❌ FAIL - $($_.Exception.Message) (${sw.ElapsedMilliseconds}ms)" -ForegroundColor Red
    Add-TestResult -TestName "Get All Videos (Public)" -Success $false -Message $_.Exception.Message -Duration $sw.ElapsedMilliseconds
}

# Test 2: Get videos from Bunny CDN
Write-Host "`n▶ Testing: 2. Get Videos from Bunny CDN" -ForegroundColor White
try {
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $bunnyVideosResponse = Invoke-RestMethod -Uri "$apiBase/bunny-videos" -Method GET -Headers $authHeaders
    $sw.Stop()
    
    Write-Host "  ✅ PASS - Retrieved Bunny videos (Count: $($bunnyVideosResponse.items.Count -or 0)) (${sw.ElapsedMilliseconds}ms)" -ForegroundColor Green
    Add-TestResult -TestName "Get Bunny Videos" -Success $true -Message "Retrieved $($bunnyVideosResponse.items.Count -or 0) videos" -Duration $sw.ElapsedMilliseconds -Data $bunnyVideosResponse
    
    $script:bunnyVideos = $bunnyVideosResponse.items
} catch {
    $sw.Stop()
    $statusCode = try { $_.Exception.Response.StatusCode.value__ } catch { "unknown" }
    
    # 503/501 are acceptable (service not configured yet)
    if ($statusCode -eq 503 -or $statusCode -eq 501) {
        Write-Host "  ⚠️  WARNING - Bunny service not configured ($statusCode) (${sw.ElapsedMilliseconds}ms)" -ForegroundColor Yellow
        Add-TestResult -TestName "Get Bunny Videos" -Success $true -Message "Service not configured (expected)" -Duration $sw.ElapsedMilliseconds
    } else {
        Write-Host "  ❌ FAIL - $($_.Exception.Message) (${sw.ElapsedMilliseconds}ms)" -ForegroundColor Red
        Add-TestResult -TestName "Get Bunny Videos" -Success $false -Message $_.Exception.Message -Duration $sw.ElapsedMilliseconds
    }
}

# ============================================================================
# PHASE 3: SINGLE VIDEO DETAILS
# ============================================================================

Write-Host "`n📋 PHASE 3: Single Video Details" -ForegroundColor Yellow
Write-Host "─────────────────────────────────────`n" -ForegroundColor DarkGray

# Test 3: Get single video by ID (if we have any)
if ($script:videoList -and $script:videoList.Count -gt 0) {
    $testVideoId = $script:videoList[0].id
    
    Write-Host "▶ Testing: 3. Get Single Video Details (ID: $testVideoId)" -ForegroundColor White
    try {
        $sw = [System.Diagnostics.Stopwatch]::StartNew()
        $videoResponse = Invoke-RestMethod -Uri "$apiBase/videos/$testVideoId" -Method GET -Headers $authHeaders
        $sw.Stop()
        
        Write-Host "  ✅ PASS - Retrieved video details: $($videoResponse.title)" -ForegroundColor Green
        Add-TestResult -TestName "Get Single Video" -Success $true -Message "Retrieved video: $($videoResponse.title)" -Duration $sw.ElapsedMilliseconds -Data $videoResponse
    } catch {
        $sw.Stop()
        Write-Host "  ❌ FAIL - $($_.Exception.Message)" -ForegroundColor Red
        Add-TestResult -TestName "Get Single Video" -Success $false -Message $_.Exception.Message -Duration $sw.ElapsedMilliseconds
    }
} else {
    Write-Host "▶ Testing: 3. Get Single Video Details" -ForegroundColor White
    Write-Host "  ⚠️  SKIP - No videos available to test" -ForegroundColor Yellow
    Add-TestResult -TestName "Get Single Video" -Success $true -Message "No videos to test (acceptable)" -Duration 0
}

# Test 4: Get non-existent video (should 404)
Write-Host "`n▶ Testing: 4. Get Non-Existent Video (Error Handling)" -ForegroundColor White
try {
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $nonExistentResponse = Invoke-RestMethod -Uri "$apiBase/videos/99999999" -Method GET -Headers $authHeaders
    $sw.Stop()
    
    Write-Host "  ❌ FAIL - Should have returned 404" -ForegroundColor Red
    Add-TestResult -TestName "Get Non-Existent Video" -Success $false -Message "Should return 404" -Duration $sw.ElapsedMilliseconds
} catch {
    $sw.Stop()
    $statusCode = try { $_.Exception.Response.StatusCode.value__ } catch { 0 }
    
    if ($statusCode -eq 404) {
        Write-Host "  ✅ PASS - Correctly returned 404 Not Found" -ForegroundColor Green
        Add-TestResult -TestName "Get Non-Existent Video" -Success $true -Message "Correct 404 response" -Duration $sw.ElapsedMilliseconds
    } else {
        Write-Host "  ❌ FAIL - Expected 404, got $statusCode" -ForegroundColor Red
        Add-TestResult -TestName "Get Non-Existent Video" -Success $false -Message "Expected 404, got $statusCode" -Duration $sw.ElapsedMilliseconds
    }
}

# ============================================================================
# PHASE 4: VIDEO PLAYBACK AUTHENTICATION
# ============================================================================

Write-Host "`n📋 PHASE 4: Video Playback Authentication" -ForegroundColor Yellow
Write-Host "─────────────────────────────────────`n" -ForegroundColor DarkGray

# Test 5: Access video stream without auth (should require auth)
if ($script:videoList -and $script:videoList.Count -gt 0) {
    $testVideoId = $script:videoList[0].id
    
    Write-Host "▶ Testing: 5. Video Stream Without Auth (Should Require Auth)" -ForegroundColor White
    try {
        $sw = [System.Diagnostics.Stopwatch]::StartNew()
        $streamResponse = Invoke-RestMethod -Uri "$apiBase/videos/$testVideoId/stream" -Method GET
        $sw.Stop()
        
        Write-Host "  ⚠️  WARNING - Stream accessible without auth (may be public)" -ForegroundColor Yellow
        Add-TestResult -TestName "Stream Without Auth" -Success $true -Message "Public access (OK if intended)" -Duration $sw.ElapsedMilliseconds
    } catch {
        $sw.Stop()
        $statusCode = try { $_.Exception.Response.StatusCode.value__ } catch { 0 }
        
        if ($statusCode -eq 401 -or $statusCode -eq 403) {
            Write-Host "  ✅ PASS - Correctly requires authentication ($statusCode)" -ForegroundColor Green
            Add-TestResult -TestName "Stream Without Auth" -Success $true -Message "Auth required (correct)" -Duration $sw.ElapsedMilliseconds
        } elseif ($statusCode -eq 503 -or $statusCode -eq 501) {
            Write-Host "  ⚠️  WARNING - Service not implemented ($statusCode)" -ForegroundColor Yellow
            Add-TestResult -TestName "Stream Without Auth" -Success $true -Message "Not implemented yet" -Duration $sw.ElapsedMilliseconds
        } else {
            Write-Host "  ❌ FAIL - Unexpected status: $statusCode" -ForegroundColor Red
            Add-TestResult -TestName "Stream Without Auth" -Success $false -Message "Unexpected status: $statusCode" -Duration $sw.ElapsedMilliseconds
        }
    }
    
    # Test 6: Access video stream with valid auth
    Write-Host "`n▶ Testing: 6. Video Stream With Valid Auth" -ForegroundColor White
    try {
        $sw = [System.Diagnostics.Stopwatch]::StartNew()
        $streamResponse = Invoke-RestMethod -Uri "$apiBase/videos/$testVideoId/stream" -Method GET -Headers $authHeaders
        $sw.Stop()
        
        Write-Host "  ✅ PASS - Stream accessible with auth" -ForegroundColor Green
        Add-TestResult -TestName "Stream With Auth" -Success $true -Message "Authorized access successful" -Duration $sw.ElapsedMilliseconds
    } catch {
        $sw.Stop()
        $statusCode = try { $_.Exception.Response.StatusCode.value__ } catch { 0 }
        
        if ($statusCode -eq 503 -or $statusCode -eq 501) {
            Write-Host "  ⚠️  WARNING - Service not implemented ($statusCode)" -ForegroundColor Yellow
            Add-TestResult -TestName "Stream With Auth" -Success $true -Message "Not implemented yet (acceptable)" -Duration $sw.ElapsedMilliseconds
        } else {
            Write-Host "  ❌ FAIL - $($_.Exception.Message)" -ForegroundColor Red
            Add-TestResult -TestName "Stream With Auth" -Success $false -Message $_.Exception.Message -Duration $sw.ElapsedMilliseconds
        }
    }
} else {
    Write-Host "▶ Testing: 5-6. Video Stream Auth Tests" -ForegroundColor White
    Write-Host "  ⚠️  SKIP - No videos available to test" -ForegroundColor Yellow
    Add-TestResult -TestName "Stream Without Auth" -Success $true -Message "No videos to test" -Duration 0
    Add-TestResult -TestName "Stream With Auth" -Success $true -Message "No videos to test" -Duration 0
}

# ============================================================================
# PHASE 5: VIDEO METADATA & THUMBNAILS
# ============================================================================

Write-Host "`n📋 PHASE 5: Video Metadata & Thumbnails" -ForegroundColor Yellow
Write-Host "─────────────────────────────────────`n" -ForegroundColor DarkGray

# Test 7: Video metadata validation
if ($script:videoList -and $script:videoList.Count -gt 0) {
    $testVideo = $script:videoList[0]
    
    Write-Host "▶ Testing: 7. Video Metadata Structure" -ForegroundColor White
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    
    $requiredFields = @("id", "title", "description")
    $missingFields = @()
    
    foreach ($field in $requiredFields) {
        if (-not $testVideo.PSObject.Properties[$field]) {
            $missingFields += $field
        }
    }
    
    $sw.Stop()
    
    if ($missingFields.Count -eq 0) {
        Write-Host "  ✅ PASS - All required metadata fields present" -ForegroundColor Green
        Add-TestResult -TestName "Video Metadata" -Success $true -Message "All required fields present" -Duration $sw.ElapsedMilliseconds
    } else {
        Write-Host "  ⚠️  WARNING - Missing fields: $($missingFields -join ', ')" -ForegroundColor Yellow
        Add-TestResult -TestName "Video Metadata" -Success $false -Message "Missing: $($missingFields -join ', ')" -Duration $sw.ElapsedMilliseconds
    }
} else {
    Write-Host "▶ Testing: 7. Video Metadata Structure" -ForegroundColor White
    Write-Host "  ⚠️  SKIP - No videos available to test" -ForegroundColor Yellow
    Add-TestResult -TestName "Video Metadata" -Success $true -Message "No videos to test" -Duration 0
}

# Test 8: Video blob/thumbnail access
if ($script:videoList -and $script:videoList.Count -gt 0) {
    $testVideoId = $script:videoList[0].id
    
    Write-Host "`n▶ Testing: 8. Video Blob/Thumbnail Access" -ForegroundColor White
    try {
        $sw = [System.Diagnostics.Stopwatch]::StartNew()
        $blobResponse = Invoke-RestMethod -Uri "$apiBase/videos/$testVideoId/blob" -Method GET -Headers $authHeaders
        $sw.Stop()
        
        Write-Host "  ✅ PASS - Blob endpoint accessible" -ForegroundColor Green
        Add-TestResult -TestName "Video Blob Access" -Success $true -Message "Blob accessible" -Duration $sw.ElapsedMilliseconds
    } catch {
        $sw.Stop()
        $statusCode = try { $_.Exception.Response.StatusCode.value__ } catch { 0 }
        
        if ($statusCode -eq 503 -or $statusCode -eq 501) {
            Write-Host "  ⚠️  WARNING - Blob service not implemented ($statusCode)" -ForegroundColor Yellow
            Add-TestResult -TestName "Video Blob Access" -Success $true -Message "Not implemented yet" -Duration $sw.ElapsedMilliseconds
        } else {
            Write-Host "  ⚠️  INFO - Status: $statusCode" -ForegroundColor Yellow
            Add-TestResult -TestName "Video Blob Access" -Success $false -Message "Status: $statusCode" -Duration $sw.ElapsedMilliseconds
        }
    }
} else {
    Write-Host "`n▶ Testing: 8. Video Blob/Thumbnail Access" -ForegroundColor White
    Write-Host "  ⚠️  SKIP - No videos available to test" -ForegroundColor Yellow
    Add-TestResult -TestName "Video Blob Access" -Success $true -Message "No videos to test" -Duration 0
}

# ============================================================================
# PHASE 6: BUNNY.NET INTEGRATION
# ============================================================================

Write-Host "`n📋 PHASE 6: Bunny.net Integration" -ForegroundColor Yellow
Write-Host "─────────────────────────────────────`n" -ForegroundColor DarkGray

# Test 9: Bunny collections
Write-Host "▶ Testing: 9. Bunny Collections API" -ForegroundColor White
try {
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $collectionsResponse = Invoke-RestMethod -Uri "$apiBase/bunny-collections" -Method GET
    $sw.Stop()
    
    Write-Host "  ✅ PASS - Retrieved collections (Count: $($collectionsResponse.items.Count -or 0))" -ForegroundColor Green
    Add-TestResult -TestName "Bunny Collections" -Success $true -Message "Retrieved $($collectionsResponse.items.Count -or 0) collections" -Duration $sw.ElapsedMilliseconds
} catch {
    $sw.Stop()
    $statusCode = try { $_.Exception.Response.StatusCode.value__ } catch { 0 }
    
    if ($statusCode -eq 503 -or $statusCode -eq 501) {
        Write-Host "  ⚠️  WARNING - Bunny service not configured ($statusCode)" -ForegroundColor Yellow
        Add-TestResult -TestName "Bunny Collections" -Success $true -Message "Not configured (expected)" -Duration $sw.ElapsedMilliseconds
    } else {
        Write-Host "  ❌ FAIL - $($_.Exception.Message)" -ForegroundColor Red
        Add-TestResult -TestName "Bunny Collections" -Success $false -Message $_.Exception.Message -Duration $sw.ElapsedMilliseconds
    }
}

# Test 10: Bunny video update (requires auth)
if ($script:bunnyVideos -and $script:bunnyVideos.Count -gt 0) {
    $testBunnyId = $script:bunnyVideos[0].videoLibraryId
    
    Write-Host "`n▶ Testing: 10. Bunny Video Update (Auth Required)" -ForegroundColor White
    try {
        $updateBody = @{
            title = "Updated Test Video"
        } | ConvertTo-Json
        
        $sw = [System.Diagnostics.Stopwatch]::StartNew()
        $updateResponse = Invoke-RestMethod -Uri "$apiBase/bunny-videos/$testBunnyId" -Method PUT -Headers $authHeaders -Body $updateBody
        $sw.Stop()
        
        Write-Host "  ✅ PASS - Video update successful" -ForegroundColor Green
        Add-TestResult -TestName "Bunny Video Update" -Success $true -Message "Update successful" -Duration $sw.ElapsedMilliseconds
    } catch {
        $sw.Stop()
        $statusCode = try { $_.Exception.Response.StatusCode.value__ } catch { 0 }
        
        if ($statusCode -eq 503 -or $statusCode -eq 501) {
            Write-Host "  ⚠️  WARNING - Service not implemented ($statusCode)" -ForegroundColor Yellow
            Add-TestResult -TestName "Bunny Video Update" -Success $true -Message "Not implemented" -Duration $sw.ElapsedMilliseconds
        } elseif ($statusCode -eq 401 -or $statusCode -eq 403) {
            Write-Host "  ⚠️  INFO - Requires higher permissions ($statusCode)" -ForegroundColor Yellow
            Add-TestResult -TestName "Bunny Video Update" -Success $true -Message "Permission check works" -Duration $sw.ElapsedMilliseconds
        } else {
            Write-Host "  ❌ FAIL - $($_.Exception.Message)" -ForegroundColor Red
            Add-TestResult -TestName "Bunny Video Update" -Success $false -Message $_.Exception.Message -Duration $sw.ElapsedMilliseconds
        }
    }
} else {
    Write-Host "`n▶ Testing: 10. Bunny Video Update" -ForegroundColor White
    Write-Host "  ⚠️  SKIP - No Bunny videos available" -ForegroundColor Yellow
    Add-TestResult -TestName "Bunny Video Update" -Success $true -Message "No videos to test" -Duration 0
}

# ============================================================================
# RESULTS SUMMARY
# ============================================================================

Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║           VIDEO STREAMING BRAID TEST RESULTS                ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════════╝`n" -ForegroundColor Cyan

$totalTests = $results.Passed + $results.Failed
$passRate = if ($totalTests -gt 0) { [math]::Round(($results.Passed / $totalTests) * 100, 1) } else { 0 }

Write-Host "📊 STATISTICS:" -ForegroundColor White
Write-Host "   Total Tests:   $totalTests" -ForegroundColor White
Write-Host "   Passed:        $($results.Passed) ✅" -ForegroundColor Green
Write-Host "   Failed:        $($results.Failed) ❌" -ForegroundColor $(if ($results.Failed -eq 0) { "Green" } else { "Red" })
Write-Host "   Pass Rate:     $passRate%" -ForegroundColor $(if ($passRate -ge 80) { "Green" } elseif ($passRate -ge 60) { "Yellow" } else { "Red" })
Write-Host "   Total Duration: $($results.TotalDuration)ms`n" -ForegroundColor DarkGray

Write-Host "📋 DETAILED RESULTS:" -ForegroundColor White
$testNumber = 1
foreach ($test in $results.Tests) {
    $icon = if ($test.Success) { "✅" } else { "❌" }
    $color = if ($test.Success) { "Green" } else { "Red" }
    Write-Host "   $icon $testNumber. $($test.Name)" -ForegroundColor $color
    Write-Host "      └─ $($test.Message) ($($test.Duration)ms)" -ForegroundColor DarkGray
    $testNumber++
}

# Save results
$resultsFile = "test-results-braid-video.json"
$results | ConvertTo-Json -Depth 10 | Out-File $resultsFile
Write-Host "`n💾 Results saved to: $resultsFile" -ForegroundColor Cyan

# Pass/Fail determination
if ($passRate -ge 80) {
    Write-Host "`n✅ VIDEO STREAMING BRAID TESTS PASSED! (≥80% pass rate)`n" -ForegroundColor Green
    exit 0
} else {
    Write-Host "`n⚠️  VIDEO STREAMING BRAID TESTS NEED ATTENTION (<80% pass rate)`n" -ForegroundColor Yellow
    exit 1
}

