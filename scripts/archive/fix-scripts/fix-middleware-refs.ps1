# Fix middleware type references

Write-Host "🔄 Fixing middleware references..."

$file = "backend/authentication/middleware/middleware.go"
$content = Get-Content $file -Raw

# Fix type references from models.DB to database.DB
$content = $content -replace '\*models\.DB', '*database.DB'

Set-Content -Path $file -Value $content -NoNewline

Write-Host "✅ Middleware references fixed!"

