# Comb Remaining Authentication Strands

Write-Host ""
Write-Host "=================================" -ForegroundColor Cyan
Write-Host "COMBING REMAINING AUTH STRANDS" -ForegroundColor Cyan
Write-Host "=================================" -ForegroundColor Cyan
Write-Host ""

# STRAND 3: Email Verification
Write-Host "STRAND 3: Email Verification" -ForegroundColor Yellow
Write-Host "-----------------------------" -ForegroundColor Yellow

$hasVerifyHandler = Select-String -Path "authentication/handlers/auth.go" -Pattern "func.*VerifyEmail" -Quiet
if ($hasVerifyHandler) { 
    Write-Host "  [OK] VerifyEmailHandler found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] VerifyEmailHandler missing" -ForegroundColor Red 
}

if (Test-Path "../frontend/src/routes/verify-email/+page.svelte") {
    Write-Host "  [OK] Email verification page exists" -ForegroundColor Green
} else {
    Write-Host "  [SPLIT-END] Email verification page missing" -ForegroundColor Red
}

Write-Host ""

# STRAND 4: Session Management
Write-Host "STRAND 4: Session Management" -ForegroundColor Yellow
Write-Host "-----------------------------" -ForegroundColor Yellow

if (Test-Path "authentication/models/session.go") {
    Write-Host "  [OK] session.go model exists" -ForegroundColor Green
} else {
    Write-Host "  [SPLIT-END] session.go model missing" -ForegroundColor Red
}

$hasLogoutHandler = Select-String -Path "authentication/handlers/auth.go" -Pattern "func.*Logout" -Quiet
if ($hasLogoutHandler) { 
    Write-Host "  [OK] LogoutHandler found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] LogoutHandler missing" -ForegroundColor Red 
}

$hasCreateSession = Select-String -Path "authentication/models/*.go" -Pattern "func.*CreateSession" -Quiet
if ($hasCreateSession) { 
    Write-Host "  [OK] CreateSession found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] CreateSession function missing" -ForegroundColor Red 
}

$hasInvalidateSession = Select-String -Path "authentication/models/*.go" -Pattern "func.*InvalidateSession" -Quiet
if ($hasInvalidateSession) { 
    Write-Host "  [OK] InvalidateSession found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] InvalidateSession function missing" -ForegroundColor Red 
}

Write-Host ""

# STRAND 5: OAuth2 Integration
Write-Host "STRAND 5: OAuth2 Integration (Enhancement)" -ForegroundColor Yellow
Write-Host "-----------------------------" -ForegroundColor Yellow

if (Test-Path "authentication/handlers/oauth2_routes.go") {
    Write-Host "  [OK] oauth2_routes.go exists" -ForegroundColor Green
} else {
    Write-Host "  [SPLIT-END] oauth2_routes.go missing" -ForegroundColor Yellow
}

if (Test-Path "authentication/services/oauth2.go") {
    Write-Host "  [OK] oauth2.go service exists" -ForegroundColor Green
} else {
    Write-Host "  [SPLIT-END] oauth2.go service missing" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "=================================" -ForegroundColor Green
Write-Host "ALL STRANDS COMBED!" -ForegroundColor Green
Write-Host "=================================" -ForegroundColor Green
Write-Host ""

