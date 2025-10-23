# 🧹 FINAL ROOT CLEANUP - Move & Delete
# Organizes root folder by moving recovery docs and archiving scripts

Write-Host "🧹 Starting Final Root Cleanup..." -ForegroundColor Cyan
Write-Host ""

$moved = 0
$deleted = 0
$archived = 0

# ============================================================================
# CREATE FOLDERS
# ============================================================================
Write-Host "📁 Creating folders..." -ForegroundColor Cyan
New-Item -ItemType Directory -Path "CONTEXT/0-RECOVERY" -Force | Out-Null
New-Item -ItemType Directory -Path "CONTEXT/scripts/recovery" -Force | Out-Null
New-Item -ItemType Directory -Path "scripts/archive/fix-scripts" -Force | Out-Null
New-Item -ItemType Directory -Path "scripts/testing" -Force | Out-Null
Write-Host "  ✅ Folders created" -ForegroundColor Green
Write-Host ""

# ============================================================================
# MOVE RECOVERY DOCUMENTATION TO CONTEXT
# ============================================================================
Write-Host "📝 Moving recovery documentation to CONTEXT/0-RECOVERY..." -ForegroundColor Cyan

$recoveryDocs = @(
    "CONTEXT_RECOVERY_PROGRESS.md",
    "DELETED_FILES_MANIFEST.md",
    "DOCUMENTATION_ORGANIZATION_COMPLETE.md",
    "DOCUMENTATION_ORGANIZATION_GUIDE.md",
    "RECOVERY_COMPLETE_SUMMARY.md",
    "RECOVERY_STATUS.md",
    "SESSION_2_COMPLETE_SUMMARY.md",
    "URGENT_FILE_RECOVERY_GUIDE.md"
)

foreach ($file in $recoveryDocs) {
    if (Test-Path $file) {
        Move-Item $file "CONTEXT/0-RECOVERY/$file" -Force
        Write-Host "  ✅ Moved: $file" -ForegroundColor Green
        $moved++
    }
}

Write-Host ""

# ============================================================================
# MOVE RECOVERY SCRIPTS TO CONTEXT
# ============================================================================
Write-Host "🔧 Moving recovery scripts to CONTEXT/scripts/recovery..." -ForegroundColor Cyan

$recoveryScripts = @(
    "CREATE_PLACEHOLDER_FILES.ps1",
    "COPY_EXISTING_FILES.ps1",
    "ORGANIZE_CONTEXT_COMPLETE_V2.ps1",
    "ORGANIZE_CONTEXT_COMPLETE.ps1",
    "CLEANUP_ROOT_DOCUMENTATION.ps1"
)

foreach ($script in $recoveryScripts) {
    if (Test-Path $script) {
        Move-Item $script "CONTEXT/scripts/recovery/$script" -Force
        Write-Host "  ✅ Moved: $script" -ForegroundColor Green
        $moved++
    }
}

Write-Host ""

# ============================================================================
# ARCHIVE FIX/DEVELOPMENT SCRIPTS
# ============================================================================
Write-Host "📦 Archiving development/fix scripts to scripts/archive..." -ForegroundColor Cyan

$fixScripts = @(
    "consolidate-braids.ps1",
    "consolidate-braids-simple.ps1",
    "convert-all-model-methods.ps1",
    "convert-remaining-models.ps1",
    "create-stub-functions.ps1",
    "COMPLETE-TO-100-PERCENT.ps1",
    "final-100-percent.ps1",
    "final-final-fixes.ps1",
    "FINAL-FIX-TO-100.ps1",
    "final-fixes.ps1",
    "fix-all-import-syntax.ps1",
    "fix-all-models.ps1",
    "fix-all-remaining.ps1",
    "fix-all-services.ps1",
    "fix-analytics-100-percent.ps1",
    "fix-analytics-calls.ps1",
    "fix-analytics-syntax.ps1",
    "fix-crypto-refs.ps1",
    "fix-double-db.ps1",
    "fix-email-calls.ps1",
    "fix-final-errors.ps1",
    "fix-helper-imports.ps1",
    "fix-main-services.ps1",
    "fix-middleware-cross-braid.ps1",
    "fix-middleware-refs.ps1",
    "fix-middleware.ps1",
    "fix-oauth-calls.ps1",
    "fix-remaining-errors.ps1",
    "fix-routing.ps1",
    "fix-syntax-errors.ps1",
    "fix-user-calls.ps1",
    "fix-user-functions.ps1",
    "fix-video-calls.ps1",
    "fix-video-final.ps1",
    "HIT-100-PERCENT.ps1",
    "LAST-1-PERCENT.ps1",
    "stub-analytics-problems.ps1",
    "update-imports.ps1"
)

