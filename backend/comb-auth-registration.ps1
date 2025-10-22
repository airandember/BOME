# Comb User Registration Strand

Write-Host "`n🔍 COMBING: User Registration Strand`n" -ForegroundColor Cyan

Write-Host "Checking handler functions..." -ForegroundColor Yellow
if (Select-String -Path "authentication/handlers/auth.go" -Pattern "func.*RegisterHandler" -Quiet) {
    Write-Host "  [OK] RegisterHandler found" -ForegroundColor Green
} else {
    Write-Host "  [SPLIT-END] RegisterHandler missing" -ForegroundColor Red
}

Write-Host "`nChecking model functions..." -ForegroundColor Yellow
if (Select-String -Path "authentication/models/user.go" -Pattern "func.*CreateUser" -Quiet) {
    Write-Host "  [OK] CreateUser found" -ForegroundColor Green
} else {
    Write-Host "  [SPLIT-END] CreateUser missing" -ForegroundColor Red
}

Write-Host "`nChecking service functions..." -ForegroundColor Yellow
if (Select-String -Path "authentication/services/password.go" -Pattern "func.*HashPassword" -Quiet) {
    Write-Host "  [OK] HashPassword found" -ForegroundColor Green
} else {
    Write-Host "  [SPLIT-END] HashPassword missing" -ForegroundColor Red
}

if (Select-String -Path "authentication/services/password.go" -Pattern "func.*ValidatePassword" -Quiet) {
    Write-Host "  [OK] ValidatePassword found" -ForegroundColor Green
} else {
    Write-Host "  [SPLIT-END] ValidatePassword missing (CRITICAL)" -ForegroundColor Red
}

if (Select-String -Path "authentication/services/email.go" -Pattern "func.*SendVerificationEmail" -Quiet) {
    Write-Host "  [OK] SendVerificationEmail found" -ForegroundColor Green
} else {
    Write-Host "  [SPLIT-END] SendVerificationEmail missing" -ForegroundColor Red
}

Write-Host "`nChecking database schema..." -ForegroundColor Yellow
if (Test-Path "../backend_original/migrations/*verification*.sql") {
    Write-Host "  [OK] Email verification table migrations exist" -ForegroundColor Green
} else {
    Write-Host "  [SPLIT-END] Email verification migrations missing" -ForegroundColor Red
}

Write-Host "`nChecking frontend..." -ForegroundColor Yellow
if (Test-Path "../frontend/src/routes/register/+page.svelte") {
    Write-Host "  [OK] Registration page exists" -ForegroundColor Green
} else {
    Write-Host "  [SPLIT-END] Registration page missing" -ForegroundColor Red
}

Write-Host "`nChecking RegisterHandler implementation..." -ForegroundColor Yellow
$regHandlerContent = Get-Content "authentication/handlers/auth.go" -Raw
if ($regHandlerContent -match "ValidatePassword") {
    Write-Host "  [DEPENDENCY] RegisterHandler calls ValidatePassword" -ForegroundColor Yellow
    Write-Host "  [NOTE] This split-end blocks registration flow!" -ForegroundColor Red
}

Write-Host "`nUser Registration strand combing complete!`n" -ForegroundColor Green

