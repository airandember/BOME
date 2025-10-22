# Fix main.go service references

Write-Host "🔄 Updating main.go service references..."

$file = "backend/main.go"
$content = Get-Content $file -Raw

# Update service references to use braid-specific imports
$content = $content -replace 'services\.NewCryptoServiceFromEnv', 'authServices.NewCryptoServiceFromEnv'
$content = $content -replace 'services\.SetGlobalCryptoService', 'authServices.SetGlobalCryptoService'
$content = $content -replace 'services\.NewBunnyService', 'videoServices.NewBunnyService'
$content = $content -replace 'services\.NewOptimizedBunnyService', 'videoServices.NewOptimizedBunnyService'
$content = $content -replace 'services\.SetGlobalOptimizedBunnyService', 'videoServices.SetGlobalOptimizedBunnyService'
$content = $content -replace 'services\.NewAnalyticsService', 'analyticsServices.NewAnalyticsService'
$content = $content -replace '\*services\.AnalyticsService', '*analyticsServices.AnalyticsService'
$content = $content -replace 'services\.NewStripeService', 'subServices.NewStripeService'
$content = $content -replace 'services\.NewSpacesService', 'videoServices.NewSpacesService'
$content = $content -replace 'services\.NewEmailService', 'commServices.NewEmailService'
$content = $content -replace '\*services\.EmailService', '*commServices.EmailService'
$content = $content -replace 'services\.NewBusinessIntelligenceService', 'analyticsServices.NewBusinessIntelligenceService'
$content = $content -replace '\*services\.BusinessIntelligenceService', '*analyticsServices.BusinessIntelligenceService'
$content = $content -replace 'services\.StartTokenBlacklistCleanup', 'authServices.StartTokenBlacklistCleanup'

Set-Content -Path $file -Value $content -NoNewline

Write-Host "✅ Service references updated in main.go"

