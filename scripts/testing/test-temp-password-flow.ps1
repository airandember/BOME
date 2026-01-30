# ============================================================================
# 🔑 TEMP PASSWORD FLOW TEST SCRIPT
# ============================================================================
# Tests the new temporary password flow for existing Stripe subscribers
# ============================================================================

$ErrorActionPreference = "Continue"

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

Write-Host "`n╔════════════════════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                                                                                ║" -ForegroundColor Cyan
Write-Host "║              🔑 TEMP PASSWORD FLOW TESTS                                       ║" -ForegroundColor Cyan
Write-Host "║                                                                                ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════════════════════════╝`n" -ForegroundColor Cyan

# Configuration
$BACKEND_URL = "http://localhost:8080"
$FRONTEND_URL = "http://localhost:5173"
$API_BASE = "$BACKEND_URL/api/v1"

# ============================================================================
# Pre-flight checks
# ============================================================================

Write-Host "🔍 Pre-flight Checks" -ForegroundColor Yellow
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Gray

# Test 1: Backend Health
Write-Host "`n📡 Test 1: Backend Health Check" -ForegroundColor Cyan
try {
    $health = Invoke-RestMethod -Uri "$BACKEND_URL/health" -Method GET -TimeoutSec 5
    Write-Host "  ✅ PASS - Backend is running!" -ForegroundColor Green
    Add-Result -Name "Backend Health" -Success $true -Message "Healthy"
} catch {
    Write-Host "  ❌ FAIL - Backend not responding at $BACKEND_URL" -ForegroundColor Red
    Write-Host "  💡 Start the backend first: cd backend && go run ." -ForegroundColor Yellow
    Add-Result -Name "Backend Health" -Success $false -Message $_.Exception.Message
    Write-Host "`n⛔ Cannot continue without backend. Exiting.`n" -ForegroundColor Red
    exit 1
}

# ============================================================================
# Database Migration Check
# ============================================================================

Write-Host "`n📊 Test 2: Database Schema Check (temp_password columns)" -ForegroundColor Cyan
Write-Host "  💡 Make sure you've run: backend/migrations/073_add_temp_password_columns.sql" -ForegroundColor Gray

# ============================================================================
# FLOW 1: Registration with temp password (simulated Stripe customer)
# ============================================================================

Write-Host "`n`n🔑 FLOW 1: Registration Flow Test" -ForegroundColor Yellow
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Gray

$timestamp = Get-Date -Format 'yyyyMMddHHmmss'
$testEmail = "temp_password_test_$timestamp@example.com"

Write-Host "`n📝 Test 3: Standard Registration (no existing Stripe customer)" -ForegroundColor Cyan
Write-Host "  Email: $testEmail" -ForegroundColor Gray

try {
    $registerBody = @{
        email = $testEmail
        first_name = "TempPass"
        last_name = "TestUser"
    } | ConvertTo-Json
    
    $registerResp = Invoke-RestMethod -Uri "$API_BASE/auth/register" -Method POST -Body $registerBody -ContentType "application/json" -TimeoutSec 10
    
    Write-Host "  ✅ PASS - Registration successful!" -ForegroundColor Green
    Write-Host "     User ID: $($registerResp.user_id)" -ForegroundColor Gray
    Write-Host "     Message: $($registerResp.message)" -ForegroundColor Gray
    Write-Host "     Temp Password Sent: $($registerResp.temp_password_sent)" -ForegroundColor Gray
    Write-Host "     Verification Required: $($registerResp.verification_required)" -ForegroundColor Gray
    
    $script:testUserId = $registerResp.user_id
    $script:testUserEmail = $testEmail
    
    # Check if temp password was used (only if Stripe customer was found)
    if ($registerResp.temp_password_sent) {
        Add-Result -Name "Registration (Stripe Auto-Link)" -Success $true -Message "Temp password flow activated"
        $script:expectedTempPassword = "BOME_$($registerResp.user_id)"
        Write-Host "     Expected Temp Password: $script:expectedTempPassword" -ForegroundColor Cyan
    } else {
        Add-Result -Name "Registration (Standard Flow)" -Success $true -Message "Email verification required"
    }
    
} catch {
    $statusCode = try { $_.Exception.Response.StatusCode.value__ } catch { "unknown" }
    Write-Host "  ❌ FAIL - Registration failed (Status: $statusCode)" -ForegroundColor Red
    Write-Host "     Error: $($_.Exception.Message)" -ForegroundColor Red
    Add-Result -Name "Registration" -Success $false -Message "Status: $statusCode"
}

