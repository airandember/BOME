# Fix analytics syntax error - close unterminated comment

Write-Host "🔧 Fixing analytics syntax..."

$file = "backend/analytics/services/analytics.go"
$content = Get-Content $file -Raw

# Find and close the unterminated comment before the functions
$content = $content -replace '(\s+)// Original code below - commented out for now(\s+)// Helper functions', '$1// Original code below - commented out for now$1*/$2// Helper functions'

Set-Content -Path $file -Value $content -NoNewline
Write-Host "  ✅ Fixed analytics syntax"

