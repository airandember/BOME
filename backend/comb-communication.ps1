# Comb All Communication Strands - FAST CHECK!

Write-Host ""
Write-Host "COMMUNICATION BRAID - COMPLETE COMBING" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

# Check if communication directory exists
if (-not (Test-Path "communication")) {
    Write-Host "[INFO] No dedicated communication/ directory" -ForegroundColor Yellow
    Write-Host "[INFO] Communication likely in authentication/services" -ForegroundColor Gray
    Write-Host ""
}

# STRAND 1: Email Template Management
Write-Host "STRAND 1: Email Template Management" -ForegroundColor Yellow
Write-Host "-----------------------------------" -ForegroundColor Yellow

# Check authentication services (where email.go is)
if (Test-Path "authentication/services/email.go") {
    Write-Host "  [OK] email.go service found in authentication" -ForegroundColor Green
    
    $hasTemplates = Select-String -Path "authentication/services/email.go" -Pattern "template|Template" -Quiet
    if ($hasTemplates) {
        Write-Host "  [OK] Email templates found" -ForegroundColor Green
    } else {
        Write-Host "  [SPLIT-END] Templates missing" -ForegroundColor Yellow
    }
} else {
    Write-Host "  [SPLIT-END] email.go service missing" -ForegroundColor Red
}

Write-Host ""

# STRAND 2: Notification System
Write-Host "STRAND 2: Notification System" -ForegroundColor Yellow
Write-Host "-----------------------------" -ForegroundColor Yellow

$hasNotifications = Select-String -Path "**/*.go" -Pattern "func.*Notify|func.*Notification" -Quiet 2>$null
if ($hasNotifications) { 
    Write-Host "  [OK] Notification functions found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Notifications missing" -ForegroundColor Yellow 
}

Write-Host ""

# STRAND 3: Email Delivery
Write-Host "STRAND 3: Email Delivery and Tracking" -ForegroundColor Yellow
Write-Host "--------------------------------------" -ForegroundColor Yellow

if (Test-Path "authentication/services/email.go") {
    $hasSend = Select-String -Path "authentication/services/email.go" -Pattern "func.*Send|SendEmail" -Quiet
    if ($hasSend) {
        Write-Host "  [OK] Email sending functions found" -ForegroundColor Green
        
        # Count send functions
        $sendFuncs = Select-String -Path "authentication/services/email.go" -Pattern "func.*Send" | Measure-Object
        Write-Host "  [INFO] Found $($sendFuncs.Count) email sending functions" -ForegroundColor Cyan
    } else {
        Write-Host "  [SPLIT-END] Email sending missing" -ForegroundColor Red
    }
}

Write-Host ""

# STRAND 4: User Preferences
Write-Host "STRAND 4: User Communication Preferences" -ForegroundColor Yellow
Write-Host "----------------------------------------" -ForegroundColor Yellow

$hasPreferences = Select-String -Path "**/*.go" -Pattern "email.*preference|notification.*preference" -Quiet 2>$null
if ($hasPreferences) { 
    Write-Host "  [OK] Preferences found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Preferences missing (optional)" -ForegroundColor Yellow 
}

Write-Host ""

# STRAND 5: Communication Analytics
Write-Host "STRAND 5: Communication Analytics" -ForegroundColor Yellow
Write-Host "---------------------------------" -ForegroundColor Yellow

$hasAnalytics = Select-String -Path "**/*.go" -Pattern "email.*analytics|email.*tracking" -Quiet 2>$null
if ($hasAnalytics) { 
    Write-Host "  [OK] Email analytics found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Analytics missing (optional)" -ForegroundColor Yellow 
}

Write-Host ""

# Check communication services
Write-Host "COMMUNICATION SERVICES CHECK:" -ForegroundColor Cyan

if (Test-Path "communication/services") {
    Write-Host "  [OK] communication/services/ directory exists" -ForegroundColor Green
    $serviceCount = (Get-ChildItem "communication/services/*.go" -File 2>$null | Measure-Object).Count
    Write-Host "  [INFO] Found $serviceCount service files" -ForegroundColor Cyan
} else {
    Write-Host "  [INFO] No dedicated communication/services directory" -ForegroundColor Gray
    Write-Host "  [INFO] Email services likely in authentication package" -ForegroundColor Gray
}

Write-Host ""
Write-Host "ALL 5 STRANDS COMBED" -ForegroundColor Green
Write-Host ""

