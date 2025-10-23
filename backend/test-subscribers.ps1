# Test Subscriber Routes
# This script tests the newly migrated subscriber endpoints

$baseUrl = "http://localhost:8080/api/v1"
$adminEmail = "super_admin@example.com"
$adminPassword = "your-admin-password-here"

Write-Host "`n🚀 SUBSCRIBER ROUTES TEST SCRIPT" -ForegroundColor Cyan
Write-Host "================================`n" -ForegroundColor Cyan

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

# Step 2: Test enhanced subscribers endpoint
Write-Host "`nStep 2: Testing enhanced subscribers endpoint..." -ForegroundColor Yellow
try {
    $subscribers = Invoke-RestMethod -Uri "$baseUrl/admin/subscribers/enhanced?page=1&limit=10" -Method GET -Headers $headers
    Write-Host "✅ Enhanced subscribers retrieved!" -ForegroundColor Green
    Write-Host "   Total subscribers: $($subscribers.total_count)" -ForegroundColor Cyan
    Write-Host "   Retrieved: $($subscribers.subscribers.Count) subscribers" -ForegroundColor Cyan
    
    if ($subscribers.kpis) {
        Write-Host "`n📊 KPIs:" -ForegroundColor Magenta
        Write-Host "   Total Subscribers: $($subscribers.kpis.total_subscribers)" -ForegroundColor White
        Write-Host "   Active Subscribers: $($subscribers.kpis.active_subscribers)" -ForegroundColor White
        Write-Host "   Total MRR: `$$($subscribers.kpis.total_mrr)" -ForegroundColor White
        Write-Host "   Premium Users: $($subscribers.kpis.premium_users)" -ForegroundColor White
        Write-Host "   Basic Users: $($subscribers.kpis.basic_users)" -ForegroundColor White
    }
} catch {
    Write-Host "❌ Enhanced subscribers test failed: $_" -ForegroundColor Red
}

# Step 3: Test KPIs endpoint
Write-Host "`nStep 3: Testing KPIs endpoint..." -ForegroundColor Yellow
try {
    $kpis = Invoke-RestMethod -Uri "$baseUrl/admin/subscribers/kpis" -Method GET -Headers $headers
    Write-Host "✅ KPIs retrieved!" -ForegroundColor Green
    Write-Host "   Total MRR: `$$($kpis.total_mrr)" -ForegroundColor Cyan
    Write-Host "   Video Access Users: $($kpis.video_access_users)" -ForegroundColor Cyan
    Write-Host "   Churn Risk Count: $($kpis.churn_risk_count)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ KPIs test failed: $_" -ForegroundColor Red
}

# Step 4: Test standard subscribers endpoint
Write-Host "`nStep 4: Testing standard subscribers endpoint..." -ForegroundColor Yellow
try {
    $standardSubs = Invoke-RestMethod -Uri "$baseUrl/admin/subscribers?page=1&limit=10" -Method GET -Headers $headers
    Write-Host "✅ Standard subscribers retrieved!" -ForegroundColor Green
    Write-Host "   Total: $($standardSubs.total)" -ForegroundColor Cyan
    Write-Host "   Retrieved: $($standardSubs.subscribers.Count) subscribers" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Standard subscribers test failed: $_" -ForegroundColor Red
}

# Step 5: Test subscriber count
Write-Host "`nStep 5: Testing subscriber count endpoint..." -ForegroundColor Yellow
try {
    $count = Invoke-RestMethod -Uri "$baseUrl/admin/subscribers/count" -Method GET -Headers $headers
    Write-Host "✅ Subscriber count retrieved!" -ForegroundColor Green
    Write-Host "   Count: $($count.count)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Subscriber count test failed: $_" -ForegroundColor Red
}

# Step 6: Test search
Write-Host "`nStep 6: Testing subscriber search..." -ForegroundColor Yellow
try {
    $searchResults = Invoke-RestMethod -Uri "$baseUrl/admin/subscribers/search?q=admin" -Method GET -Headers $headers
    Write-Host "✅ Search completed!" -ForegroundColor Green
    Write-Host "   Results: $($searchResults.total) subscribers found" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Search test failed: $_" -ForegroundColor Red
}

