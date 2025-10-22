# Fix function calls in user.go to use new function syntax

Write-Host "🔄 Fixing function calls in user.go..."

$file = "backend/authentication/models/user.go"
$content = Get-Content $file -Raw

# Fix function calls from method style to function style
# db.GetUserByID(x) -> GetUserByID(db, x)
$content = $content -replace 'db\.GetUserByID\(', 'GetUserByID(db, '
$content = $content -replace 'db\.GetSessionCount\(', 'GetSessionCount(db, '

Set-Content -Path $file -Value $content -NoNewline

Write-Host "✅ Fixed function calls in user.go"

