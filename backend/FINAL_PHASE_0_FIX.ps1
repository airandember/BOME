# FINAL PHASE 0 FIX - Complete all remaining import issues

Write-Host "`n=== FINAL PHASE 0 COMPREHENSIVE FIX ===" -ForegroundColor Cyan

# 1. Fix communication/services/email-service.go
Write-Host "`n1. Fixing communication/services/email-service.go..." -ForegroundColor Yellow
$content = Get-Content "communication/services/email-service.go" -Raw
$content = $content -replace 'authServices\.CryptoService', 'cryptoSvc.CryptoService'
$content = $content -replace 'authServices\.GetGlobalCryptoService', 'cryptoSvc.GetGlobalCryptoService'
$content = $content -replace 'authServices "bome-backend/authentication/services"', 'cryptoSvc "bome-backend/services/crypto"'
Set-Content "communication/services/email-service.go" $content -NoNewline
Write-Host "   ✅ Fixed communication email service"

# 2. Fix authentication/middleware/middleware.go
Write-Host "`n2. Fixing authentication/middleware/middleware.go..." -ForegroundColor Yellow
$content = Get-Content "authentication/middleware/middleware.go" -Raw
$content = $content -replace 'services\.ParseToken', 'crypto.ParseToken'
$content = $content -replace 'services\.ValidateTokenClaims', 'crypto.ValidateTokenClaims'
$content = $content -replace 'services\.Claims', 'crypto.Claims'
# Add crypto import if not present
if ($content -notmatch '"bome-backend/services/crypto"') {
    $content = $content -replace '("bome-backend/infrastructure/database")', '$1`n`t"bome-backend/services/crypto"'
}
Set-Content "authentication/middleware/middleware.go" $content -NoNewline
Write-Host "   ✅ Fixed authentication middleware"

# 3. Fix subscription/services/stripe.go
Write-Host "`n3. Fixing subscription/services/stripe.go..." -ForegroundColor Yellow
$content = Get-Content "subscription/services/stripe.go" -Raw
$content = $content -replace 'authServices\.GetGlobalCryptoService', 'cryptoSvc.GetGlobalCryptoService'
$content = $content -replace 'authServices "bome-backend/authentication/services"', 'cryptoSvc "bome-backend/services/crypto"'
Set-Content "subscription/services/stripe.go" $content -NoNewline
Write-Host "   ✅ Fixed subscription stripe service"

# 4. Fix services/stripe/stripe.go  
Write-Host "`n4. Fixing services/stripe/stripe.go..." -ForegroundColor Yellow
$content = Get-Content "services/stripe/stripe.go" -Raw
$content = $content -replace 'authServices\.GetGlobalCryptoService', 'cryptoSvc.GetGlobalCryptoService'
$content = $content -replace 'authServices "bome-backend/authentication/services"', 'cryptoSvc "bome-backend/services/crypto"'
Set-Content "services/stripe/stripe.go" $content -NoNewline
Write-Host "   ✅ Fixed shared stripe service"

# 5. Fix oauth2_routes.go - change crypto to authServices for OAuth2Service
Write-Host "`n5. Fixing authentication/handlers/oauth2_routes.go..." -ForegroundColor Yellow
$content = Get-Content "authentication/handlers/oauth2_routes.go" -Raw
$content = $content -replace '\*crypto\.OAuth2Service', '*authServices.OAuth2Service'
# Add authServices import
$content = $content -replace '("bome-backend/services/crypto")', '$1`n`tauthServices "bome-backend/authentication/services"'
Set-Content "authentication/handlers/oauth2_routes.go" $content -NoNewline
Write-Host "   ✅ Fixed oauth2 routes"

# 6. Fix auth.go - db methods should call authModels functions
Write-Host "`n6. Fixing authentication/handlers/auth.go..." -ForegroundColor Yellow
$content = Get-Content "authentication/handlers/auth.go" -Raw
$content = $content -replace 'db\.GetUserByEmail\(', 'authModels.GetUserByEmail(db, '
$content = $content -replace 'db\.SetVerificationToken\(', 'authModels.SetVerificationToken(db, '
$content = $content -replace 'db\.CreateUser\(', 'authModels.CreateUser(db, '
$content = $content -replace 'db\.GetUserByVerificationToken\(', 'authModels.GetUserByVerificationToken(db, '
$content = $content -replace 'db\.VerifyUserEmail\(', 'authModels.VerifyUserEmail(db, '
$content = $content -replace 'db\.SetPasswordResetToken\(', 'authModels.SetPasswordResetToken(db, '
$content = $content -replace 'db\.GetUserByResetToken\(', 'authModels.GetUserByResetToken(db, '
$content = $content -replace 'db\.UpdateUserPassword\(', 'authModels.UpdateUserPassword(db, '
$content = $content -replace 'db\.SetupPassword\(', 'authModels.SetupPassword(db, '
$content = $content -replace 'db\.CreateSession\(', 'authModels.CreateSession(db, '
$content = $content -replace 'db\.CreateAuditLog\(', 'authModels.CreateAuditLog(db, '
$content = $content -replace 'db\.GetActiveSessions\(', 'authModels.GetActiveSessions(db, '
$content = $content -replace 'db\.DeactivateSession\(', 'authModels.DeactivateSession(db, '
$content = $content -replace 'db\.DeactivateAllUserSessions\(', 'authModels.DeactivateAllUserSessions(db, '
$content = $content -replace 'db\.GetUserByID\(', 'authModels.GetUserByID(db, '
$content = $content -replace 'db\.UpdateUser\(', 'authModels.UpdateUser(db, '
Set-Content "authentication/handlers/auth.go" $content -NoNewline
Write-Host "   ✅ Fixed auth handler DB calls"

Write-Host "`n✅ ALL FIXES COMPLETE!" -ForegroundColor Green
Write-Host "Testing build..." -ForegroundColor Cyan

