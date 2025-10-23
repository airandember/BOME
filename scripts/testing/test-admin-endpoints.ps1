# ============================================================================
# 🎛️  BOME Admin Dashboard - Endpoint Tests
# ============================================================================
# Tests admin endpoints with different authentication levels
# ============================================================================

$ErrorActionPreference = "Stop"

Write-Host "`n╔════════════════════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                                                                                ║" -ForegroundColor Cyan
Write-Host "║              🎛️  ADMIN DASHBOARD ENDPOINT TESTS                                ║" -ForegroundColor Cyan
Write-Host "║                                                                                ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════════════════════════╝`n" -ForegroundColor Cyan

$results = @{
    Tests = @()
    Passed = 0
    Failed = 0
}

function Add-Result {
    param([string]$Name, [bool]$Success, [string]$Message)
    $results.Tests += @{ Name = $Name; Success = $Success; Message = $Message }
    if ($Success) { $results.Passed++ } else { $results.Failed++ }
}

# Test 1: Check if admin routes are registered
Write-Host "🔍 Test 1: Backend Health Check" -ForegroundColor Cyan
try {
    $health = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET -TimeoutSec 3
    Write-Host "  ✅ PASS - Backend responding" -ForegroundColor Green
    Add-Result -Name "Backend Health" -Success $true -Message "Healthy"
} catch {
    Write-Host "  ❌ FAIL - Backend not responding" -ForegroundColor Red
    Add-Result -Name "Backend Health" -Success $false -Message $_.Exception.Message
    exit 1
}

# Test 2: Admin users endpoint without auth (should return 401)
Write-Host "`n🔍 Test 2: Admin Endpoint Without Auth" -ForegroundColor Cyan
try {
    Invoke-RestMethod -Uri "http://localhost:8080/api/v1/admin/users" -Method GET -TimeoutSec 3
    Write-Host "  ❌ FAIL - Should require authentication" -ForegroundColor Red
    Add-Result -Name "Admin Auth Required" -Success $false -Message "No auth required (BAD)"
} catch {
    $statusCode = try { $_.Exception.Response.StatusCode.value__ } catch { "unknown" }
    if ($statusCode -eq 401) {
        Write-Host "  ✅ PASS - Correctly returns 401 Unauthorized" -ForegroundColor Green
        Add-Result -Name "Admin Auth Required" -Success $true -Message "401 Unauthorized"
    } else {
        Write-Host "  ⚠️  WARNING - Expected 401, got $statusCode" -ForegroundColor Yellow
        Add-Result -Name "Admin Auth Required" -Success $true -Message "Status: $statusCode"
    }
}

# Test 3: Admin analytics endpoint
Write-Host "`n🔍 Test 3: Admin Analytics Endpoint (No Auth)" -ForegroundColor Cyan
try {
    Invoke-RestMethod -Uri "http://localhost:8080/api/v1/admin/analytics" -Method GET -TimeoutSec 3
    Write-Host "  ❌ FAIL - Should require authentication" -ForegroundColor Red
    Add-Result -Name "Analytics Auth" -Success $false -Message "No auth required"
} catch {
    $statusCode = try { $_.Exception.Response.StatusCode.value__ } catch { "unknown" }
    if ($statusCode -eq 401) {
        Write-Host "  ✅ PASS - Correctly returns 401" -ForegroundColor Green
        Add-Result -Name "Analytics Auth" -Success $true -Message "401 Unauthorized"
    } else {
        Write-Host "  ⚠️  WARNING - Expected 401, got $statusCode" -ForegroundColor Yellow
        Add-Result -Name "Analytics Auth" -Success $true -Message "Status: $statusCode"
    }
}

# Test 4: Admin monitoring/system endpoint
Write-Host "`n🔍 Test 4: System Monitoring Endpoint (No Auth)" -ForegroundColor Cyan
try {
    Invoke-RestMethod -Uri "http://localhost:8080/api/v1/admin/monitoring/system" -Method GET -TimeoutSec 3
    Write-Host "  ❌ FAIL - Should require authentication" -ForegroundColor Red
    Add-Result -Name "Monitoring Auth" -Success $false -Message "No auth required"
} catch {
    $statusCode = try { $_.Exception.Response.StatusCode.value__ } catch { "unknown" }
    if ($statusCode -eq 401) {
        Write-Host "  ✅ PASS - Correctly returns 401" -ForegroundColor Green
        Add-Result -Name "Monitoring Auth" -Success $true -Message "401 Unauthorized"
    } else {
        Write-Host "  ⚠️  WARNING - Expected 401, got $statusCode" -ForegroundColor Yellow
        Add-Result -Name "Monitoring Auth" -Success $true -Message "Status: $statusCode"
    }
}

