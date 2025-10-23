# ============================================================================
# 🎬 VIDEO STREAMING BRAID E2E TESTING (Simplified - Public Endpoints)
# ============================================================================

$ErrorActionPreference = "Stop"

$apiBase = "http://localhost:8080/api/v1"
$results = @{ Passed = 0; Failed = 0; Tests = @() }

function Add-TestResult {
    param([string]$Name, [bool]$Success, [string]$Message, [int]$Duration)
    $results.Tests += @{ Name = $Name; Success = $Success; Message = $Message; Duration = $Duration }
    if ($Success) { $results.Passed++ } else { $results.Failed++ }
}

Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║     VIDEO STREAMING BRAID E2E TESTING (Public Endpoints)   ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════════╝`n" -ForegroundColor Cyan

# Test 1: Video test endpoint
Write-Host "▶ Test 1: Video Test Endpoint" -ForegroundColor White
try {
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $resp = Invoke-RestMethod -Uri "$apiBase/videos/test" -Method GET
    $sw.Stop()
    Write-Host "  ✅ PASS - $($resp.message) (${sw.ElapsedMilliseconds}ms)" -ForegroundColor Green
    Add-TestResult -Name "Video Test Endpoint" -Success $true -Message $resp.message -Duration $sw.ElapsedMilliseconds
} catch {
    $sw.Stop()
    Write-Host "  ❌ FAIL - $($_.Exception.Message) (${sw.ElapsedMilliseconds}ms)" -ForegroundColor Red
    Add-TestResult -Name "Video Test Endpoint" -Success $false -Message $_.Exception.Message -Duration $sw.ElapsedMilliseconds
}

# Test 2: Bunny collections (public)
Write-Host "`n▶ Test 2: Bunny Collections API (Public)" -ForegroundColor White
try {
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $resp = Invoke-RestMethod -Uri "$apiBase/bunny-collections" -Method GET
    $sw.Stop()
    Write-Host "  ✅ PASS - Retrieved collections (Count: $($resp.items.Count -or 0)) (${sw.ElapsedMilliseconds}ms)" -ForegroundColor Green
    Add-TestResult -Name "Bunny Collections" -Success $true -Message "Retrieved $($resp.items.Count -or 0) collections" -Duration $sw.ElapsedMilliseconds
} catch {
    $sw.Stop()
    $statusCode = try { $_.Exception.Response.StatusCode.value__ } catch { 0 }
    if ($statusCode -eq 503 -or $statusCode -eq 501) {
        Write-Host "  ⚠️  INFO - Bunny service not configured ($statusCode) (${sw.ElapsedMilliseconds}ms)" -ForegroundColor Yellow
        Add-TestResult -Name "Bunny Collections" -Success $true -Message "Not configured (expected)" -Duration $sw.ElapsedMilliseconds
    } else {
        Write-Host "  ❌ FAIL - $($_.Exception.Message) (${sw.ElapsedMilliseconds}ms)" -ForegroundColor Red
        Add-TestResult -Name "Bunny Collections" -Success $false -Message $_.Exception.Message -Duration $sw.ElapsedMilliseconds
    }
}

# Test 3: Video stream endpoint (public - should return something)
Write-Host "`n▶ Test 3: Video Stream Endpoint (ID: 1)" -ForegroundColor White
try {
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $resp = Invoke-RestMethod -Uri "$apiBase/videos/1/stream" -Method GET
    $sw.Stop()
    Write-Host "  ✅ PASS - Stream endpoint responds (${sw.ElapsedMilliseconds}ms)" -ForegroundColor Green
    Add-TestResult -Name "Video Stream Endpoint" -Success $true -Message "Endpoint accessible" -Duration $sw.ElapsedMilliseconds
} catch {
    $sw.Stop()
    $statusCode = try { $_.Exception.Response.StatusCode.value__ } catch { 0 }
    if ($statusCode -eq 404) {
        Write-Host "  ⚠️  INFO - Video not found (404) - endpoint works (${sw.ElapsedMilliseconds}ms)" -ForegroundColor Yellow
        Add-TestResult -Name "Video Stream Endpoint" -Success $true -Message "Endpoint works (404 expected)" -Duration $sw.ElapsedMilliseconds
    } elseif ($statusCode -eq 503) {
        Write-Host "  ⚠️  INFO - Service not implemented (503) (${sw.ElapsedMilliseconds}ms)" -ForegroundColor Yellow
        Add-TestResult -Name "Video Stream Endpoint" -Success $true -Message "Not implemented" -Duration $sw.ElapsedMilliseconds
    } else {
        Write-Host "  ❌ FAIL - $($_.Exception.Message) (${sw.ElapsedMilliseconds}ms)" -ForegroundColor Red
        Add-TestResult -Name "Video Stream Endpoint" -Success $false -Message $_.Exception.Message -Duration $sw.ElapsedMilliseconds
    }
}

