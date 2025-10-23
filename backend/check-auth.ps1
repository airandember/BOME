# Simple Authentication Braid Check

Write-Host ""
Write-Host "AUTHENTICATION BRAID - FILE CHECK" -ForegroundColor Cyan
Write-Host ""

Write-Host "Layer 5: Persistence (Database)" -ForegroundColor Yellow
Test-Path "../backend_original/migrations/*user*.sql" | ForEach-Object { if ($_) { Write-Host "  [OK] User migrations" -ForegroundColor Green } else { Write-Host "  [MISSING] User migrations" -ForegroundColor Red } }
Test-Path "../backend_original/migrations/*session*.sql" | ForEach-Object { if ($_) { Write-Host "  [OK] Session migrations" -ForegroundColor Green } else { Write-Host "  [MISSING] Session migrations" -ForegroundColor Yellow } }

Write-Host ""
Write-Host "Layer 4: Data Access (Models)" -ForegroundColor Yellow
Test-Path "authentication/models/user.go" | ForEach-Object { if ($_) { Write-Host "  [OK] user.go" -ForegroundColor Green } else { Write-Host "  [MISSING] user.go" -ForegroundColor Red } }
Test-Path "authentication/models/session.go" | ForEach-Object { if ($_) { Write-Host "  [OK] session.go" -ForegroundColor Green } else { Write-Host "  [MISSING] session.go" -ForegroundColor Yellow } }

Write-Host ""
Write-Host "Layer 3: Business Logic" -ForegroundColor Yellow
Test-Path "authentication/handlers/auth.go" | ForEach-Object { if ($_) { Write-Host "  [OK] auth.go handler" -ForegroundColor Green } else { Write-Host "  [MISSING] auth.go handler" -ForegroundColor Red } }
Test-Path "authentication/services/jwt.go" | ForEach-Object { if ($_) { Write-Host "  [OK] jwt.go service" -ForegroundColor Green } else { Write-Host "  [MISSING] jwt.go service" -ForegroundColor Red } }
Test-Path "authentication/services/password.go" | ForEach-Object { if ($_) { Write-Host "  [OK] password.go service" -ForegroundColor Green } else { Write-Host "  [MISSING] password.go service" -ForegroundColor Red } }
Test-Path "authentication/services/email.go" | ForEach-Object { if ($_) { Write-Host "  [OK] email.go service" -ForegroundColor Green } else { Write-Host "  [MISSING] email.go service" -ForegroundColor Yellow } }
Test-Path "authentication/middleware/middleware.go" | ForEach-Object { if ($_) { Write-Host "  [OK] middleware.go" -ForegroundColor Green } else { Write-Host "  [MISSING] middleware.go" -ForegroundColor Red } }

Write-Host ""
Write-Host "Layer 1: Presentation (Frontend)" -ForegroundColor Yellow
Test-Path "../frontend/src/routes/login/+page.svelte" | ForEach-Object { if ($_) { Write-Host "  [OK] Login page" -ForegroundColor Green } else { Write-Host "  [MISSING] Login page" -ForegroundColor Red } }
Test-Path "../frontend/src/lib/auth.ts" | ForEach-Object { if ($_) { Write-Host "  [OK] Auth store" -ForegroundColor Green } else { Write-Host "  [MISSING] Auth store" -ForegroundColor Red } }

Write-Host ""
Write-Host "Checking key functions..." -ForegroundColor Cyan
Write-Host ""

$foundGetUser = Select-String -Path "authentication/models/user.go" -Pattern "GetUserByEmail" -Quiet
$foundGenToken = Select-String -Path "authentication/services/jwt.go" -Pattern "GenerateToken" -Quiet
$foundValidatePwd = Select-String -Path "authentication/services/password.go" -Pattern "ValidatePassword" -Quiet
$foundLoginHandler = Select-String -Path "authentication/handlers/auth.go" -Pattern "LoginHandler" -Quiet

if ($foundGetUser) { Write-Host "  [OK] GetUserByEmail found" -ForegroundColor Green } else { Write-Host "  [SPLIT-END] GetUserByEmail missing" -ForegroundColor Red }
if ($foundGenToken) { Write-Host "  [OK] GenerateToken found" -ForegroundColor Green } else { Write-Host "  [SPLIT-END] GenerateToken missing" -ForegroundColor Red }
if ($foundValidatePwd) { Write-Host "  [OK] ValidatePassword found" -ForegroundColor Green } else { Write-Host "  [SPLIT-END] ValidatePassword missing" -ForegroundColor Red }
if ($foundLoginHandler) { Write-Host "  [OK] LoginHandler found" -ForegroundColor Green } else { Write-Host "  [SPLIT-END] LoginHandler missing" -ForegroundColor Red }

Write-Host ""
Write-Host "Check complete!" -ForegroundColor Green
Write-Host ""