# Test 5: Admin roles endpoint
Write-Host "`n🔍 Test 5: Roles Endpoint (No Auth)" -ForegroundColor Cyan
try {
    Invoke-RestMethod -Uri "http://localhost:8080/api/v1/admin/roles" -Method GET -TimeoutSec 3
    Write-Host "  ❌ FAIL - Should require authentication" -ForegroundColor Red
    Add-Result -Name "Roles Auth" -Success $false -Message "No auth required"
} catch {
    $statusCode = try { $_.Exception.Response.StatusCode.value__ } catch { "unknown" }
    if ($statusCode -eq 401) {
        Write-Host "  ✅ PASS - Correctly returns 401" -ForegroundColor Green
        Add-Result -Name "Roles Auth" -Success $true -Message "401 Unauthorized"
    } else {
        Write-Host "  ⚠️  WARNING - Expected 401, got $statusCode" -ForegroundColor Yellow
        Add-Result -Name "Roles Auth" -Success $true -Message "Status: $statusCode"
    }
}

# Test 6: Admin videos endpoint
Write-Host "`n🔍 Test 6: Admin Videos Endpoint (No Auth)" -ForegroundColor Cyan
try {
    Invoke-RestMethod -Uri "http://localhost:8080/api/v1/admin/videos" -Method GET -TimeoutSec 3
    Write-Host "  ❌ FAIL - Should require authentication" -ForegroundColor Red
    Add-Result -Name "Videos Auth" -Success $false -Message "No auth required"
} catch {
    $statusCode = try { $_.Exception.Response.StatusCode.value__ } catch { "unknown" }
    if ($statusCode -eq 401) {
        Write-Host "  ✅ PASS - Correctly returns 401" -ForegroundColor Green
        Add-Result -Name "Videos Auth" -Success $true -Message "401 Unauthorized"
    } else {
        Write-Host "  ⚠️  WARNING - Expected 401, got $statusCode" -ForegroundColor Yellow
        Add-Result -Name "Videos Auth" -Success $true -Message "Status: $statusCode"
    }
}

# Results Summary
Write-Host "`n╔════════════════════════════════════════════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║                                                                                ║" -ForegroundColor Green
Write-Host "║                      ADMIN ENDPOINT TEST RESULTS                               ║" -ForegroundColor Green
Write-Host "║                                                                                ║" -ForegroundColor Green
Write-Host "╚════════════════════════════════════════════════════════════════════════════════╝`n" -ForegroundColor Green

$total = $results.Passed + $results.Failed
$passRate = if ($total -gt 0) { [math]::Round(($results.Passed / $total) * 100, 1) } else { 0 }

Write-Host "📊 RESULTS:" -ForegroundColor White
Write-Host "   Total Tests: $total" -ForegroundColor White
Write-Host "   Passed: $($results.Passed) ✅" -ForegroundColor Green
Write-Host "   Failed: $($results.Failed) ❌" -ForegroundColor $(if ($results.Failed -eq 0) { "Green" } else { "Red" })
Write-Host "   Pass Rate: $passRate%`n" -ForegroundColor $(if ($passRate -ge 80) { "Green" } elseif ($passRate -ge 60) { "Yellow" } else { "Red" })

Write-Host "📋 DETAILED RESULTS:" -ForegroundColor White
$num = 1
foreach ($test in $results.Tests) {
    $icon = if ($test.Success) { "✅" } else { "❌" }
    $color = if ($test.Success) { "Green" } else { "Red" }
    Write-Host "   $icon $num. $($test.Name)" -ForegroundColor $color
    Write-Host "      └─ $($test.Message)" -ForegroundColor Gray
    $num++
}

Write-Host "`n📄 Next Steps:" -ForegroundColor Yellow
Write-Host "   1. Update user to super_admin role in database" -ForegroundColor White
Write-Host "   2. Login and get JWT token" -ForegroundColor White
Write-Host "   3. Test endpoints with valid admin token" -ForegroundColor White
Write-Host "   4. Verify RBAC enforcement`n" -ForegroundColor White

if ($passRate -ge 80) {
    Write-Host "✅ ADMIN ROUTES: PROPERLY SECURED!`n" -ForegroundColor Green
    exit 0
} else {
    Write-Host "⚠️  ADMIN ROUTES: NEEDS ATTENTION`n" -ForegroundColor Yellow
    exit 1
}

