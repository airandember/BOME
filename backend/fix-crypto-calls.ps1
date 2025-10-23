# Fix Crypto Service Calls in Handlers
# This script updates all crypto package function calls to use CryptoService methods

Write-Host "`n🔧 Fixing crypto service calls..." -ForegroundColor Yellow

$files = @(
    "authentication/handlers/auth.go",
    "authentication/handlers/oauth2_routes.go"
)

foreach ($file in $files) {
    Write-Host "  Processing $file..." -ForegroundColor Cyan
    
    $content = Get-Content $file -Raw
    
    # Add crypto service retrieval at the start of each handler
    # This is a bit complex, so we'll do it manually for now
    
    # Replace direct package-level function calls with service method calls
    # Note: We keep using crypto.GetGlobalCryptoService() which returns the singleton
    
    # The key insight: Most handlers already use crypto.GetGlobalCryptoService()
    # We just need to ensure all calls go through a service instance
    
    # Replace crypto.PackageFunction() with cryptoService.Method()
    $content = $content -replace 'crypto\.GetClientIP\(', 'crypto.GetGlobalCryptoService().GetClientIP('
    $content = $content -replace 'crypto\.SanitizeString\(', 'crypto.GetGlobalCryptoService().SanitizeString('
    $content = $content -replace 'crypto\.ValidateEmail\(', 'crypto.GetGlobalCryptoService().ValidateEmail('
    $content = $content -replace 'crypto\.ValidateName\(', 'crypto.GetGlobalCryptoService().ValidateName('
    $content = $content -replace 'crypto\.GenerateSecureToken\(', 'crypto.GetGlobalCryptoService().GenerateSecureToken('
    $content = $content -replace 'crypto\.GenerateJWT\(', 'crypto.GetGlobalCryptoService().GenerateJWT('
    $content = $content -replace 'crypto\.HashPassword\(', 'crypto.GetGlobalCryptoService().HashPassword('
    $content = $content -replace 'crypto\.CheckPassword\(', 'crypto.GetGlobalCryptoService().CheckPassword('
    $content = $content -replace 'crypto\.ValidatePassword\(', 'crypto.GetGlobalCryptoService().ValidatePassword('
    $content = $content -replace 'crypto\.ParseToken\(', 'crypto.GetGlobalCryptoService().ParseToken('
    $content = $content -replace 'crypto\.ParseRefreshToken\(', 'crypto.GetGlobalCryptoService().ParseRefreshToken('
    $content = $content -replace 'crypto\.BlacklistToken\(', 'crypto.GetGlobalCryptoService().BlacklistToken('
    $content = $content -replace 'crypto\.EncryptString\(', 'crypto.GetGlobalCryptoService().EncryptString('
    $content = $content -replace 'crypto\.DecryptString\(', 'crypto.GetGlobalCryptoService().DecryptString('
    
    Set-Content -Path $file -Value $content -NoNewline
    Write-Host "    ✅ Updated $file" -ForegroundColor Green
}

Write-Host "`n✅ Crypto service calls fixed!`n" -ForegroundColor Green

