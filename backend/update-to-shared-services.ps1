# Update all handler imports to use shared services layer

Write-Host "`nUpdating imports to use shared services...`n" -ForegroundColor Cyan

$replacements = @{
    '"bome-backend/authentication/services"' = 'authServices "bome-backend/services/crypto"`n`t"bome-backend/services/email"'
    '"bome-backend/subscription/services"' = 'stripeServices "bome-backend/services/stripe"`n`t"bome-backend/services/analytics"'
    '"bome-backend/video-streaming/services"' = '"bome-backend/services/bunny"'
    '"bome-backend/communication/services"' = '"bome-backend/services/email"'
}

$files = Get-ChildItem -Path "." -Recurse -Filter "*.go" | Where-Object { $_.FullName -notmatch 'services\\' }

$count = 0
foreach ($file in $files) {
    $content = Get-Content $file.FullName -Raw
    $original = $content
    
    # Replace old service imports with shared services
    $content = $content -replace 'import \("bome-backend/authentication/services"\)', 'import ("bome-backend/services/crypto")'
    $content = $content -replace 'import \("bome-backend/subscription/services"\)', 'import ("bome-backend/services/stripe")'
    $content = $content -replace 'import \("bome-backend/video-streaming/services"\)', 'import ("bome-backend/services/bunny")'
    $content = $content -replace 'import \("bome-backend/communication/services"\)', 'import ("bome-backend/services/email")'
    
    # Fix service references
    $content = $content -replace 'services\.StripeService', 'stripe.StripeService'
    $content = $content -replace 'services\.BunnyService', 'bunny.BunnyService'
    $content = $content -replace 'services\.EmailService', 'email.EmailService'
    $content = $content -replace 'services\.CryptoService', 'crypto.CryptoService'
    $content = $content -replace 'services\.GetGlobalCryptoService', 'crypto.GetGlobalCryptoService'
    $content = $content -replace 'services\.SubscriptionAnalyticsService', 'analytics.SubscriptionAnalyticsService'
    $content = $content -replace 'services\.BusinessIntelligenceService', 'analytics.BusinessIntelligenceService'
    
    if ($content -ne $original) {
        Set-Content $file.FullName $content -NoNewline
        Write-Host "  Updated: $($file.Name)" -ForegroundColor Green
        $count++
    }
}

Write-Host "`n✅ Updated $count files`n" -ForegroundColor Green

