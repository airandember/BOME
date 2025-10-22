# Fix CryptoService references

Write-Host "🔄 Fixing CryptoService references..."

$files = @(
    "backend/communication/services/email-service.go",
    "backend/subscription/services/stripe.go"
)

foreach ($file in $files) {
    $content = Get-Content $file -Raw
    
    # Fix CryptoService and GetGlobalCryptoService references
    $content = $content -replace 'GetGlobalCryptoService\(\)', 'authServices.GetGlobalCryptoService()'
    $content = $content -replace '\*CryptoService', '*authServices.CryptoService'
    
    Set-Content -Path $file -Value $content -NoNewline
    Write-Host "  ✅ Fixed: $file"
}

Write-Host "✅ CryptoService references fixed!"

