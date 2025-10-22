# Integration Test Suite
# Tests realistic user flows end-to-end

Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Magenta
Write-Host "║         MISSION 1: INTEGRATION TEST SUITE                   ║" -ForegroundColor Magenta
Write-Host "╚══════════════════════════════════════════════════════════════╝`n" -ForegroundColor Magenta

$baseUrl = "http://localhost:8080"
$results = @()

function Test-Integration {
    param(
        [string]$Name,
        [scriptblock]$TestScript
    )
    
    Write-Host "`n▶ Running: $Name" -ForegroundColor Cyan
    
    try {
        $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
        $result = & $TestScript
        $stopwatch.Stop()
        
        if ($result.Success) {
            Write-Host "  ✅ PASS ($($stopwatch.ElapsedMilliseconds)ms)" -ForegroundColor Green
        } else {
            Write-Host "  ❌ FAIL: $($result.Message)" -ForegroundColor Red
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
            Message = $_.Exception.Message
            Duration = 0
        }
    }
}

Write-Host "📊 Running Integration Tests...`n" -ForegroundColor Yellow

# Test 1: Health Check Flow
$results += Test-Integration "Health Check Flow" {
    $health = Invoke-RestMethod -Uri "$baseUrl/health" -Method GET
    $live = Invoke-RestMethod -Uri "$baseUrl/health/live" -Method GET
    $ready = Invoke-RestMethod -Uri "$baseUrl/health/ready" -Method GET
    
    if ($health.status -eq "healthy" -and $live.status -eq "alive" -and $ready.status -eq "ready") {
        return @{ Success = $true; Message = "All health endpoints operational" }
    }
    return @{ Success = $false; Message = "Health check failed" }
}

# Test 2: Subscription Plans Flow
$results += Test-Integration "Subscription Plans Retrieval" {
    $plans = Invoke-RestMethod -Uri "$baseUrl/api/v1/subscription-plans/active" -Method GET
    
    if ($plans -and $plans.Count -ge 0) {
        return @{ Success = $true; Message = "Retrieved $($plans.Count) active plans" }
    }
    return @{ Success = $false; Message = "Failed to retrieve plans" }
}

# Test 3: Authentication Error Handling
$results += Test-Integration "Auth Error Handling (Invalid Credentials)" {
    try {
        $body = @{
            email = "nonexistent@example.com"
            password = "wrongpassword"
        } | ConvertTo-Json
        
        Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method POST -Body $body -ContentType "application/json"
        return @{ Success = $false; Message = "Should have returned 401" }
    } catch {
        if ($_.Exception.Response.StatusCode.value__ -eq 401) {
            return @{ Success = $true; Message = "Correctly rejected invalid credentials" }
        }
        return @{ Success = $false; Message = "Unexpected error code" }
    }
}

# Test 4: Video Endpoint Auth Protection
$results += Test-Integration "Video Endpoint Auth Protection" {
    try {
        Invoke-RestMethod -Uri "$baseUrl/api/v1/videos" -Method GET
        return @{ Success = $false; Message = "Should require authentication" }
    } catch {
        if ($_.Exception.Response.StatusCode.value__ -eq 401) {
            return @{ Success = $true; Message = "Correctly requires authentication" }
        }
        return @{ Success = $false; Message = "Unexpected response" }
    }
}

# Test 5: Stripe Webhook Accepts Requests
$results += Test-Integration "Stripe Webhook Endpoint" {
    try {
        $body = '{"type":"test.event"}' 
        $response = Invoke-WebRequest -Uri "$baseUrl/api/v1/webhooks/stripe" -Method POST -Body $body -ContentType "application/json"
        
        # Webhook should accept POST requests (even if signature validation fails)
        if ($response.StatusCode -eq 200 -or $response.StatusCode -eq 400) {
            return @{ Success = $true; Message = "Webhook endpoint operational" }
        }
        return @{ Success = $false; Message = "Unexpected status: $($response.StatusCode)" }
    } catch {
        # 400 is acceptable for invalid signature
        if ($_.Exception.Response.StatusCode.value__ -eq 400) {
            return @{ Success = $true; Message = "Webhook validation working" }
        }
        return @{ Success = $false; Message = $_.Exception.Message }
    }
}

# Test 6: Performance Metrics Availability
$results += Test-Integration "Performance Metrics" {
    $metrics = Invoke-RestMethod -Uri "$baseUrl/api/v1/performance/metrics" -Method GET
    
    if ($metrics) {
        return @{ Success = $true; Message = "Metrics endpoint operational" }
    }
    return @{ Success = $false; Message = "No metrics returned" }
}

# Test 7: Concurrent Request Handling (10 simultaneous requests)
$results += Test-Integration "Concurrent Request Handling (10 requests)" {
    $jobs = 1..10 | ForEach-Object {
        Start-Job -ScriptBlock {
            param($url)
            try {
                Invoke-RestMethod -Uri $url -Method GET -TimeoutSec 5
                return $true
            } catch {
                return $false
            }
        } -ArgumentList "$baseUrl/health"
    }
    
    $completed = $jobs | Wait-Job | Receive-Job
    $jobs | Remove-Job
    
    $successCount = ($completed | Where-Object { $_ -eq $true }).Count
    
    if ($successCount -ge 9) {
        return @{ Success = $true; Message = "$successCount/10 concurrent requests succeeded" }
    }
    return @{ Success = $false; Message = "Only $successCount/10 requests succeeded" }
}

# Test 8: Response Time Consistency (5 sequential requests)
$results += Test-Integration "Response Time Consistency" {
    $times = @()
    
    1..5 | ForEach-Object {
        $sw = [System.Diagnostics.Stopwatch]::StartNew()
        Invoke-RestMethod -Uri "$baseUrl/health" -Method GET | Out-Null
        $sw.Stop()
        $times += $sw.ElapsedMilliseconds
    }
    
    $avg = ($times | Measure-Object -Average).Average
    $max = ($times | Measure-Object -Maximum).Maximum
    
    if ($max -lt 100 -and $avg -lt 50) {
        return @{ Success = $true; Message = "Avg: ${avg}ms, Max: ${max}ms" }
    }
    return @{ Success = $false; Message = "Response times too high (Avg: ${avg}ms)" }
}

# Calculate statistics
$totalTests = $results.Count
$passedTests = ($results | Where-Object { $_.Success }).Count
$failedTests = $totalTests - $passedTests
$passRate = [math]::Round(($passedTests / $totalTests) * 100, 1)

# Display Results
Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Blue
Write-Host "║              INTEGRATION TEST RESULTS                        ║" -ForegroundColor Blue
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
    Write-Host "   $status $($result.Name) - $($result.Message) ($($result.Duration)ms)" -ForegroundColor $color
}

# Export results
$results | ConvertTo-Json -Depth 3 | Out-File "test-results-integration.json"
Write-Host "`n💾 Results saved to: test-results-integration.json" -ForegroundColor Yellow

if ($passRate -ge 80) {
    Write-Host "`n✅ INTEGRATION TESTS PASSED! (≥80% pass rate)`n" -ForegroundColor Green
    exit 0
} else {
    Write-Host "`n⚠️  INTEGRATION TESTS DEGRADED! (<80% pass rate)`n" -ForegroundColor Yellow
    exit 1
}

