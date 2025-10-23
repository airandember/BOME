# Test WebSocket Real-Time Updates
# This script tests the WebSocket connection and real-time events

$baseUrl = "http://localhost:8080/api/v1"
$adminEmail = "super_admin@example.com"
$adminPassword = "your-admin-password-here"

Write-Host "`n🌐 WEBSOCKET REAL-TIME TEST SCRIPT" -ForegroundColor Cyan
Write-Host "===================================`n" -ForegroundColor Cyan

# Step 1: Login as admin
Write-Host "Step 1: Logging in as admin..." -ForegroundColor Yellow
$loginBody = @{
    email = $adminEmail
    password = $adminPassword
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $token = $loginResponse.access_token
    Write-Host "✅ Login successful! Token received." -ForegroundColor Green
} catch {
    Write-Host "❌ Login failed: $_" -ForegroundColor Red
    Write-Host "Please update admin credentials in this script" -ForegroundColor Yellow
    exit 1
}

# Create headers with auth token
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# Step 2: Check WebSocket stats
Write-Host "`nStep 2: Checking WebSocket hub stats..." -ForegroundColor Yellow
try {
    $stats = Invoke-RestMethod -Uri "$baseUrl/ws/admin/stats" -Method GET -Headers $headers
    Write-Host "✅ WebSocket hub is running!" -ForegroundColor Green
    Write-Host "   Active connections: $($stats.stats.active_connections)" -ForegroundColor Cyan
    Write-Host "   Total connections: $($stats.stats.total_connections)" -ForegroundColor Cyan
    Write-Host "   Total messages: $($stats.stats.total_messages)" -ForegroundColor Cyan
    Write-Host "   Total broadcasts: $($stats.stats.total_broadcasts)" -ForegroundColor Cyan
    Write-Host "   Uptime: $([math]::Round($stats.stats.uptime_seconds, 2)) seconds" -ForegroundColor Cyan
} catch {
    Write-Host "❌ WebSocket stats check failed: $_" -ForegroundColor Red
}

Write-Host "`n" -ForegroundColor White
Write-Host "====================================" -ForegroundColor Cyan
Write-Host "📝 WEBSOCKET CONNECTION INSTRUCTIONS" -ForegroundColor Yellow
Write-Host "====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "To test WebSocket in your browser:" -ForegroundColor White
Write-Host ""
Write-Host "1. Open browser console (F12)" -ForegroundColor Yellow
Write-Host "2. Navigate to admin subscribers page" -ForegroundColor Yellow
Write-Host "3. The WebSocket should auto-connect" -ForegroundColor Yellow
Write-Host "4. Check console for connection logs" -ForegroundColor Yellow
Write-Host ""
Write-Host "WebSocket URL:" -ForegroundColor Magenta
Write-Host "   ws://localhost:8080/api/v1/ws/admin" -ForegroundColor White
Write-Host ""
Write-Host "Expected console logs:" -ForegroundColor Magenta
Write-Host "   ✅ AdminWS: Connected!" -ForegroundColor Green
Write-Host "   📨 AdminWS: Event received - connected" -ForegroundColor Cyan
Write-Host "   🎉 AdminWS: Connection confirmed by server" -ForegroundColor Green
Write-Host ""

# Step 3: Trigger a real-time event by updating a subscriber
Write-Host "`nStep 3: Testing real-time events..." -ForegroundColor Yellow
Write-Host "Getting first subscriber to update..." -ForegroundColor White

try {
    # Get subscribers
    $subscribers = Invoke-RestMethod -Uri "$baseUrl/admin/subscribers?page=1&limit=1" -Method GET -Headers $headers
    
    if ($subscribers.subscribers.Count -gt 0) {
        $firstSub = $subscribers.subscribers[0]
        $subId = $firstSub.id
        
        Write-Host "Found subscriber: $($firstSub.email) (ID: $subId)" -ForegroundColor Cyan
        Write-Host "`nUpdating subscriber to trigger WebSocket event..." -ForegroundColor Yellow
        
        # Update subscriber (just touch the updated_at field)
        $updateBody = @{
            last_login = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ss")
        } | ConvertTo-Json
        
        try {
            $updated = Invoke-RestMethod -Uri "$baseUrl/admin/subscribers/$subId" -Method PUT -Body $updateBody -Headers $headers -ContentType "application/json"
            Write-Host "✅ Subscriber updated!" -ForegroundColor Green
            Write-Host "📡 WebSocket event should have been broadcasted!" -ForegroundColor Magenta
            Write-Host "   Event type: subscriber.updated" -ForegroundColor Cyan
            Write-Host "   Check your browser console for the update!" -ForegroundColor Yellow
        } catch {
            Write-Host "⚠️  Update failed: $_" -ForegroundColor Yellow
        }
    } else {
        Write-Host "⚠️  No subscribers found to test with" -ForegroundColor Yellow
    }
} catch {
    Write-Host "❌ Failed to get subscribers: $_" -ForegroundColor Red
}

Write-Host "`n====================================" -ForegroundColor Cyan
Write-Host "🧪 MANUAL WEBSOCKET TEST WITH WSCAT" -ForegroundColor Yellow
Write-Host "====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Install wscat globally:" -ForegroundColor White
Write-Host "   npm install -g wscat" -ForegroundColor Cyan
Write-Host ""
Write-Host "Connect to WebSocket (note: requires valid token):" -ForegroundColor White
Write-Host "   wscat -c 'ws://localhost:8080/api/v1/ws/admin'" -ForegroundColor Cyan
Write-Host ""
Write-Host "Send ping:" -ForegroundColor White
Write-Host '   {"type":"ping"}' -ForegroundColor Cyan
Write-Host ""
Write-Host "Expected response:" -ForegroundColor White
Write-Host '   {"type":"pong","timestamp":"...","data":{}}' -ForegroundColor Green
Write-Host ""

Write-Host "====================================" -ForegroundColor Cyan
Write-Host "✅ WEBSOCKET TEST COMPLETE!" -ForegroundColor Green
Write-Host "====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Open browser dev console" -ForegroundColor White
Write-Host "2. Navigate to /admin/streaming/subscribers" -ForegroundColor White
Write-Host "3. WebSocket will auto-connect" -ForegroundColor White
Write-Host "4. Update a subscriber to see real-time event" -ForegroundColor White
Write-Host "5. Check console for event logs" -ForegroundColor White
Write-Host ""
Write-Host "Event types to watch for:" -ForegroundColor Magenta
Write-Host "   • subscriber.created" -ForegroundColor Cyan
Write-Host "   • subscriber.updated" -ForegroundColor Cyan
Write-Host "   • kpi.update" -ForegroundColor Cyan
Write-Host "   • payment.received" -ForegroundColor Cyan
Write-Host "   • payment.failed" -ForegroundColor Cyan
Write-Host ""

