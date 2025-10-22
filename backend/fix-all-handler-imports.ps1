# Comprehensive fix for all handler imports to use shared services

Write-Host "`n🔧 Fixing ALL handler imports systematically...`n" -ForegroundColor Cyan

# Fix admin_streaming.go
Write-Host "Fixing admin/handlers/admin_streaming.go..." -ForegroundColor Yellow
$adminStreaming = Get-Content "admin/handlers/admin_streaming.go" -Raw

# Replace imports
$adminStreaming = $adminStreaming -replace 'import \([^)]+\)', @'
import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bome-backend/authentication/middleware"
	"bome-backend/infrastructure/database"
	"bome-backend/services/analytics"
	"bome-backend/services/bunny"
	"bome-backend/services/crypto"
	"bome-backend/services/stripe"

	"github.com/gin-gonic/gin"
	stripeAPI "github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/coupon"
)
'@

# Fix function signature
$adminStreaming = $adminStreaming -replace 'func SetupAdminStreamingRoutes\([^)]+\)', 'func SetupAdminStreamingRoutes(admin *gin.RouterGroup, db *database.DB, stripeService *stripe.StripeService, analyticsService *analytics.SubscriptionAnalyticsService, biService *analytics.BusinessIntelligenceService, subscriptionPlanStripeService *stripe.SubscriptionPlanStripeService, subscriptionOffersStripeService *stripe.SubscriptionOffersStripeService, bunnyService *bunny.BunnyService)'

# Fix crypto service references
$adminStreaming = $adminStreaming -replace 'crypto := crypto\.GetGlobalCryptoService\(\)', 'cryptoService := crypto.GetGlobalCryptoService()'
$adminStreaming = $adminStreaming -replace 'if crypto == nil', 'if cryptoService == nil'
$adminStreaming = $adminStreaming -replace 'crypto\.Encrypt', 'cryptoService.Encrypt'
$adminStreaming = $adminStreaming -replace 'crypto\.Decrypt', 'cryptoService.Decrypt'

Set-Content "admin/handlers/admin_streaming.go" $adminStreaming -NoNewline
Write-Host "  ✅ Fixed admin_streaming.go" -ForegroundColor Green

# Fix auth.go
Write-Host "`nFixing authentication/handlers/auth.go..." -ForegroundColor Yellow
$authFile = Get-Content "authentication/handlers/auth.go" -Raw
$authFile = $authFile -replace '"bome-backend/authentication/models"', '"bome-backend/infrastructure/database"`n`tauthModels "bome-backend/authentication/models"'
$authFile = $authFile -replace '\*models\.DB', '*database.DB'
$authFile = $authFile -replace 'email\.', 'emailService.'
Set-Content "authentication/handlers/auth.go" $authFile -NoNewline
Write-Host "  ✅ Fixed auth.go" -ForegroundColor Green

Write-Host "`n✅ All handler imports fixed!`n" -ForegroundColor Green

