# Token Refresh & Logout Testing
# Tests advanced auth flows

Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Magenta
Write-Host "║      MISSION 2: TOKEN REFRESH & LOGOUT TESTING              ║" -ForegroundColor Magenta
Write-Host "╚══════════════════════════════════════════════════════════════╝`n" -ForegroundColor Magenta

$baseUrl = "http://localhost:8080"
$apiBase = "$baseUrl/api/v1"
$results = @()

# Test context
$testContext = @{
    TestEmail = "refresh_test_$(Get-Random -Maximum 99999)@example.com"
    TestPassword = "TestPass123!"
    FirstName = "Refresh"
    LastName = "Tester"
    AccessToken = $null
    RefreshToken = $null
    UserId = $null
}

function Test-Flow {
    param(
        [string]$Name,
        [scriptblock]$TestScript
    )
    
    Write-Host "`n▶ Testing: $Name" -ForegroundColor Cyan
    
    try {
        $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
        $result = & $TestScript
        $stopwatch.Stop()
        
        if ($result.Success) {
            Write-Host "  ✅ PASS - $($result.Message) ($($stopwatch.ElapsedMilliseconds)ms)" -ForegroundColor Green
        } else {
            Write-Host "  ❌ FAIL - $($result.Message)" -ForegroundColor Red
        }
        
        return @{
            Name = $Name
            Success = $result.Success
            Message = $result.Message
            Duration = $stopwatch.ElapsedMilliseconds
        }
        
    } catch {
        Write-Host "  ❌ ERROR: $($_.Exception.Message)" -ForegroundColor Red
        return @{
            Name = $Name
            Success = $false
            Message = "Exception: $($_.Exception.Message)"
            Duration = 0
        }
    }
}

Write-Host "📧 Test User: $($testContext.TestEmail)" -ForegroundColor Yellow
Write-Host "🔐 Test Password: $($testContext.TestPassword)`n" -ForegroundColor Yellow

# ============================================================================
# SETUP: Register and verify user to enable login
# ============================================================================
Write-Host "🔧 SETUP: Creating test user...`n" -ForegroundColor Gray

$setupBody = @{
    email = $testContext.TestEmail
    first_name = $testContext.FirstName
    last_name = $testContext.LastName
} | ConvertTo-Json

try {
    $setupResponse = Invoke-RestMethod -Uri "$apiBase/auth/register" -Method POST -Body $setupBody -ContentType "application/json"
    $testContext.UserId = $setupResponse.user_id
    Write-Host "✅ User registered (ID: $($testContext.UserId))" -ForegroundColor Gray
} catch {
    Write-Host "⚠️  Setup failed - user may already exist" -ForegroundColor Yellow
}

# NOTE: In a real E2E test, we'd verify email here. For now, we'll test what we can without email verification.

# ============================================================================
# TEST 1: TOKEN REFRESH ENDPOINT EXISTS
# ============================================================================
$results += Test-Flow "1. Token Refresh Endpoint (Structure Test)" {
    # Try to refresh with invalid token to see if endpoint exists
    $body = @{
        refresh_token = "fake.refresh.token"
    } | ConvertTo-Json
    
    try {
        Invoke-RestMethod -Uri "$apiBase/auth/refresh" -Method POST -Body $body -ContentType "application/json"
        return @{ Success = $false; Message = "Should have rejected fake token" }
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        
        # 401 or 400 means endpoint exists and is validating tokens
        if ($statusCode -eq 401 -or $statusCode -eq 400) {
            return @{ Success = $true; Message = "Endpoint exists and validates tokens ($statusCode)" }
        }
        
        # 404 means not implemented
        if ($statusCode -eq 404) {
            return @{ Success = $false; Message = "Endpoint not implemented (404)" }
        }
        
        return @{ Success = $false; Message = "Unexpected status: $statusCode" }
    }
}

# ============================================================================
# TEST 2: LOGOUT ENDPOINT EXISTS
# ============================================================================
$results += Test-Flow "2. Logout Endpoint (Structure Test)" {
    try {
        # Try logout without auth header
        Invoke-RestMethod -Uri "$apiBase/auth/logout" -Method POST
        return @{ Success = $false; Message = "Should require authentication" }
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        
        # 401 or 400 means endpoint exists and requires auth
        if ($statusCode -eq 401 -or $statusCode -eq 400) {
            return @{ Success = $true; Message = "Endpoint exists and requires auth ($statusCode)" }
        }
        
        # 404 means not implemented
        if ($statusCode -eq 404) {
            return @{ Success = $false; Message = "Endpoint not implemented (404)" }
        }
        
        return @{ Success = $false; Message = "Unexpected status: $statusCode" }
    }
}