foreach ($script in $fixScripts) {
    if (Test-Path $script) {
        Move-Item $script "scripts/archive/fix-scripts/$script" -Force
        Write-Host "  ✅ Archived: $script" -ForegroundColor Yellow
        $archived++
    }
}

Write-Host ""

# ============================================================================
# MOVE TEST SCRIPTS
# ============================================================================
Write-Host "🧪 Moving test scripts to scripts/testing..." -ForegroundColor Cyan

$testScripts = @(
    "test_sync.ps1",
    "test-admin-endpoints.ps1",
    "test-fullstack-integration.ps1",
    "test-production-endpoints.md"
)

foreach ($script in $testScripts) {
    if (Test-Path $script) {
        Move-Item $script "scripts/testing/$script" -Force
        Write-Host "  ✅ Moved: $script" -ForegroundColor Green
        $moved++
    }
}

Write-Host ""

# ============================================================================
# DELETE DUPLICATE .MD FILES (Already in CONTEXT)
# ============================================================================
Write-Host "🗑️  Deleting duplicate .md files (already in CONTEXT)..." -ForegroundColor Cyan

# Helper function to check if file exists in CONTEXT
function Test-InContext {
    param($FileName)
    
    # Check common locations in CONTEXT
    $locations = @(
        "CONTEXT/1-ARCHITECTURE/$FileName",
        "CONTEXT/2-DATABASE/$FileName",
        "CONTEXT/3-TESTING/$FileName",
        "CONTEXT/4-FRONTEND/$FileName",
        "CONTEXT/5-DEPLOYMENT/$FileName",
        "CONTEXT/6-MIGRATIONS/admin/$FileName",
        "CONTEXT/6-MIGRATIONS/stripe/$FileName",
        "CONTEXT/6-MIGRATIONS/videos/$FileName",
        "CONTEXT/6-MIGRATIONS/youtube/$FileName",
        "CONTEXT/6-MIGRATIONS/subscriptions/$FileName",
        "CONTEXT/8-STATUS/$FileName",
        "CONTEXT/9-BRAIDS/authentication/$FileName",
        "CONTEXT/10-FEATURES/$FileName",
        "CONTEXT/11-GUIDES/$FileName"
    )
    
    foreach ($loc in $locations) {
        if (Test-Path $loc) {
            return $true
        }
    }
    return $false
}

