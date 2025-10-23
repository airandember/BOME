# Fix remaining stripe service references

Write-Host "`n🔧 Fixing remaining stripe references...`n" -ForegroundColor Cyan

# Fix stripe_webhook_routes.go
$webhookFile = "subscription/handlers/stripe_webhook_routes.go"
$content = Get-Content $webhookFile -Raw

# Replace substripe with stripe
$content = $content -replace 'substripe\.', 'stripe.'

Set-Content $webhookFile $content -NoNewline
Write-Host "✅ Fixed stripe_webhook_routes.go" -ForegroundColor Green

Write-Host "`n✅ All stripe references fixed!`n" -ForegroundColor Green

