# Fix final Phase 0 issues

Write-Host "`n🔧 Fixing final issues...`n" -ForegroundColor Cyan

# 1. Move stripe_logger.go
Write-Host "1. Moving stripe_logger.go..." -ForegroundColor Yellow
Copy-Item "subscription/services/stripe_logger.go" "services/stripe/" -Force
$content = Get-Content "services/stripe/stripe_logger.go" -Raw
$content = $content -replace "package services", "package stripe"
Set-Content "services/stripe/stripe_logger.go" $content -NoNewline
Write-Host "  ✅ Moved and fixed stripe_logger.go`n" -ForegroundColor Green

# 2. Fix email.go crypto import
Write-Host "2. Fixing email.go imports..." -ForegroundColor Yellow
$emailContent = Get-Content "services/email/email.go" -Raw

# Remove the malformed import line
$emailContent = $emailContent -replace '\s+"bome-backend/services/crypto"\s+imported\s+as\s+stripe\s+and\s+not\s+used', ''

# Ensure proper imports
$emailContent = $emailContent -replace 'import \(\s+authModels "bome-backend/authentication/models"\s+"bome-backend/infrastructure/database"', 
'import (
	authModels "bome-backend/authentication/models"
	"bome-backend/infrastructure/database"
	"bome-backend/services/crypto"'

Set-Content "services/email/email.go" $emailContent -NoNewline
Write-Host "  ✅ Fixed email.go imports`n" -ForegroundColor Green

Write-Host "✅ All fixes applied!`n" -ForegroundColor Green

