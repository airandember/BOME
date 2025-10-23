# Fix email services to use auth models

Write-Host "🔄 Fixing email service calls..."

$files = @(
    "backend/authentication/services/email.go",
    "backend/communication/services/email-service.go"
)

foreach ($file in $files) {
    $content = Get-Content $file -Raw
    
    # Fix SetVerificationToken calls
    $content = $content -replace 's\.db\.SetVerificationToken\(', 'authModels.SetVerificationToken(s.db, '
    
    Set-Content -Path $file -Value $content -NoNewline
    Write-Host "  ✅ Fixed: $file"
}

Write-Host "✅ Email services fixed!"