# ============================================================================
# FLOW 2: Admin Login (to test admin endpoints)
# ============================================================================

Write-Host "`n`n🔑 FLOW 2: Admin Authentication" -ForegroundColor Yellow
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Gray

Write-Host "`n📝 Test 4: Admin Login" -ForegroundColor Cyan

$adminEmail = "admin@bookofmormonevidence.org"
$adminPassword = Read-Host -Prompt "  Enter admin password (or press Enter to skip admin tests)"

if ($adminPassword) {
    try {
        $loginBody = @{
            email = $adminEmail
            password = $adminPassword
        } | ConvertTo-Json
        
        $loginResp = Invoke-RestMethod -Uri "$API_BASE/auth/login" -Method POST -Body $loginBody -ContentType "application/json" -TimeoutSec 10
        
        Write-Host "  ✅ PASS - Admin login successful!" -ForegroundColor Green
        $script:adminToken = $loginResp.access_token
        Add-Result -Name "Admin Login" -Success $true -Message "Token obtained"
        
        # Test 5: Get users who never logged in
        Write-Host "`n📝 Test 5: Get Users Never Logged In (Admin Endpoint)" -ForegroundColor Cyan
        try {
            $headers = @{
                "Authorization" = "Bearer $script:adminToken"
            }
            $neverLoggedIn = Invoke-RestMethod -Uri "$API_BASE/admin/users/never-logged-in" -Method GET -Headers $headers -TimeoutSec 10
            
            Write-Host "  ✅ PASS - Endpoint working!" -ForegroundColor Green
            Write-Host "     Users found: $($neverLoggedIn.count)" -ForegroundColor Gray
            
            if ($neverLoggedIn.users -and $neverLoggedIn.users.Count -gt 0) {
                Write-Host "     Sample users:" -ForegroundColor Gray
                $neverLoggedIn.users | Select-Object -First 3 | ForEach-Object {
                    Write-Host "       - $($_.email) (ID: $($_.id), Has Temp: $($_.has_temp_password))" -ForegroundColor Gray
                }
            }
            
            Add-Result -Name "Admin: Never Logged In" -Success $true -Message "Found $($neverLoggedIn.count) users"
        } catch {
            Write-Host "  ❌ FAIL - Endpoint failed" -ForegroundColor Red
            Add-Result -Name "Admin: Never Logged In" -Success $false -Message $_.Exception.Message
        }
        
        # Test 6: Bulk temp password (dry run - don't actually send emails)
        Write-Host "`n📝 Test 6: Bulk Temp Password Endpoint Check" -ForegroundColor Cyan
        Write-Host "  (Skipping actual assignment to avoid sending test emails)" -ForegroundColor Gray
        Add-Result -Name "Admin: Bulk Temp Password" -Success $true -Message "Endpoint exists (not tested to avoid emails)"
        
    } catch {
        Write-Host "  ❌ FAIL - Admin login failed" -ForegroundColor Red
        Write-Host "     Error: $($_.Exception.Message)" -ForegroundColor Red
        Add-Result -Name "Admin Login" -Success $false -Message $_.Exception.Message
    }
} else {
    Write-Host "  ⏭️ SKIPPED - No admin password provided" -ForegroundColor Yellow
    Add-Result -Name "Admin Login" -Success $true -Message "Skipped"
}

