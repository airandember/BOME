# Phase 0: Fix authentication handlers for shared services

Write-Host "`n🔧 Fixing authentication/handlers/auth.go...`n" -ForegroundColor Cyan

$authFile = "authentication/handlers/auth.go"
$content = Get-Content $authFile -Raw

# Fix all function signatures with models.DB
$content = $content -replace 'func (\w+Handler)\(db \*models\.DB', 'func $1(db *database.DB'

# Fix models.User references to use authModels
$content = $content -replace '(\s+)(models\.User\b)', '$1authModels.User'
$content = $content -replace '(\s+)(\*models\.User\b)', '$1*authModels.User'
$content = $content -replace '(\s+)(models\.Session\b)', '$1authModels.Session'

# Fix services.something references to use appropriate shared service
$content = $content -replace 'services\.GetClientIP', 'crypto.GetClientIP'
$content = $content -replace 'services\.RegisterRateLimiter', 'crypto.RegisterRateLimiter'
$content = $content -replace 'services\.LoginRateLimiter', 'crypto.LoginRateLimiter'
$content = $content -replace 'services\.GenerateJWT', 'crypto.GenerateJWT'
$content = $content -replace 'services\.ValidatePassword', 'crypto.ValidatePassword'
$content = $content -replace 'services\.HashPassword', 'crypto.HashPassword'
$content = $content -replace 'services\.CheckPassword', 'crypto.CheckPassword'

Set-Content $authFile $content -NoNewline

Write-Host "✅ Fixed auth.go" -ForegroundColor Green
Write-Host "  - Updated all function signatures" -ForegroundColor Gray
Write-Host "  - Fixed model references" -ForegroundColor Gray
Write-Host "  - Fixed service references`n" -ForegroundColor Gray

# Now fix subscription handlers
Write-Host "🔧 Fixing subscription/handlers/subscription.go...`n" -ForegroundColor Cyan

$subFile = "subscription/handlers/subscription.go"
if (Test-Path $subFile) {
    $content = Get-Content $subFile -Raw
    
    # Fix service references
    $content = $content -replace 'substripe\.', 'stripe.'
    $content = $content -replace 'subanalytics\.', 'analytics.'
    
    Set-Content $subFile $content -NoNewline
    Write-Host "✅ Fixed subscription.go`n" -ForegroundColor Green
}

Write-Host "✅ Phase 0 handler fixes complete!`n" -ForegroundColor Green

