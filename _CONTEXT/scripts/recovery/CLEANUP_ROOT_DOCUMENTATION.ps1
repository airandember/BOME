# 🧹 CLEANUP ROOT DOCUMENTATION FILES
# Removes documentation files that have been copied to CONTEXT/
# Keeps essential project files (README.md, .gitignore, etc.)

Write-Host "🧹 Starting Root Directory Cleanup..." -ForegroundColor Cyan
Write-Host ""

# Check if CONTEXT folder exists
if (-not (Test-Path "CONTEXT")) {
    Write-Host "❌ Error: CONTEXT folder not found!" -ForegroundColor Red
    Write-Host "Please run ORGANIZE_CONTEXT_COMPLETE_V2.ps1 first." -ForegroundColor Yellow
    exit 1
}

# Files to DELETE (all documentation files that were copied to CONTEXT/)
$filesToDelete = @(
    # Architecture
    "BOME_CONTEXT_STANDARD.md",
    "BOME_BRAIDS_SUMMARY.md",
    "BOME_NAMING_CONVENTIONS.md",
    "BRAIDS_INDEX.md",
    "ARCHITECTURAL_REFINEMENT_PLAN.md",
    "ARCHITECTURE_COMPARISON.md",
    "MASTER_BRAID_IMPLEMENTATION_PLAN.md",
    "PROJECT_STRUCTURE.md",
    "SHARED_SERVICES_IMPLEMENTATION_STATUS.md",
    
    # Database
    "DATABASE_SCHEMA.md",
    "POSTGRESQL_MIGRATION_SUMMARY.md",
    "POSTGRESQL_MIGRATION.md",
    "DEPARTMENT_ROLES_MIGRATION.md",
    "RBAC_ASSESSMENT.md",
    
    # Testing
    "BRAID_COMBING_STANDARD.md",
    "BRAID_COMB_CHECKLIST.md",
    "BRAID_COMBING_METHODOLOGY.md",
    "TESTING_STANDARD_COMPLETE.md",
    "TESTING_CHECKLIST.md",
    "QUICK_TEST_REFERENCE.md",
    "SPLIT_END_NAMING_CONVENTION.md",
    
    # Frontend
    "SVELTE5_REACTIVITY_GUIDE.md",
    "SVELTE5_REACTIVITY_FINAL_FIX.md",
    "REACTIVITY_FIX.md",
    "FIGMA_DESIGN_SYSTEM_IMPLEMENTATION.md",
    "JSON_MOCK_DATA_ARCHITECTURE.md",
    "VIDEO_PLAYER_FINAL_SOLUTION.md",
    "VIDEO_PLAYER_IMPROVEMENTS.md",
    "LOADING_WHEEL_FIX.md",
    "INFINITE_SPINNER_DEBUG.md",
    "IFRAME_URL_FIX.md",
    "NEUMORPHIC_SUBSITE_ICONS.md",
    "ADMIN_SIDEBAR_COLLAPSE_IMPLEMENTATION.md",
    
    # Deployment
    "DEPLOYMENT_GUIDE.md",
    "DEPLOYMENT_QUICK_REFERENCE.md",
    "DEPLOYMENT_SUMMARY.md",
    "DOCKER_README.md",
    "GIT_WORKFLOW.md",
    
    # Stripe
    "STRIPE_MIGRATION_ARCHITECTURE.md",
    "STRIPE_PHASE_1_COMPLETE.md",
    "STRIPE_PHASE_2_COMPLETE.md",
    "STRIPE_PHASE_3_COMPLETE.md",
    "STRIPE_PHASE_4_COMPLETE.md",
    "STRIPE_COMPLETE_ALL_PHASES.md",
    "STRIPE_SUBSCRIPTION_INTEGRATION.md",
    "STRIPE_SYNC_MIGRATION_PLAN.md",
    "STRIPE_GHOSTS_MIGRATION_GUIDE.md",
    "STRIPE_TESTING_GUIDE.md",
    "STRIPE_COUPON_ENHANCEMENTS.md",
    "STRIPE_QUERY_OPTIMIZATION_PLAN.md",
    "STRIPE_REAL_SCHEMA_ASSESSMENT.md",
    "STRIPE_SUBS_TAB_ANALYSIS.md",
    
    # Videos
    "VIDEOS_MIGRATION_COMPLETE.md",
    "VIDEOS_MIGRATION_FINAL_SUMMARY.md",
    "VIDEOS_COMPLETE_INTEGRATION.md",
    "VIDEOS_30_ENDPOINTS_LIST.md",
    "VIDEOS_STRAND_ANALYSIS.md",
    "VIDEOS_INTEGRATION_STATUS.md",
    "VIDEOS_FRONTEND_INTEGRATION_ANALYSIS.md",
    "BUNNY_STATUS_MAPPING.md",
    
    # YouTube
    "YOUTUBE_MIGRATION_COMPLETE.md",
    "YOUTUBE_IMPLEMENTATION_SUMMARY.md",
    "YOUTUBE_PHASES_1-3_COMPLETE.md",
    "YOUTUBE_SETUP.md",
    "YOUTUBE_STRAND_ANALYSIS.md",
    "YOUTUBE_BACKEND_INTEGRATION.md",
    "YOUTUBE_DATABASE_MIGRATION.md",
    
    # Subscriptions
    "SUBSCRIBERS_MIGRATION_COMPLETE.md",
    "SUBSCRIBERS_SUBSCRIPTIONS_MIGRATION.md",
    "SUBSCRIPTION_PLANS_OFFERS_COMPLETE.md",
    "SUBSCRIPTION_SYSTEM_ARCHITECTURE.md",
    "SUBSCRIPTION_SYSTEM_IMPLEMENTATION.md",
    
    # Creator Payouts
    "CREATOR_PAYOUT_MIGRATION_GUIDE.md",
    "CREATOR_PAYOUTS_500_ERROR_FIX.md",
    "CREATOR_PAYOUTS_NAVIGATION_ADDED.md",
    "PHASE_7_COMPREHENSIVE_PLAN.md",
    "PHASE_7_SQL_COMPLETE.md",
    "PHASE_7B_CREATOR_PAYOUTS_COMPLETE.md",
    "PHASE_7B_QUICK_START.md",
    
    # Admin
    "ADMIN_ROUTES_MIGRATION_PLAN.md",
    "ADMIN_ROUTES_STATUS.md",
    "ADMIN_DASHBOARD_CLEANUP_PLAN.md",
    "ADMIN_DASHBOARD_SESSION_STATUS.md",
    "ADMIN_PANEL_README.md",
    
    # Braid Migrations
    "BRAID_MIGRATION_COMPLETE.md",
    "BRAID_MIGRATION_STATUS.md",
    "FINAL_MIGRATION_STATUS.md",
    "IMPORT_MAPPING.md",
    "IMPORT_PATH_FIX.md",
    
    # Phases
    "50_PERCENT_MILESTONE.md",
    "90-PERCENT-STATUS.md",
    "99-PERCENT-STATUS.md",
    "100-PERCENT-FINAL.md",
    "🎊_100_PERCENT_COMPLETE.md",
    "100_PERCENT_COMPLETE_FINAL_REPORT.md",
    "🎊_TODAY_ACHIEVEMENTS.md",
    "TODAYS_ACCOMPLISHMENTS.md",
    "PHASE_6_COMPLETE.md",
    
    # Status
    "PRODUCTION_READINESS_REPORT.md",
    "FINAL_STATUS_REPORT.md",
    "BUGS_FIXED_SUMMARY.md",
    "ERROR_HANDLING_IMPROVEMENTS.md",
    "PROJECT_TASK_LIST.md",
    "FRONTEND_BACKEND_FIXES.md",
    "ROUTING_FIXES_COMPLETE.md",
    "REVERT_STREAMING_FIXES.md",
    
    # Braids
    "AUTH_BRAID_COMPLETE_REPORT.md",
    "AUTHENTICATION_BRAID_COMPLETE.md",
    "AUTHENTICATION_BRAID_COMBING.md",
    "AUTHENTICATION_DEBUG_PLAN.md",
    "AUTHENTICATION_IMPLEMENTATION.md",
    "SUBSCRIPTION_BILLING_BRAID_COMPLETE.md",
    "SUBSCRIPTION_BRAID_COMPLETE.md",
    "USER_MANAGEMENT_BRAID_COMPLETE.md",
    "CONTENT_MANAGEMENT_BRAID_COMPLETE.md",
    "COMMUNICATION_BRAID_COMPLETE.md",
    "ADVERTISEMENT_SYSTEM_BRAID_COMPLETE.md",
    
    # Split Ends
    "SPLIT_END_TRACKER_AuthBraid.md",
    "SPLIT_END_TRACKER_SubBraid.md",
    "SPLIT_END_TRACKER_VideoBraid.md",
    "SPLIT_END_TRACKER_UserMgmtBraid.md",
    "SPLIT_END_TRACKER_ContentBraid.md",
    "SPLIT_END_TRACKER_CommunicationBraid.md",
    "SPLIT_END_TRACKER_AnalyticsBraid.md",
    "SPLIT_END_TRACKER_AdvertisementBraid.md",
    "SPLIT_END_TRACKER_AdminBraid.md",
    "SPLIT_END_TRACKER_InfrastructureBraid.md",
    "SPLIT_END_REPAIR_AuthBraid_001.md",
    "SPLIT_END_REPAIR_AuthBraid_002.md",
    "SPLIT_END_REPAIR_AuthBraid_003.md",
    "SPLIT_ENDS_REPAIRED.md",
    
    # Features
    "PASSWORD_CHANGE_FEATURE.md",
    "WEBSOCKET_REALTIME_COMPLETE.md",
    "STREAMING_OPTIMIZATION_SUMMARY.md",
    "STREAMING_SYSTEM_OPTIMIZATION_GUIDE.md",
    "ADVERTISER_WORKFLOW_TEST.md",
    "BRAND_EXTRACTION_STATUS.md",
    "BRAND_INFRASTRUCTURE_COMPLETE.md",
    "TOGGLE_FIX_COMPLETE.md",
    "MISSION_3_FULLSTACK_INTEGRATION.md",
    
    # Guides
    "TEST_ACCOUNTS.md",
    "TEST_USERS_GUIDE.md",
    "Taskfile.md",
    
    # Organization files (these are now in CONTEXT/)
    "FILE_ORGANIZATION_PLAN.md",
    "README_CONTEXT_UPDATE.md",
    "ORGANIZE_CONTEXT.ps1"
)

