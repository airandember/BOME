# Fix oauth2 service calls

Write-Host "🔄 Fixing oauth2 service..."

$file = "backend/authentication/services/oauth2.go"
$content = Get-Content $file -Raw

# Fix function calls to use authModels
$content = $content -replace 's\.db\.GetUserByEmail\(', 'authModels.GetUserByEmail(s.db, '
$content = $content -replace 's\.db\.SetUserEmailVerified\(', 'authModels.SetUserEmailVerified(s.db, '
$content = $content -replace 's\.db\.CreateUser\(', 'authModels.CreateUser(s.db, '

Set-Content -Path $file -Value $content -NoNewline

Write-Host "✅ OAuth2 service fixed!"

