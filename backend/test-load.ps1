# Load Testing Script
# Simulates 100 concurrent users

Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Red
Write-Host "║            MISSION 1: LOAD TESTING (100 USERS)              ║" -ForegroundColor Red
Write-Host "╚══════════════════════════════════════════════════════════════╝`n" -ForegroundColor Red

$baseUrl = "http://localhost:8080"
$totalUsers = 100
$duration = 30 # seconds

Write-Host "🔥 Load Test Configuration:" -ForegroundColor Yellow
Write-Host "   Concurrent Users: $totalUsers" -ForegroundColor White
Write-Host "   Test Duration:    ${duration}s" -ForegroundColor White
Write-Host "   Target:           $baseUrl`n" -ForegroundColor White

$endpoints = @(
    @{ Path = "/health"; Weight = 30 },
    @{ Path = "/api/v1/subscription-plans/active"; Weight = 20 },
    @{ Path = "/api/v1/videos/test"; Weight = 20 },
    @{ Path = "/api/v1/performance/metrics"; Weight = 15 },
    @{ Path = "/health/ready"; Weight = 15 }
)

Write-Host "📊 Starting load test..." -ForegroundColor Cyan
Write-Host "⏰ Running for ${duration} seconds...`n" -ForegroundColor Yellow

$startTime = Get-Date

# Create worker function that returns results
$workerScript = {
    param($baseUrl, $endpoints, $duration)
    
    $results = @()
    $endTime = (Get-Date).AddSeconds($duration)
    $random = New-Object System.Random
    
    while ((Get-Date) -lt $endTime) {
        # Select random endpoint based on weight
        $rand = $random.Next(0, 100)
        $cumulative = 0
        $selectedEndpoint = $endpoints[0].Path
        
        foreach ($ep in $endpoints) {
            $cumulative += $ep.Weight
            if ($rand -lt $cumulative) {
                $selectedEndpoint = $ep.Path
                break
            }
        }
        
        $url = "$baseUrl$selectedEndpoint"
        
        try {
            $sw = [System.Diagnostics.Stopwatch]::StartNew()
            $response = Invoke-WebRequest -Uri $url -Method GET -TimeoutSec 5 -UseBasicParsing
            $sw.Stop()
            
            $results += @{
                Success = $true
                StatusCode = $response.StatusCode
                ResponseTime = $sw.ElapsedMilliseconds
                Endpoint = $selectedEndpoint
            }
        } catch {
            $sw.Stop()
            $results += @{
                Success = $false
                StatusCode = if ($_.Exception.Response) { $_.Exception.Response.StatusCode.value__ } else { 0 }
                ResponseTime = $sw.ElapsedMilliseconds
                Endpoint = $selectedEndpoint
            }
        }
        
        # Small random delay between requests (50-200ms)
        Start-Sleep -Milliseconds ($random.Next(50, 200))
    }
    
    return $results
}

# Launch concurrent users
Write-Host "🚀 Launching $totalUsers concurrent users..." -ForegroundColor Green
$jobs = 1..$totalUsers | ForEach-Object {
    Start-Job -ScriptBlock $workerScript -ArgumentList $baseUrl, $endpoints, $duration
}

# Monitor progress
$progressInterval = 5
for ($i = 0; $i -lt $duration; $i += $progressInterval) {
    Start-Sleep -Seconds $progressInterval
    $elapsed = [math]::Min($i + $progressInterval, $duration)
    $percent = [math]::Round(($elapsed / $duration) * 100)
    
    Write-Host "⏱️  Progress: ${elapsed}s / ${duration}s ($percent%) | Active Jobs: $($jobs.Count)" -ForegroundColor Cyan
}

# Wait for all jobs to complete and collect results
Write-Host "`n⏳ Waiting for all users to complete..." -ForegroundColor Yellow
$allResults = @()
$jobs | ForEach-Object {
    $jobResults = Receive-Job -Job $_
    if ($jobResults) {
        $allResults += $jobResults
    }
}
$jobs | Remove-Job

$endTime = Get-Date
$actualDuration = ($endTime - $startTime).TotalSeconds

# Analyze results
Write-Host "`n📊 Analyzing results...`n" -ForegroundColor Cyan

$totalRequests = $allResults.Count
if ($totalRequests -eq 0) {
    Write-Host "❌ NO REQUESTS COMPLETED - Test failed to execute`n" -ForegroundColor Red
    exit 1
}
$successfulRequests = ($allResults | Where-Object { $_.Success }).Count
$failedRequests = $totalRequests - $successfulRequests
$successRate = [math]::Round(($successfulRequests / $totalRequests) * 100, 2)