# Files to KEEP in root (essential project files)
$filesToKeep = @(
    "README.md",
    ".gitignore",
    ".env",
    ".env.example",
    "package.json",
    "package-lock.json",
    "go.mod",
    "go.sum",
    "docker-compose.yml",
    "Dockerfile",
    # Keep the cleanup and organization scripts
    "ORGANIZE_CONTEXT_COMPLETE_V2.ps1",
    "CLEANUP_ROOT_DOCUMENTATION.ps1",
    "DOCUMENTATION_ORGANIZATION_COMPLETE.md"
)

Write-Host "📋 Checking files to delete..." -ForegroundColor Cyan
Write-Host ""

$existingFiles = @()
$missingFiles = @()

foreach ($file in $filesToDelete) {
    if (Test-Path $file) {
        $existingFiles += $file
    } else {
        $missingFiles += $file
    }
}

Write-Host "Found:" -ForegroundColor White
Write-Host "  ✅ $($existingFiles.Count) files to delete" -ForegroundColor Green
Write-Host "  ⚠️  $($missingFiles.Count) files already deleted/missing" -ForegroundColor Yellow
Write-Host ""

if ($existingFiles.Count -eq 0) {
    Write-Host "✨ No files to delete - root is already clean!" -ForegroundColor Green
    exit 0
}