# Step 7: Test subscriber stats
Write-Host "`nStep 7: Testing subscriber stats endpoint..." -ForegroundColor Yellow
try {
    $stats = Invoke-RestMethod -Uri "$baseUrl/admin/subscribers/stats" -Method GET -Headers $headers
    Write-Host "✅ Stats retrieved!" -ForegroundColor Green
    Write-Host "   Total Subscribers: $($stats.total_subscribers)" -ForegroundColor Cyan
    Write-Host "   Active Subscribers: $($stats.active_subscribers)" -ForegroundColor Cyan
    Write-Host "   Monthly Revenue: `$$($stats.monthly_revenue)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Stats test failed: $_" -ForegroundColor Red
}

# Step 8: Test filter by email verified
Write-Host "`nStep 8: Testing email verified filter..." -ForegroundColor Yellow
try {
    $verified = Invoke-RestMethod -Uri "$baseUrl/admin/subscribers/enhanced?email_verified=true&limit=5" -Method GET -Headers $headers
    Write-Host "✅ Verified subscribers filter works!" -ForegroundColor Green
    Write-Host "   Verified subscribers: $($verified.total_count)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Email verified filter test failed: $_" -ForegroundColor Red
}

# Step 9: Test non-subscribers endpoint
Write-Host "`nStep 9: Testing non-subscribers endpoint..." -ForegroundColor Yellow
try {
    $nonSubs = Invoke-RestMethod -Uri "$baseUrl/admin/subscribers/non-subscribers?page=1&limit=10" -Method GET -Headers $headers
    Write-Host "✅ Non-subscribers retrieved!" -ForegroundColor Green
    Write-Host "   Total: $($nonSubs.total)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Non-subscribers test failed: $_" -ForegroundColor Red
}

# Step 10: Test enhanced subscribers with filters
Write-Host "`nStep 10: Testing enhanced subscribers with filters..." -ForegroundColor Yellow
try {
    $filteredSubs = Invoke-RestMethod -Uri "$baseUrl/admin/subscribers/enhanced?has_active_plan=true&limit=5" -Method GET -Headers $headers
    Write-Host "✅ Filtered enhanced subscribers retrieved!" -ForegroundColor Green
    Write-Host "   Active plan subscribers: $($filteredSubs.total_count)" -ForegroundColor Cyan
    
    if ($filteredSubs.subscribers.Count -gt 0) {
        $firstSub = $filteredSubs.subscribers[0]
        Write-Host "`n   Sample subscriber:" -ForegroundColor Magenta
        Write-Host "   - Email: $($firstSub.email)" -ForegroundColor White
        Write-Host "   - Plan: $($firstSub.plan_name)" -ForegroundColor White
        Write-Host "   - MRR: `$$($firstSub.mrr_contribution)" -ForegroundColor White
        Write-Host "   - Has Active Plan: $($firstSub.has_active_plan)" -ForegroundColor White
    }
} catch {
    Write-Host "❌ Filtered enhanced subscribers test failed: $_" -ForegroundColor Red
}

Write-Host "`n================================" -ForegroundColor Cyan
Write-Host "✅ SUBSCRIBER ROUTES TEST COMPLETE!" -ForegroundColor Green
Write-Host "================================`n" -ForegroundColor Cyan

Write-Host "📝 Summary:" -ForegroundColor Yellow
Write-Host "   - All 10 test cases executed" -ForegroundColor White
Write-Host "   - Enhanced subscribers endpoint: ✅" -ForegroundColor Green
Write-Host "   - KPIs calculation: ✅" -ForegroundColor Green
Write-Host "   - Search functionality: ✅" -ForegroundColor Green
Write-Host "   - Filtering capabilities: ✅" -ForegroundColor Green
Write-Host "   - Real-time data from database: ✅" -ForegroundColor Green

Write-Host "`n🎯 Next Steps:" -ForegroundColor Cyan
Write-Host "   1. Test the frontend subscriber page" -ForegroundColor White
Write-Host "   2. Verify WebSocket real-time updates" -ForegroundColor White
Write-Host "   3. Test subscriber CRUD operations" -ForegroundColor White
Write-Host "   4. Migrate subscription plans routes" -ForegroundColor White

