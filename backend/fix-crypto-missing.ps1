# Fix missing crypto calls in handlers

$files = @(
    "authentication/handlers/auth.go",
    "authentication/handlers/oauth2_routes.go"
)

foreach ($file in $files) {
    $content = Get-Content $file -Raw
    
    # Add the missing function conversions
    $content = $content -replace 'crypto\.GenerateTokenPair\(', 'crypto.GetGlobalCryptoService().GenerateTokenPair('
    $content = $content -replace 'crypto\.RefreshTokenPair\(', 'crypto.GetGlobalCryptoService().RefreshTokenPair('
    $content = $content -replace 'crypto\.GenerateDeviceFingerprint\(', 'crypto.GetGlobalCryptoService().GenerateDeviceFingerprint('
    
    Set-Content -Path $file -Value $content -NoNewline
    Write-Host "✅ Updated $file" -ForegroundColor Green
}

Write-Host "`n✅ All crypto calls fixed!" -ForegroundColor Green

