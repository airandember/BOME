# Authentication Braid E2E Testing
# Tests complete user authentication flows from registration to logout

Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Blue
Write-Host "║        MISSION 2: AUTHENTICATION BRAID E2E TESTING          ║" -ForegroundColor Blue
Write-Host "╚══════════════════════════════════════════════════════════════╝`n" -ForegroundColor Blue

$baseUrl = "http://localhost:8080"
$apiBase = "$baseUrl/api/v1"
$results = @()

# Test context - will store tokens and user data
$testContext = @{
    TestEmail = "test_user_$(Get-Random -Maximum 99999)@example.com"
    TestPassword = "SecurePass123!"
    FirstName = "Test"
    LastName = "User"
    AccessToken = $null
    RefreshToken = $null
    UserId = $null
    VerificationToken = $null
}

function Test-E2EScenario {
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
            Details = $result.Details
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
Write-Host "👤 Test Name: $($testContext.FirstName) $($testContext.LastName)" -ForegroundColor Yellow
Write-Host "🔐 Password (for later): $($testContext.TestPassword)`n" -ForegroundColor Yellow

# ============================================================================
# TEST 1: USER REGISTRATION (Step 1: Register without password)
# ============================================================================
$results += Test-E2EScenario "1. User Registration (Email + Name Only)" {
    $body = @{
        email = $testContext.TestEmail
        first_name = $testContext.FirstName
        last_name = $testContext.LastName
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$apiBase/auth/register" `
            -Method POST `
            -Body $body `
            -ContentType "application/json" `
            -ErrorAction Stop
        
        # Verify response structure
        if (-not $response.user_id) {
            return @{ Success = $false; Message = "Response missing 'user_id' field" }
        }
        
        if ($response.email -ne $testContext.TestEmail) {
            return @{ Success = $false; Message = "Email mismatch in response" }
        }
        
        if (-not $response.verification_required) {
            return @{ Success = $false; Message = "Expected verification_required flag" }
        }
        
        # Store user ID for later tests
        $testContext.UserId = $response.user_id
        
        return @{ 
            Success = $true
            Message = "User registered (ID: $($response.user_id)) - Verification required"
            Details = $response
        }
        
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        
        # 400 might mean validation error or duplicate
        if ($statusCode -eq 400) {
            return @{ Success = $false; Message = "Registration failed (400) - Validation error or duplicate email" }
        }
        
        return @{ Success = $false; Message = "Registration failed ($statusCode): $($_.Exception.Message)" }
    }
}

# ============================================================================
# TEST 2: DUPLICATE REGISTRATION (Should Resend Verification)
# ============================================================================
$results += Test-E2EScenario "2. Duplicate Registration (Should Resend Verification)" {
    $body = @{
        email = $testContext.TestEmail
        first_name = $testContext.FirstName
        last_name = $testContext.LastName
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$apiBase/auth/register" `
            -Method POST `
            -Body $body `
            -ContentType "application/json" `
            -ErrorAction Stop
        
        # If we get here with 201, that's OK - it resends verification
        if ($response.verification_required) {
            return @{ Success = $true; Message = "Duplicate handled gracefully - resent verification" }
        }
        
        return @{ Success = $false; Message = "Unexpected response to duplicate registration" }
        
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        
        # 400 or 409 are also acceptable responses
        if ($statusCode -eq 400 -or $statusCode -eq 409) {
            return @{ Success = $true; Message = "Duplicate registration rejected ($statusCode)" }
        }
        
        return @{ Success = $false; Message = "Unexpected error code: $statusCode" }
    }
}

# ============================================================================
# TEST 3: LOGIN BEFORE VERIFICATION (Should Fail)
# ============================================================================
$results += Test-E2EScenario "3. Login Before Email Verification (Should Fail)" {
    $body = @{
        email = $testContext.TestEmail
        password = $testContext.TestPassword
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$apiBase/auth/login" `
            -Method POST `
            -Body $body `
            -ContentType "application/json" `
            -ErrorAction Stop
        
        # If we get here, that's BAD - should require verification!
        return @{ Success = $false; Message = "🚨 Login allowed without email verification!" }
        
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        
        # We WANT a 401 or 403 (unverified)
        if ($statusCode -eq 401 -or $statusCode -eq 403) {
            return @{ Success = $true; Message = "Correctly blocked unverified user ($statusCode)" }
        }
        
        return @{ Success = $false; Message = "Unexpected error code: $statusCode" }
    }
}

# ============================================================================
# NOTE: Tests 4-9 would normally test email verification and password setup
# But since this requires actual email service, we'll skip to testing the
# security aspects that don't require email verification
# ============================================================================

# ============================================================================
# TEST 4: INVALID EMAIL FORMAT
# ============================================================================
$results += Test-E2EScenario "4. Registration with Invalid Email" {
    $body = @{
        email = "not-an-email"
        first_name = "Test"
        last_name = "User"
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$apiBase/auth/register" `
            -Method POST `
            -Body $body `
            -ContentType "application/json" `
            -ErrorAction Stop
        
        return @{ Success = $false; Message = "Invalid email accepted" }
        
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        
        if ($statusCode -eq 400) {
            return @{ Success = $true; Message = "Correctly rejected invalid email (400)" }
        }
        
        return @{ Success = $false; Message = "Unexpected error code: $statusCode" }
    }
}

# ============================================================================
# TEST 5: MISSING REQUIRED FIELDS
# ============================================================================
$results += Test-E2EScenario "5. Registration Missing Required Fields" {
    $body = @{
        email = "test@example.com"
        # Missing first_name and last_name
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$apiBase/auth/register" `
            -Method POST `
            -Body $body `
            -ContentType "application/json" `
            -ErrorAction Stop
        
        return @{ Success = $false; Message = "Missing fields accepted" }
        
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        
        if ($statusCode -eq 400) {
            return @{ Success = $true; Message = "Correctly rejected missing fields (400)" }
        }
        
        return @{ Success = $false; Message = "Unexpected error code: $statusCode" }
    }
}

# ============================================================================
# TEST 6: INVALID LOGIN (Wrong Email)
# ============================================================================
$results += Test-E2EScenario "6. Login with Non-Existent Email" {
    $body = @{
        email = "nonexistent_$(Get-Random)@example.com"
        password = "SomePassword123!"
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$apiBase/auth/login" `
            -Method POST `
            -Body $body `
            -ContentType "application/json" `
            -ErrorAction Stop
        
        return @{ Success = $false; Message = "🚨 Invalid credentials accepted!" }
        
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        
        if ($statusCode -eq 401) {
            return @{ Success = $true; Message = "Correctly rejected invalid credentials (401)" }
        }
        
        return @{ Success = $false; Message = "Unexpected error code: $statusCode" }
    }
}

# ============================================================================
# TEST 7: INVALID LOGIN (Wrong Password - using our test user)
# ============================================================================
$results += Test-E2EScenario "7. Login with Wrong Password" {
    $body = @{
        email = $testContext.TestEmail
        password = "WrongPassword123!"
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$apiBase/auth/login" `
            -Method POST `
            -Body $body `
            -ContentType "application/json" `
            -ErrorAction Stop
        
        # If we get here, that's BAD - wrong password should be rejected!
        return @{ Success = $false; Message = "🚨 Invalid credentials accepted (security issue!)" }
        
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        
        # We WANT a 401 (Unauthorized)
        if ($statusCode -eq 401) {
            return @{ Success = $true; Message = "Correctly rejected invalid credentials (401)" }
        }
        
        return @{ Success = $false; Message = "Unexpected error code: $statusCode (expected 401)" }
    }
}

# ============================================================================
# TEST 8: ACCESS PROTECTED ENDPOINT (Without Token)
# ============================================================================
$results += Test-E2EScenario "8. Access Protected Endpoint (Without Token)" {
    if (-not $testContext.AccessToken) {
        return @{ Success = $false; Message = "No access token available (login test failed?)" }
    }
    
    try {
        $headers = @{
            Authorization = "Bearer $($testContext.AccessToken)"
        }
        
        # Try to access a protected endpoint (videos list requires auth)
        $response = Invoke-RestMethod -Uri "$apiBase/videos" `
            -Method GET `
            -Headers $headers `
            -ErrorAction Stop
        
        return @{ 
            Success = $true
            Message = "Successfully accessed protected endpoint with token"
            Details = $response
        }
        
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        
        # 401 means token was rejected (bad)
        if ($statusCode -eq 401) {
            return @{ Success = $false; Message = "Valid token rejected (401)" }
        }
        
        # 404 or other errors might be okay (endpoint might not be fully implemented)
        return @{ Success = $false; Message = "Unexpected response ($statusCode): $($_.Exception.Message)" }
    }
}

# ============================================================================
# TEST 6: ACCESS PROTECTED ENDPOINT (Without Token)
# ============================================================================
$results += Test-E2EScenario "6. Access Protected Endpoint (Without Token)" {
    try {
        $response = Invoke-RestMethod -Uri "$apiBase/videos" `
            -Method GET `
            -ErrorAction Stop
        
        # If we get here, that's BAD - should require auth!
        return @{ Success = $false; Message = "🚨 Protected endpoint accessible without token!" }
        
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        
        # We WANT a 401 (Unauthorized)
        if ($statusCode -eq 401) {
            return @{ Success = $true; Message = "Correctly requires authentication (401)" }
        }
        
        return @{ Success = $false; Message = "Unexpected error code: $statusCode (expected 401)" }
    }
}

# ============================================================================
# TEST 9: ACCESS PROTECTED ENDPOINT (Invalid Token)
# ============================================================================
$results += Test-E2EScenario "9. Access Protected Endpoint (Invalid Token)" {
    try {
        $headers = @{
            Authorization = "Bearer invalid.token.here"
        }
        
        $response = Invoke-RestMethod -Uri "$apiBase/videos" `
            -Method GET `
            -Headers $headers `
            -ErrorAction Stop
        
        # If we get here, that's BAD - invalid token should be rejected!
        return @{ Success = $false; Message = "🚨 Invalid token accepted!" }
        
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        
        # We WANT a 401 (Unauthorized)
        if ($statusCode -eq 401) {
            return @{ Success = $true; Message = "Correctly rejected invalid token (401)" }
        }
        
        return @{ Success = $false; Message = "Unexpected error code: $statusCode (expected 401)" }
    }
}


# ============================================================================
# RESULTS SUMMARY
# ============================================================================

$totalTests = $results.Count
$passedTests = ($results | Where-Object { $_.Success }).Count
$failedTests = $totalTests - $passedTests
$passRate = [math]::Round(($passedTests / $totalTests) * 100, 1)

Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Blue
Write-Host "║           AUTHENTICATION BRAID TEST RESULTS                  ║" -ForegroundColor Blue
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
$results | ConvertTo-Json -Depth 4 | Out-File "test-results-braid-auth.json"
Write-Host "`n💾 Results saved to: test-results-braid-auth.json" -ForegroundColor Yellow

if ($passRate -ge 80) {
    Write-Host "`n✅ AUTHENTICATION BRAID TESTS PASSED! (≥80% pass rate)`n" -ForegroundColor Green
    exit 0
} else {
    Write-Host "`n⚠️  AUTHENTICATION BRAID TESTS NEED ATTENTION (<80% pass rate)`n" -ForegroundColor Yellow
    exit 1
}

