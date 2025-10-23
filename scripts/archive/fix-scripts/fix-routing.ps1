# Fix routing/setup.go type references

Write-Host "🔄 Updating routing/setup.go type references..."

$file = "backend/routing/setup.go"
$content = Get-Content $file -Raw

# Replace type references
$content = $content -replace '\*database\.DB', '*models.DB'
$content = $content -replace 'database\.', 'models.'
$content = $content -replace '\(\*database\.', '(*models.'

Set-Content -Path $file -Value $content -NoNewline

Write-Host "✅ Type references updated in routing/setup.go"

