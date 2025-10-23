# Comb User Registration Strand
Write-Host ""
Write-Host "COMBING: User Registration Strand" -ForegroundColor Cyan
Write-Host ""

Write-Host "Checking handler functions..." -ForegroundColor Yellow
Select-String -Path "authentication/handlers/auth.go" -Pattern "func.*RegisterHandler" -Quiet | ForEach-Object { 
    if ($_) { Write-Host "  [OK] RegisterHandler found" -ForegroundColor Green } 
    else { Write-Host "  [SPLIT-END] RegisterHandler missing" -ForegroundColor Red } 
}

Write-Host ""
Write-Host "Checking model functions..." -ForegroundColor Yellow
Select-String -Path "authentication/models/user.go" -Pattern "func.*CreateUser" -Quiet | ForEach-Object { 
    if ($_) { Write-Host "  [OK] CreateUser found" -ForegroundColor Green } 
    else { Write-Host "  [SPLIT-END] CreateUser missing" -ForegroundColor Red } 
}

Write-Host ""
Write-Host "Checking service functions..." -ForegroundColor Yellow
Select-String -Path "authentication/services/password.go" -Pattern "func.*HashPassword" -Quiet | ForEach-Object { 
    if ($_) { Write-Host "  [OK] HashPassword found" -ForegroundColor Green } 
    else { Write-Host "  [SPLIT-END] HashPassword missing" -ForegroundColor Red } 
}

$hasValidate = Select-String -Path "authentication/services/password.go" -Pattern "func.*ValidatePassword" -Quiet
if ($hasValidate) { 
    Write-Host "  [OK] ValidatePassword found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] ValidatePassword missing (CRITICAL)" -ForegroundColor Red 
}

$hasEmailSend = Select-String -Path "authentication/services/email.go" -Pattern "func.*SendVerificationEmail" -Quiet
if ($hasEmailSend) { 
    Write-Host "  [OK] SendVerificationEmail found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] SendVerificationEmail missing" -ForegroundColor Red 
}

Write-Host ""
Write-Host "Checking database schema..." -ForegroundColor Yellow
if (Test-Path "../backend_original/migrations/*verification*.sql") {
    Write-Host "  [OK] Email verification migrations exist" -ForegroundColor Green
} else {
    Write-Host "  [SPLIT-END] Email verification migrations missing" -ForegroundColor Red
}

Write-Host ""
Write-Host "Checking frontend..." -ForegroundColor Yellow
if (Test-Path "../frontend/src/routes/register/+page.svelte") {
    Write-Host "  [OK] Registration page exists" -ForegroundColor Green
} else {
    Write-Host "  [SPLIT-END] Registration page missing" -ForegroundColor Red
}

Write-Host ""
Write-Host "Registration strand combing complete!" -ForegroundColor Green
Write-Host ""

