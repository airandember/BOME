# Fix all remaining compilation issues

Write-Host "Fixing all remaining issues..."
Write-Host ""

# 1. Fix analytics - close the unterminated comment
Write-Host "1. Fixing analytics..."
$analyticsFile = "backend/analytics/services/analytics.go"
$content = Get-Content $analyticsFile -Raw

# Replace the malformed comment close with proper one
$content = $content -replace '// Original code below - commented out for now \*/', "// Original code below - commented out for now`n*/"

Set-Content -Path $analyticsFile -Value $content -NoNewline
Write-Host "   Analytics comment fixed"

# 2. Fix routing - need to import database and service packages properly
Write-Host "2. Fixing routing..."
$routingFile = "backend/routing/setup.go"
$content = Get-Content $routingFile -Raw

# Add the infrastructure imports after the existing model imports
$content = $content -replace '(contentModels "bome-backend/content/models")', @'
$1
	infraServices "bome-backend/infrastructure/services"
	videoServices "bome-backend/video-streaming/services"
	subServices "bome-backend/subscription/services"
	commServices "bome-backend/communication/services"
	analyticsServices "bome-backend/analytics/services"
'@

# Fix the type references in function signature
$content = $content -replace '\*models\.Redis', '*database.Redis'
$content = $content -replace '\*services\.BunnyService', '*videoServices.BunnyService'
$content = $content -replace '\*services\.StripeService', '*subServices.StripeService'
$content = $content -replace '\*services\.SpacesService', '*infraServices.SpacesService'
$content = $content -replace '\*services\.EmailService', '*commServices.EmailService'
$content = $content -replace '\*services\.BusinessIntelligenceService', '*analyticsServices.BusinessIntelligenceService'

# Fix service references in function body
$content = $content -replace '([^\*\w])services\.', '$1subServices.'

Set-Content -Path $routingFile -Value $content -NoNewline
Write-Host "   Routing imports fixed"

# 3. Fix subscription handlers
Write-Host "3. Fixing subscription handlers..."
$subFile = "backend/subscription/handlers/subscription.go"
$content = Get-Content $subFile -Raw

# Add necessary imports
$content = $content -replace '(import \()', @'
$1
	"bome-backend/infrastructure/database"
	subServices "bome-backend/subscription/services"
'@

# Fix type references
$content = $content -replace 'models\.DB', 'database.DB'
$content = $content -replace 'services\.StripeService', 'subServices.StripeService'
$content = $content -replace 'services\.SubscriptionAnalyticsService', 'subServices.SubscriptionAnalyticsService'

Set-Content -Path $subFile -Value $content -NoNewline
Write-Host "   Subscription handlers fixed"

Write-Host ""
Write-Host "All fixes applied!"
Write-Host ""
