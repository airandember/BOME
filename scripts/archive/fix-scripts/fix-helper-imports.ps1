# Fix imports in helper files

Write-Host "🔄 Fixing helper file imports..."

$files = @(
    "backend/authentication/services/crypto.go",
    "backend/authentication/services/email_helpers.go",
    "backend/communication/services/email_helpers.go",
    "backend/subscription/services/stripe_logger.go"
)

foreach ($file in $files) {
    if (Test-Path $file) {
        $content = Get-Content $file -Raw
        
        # Fix imports
        $content = $content -replace '"bome-backend/internal/config"', '"bome-backend/infrastructure/config"'
        $content = $content -replace '"bome-backend/internal/database"', '"bome-backend/infrastructure/database"'
        
        Set-Content -Path $file -Value $content -NoNewline
        Write-Host "  ✅ Fixed: $file"
    }
}

Write-Host "✅ Helper imports fixed!"

