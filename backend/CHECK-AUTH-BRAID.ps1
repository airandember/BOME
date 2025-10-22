# Authentication Braid File Existence Check

Write-Host "`n🔍 AUTHENTICATION BRAID - FILE CHECK`n" -ForegroundColor Cyan

# Layer 5: Persistence
Write-Host "Layer 5: Persistence (Database)" -ForegroundColor Yellow
if (Test-Path "../backend_original/migrations/*user*.sql") {
    Write-Host "  ✅ User migrations exist" -ForegroundColor Green
} else {
    Write-Host "  ❌ User migrations missing" -ForegroundColor Red
}

if (Test-Path "../backend_original/migrations/*session*.sql") {
    Write-Host "  ✅ Session migrations exist" -ForegroundColor Green
} else {
    Write-Host "  ⚠️  Session migrations missing" -ForegroundColor Yellow
}

# Layer 4: Data Access
Write-Host "`nLayer 4: Data Access (Models)" -ForegroundColor Yellow
if (Test-Path "authentication/models/user.go") {
    Write-Host "  ✅ user.go exists" -ForegroundColor Green
} else {
    Write-Host "  ❌ user.go missing" -ForegroundColor Red
}

if (Test-Path "authentication/models/session.go") {
    Write-Host "  ✅ session.go exists" -ForegroundColor Green
} else {
    Write-Host "  ⚠️  session.go missing" -ForegroundColor Yellow
}

# Layer 3: Business Logic
Write-Host "`nLayer 3: Business Logic (Services and Handlers)" -ForegroundColor Yellow
if (Test-Path "authentication/handlers/auth.go") {
    Write-Host "  ✅ auth.go handler exists" -ForegroundColor Green
} else {
    Write-Host "  ❌ auth.go handler missing" -ForegroundColor Red
}

if (Test-Path "authentication/services/jwt.go") {
    Write-Host "  ✅ jwt.go service exists" -ForegroundColor Green
} else {
    Write-Host "  ❌ jwt.go service missing" -ForegroundColor Red
}

if (Test-Path "authentication/services/password.go") {
    Write-Host "  ✅ password.go service exists" -ForegroundColor Green
} else {
    Write-Host "  ❌ password.go service missing" -ForegroundColor Red
}

if (Test-Path "authentication/services/email.go") {
    Write-Host "  ✅ email.go service exists" -ForegroundColor Green
} else {
    Write-Host "  ⚠️  email.go service missing" -ForegroundColor Yellow
}

if (Test-Path "authentication/middleware/middleware.go") {
    Write-Host "  ✅ middleware.go exists" -ForegroundColor Green
} else {
    Write-Host "  ❌ middleware.go missing" -ForegroundColor Red
}

# Layer 1: Presentation
Write-Host "`nLayer 1: Presentation (Frontend)" -ForegroundColor Yellow
if (Test-Path "../frontend/src/routes/login/+page.svelte") {
    Write-Host "  ✅ Login page exists" -ForegroundColor Green
} else {
    Write-Host "  ❌ Login page missing" -ForegroundColor Red
}

if (Test-Path "../frontend/src/routes/register/+page.svelte") {
    Write-Host "  ✅ Register page exists" -ForegroundColor Green
} else {
    Write-Host "  ⚠️  Register page missing" -ForegroundColor Yellow
}

if (Test-Path "../frontend/src/lib/auth.ts") {
    Write-Host "  ✅ Auth store exists" -ForegroundColor Green
} else {
    Write-Host "  ❌ Auth store missing" -ForegroundColor Red
}

Write-Host "`n" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "Now checking for key functions..." -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

# Check for key functions
Write-Host "Checking user.go for GetUserByEmail..." -ForegroundColor Yellow
if (Select-String -Path "authentication/models/user.go" -Pattern "GetUserByEmail" -Quiet) {
    Write-Host "  ✅ GetUserByEmail found" -ForegroundColor Green
} else {
    Write-Host "  ❌ GetUserByEmail missing (SPLIT-END!)" -ForegroundColor Red
}

Write-Host "`nChecking jwt.go for GenerateToken..." -ForegroundColor Yellow
if (Select-String -Path "authentication/services/jwt.go" -Pattern "GenerateToken" -Quiet) {
    Write-Host "  ✅ GenerateToken found" -ForegroundColor Green
} else {
    Write-Host "  ❌ GenerateToken missing (SPLIT-END!)" -ForegroundColor Red
}

Write-Host "`nChecking password.go for ValidatePassword..." -ForegroundColor Yellow
if (Select-String -Path "authentication/services/password.go" -Pattern "ValidatePassword" -Quiet) {
    Write-Host "  ✅ ValidatePassword found" -ForegroundColor Green
} else {
    Write-Host "  ❌ ValidatePassword missing (SPLIT-END!)" -ForegroundColor Red
}

Write-Host "`nChecking auth.go for LoginHandler..." -ForegroundColor Yellow
if (Select-String -Path "authentication/handlers/auth.go" -Pattern "LoginHandler" -Quiet) {
    Write-Host "  ✅ LoginHandler found" -ForegroundColor Green
} else {
    Write-Host "  ❌ LoginHandler missing (SPLIT-END!)" -ForegroundColor Red
}

Write-Host "`nFile check complete! See results above.`n" -ForegroundColor Green

