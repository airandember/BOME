# Fix middleware references

Write-Host "🔄 Fixing middleware..."

$file = "backend/authentication/middleware/middleware.go"
$content = Get-Content $file -Raw

# Fix type references
$content = $content -replace 'models\.DB', 'database.DB'

# Fix service references
$content = $content -replace 'services\.NewRateLimiter', 'authServices.NewRateLimiter'
$content = $content -replace 'services\.GetClientIP', 'authServices.GetClientIP'

Set-Content -Path $file -Value $content -NoNewline

Write-Host "✅ Middleware fixed!"

