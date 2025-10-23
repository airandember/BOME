# Service Connection Testing Script
# Tests all backend services and endpoints

Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║          MISSION 1: SERVICE CONNECTION TESTS                ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════════╝`n" -ForegroundColor Cyan

$baseUrl = "http://localhost:8080"
$results = @()

# Test function
function Test-Endpoint {
    param(
        [string]$Name,
        [string]$Url,
        [string]$Method = "GET",
        [hashtable]$Headers = @{},
        [string]$Body = $null
    )
    
    Write-Host "Testing: $Name... " -NoNewline
    
    try {
        $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
        
        $params = @{
            Uri = $Url
            Method = $Method
            Headers = $Headers
            TimeoutSec = 5
            ErrorAction = "Stop"
        }
        
        if ($Body) {
            $params.Body = $Body
            $params.ContentType = "application/json"
        }
        
        $response = Invoke-WebRequest @params
        $stopwatch.Stop()
        
        $result = @{
            Name = $Name
            Status = "✅ PASS"
            StatusCode = $response.StatusCode
            ResponseTime = "$($stopwatch.ElapsedMilliseconds)ms"
            Success = $true
        }
        
        Write-Host "✅ PASS ($($stopwatch.ElapsedMilliseconds)ms)" -ForegroundColor Green
        
    } catch {
        $stopwatch.Stop()
        
        $statusCode = if ($_.Exception.Response) { 
            $_.Exception.Response.StatusCode.value__ 
        } else { 
            "ERROR" 
        }
        
        $result = @{
            Name = $Name
            Status = "❌ FAIL"
            StatusCode = $statusCode
            ResponseTime = "$($stopwatch.ElapsedMilliseconds)ms"
            Success = $false
            Error = $_.Exception.Message
        }
        
        Write-Host "❌ FAIL ($statusCode)" -ForegroundColor Red
    }
    
    return $result
}

Write-Host "`n📊 Testing Service Connections...`n" -ForegroundColor Yellow

# 1. Health Check Endpoints
Write-Host "`n1️⃣  HEALTH CHECK ENDPOINTS" -ForegroundColor Cyan
$results += Test-Endpoint "Main Health Check" "$baseUrl/health"
$results += Test-Endpoint "Liveness Probe" "$baseUrl/health/live"
$results += Test-Endpoint "Readiness Probe" "$baseUrl/health/ready"

# 2. Authentication Endpoints
Write-Host "`n2️⃣  AUTHENTICATION ENDPOINTS" -ForegroundColor Cyan
$results += Test-Endpoint "Login Endpoint" "$baseUrl/api/v1/auth/login" "POST" @{} '{"email":"test@example.com","password":"test123"}'
$results += Test-Endpoint "Register Endpoint" "$baseUrl/api/v1/auth/register" "POST" @{} '{"email":"new@example.com","password":"test123"}'
$results += Test-Endpoint "Logout Endpoint" "$baseUrl/api/v1/auth/logout" "POST"

# 3. Subscription Endpoints
Write-Host "`n3️⃣  SUBSCRIPTION ENDPOINTS" -ForegroundColor Cyan
$results += Test-Endpoint "Subscription Plans (Active)" "$baseUrl/api/v1/subscription-plans/active"
$results += Test-Endpoint "Subscription Plans (Promoted)" "$baseUrl/api/v1/subscription-plans/promoted"

# 4. Video Endpoints
Write-Host "`n4️⃣  VIDEO ENDPOINTS" -ForegroundColor Cyan
$results += Test-Endpoint "Videos List" "$baseUrl/api/v1/videos"
$results += Test-Endpoint "Videos Test" "$baseUrl/api/v1/videos/test"

# 5. Performance Endpoints
Write-Host "`n5️⃣  PERFORMANCE ENDPOINTS" -ForegroundColor Cyan
$results += Test-Endpoint "Optimization Test" "$baseUrl/api/v1/test/optimization"
$results += Test-Endpoint "Performance Metrics" "$baseUrl/api/v1/performance/metrics"

# 6. Stripe Webhook (Public)
Write-Host "`n6️⃣  WEBHOOK ENDPOINTS" -ForegroundColor Cyan
$results += Test-Endpoint "Stripe Webhook" "$baseUrl/api/v1/webhooks/stripe" "POST" @{} '{}'

# Calculate statistics
$totalTests = $results.Count
$passedTests = ($results | Where-Object { $_.Success }).Count
$failedTests = $totalTests - $passedTests
$passRate = [math]::Round(($passedTests / $totalTests) * 100, 1)

$avgResponseTime = ($results | Where-Object { $_.ResponseTime -match '(\d+)ms' } | ForEach-Object { 
    [int]($_.ResponseTime -replace 'ms','') 
} | Measure-Object -Average).Average

# Display Results
Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Blue
Write-Host "║                    TEST RESULTS SUMMARY                      ║" -ForegroundColor Blue
Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Blue

Write-Host "`n📊 STATISTICS:" -ForegroundColor Cyan
Write-Host "   Total Tests:        $totalTests" -ForegroundColor White
Write-Host "   Passed:             $passedTests ✅" -ForegroundColor Green
Write-Host "   Failed:             $failedTests ❌" -ForegroundColor $(if($failedTests -eq 0){"Green"}else{"Red"})
Write-Host "   Pass Rate:          $passRate%" -ForegroundColor $(if($passRate -ge 80){"Green"}elseif($passRate -ge 60){"Yellow"}else{"Red"})
Write-Host "   Avg Response Time:  $([math]::Round($avgResponseTime, 0))ms" -ForegroundColor White

Write-Host "`n📋 DETAILED RESULTS:" -ForegroundColor Cyan
foreach ($result in $results) {
    $color = if ($result.Success) { "Green" } else { "Red" }
    Write-Host "   $($result.Status) $($result.Name) - $($result.StatusCode) ($($result.ResponseTime))" -ForegroundColor $color
}

# Export results to JSON
$results | ConvertTo-Json -Depth 3 | Out-File "test-results-services.json"
Write-Host "`n💾 Results saved to: test-results-services.json" -ForegroundColor Yellow

# Return exit code based on pass rate
if ($passRate -ge 80) {
    Write-Host "`n✅ SERVICE CONNECTION TESTS PASSED! (>80% pass rate)`n" -ForegroundColor Green
    exit 0
} else {
    Write-Host "`n⚠️  SERVICE CONNECTION TESTS DEGRADED! (<80% pass rate)`n" -ForegroundColor Yellow
    exit 1
}


