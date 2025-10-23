# Fix middleware crypto calls

$file = "authentication/middleware/middleware.go"
$content = Get-Content $file -Raw

# Replace crypto package types with ports types (since CryptoService returns ports.Claims)
$content = $content -replace 'crypto\.Claims', 'ports.Claims'

# Replace crypto.ParseToken calls
$content = $content -replace 'crypto\.ParseToken\(', 'crypto.GetGlobalCryptoService().ParseToken('

# Replace crypto.ValidateTokenClaims calls  
$content = $content -replace 'crypto\.ValidateTokenClaims\(', 'crypto.GetGlobalCryptoService().ValidateTokenClaims('

# Add ports import if not already there
if ($content -notmatch 'bome-backend/ports') {
    $content = $content -replace '(import \()', "`$1`n`t`"bome-backend/ports`""
}

Set-Content -Path $file -Value $content -NoNewline

Write-Host "✅ Fixed middleware crypto calls" -ForegroundColor Green