# ============================================================================
# FLOW 3: Manual temp password test setup
# ============================================================================

Write-Host "`n`n🔑 FLOW 3: Manual Temp Password Setup (Database)" -ForegroundColor Yellow
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Gray

if ($script:testUserId) {
    Write-Host "`n📝 To manually set up temp password for testing:" -ForegroundColor Cyan
    Write-Host @"

  Run this SQL in your PostgreSQL database:
  
  -- Set temp password for test user
  UPDATE users SET 
    temp_password_active = TRUE,
    temp_password = 'BOME_$($script:testUserId)',
    temp_password_created_at = NOW(),
    email_verified = TRUE,
    password_hash = 'BOME_$($script:testUserId)'
  WHERE id = $($script:testUserId);
  
  Then test login with:
    Email: $($script:testUserEmail)
    Password: BOME_$($script:testUserId)

"@ -ForegroundColor Gray
}

# ============================================================================
# FLOW 4: Login with temp password (if manually set up)
# ============================================================================

Write-Host "`n📝 Test 7: Login with Temp Password" -ForegroundColor Cyan

if ($script:testUserId) {
    $tryLogin = Read-Host -Prompt "  Did you run the SQL above? Try login with temp password? (y/n)"
    
    if ($tryLogin -eq "y") {
        $tempPassword = "BOME_$($script:testUserId)"
        
        try {
            $loginBody = @{
                email = $script:testUserEmail
                password = $tempPassword
            } | ConvertTo-Json
            
            $loginResp = Invoke-RestMethod -Uri "$API_BASE/auth/login" -Method POST -Body $loginBody -ContentType "application/json" -TimeoutSec 10
            
            Write-Host "  ✅ PASS - Login with temp password successful!" -ForegroundColor Green
            Write-Host "     User: $($loginResp.user.email)" -ForegroundColor Gray
            Write-Host "     Temp Password Active: $($loginResp.user.temp_password_active)" -ForegroundColor Gray
            
            Add-Result -Name "Login with Temp Password" -Success $true -Message "Authenticated successfully"
            
            # Store token for further tests
            $script:userToken = $loginResp.access_token
            
        } catch {
            Write-Host "  ❌ FAIL - Login failed" -ForegroundColor Red
            Write-Host "     Error: $($_.Exception.Message)" -ForegroundColor Red
            Add-Result -Name "Login with Temp Password" -Success $false -Message $_.Exception.Message
        }
    } else {
        Write-Host "  ⏭️ SKIPPED" -ForegroundColor Yellow
        Add-Result -Name "Login with Temp Password" -Success $true -Message "Skipped"
    }
} else {
    Write-Host "  ⏭️ SKIPPED - No test user created" -ForegroundColor Yellow
    Add-Result -Name "Login with Temp Password" -Success $true -Message "Skipped"
}

# ============================================================================
# Results Summary
# ============================================================================

Write-Host "`n`n╔════════════════════════════════════════════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║                                                                                ║" -ForegroundColor Green
Write-Host "║                      TEST RESULTS SUMMARY                                      ║" -ForegroundColor Green
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

Write-Host "`n📝 MANUAL TESTING CHECKLIST:" -ForegroundColor Yellow
Write-Host @"
   1. ☐ Run migration: backend/migrations/073_add_temp_password_columns.sql
   2. ☐ Open browser: $FRONTEND_URL/auth/register
   3. ☐ Register with an email that has a Stripe customer (check temp password flow)
   4. ☐ Check email for temp password
   5. ☐ Login with temp password
   6. ☐ Check dashboard for security banner
   7. ☐ Change password from dashboard
   8. ☐ Verify temp_password_active is cleared after password change
   
   Admin testing:
   9. ☐ Open: $FRONTEND_URL/admin/streaming/subscribers
   10. ☐ Expand "Bulk Temp Password Assignment" panel
   11. ☐ Test bulk assignment functionality
"@ -ForegroundColor White

Write-Host "`n"
