# Fix middleware cross-braid calls

Write-Host "🔄 Fixing middleware cross-braid calls..."

$file = "backend/authentication/middleware/middleware.go"
$content = Get-Content $file -Raw

# Fix cross-braid function calls
$content = $content -replace 'db\.HasVideoAccess\(', 'videoModels.HasVideoAccess(db, '
$content = $content -replace 'db\.GetSubscriptionByUserID\(', 'subModels.GetSubscriptionByUserID(db, '
$content = $content -replace 'db\.GetSubscriptionPlanByID\(', 'subModels.GetSubscriptionPlanByID(db, '
$content = $content -replace 'db\.UpdateSessionActivityByTokenID\(', 'authModels.UpdateSessionActivityByTokenID(db, '

# Remove unused authModels import marker
$content = $content -replace 'authModels "bome-backend/authentication/models"', '"bome-backend/authentication/models"'

Set-Content -Path $file -Value $content -NoNewline

Write-Host "✅ Middleware cross-braid calls fixed!"