$responseTimes = $allResults | Where-Object { $_.Success } | Select-Object -ExpandProperty ResponseTime
$avgResponseTime = [math]::Round(($responseTimes | Measure-Object -Average).Average, 2)
$minResponseTime = ($responseTimes | Measure-Object -Minimum).Minimum
$maxResponseTime = ($responseTimes | Measure-Object -Maximum).Maximum
$p95ResponseTime = $responseTimes | Sort-Object | Select-Object -Index ([math]::Floor($responseTimes.Count * 0.95))

$requestsPerSecond = [math]::Round($totalRequests / $actualDuration, 2)

# Group by endpoint
$byEndpoint = $allResults | Group-Object -Property Endpoint | ForEach-Object {
    $endpointRequests = $_.Group
    $endpointSuccess = ($endpointRequests | Where-Object { $_.Success }).Count
    $endpointRate = [math]::Round(($endpointSuccess / $endpointRequests.Count) * 100, 1)
    $endpointAvg = [math]::Round((($endpointRequests | Where-Object { $_.Success } | Select-Object -ExpandProperty ResponseTime) | Measure-Object -Average).Average, 1)
    
    @{
        Endpoint = $_.Name
        Requests = $endpointRequests.Count
        Success = $endpointSuccess
        SuccessRate = "$endpointRate%"
        AvgResponseTime = "${endpointAvg}ms"
    }
}

# Display Results
Write-Host "╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Blue
Write-Host "║                   LOAD TEST RESULTS                          ║" -ForegroundColor Blue
Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Blue

Write-Host "`n🎯 OVERALL METRICS:" -ForegroundColor Cyan
Write-Host "   Concurrent Users:     $totalUsers" -ForegroundColor White
Write-Host "   Test Duration:        $([math]::Round($actualDuration, 1))s" -ForegroundColor White
Write-Host "   Total Requests:       $totalRequests" -ForegroundColor White
Write-Host "   Successful Requests:  $successfulRequests ✅" -ForegroundColor Green
Write-Host "   Failed Requests:      $failedRequests ❌" -ForegroundColor $(if($failedRequests -eq 0){"Green"}else{"Red"})
Write-Host "   Success Rate:         $successRate%" -ForegroundColor $(if($successRate -ge 95){"Green"}elseif($successRate -ge 90){"Yellow"}else{"Red"})
Write-Host "   Requests/Second:      $requestsPerSecond" -ForegroundColor White

Write-Host "`n⚡ RESPONSE TIME METRICS:" -ForegroundColor Cyan
Write-Host "   Average:    ${avgResponseTime}ms" -ForegroundColor White
Write-Host "   Minimum:    ${minResponseTime}ms" -ForegroundColor Green
Write-Host "   Maximum:    ${maxResponseTime}ms" -ForegroundColor $(if($maxResponseTime -lt 1000){"Yellow"}else{"Red"})
Write-Host "   P95:        ${p95ResponseTime}ms" -ForegroundColor $(if($p95ResponseTime -lt 500){"Green"}elseif($p95ResponseTime -lt 1000){"Yellow"}else{"Red"})

Write-Host "`n📋 BY ENDPOINT:" -ForegroundColor Cyan
$byEndpoint | ForEach-Object {
    Write-Host "   $($_.Endpoint)" -ForegroundColor White
    Write-Host "      Requests: $($_.Requests) | Success: $($_.SuccessRate) | Avg: $($_.AvgResponseTime)" -ForegroundColor Gray
}

# Save results
$reportData = @{
    Configuration = @{
        ConcurrentUsers = $totalUsers
        Duration = $actualDuration
        BaseUrl = $baseUrl
    }
    Summary = @{
        TotalRequests = $totalRequests
        SuccessfulRequests = $successfulRequests
        FailedRequests = $failedRequests
        SuccessRate = $successRate
        RequestsPerSecond = $requestsPerSecond
    }
    ResponseTimes = @{
        Average = $avgResponseTime
        Minimum = $minResponseTime
        Maximum = $maxResponseTime
        P95 = $p95ResponseTime
    }
    ByEndpoint = $byEndpoint
}

$reportData | ConvertTo-Json -Depth 4 | Out-File "test-results-load.json"
Write-Host "`n💾 Detailed results saved to: test-results-load.json" -ForegroundColor Yellow

# Determine pass/fail
if ($successRate -ge 95 -and $avgResponseTime -lt 100 -and $p95ResponseTime -lt 500) {
    Write-Host "`n✅ LOAD TEST PASSED! Backend performs excellently under load!`n" -ForegroundColor Green
    exit 0
} elseif ($successRate -ge 90) {
    Write-Host "`n⚠️  LOAD TEST ACCEPTABLE - Some performance degradation detected`n" -ForegroundColor Yellow
    exit 0
} else {
    Write-Host "`n❌ LOAD TEST FAILED - Significant performance issues detected`n" -ForegroundColor Red
    exit 1
}