Write-Host "Files to be deleted:" -ForegroundColor Yellow
foreach ($file in $existingFiles | Select-Object -First 10) {
    Write-Host "  - $file" -ForegroundColor Gray
}
if ($existingFiles.Count -gt 10) {
    Write-Host "  ... and $($existingFiles.Count - 10) more" -ForegroundColor Gray
}
Write-Host ""

Write-Host "⚠️  WARNING: This will DELETE $($existingFiles.Count) files from root!" -ForegroundColor Yellow
Write-Host "All files have been safely copied to CONTEXT/ folder." -ForegroundColor Cyan
Write-Host ""
$confirmation = Read-Host "Type 'yes' to proceed with deletion"

if ($confirmation -ne "yes") {
    Write-Host "❌ Deletion cancelled." -ForegroundColor Red
    exit 0
}

Write-Host ""
Write-Host "🗑️  Deleting files..." -ForegroundColor Cyan

$deleted = 0
$errors = 0

foreach ($file in $existingFiles) {
    try {
        Remove-Item $file -Force
        Write-Host "  ✅ Deleted: $file" -ForegroundColor Green
        $deleted++
    } catch {
        Write-Host "  ❌ Error deleting: $file - $($_.Exception.Message)" -ForegroundColor Red
        $errors++
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "✅ CLEANUP COMPLETE!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "📊 Summary:" -ForegroundColor Cyan
Write-Host "  ✅ Files deleted: $deleted" -ForegroundColor Green
Write-Host "  ❌ Errors: $errors" -ForegroundColor $(if ($errors -gt 0) { "Red" } else { "Green" })
Write-Host "  📁 All documentation now in: CONTEXT/" -ForegroundColor Cyan
Write-Host ""
Write-Host "🎉 Root directory is now clean!" -ForegroundColor Green
Write-Host ""
Write-Host "Kept in root:" -ForegroundColor White
Write-Host "  - Essential project files (README.md, .gitignore, etc.)" -ForegroundColor Gray
Write-Host "  - Source code directories (backend/, frontend/)" -ForegroundColor Gray
Write-Host "  - CONTEXT/ folder (all documentation)" -ForegroundColor Gray
Write-Host "  - Organization scripts (for reference)" -ForegroundColor Gray
Write-Host ""

