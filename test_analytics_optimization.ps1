# Video Analytics Optimization Test Script
# Tests the new async buffering and resilience features

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$Token = "",
    [int]$ConcurrentRequests = 50,
    [int]$VideoId = 12845
)

Write-Host "🧪 Video Analytics Optimization Test" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

# Test 1: Health Check
Write-Host "Test 1: Health Check" -ForegroundColor Yellow
try {
    $healthResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/analytics/health" -Method Get
    Write-Host "✅ Health Status: $($healthResponse.status)" -ForegroundColor Green
    Write-Host "   Circuit Open: $($healthResponse.resilience.circuit_open)" -ForegroundColor Cyan
    Write-Host "   Sample Rate: $($healthResponse.resilience.sample_rate)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Health check failed: $_" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Test 2: Check if token is provided
if (-not $Token) {
    Write-Host "⚠️  No JWT token provided. Skipping authenticated tests." -ForegroundColor Yellow
    Write-Host "   To test with authentication, run:" -ForegroundColor Yellow
    Write-Host "   .\test_analytics_optimization.ps1 -Token 'YOUR_JWT_TOKEN'" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "✅ Basic health check passed!" -ForegroundColor Green
    exit 0
}

# Test 3: Single Request Test
Write-Host "Test 2: Single Analytics Request" -ForegroundColor Yellow
try {
    $body = @{
        video_id = $VideoId
        watched_duration = 30
        watched_percentage = 10.0
    } | ConvertTo-Json

    $headers = @{
        "Authorization" = "Bearer $Token"
        "Content-Type" = "application/json"
    }

    $startTime = Get-Date
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/v1/analytics/video/track" -Method Post -Body $body -Headers $headers
    $duration = (Get-Date) - $startTime

    Write-Host "✅ Response: $($response.status)" -ForegroundColor Green
    Write-Host "   Duration: $($duration.TotalMilliseconds)ms" -ForegroundColor Cyan
    
    if ($duration.TotalMilliseconds -lt 50) {
        Write-Host "   ⚡ Fast response! (async buffer working)" -ForegroundColor Green
    } else {
        Write-Host "   ⚠️  Slow response (may be using direct DB write)" -ForegroundColor Yellow
    }
} catch {
    Write-Host "❌ Single request failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 4: Concurrent Load Test
Write-Host "Test 3: Concurrent Load Test ($ConcurrentRequests requests)" -ForegroundColor Yellow
Write-Host "Simulating concurrent video viewers..." -ForegroundColor Cyan

$results = @()
$startTime = Get-Date

1..$ConcurrentRequests | ForEach-Object -Parallel {
    $body = @{
        video_id = $using:VideoId
        watched_duration = (Get-Random -Minimum 10 -Maximum 300)
        watched_percentage = (Get-Random -Minimum 1 -Maximum 100)
    } | ConvertTo-Json

    $headers = @{
        "Authorization" = "Bearer $using:Token"
        "Content-Type" = "application/json"
    }

    try {
        $requestStart = Get-Date
        $response = Invoke-RestMethod -Uri "$using:BaseUrl/api/v1/analytics/video/track" -Method Post -Body $body -Headers $headers -ErrorAction Stop
        $requestDuration = ((Get-Date) - $requestStart).TotalMilliseconds
        
        return @{
            Success = $true
            Status = $response.status
            Duration = $requestDuration
        }
    } catch {
        return @{
            Success = $false
            Error = $_.Exception.Message
            Duration = 0
        }
    }
} -ThrottleLimit 25 | ForEach-Object {
    $results += $_
}

$totalDuration = ((Get-Date) - $startTime).TotalSeconds
$successCount = ($results | Where-Object { $_.Success }).Count
$failCount = $results.Count - $successCount
$avgDuration = ($results | Where-Object { $_.Success } | Measure-Object -Property Duration -Average).Average

Write-Host ""
Write-Host "Results:" -ForegroundColor Cyan
Write-Host "  Total Requests: $($results.Count)" -ForegroundColor White
Write-Host "  Successful: $successCount" -ForegroundColor Green
Write-Host "  Failed: $failCount" -ForegroundColor $(if ($failCount -gt 0) { "Red" } else { "Green" })
Write-Host "  Total Time: $([math]::Round($totalDuration, 2))s" -ForegroundColor White
Write-Host "  Throughput: $([math]::Round($results.Count / $totalDuration, 2)) req/sec" -ForegroundColor Cyan
Write-Host "  Avg Response Time: $([math]::Round($avgDuration, 2))ms" -ForegroundColor Cyan
Write-Host ""

if ($avgDuration -lt 50) {
    Write-Host "✅ Excellent! Async buffering is working optimally." -ForegroundColor Green
} elseif ($avgDuration -lt 200) {
    Write-Host "⚠️  Good, but may be using fallback to direct DB writes." -ForegroundColor Yellow
} else {
    Write-Host "❌ Slow responses. Check Redis and backend logs." -ForegroundColor Red
}

# Test 5: Check Health After Load
Write-Host ""
Write-Host "Test 4: Health Check After Load" -ForegroundColor Yellow
try {
    Start-Sleep -Seconds 2
    $healthResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/analytics/health" -Method Get
    Write-Host "   Circuit Open: $($healthResponse.resilience.circuit_open)" -ForegroundColor $(if ($healthResponse.resilience.circuit_open) { "Red" } else { "Green" })
    Write-Host "   Failure Count: $($healthResponse.resilience.failure_count)" -ForegroundColor Cyan
    Write-Host "   Request Count: $($healthResponse.resilience.request_count)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Health check failed: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "🎉 Test Complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host "1. Wait 5-10 seconds for buffer to flush" -ForegroundColor White
Write-Host "2. Check backend logs for flush messages" -ForegroundColor White
Write-Host "3. Query database: SELECT COUNT(*) FROM watch_history WHERE video_id = $VideoId;" -ForegroundColor White
Write-Host "4. Check Redis buffer: redis-cli LLEN analytics:video_tracking_buffer" -ForegroundColor White