# ============================================================================
# TEST 3: TOKEN REFRESH WITH NO TOKEN (Should Fail)
# ============================================================================
$results += Test-Flow "3. Token Refresh Without Token" {
    $body = @{} | ConvertTo-Json
    
    try {
        Invoke-RestMethod -Uri "$apiBase/auth/refresh" -Method POST -Body $body -ContentType "application/json"
        return @{ Success = $false; Message = "Should require refresh_token field" }
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        
        if ($statusCode -eq 400 -or $statusCode -eq 401) {
            return @{ Success = $true; Message = "Correctly requires refresh_token ($statusCode)" }
        }
        
        return @{ Success = $false; Message = "Unexpected status: $statusCode" }
    }
}

# ============================================================================
# TEST 4: TOKEN REFRESH WITH INVALID TOKEN (Should Fail)
# ============================================================================
$results += Test-Flow "4. Token Refresh With Invalid Token" {
    $body = @{
        refresh_token = "invalid.token.format"
    } | ConvertTo-Json
    
    try {
        Invoke-RestMethod -Uri "$apiBase/auth/refresh" -Method POST -Body $body -ContentType "application/json"
        return @{ Success = $false; Message = "Should reject invalid token" }
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        
        if ($statusCode -eq 401) {
            return @{ Success = $true; Message = "Correctly rejects invalid token (401)" }
        }
        
        if ($statusCode -eq 400) {
            return @{ Success = $true; Message = "Rejects invalid token with validation error (400)" }
        }
        
        return @{ Success = $false; Message = "Unexpected status: $statusCode" }
    }
}

# ============================================================================
# TEST 5: LOGOUT WITHOUT TOKEN (Should Fail)
# ============================================================================
$results += Test-Flow "5. Logout Without Authentication" {
    try {
        Invoke-RestMethod -Uri "$apiBase/auth/logout" -Method POST
        return @{ Success = $false; Message = "Should require authentication" }
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        
        if ($statusCode -eq 401) {
            return @{ Success = $true; Message = "Correctly requires authentication (401)" }
        }
        
        if ($statusCode -eq 400) {
            return @{ Success = $true; Message = "Requires authentication (400)" }
        }
        
        return @{ Success = $false; Message = "Unexpected status: $statusCode" }
    }
}

# ============================================================================
# TEST 6: LOGOUT WITH INVALID TOKEN (Should Fail)
# ============================================================================
$results += Test-Flow "6. Logout With Invalid Token" {
    try {
        $headers = @{
            Authorization = "Bearer invalid.token"
        }
        Invoke-RestMethod -Uri "$apiBase/auth/logout" -Method POST -Headers $headers
        return @{ Success = $false; Message = "Should reject invalid token" }
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        
        if ($statusCode -eq 401) {
            return @{ Success = $true; Message = "Correctly rejects invalid token (401)" }
        }
        
        return @{ Success = $false; Message = "Unexpected status: $statusCode" }
    }
}

# ============================================================================
# NOTE: Full token refresh flow with real tokens would require:
# 1. Email verification (needs email service)
# 2. Successful login to get real refresh token
# 3. Then test refresh and logout
# 
# For now, we're testing endpoint structure and security.
# ============================================================================

# Calculate statistics
$totalTests = $results.Count
$passedTests = ($results | Where-Object { $_.Success }).Count
$failedTests = $totalTests - $passedTests
$passRate = [math]::Round(($passedTests / $totalTests) * 100, 1)

# Display Results
Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Blue
Write-Host "║         TOKEN REFRESH & LOGOUT TEST RESULTS                  ║" -ForegroundColor Blue
Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Blue

Write-Host "`n📊 STATISTICS:" -ForegroundColor Cyan
Write-Host "   Total Tests:   $totalTests" -ForegroundColor White
Write-Host "   Passed:        $passedTests ✅" -ForegroundColor Green
Write-Host "   Failed:        $failedTests ❌" -ForegroundColor $(if($failedTests -eq 0){"Green"}else{"Red"})
Write-Host "   Pass Rate:     $passRate%" -ForegroundColor $(if($passRate -ge 80){"Green"}elseif($passRate -ge 60){"Yellow"}else{"Red"})

Write-Host "`n📋 DETAILED RESULTS:" -ForegroundColor Cyan
foreach ($result in $results) {
    $status = if ($result.Success) { "✅" } else { "❌" }
    $color = if ($result.Success) { "Green" } else { "Red" }
    Write-Host "   $status $($result.Name)" -ForegroundColor $color
    Write-Host "      └─ $($result.Message) ($($result.Duration)ms)" -ForegroundColor Gray
}

# Export results
$results | ConvertTo-Json -Depth 3 | Out-File "test-results-token-refresh-logout.json"
Write-Host "`n💾 Results saved to: test-results-token-refresh-logout.json" -ForegroundColor Yellow

if ($passRate -ge 80) {
    Write-Host "`n✅ TOKEN REFRESH & LOGOUT TESTS PASSED! (≥80% pass rate)`n" -ForegroundColor Green
    exit 0
} else {
    Write-Host "`n⚠️  TOKEN REFRESH & LOGOUT TESTS NEED ATTENTION (<80% pass rate)`n" -ForegroundColor Yellow
    Write-Host "NOTE: Full flow testing requires email verification service`n" -ForegroundColor Gray
    exit 1
}