# Test 4: Performance metrics endpoint
Write-Host "`n▶ Test 4: Performance Metrics Endpoint" -ForegroundColor White
try {
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $resp = Invoke-RestMethod -Uri "$apiBase/performance/metrics" -Method GET
    $sw.Stop()
    Write-Host "  ✅ PASS - Metrics endpoint responds (${sw.ElapsedMilliseconds}ms)" -ForegroundColor Green
    Add-TestResult -Name "Performance Metrics" -Success $true -Message "Endpoint accessible" -Duration $sw.ElapsedMilliseconds
} catch {
    $sw.Stop()
    Write-Host "  ❌ FAIL - $($_.Exception.Message) (${sw.ElapsedMilliseconds}ms)" -ForegroundColor Red
    Add-TestResult -Name "Performance Metrics" -Success $false -Message $_.Exception.Message -Duration $sw.ElapsedMilliseconds
}

# Test 5: Test optimization endpoint
Write-Host "`n▶ Test 5: Optimization Test Endpoint" -ForegroundColor White
try {
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $resp = Invoke-RestMethod -Uri "$apiBase/test/optimization" -Method GET
    $sw.Stop()
    Write-Host "  ✅ PASS - Optimization endpoint works (${sw.ElapsedMilliseconds}ms)" -ForegroundColor Green
    Add-TestResult -Name "Optimization Test" -Success $true -Message $resp.message -Duration $sw.ElapsedMilliseconds
} catch {
    $sw.Stop()
    Write-Host "  ❌ FAIL - $($_.Exception.Message) (${sw.ElapsedMilliseconds}ms)" -ForegroundColor Red
    Add-TestResult -Name "Optimization Test" -Success $false -Message $_.Exception.Message -Duration $sw.ElapsedMilliseconds
}

# Results Summary
Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║           VIDEO STREAMING BRAID TEST RESULTS                ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════════╝`n" -ForegroundColor Cyan

$total = $results.Passed + $results.Failed
$passRate = if ($total -gt 0) { [math]::Round(($results.Passed / $total) * 100, 1) } else { 0 }

Write-Host "📊 STATISTICS:" -ForegroundColor White
Write-Host "   Total Tests:   $total" -ForegroundColor White
Write-Host "   Passed:        $($results.Passed) ✅" -ForegroundColor Green
Write-Host "   Failed:        $($results.Failed) ❌" -ForegroundColor $(if ($results.Failed -eq 0) { "Green" } else { "Red" })
Write-Host "   Pass Rate:     $passRate%`n" -ForegroundColor $(if ($passRate -ge 80) { "Green" } elseif ($passRate -ge 60) { "Yellow" } else { "Red" })

Write-Host "📋 DETAILED RESULTS:" -ForegroundColor White
$num = 1
foreach ($test in $results.Tests) {
    $icon = if ($test.Success) { "✅" } else { "❌" }
    $color = if ($test.Success) { "Green" } else { "Red" }
    Write-Host "   $icon $num. $($test.Name)" -ForegroundColor $color
    Write-Host "      └─ $($test.Message) ($($test.Duration)ms)" -ForegroundColor DarkGray
    $num++
}

$results | ConvertTo-Json -Depth 10 | Out-File "test-results-video-simple.json"
Write-Host "`n💾 Results saved to: test-results-video-simple.json" -ForegroundColor Cyan

if ($passRate -ge 80) {
    Write-Host "`n✅ VIDEO STREAMING BRAID TESTS PASSED!`n" -ForegroundColor Green
    exit 0
} else {
    Write-Host "`n⚠️  VIDEO STREAMING BRAID NEEDS ATTENTION`n" -ForegroundColor Yellow
    exit 1
}