$duplicates = @(
    "ADMIN_PANEL_README.md",
    "ADVERTISER_WORKFLOW_TEST.md",
    "AUTHENTICATION_DEBUG_PLAN.md",
    "AUTHENTICATION_IMPLEMENTATION.md",
    "BOME_NAMING_CONVENTIONS.md",
    "BUNNY_STATUS_MAPPING.md",
    "DEPARTMENT_ROLES_MIGRATION.md",
    "DEPLOYMENT_GUIDE.md",
    "DEPLOYMENT_QUICK_REFERENCE.md",
    "DEPLOYMENT_SUMMARY.md",
    "DOCKER_README.md",
    "ERROR_HANDLING_IMPROVEMENTS.md",
    "FIGMA_DESIGN_SYSTEM_IMPLEMENTATION.md",
    "GIT_WORKFLOW.md",
    "IFRAME_URL_FIX.md",
    "JSON_MOCK_DATA_ARCHITECTURE.md",
    "MASTER_BRAID_IMPLEMENTATION_PLAN.md",
    "PASSWORD_CHANGE_FEATURE.md",
    "POSTGRESQL_MIGRATION_SUMMARY.md",
    "POSTGRESQL_MIGRATION.md",
    "PROJECT_STRUCTURE.md",
    "PROJECT_TASK_LIST.md",
    "QUICK_TEST_REFERENCE.md",
    "RBAC_ASSESSMENT.md",
    "REVERT_STREAMING_FIXES.md",
    "STREAMING_OPTIMIZATION_SUMMARY.md",
    "STREAMING_SYSTEM_OPTIMIZATION_GUIDE.md",
    "STRIPE_COUPON_ENHANCEMENTS.md",
    "STRIPE_SUBSCRIPTION_INTEGRATION.md",
    "STRIPE_TESTING_GUIDE.md",
    "SUBSCRIPTION_SYSTEM_ARCHITECTURE.md",
    "SUBSCRIPTION_SYSTEM_IMPLEMENTATION.md",
    "TAGS_CATEGORIES_STRAND_ASSESSMENT.md",
    "Taskfile.md",
    "TEST_ACCOUNTS.md",
    "TEST_USERS_GUIDE.md",
    "VIDEO_PLAYER_FINAL_SOLUTION.md",
    "VIDEO_PLAYER_IMPROVEMENTS.md",
    "YOUTUBE_BACKEND_INTEGRATION.md",
    "YOUTUBE_IMPLEMENTATION_SUMMARY.md",
    "YOUTUBE_SETUP.md"
)

foreach ($file in $duplicates) {
    if ((Test-Path $file) -and (Test-InContext $file)) {
        Remove-Item $file -Force
        Write-Host "  ✅ Deleted: $file (copy in CONTEXT)" -ForegroundColor Red
        $deleted++
    } elseif (Test-Path $file) {
        Write-Host "  ⚠️  Skipped: $file (not found in CONTEXT)" -ForegroundColor Yellow
    }
}

# Delete temporary/manual fix docs
$tempDocs = @(
    "FINAL-COMPREHENSIVE-FIX.md",
    "MANUAL-FINAL-FIXES.md"
)

foreach ($file in $tempDocs) {
    if (Test-Path $file) {
        Remove-Item $file -Force
        Write-Host "  ✅ Deleted: $file (temporary doc)" -ForegroundColor Red
        $deleted++
    }
}

Write-Host ""

# ============================================================================
# SUMMARY
# ============================================================================
Write-Host "========================================" -ForegroundColor Green
Write-Host "✅ CLEANUP COMPLETE!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "📊 Summary:" -ForegroundColor Cyan
Write-Host "  ✅ Files moved: $moved" -ForegroundColor Green
Write-Host "  📦 Files archived: $archived" -ForegroundColor Yellow
Write-Host "  🗑️  Files deleted: $deleted" -ForegroundColor Red
Write-Host ""
Write-Host "📁 Structure:" -ForegroundColor Cyan
Write-Host "  ✅ CONTEXT/0-RECOVERY/ - Recovery documentation" -ForegroundColor White
Write-Host "  ✅ CONTEXT/scripts/recovery/ - Recovery scripts" -ForegroundColor White
Write-Host "  ✅ scripts/archive/fix-scripts/ - Development scripts" -ForegroundColor White
Write-Host "  ✅ scripts/testing/ - Test scripts" -ForegroundColor White
Write-Host ""
Write-Host "🎉 Root folder is now clean!" -ForegroundColor Green
Write-Host ""
Write-Host "Remaining in root:" -ForegroundColor Cyan
Write-Host "  ✅ README.md" -ForegroundColor White
Write-Host "  ✅ START_SERVERS.ps1" -ForegroundColor White
Write-Host "  ✅ start-dev-fullstack.ps1" -ForegroundColor White
Write-Host "  ✅ FINAL_CLEANUP_ROOT.ps1 (this script)" -ForegroundColor White
Write-Host ""

