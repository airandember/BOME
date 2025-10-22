# 📋 COPY EXISTING FILES TO CONTEXT
# Copies all documentation files that still exist in root to their CONTEXT locations

Write-Host "📋 Copying existing files to CONTEXT..." -ForegroundColor Cyan
Write-Host ""

$copied = 0

# Helper function
function Copy-ToContext {
    param($Source, $Destination)
    
    if (Test-Path $Source) {
        $destDir = Split-Path -Parent $Destination
        if (-not (Test-Path $destDir)) {
            New-Item -ItemType Directory -Path $destDir -Force | Out-Null
        }
        Copy-Item $Source $Destination -Force
        Write-Host "  ✅ $Source → $Destination" -ForegroundColor Green
        return $true
    } else {
        Write-Host "  ⚠️  $Source (not found)" -ForegroundColor Yellow
        return $false
    }
}

# DEPLOYMENT
Write-Host "🚀 Deployment Files..." -ForegroundColor Cyan
if (Copy-ToContext "DEPLOYMENT_GUIDE.md" "CONTEXT/5-DEPLOYMENT/DEPLOYMENT_GUIDE.md") { $copied++ }
if (Copy-ToContext "DEPLOYMENT_QUICK_REFERENCE.md" "CONTEXT/5-DEPLOYMENT/DEPLOYMENT_QUICK_REFERENCE.md") { $copied++ }
if (Copy-ToContext "DEPLOYMENT_SUMMARY.md" "CONTEXT/5-DEPLOYMENT/DEPLOYMENT_SUMMARY.md") { $copied++ }
if (Copy-ToContext "DOCKER_README.md" "CONTEXT/5-DEPLOYMENT/DOCKER_README.md") { $copied++ }
if (Copy-ToContext "GIT_WORKFLOW.md" "CONTEXT/5-DEPLOYMENT/GIT_WORKFLOW.md") { $copied++ }

# MIGRATIONS - Stripe
Write-Host "`n💳 Stripe Files..." -ForegroundColor Cyan
if (Copy-ToContext "STRIPE_SUBSCRIPTION_INTEGRATION.md" "CONTEXT/6-MIGRATIONS/stripe/STRIPE_SUBSCRIPTION_INTEGRATION.md") { $copied++ }
if (Copy-ToContext "STRIPE_TESTING_GUIDE.md" "CONTEXT/6-MIGRATIONS/stripe/STRIPE_TESTING_GUIDE.md") { $copied++ }
if (Copy-ToContext "STRIPE_COUPON_ENHANCEMENTS.md" "CONTEXT/6-MIGRATIONS/stripe/STRIPE_COUPON_ENHANCEMENTS.md") { $copied++ }

# MIGRATIONS - Videos
Write-Host "`n📹 Video Files..." -ForegroundColor Cyan
if (Copy-ToContext "BUNNY_STATUS_MAPPING.md" "CONTEXT/6-MIGRATIONS/videos/BUNNY_STATUS_MAPPING.md") { $copied++ }

# MIGRATIONS - YouTube
Write-Host "`n📺 YouTube Files..." -ForegroundColor Cyan
if (Copy-ToContext "YOUTUBE_IMPLEMENTATION_SUMMARY.md" "CONTEXT/6-MIGRATIONS/youtube/YOUTUBE_IMPLEMENTATION_SUMMARY.md") { $copied++ }
if (Copy-ToContext "YOUTUBE_SETUP.md" "CONTEXT/6-MIGRATIONS/youtube/YOUTUBE_SETUP.md") { $copied++ }
if (Copy-ToContext "YOUTUBE_BACKEND_INTEGRATION.md" "CONTEXT/6-MIGRATIONS/youtube/YOUTUBE_BACKEND_INTEGRATION.md") { $copied++ }

# MIGRATIONS - Subscriptions
Write-Host "`n💰 Subscription Files..." -ForegroundColor Cyan
if (Copy-ToContext "SUBSCRIPTION_SYSTEM_ARCHITECTURE.md" "CONTEXT/6-MIGRATIONS/subscriptions/SUBSCRIPTION_SYSTEM_ARCHITECTURE.md") { $copied++ }
if (Copy-ToContext "SUBSCRIPTION_SYSTEM_IMPLEMENTATION.md" "CONTEXT/6-MIGRATIONS/subscriptions/SUBSCRIPTION_SYSTEM_IMPLEMENTATION.md") { $copied++ }

# MIGRATIONS - Admin
Write-Host "`n🔧 Admin Files..." -ForegroundColor Cyan
if (Copy-ToContext "ADMIN_PANEL_README.md" "CONTEXT/6-MIGRATIONS/admin/ADMIN_PANEL_README.md") { $copied++ }

# STATUS
Write-Host "`n📊 Status Files..." -ForegroundColor Cyan
if (Copy-ToContext "ERROR_HANDLING_IMPROVEMENTS.md" "CONTEXT/8-STATUS/ERROR_HANDLING_IMPROVEMENTS.md") { $copied++ }
if (Copy-ToContext "PROJECT_TASK_LIST.md" "CONTEXT/8-STATUS/PROJECT_TASK_LIST.md") { $copied++ }
if (Copy-ToContext "REVERT_STREAMING_FIXES.md" "CONTEXT/8-STATUS/REVERT_STREAMING_FIXES.md") { $copied++ }

# BRAIDS - Authentication
Write-Host "`n🔐 Authentication Files..." -ForegroundColor Cyan
if (Copy-ToContext "AUTHENTICATION_DEBUG_PLAN.md" "CONTEXT/9-BRAIDS/authentication/AUTHENTICATION_DEBUG_PLAN.md") { $copied++ }
if (Copy-ToContext "AUTHENTICATION_IMPLEMENTATION.md" "CONTEXT/9-BRAIDS/authentication/AUTHENTICATION_IMPLEMENTATION.md") { $copied++ }

# FEATURES
Write-Host "`n✨ Feature Files..." -ForegroundColor Cyan
if (Copy-ToContext "PASSWORD_CHANGE_FEATURE.md" "CONTEXT/10-FEATURES/PASSWORD_CHANGE_FEATURE.md") { $copied++ }
if (Copy-ToContext "STREAMING_OPTIMIZATION_SUMMARY.md" "CONTEXT/10-FEATURES/STREAMING_OPTIMIZATION_SUMMARY.md") { $copied++ }
if (Copy-ToContext "STREAMING_SYSTEM_OPTIMIZATION_GUIDE.md" "CONTEXT/10-FEATURES/STREAMING_SYSTEM_OPTIMIZATION_GUIDE.md") { $copied++ }
if (Copy-ToContext "ADVERTISER_WORKFLOW_TEST.md" "CONTEXT/10-FEATURES/ADVERTISER_WORKFLOW_TEST.md") { $copied++ }

# GUIDES
Write-Host "`n📖 Guide Files..." -ForegroundColor Cyan
if (Copy-ToContext "TEST_ACCOUNTS.md" "CONTEXT/11-GUIDES/TEST_ACCOUNTS.md") { $copied++ }
if (Copy-ToContext "TEST_USERS_GUIDE.md" "CONTEXT/11-GUIDES/TEST_USERS_GUIDE.md") { $copied++ }
if (Copy-ToContext "Taskfile.md" "CONTEXT/11-GUIDES/Taskfile.md") { $copied++ }

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "✅ FILE COPYING COMPLETE!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "📊 Summary:" -ForegroundColor Cyan
Write-Host "  ✅ Files copied: $copied" -ForegroundColor Green
Write-Host ""

