# 📁 COMPLETE CONTEXT FOLDER ORGANIZATION SCRIPT V2
# Organizes ALL documentation including milestone and feature docs

Write-Host "📚 Starting COMPLETE CONTEXT organization (V2)..." -ForegroundColor Cyan
Write-Host ""

# Check if we're in the right directory
if (-not (Test-Path "BOME_CONTEXT_STANDARD.md")) {
    Write-Host "❌ Error: Must be run from BOME root directory" -ForegroundColor Red
    exit 1
}

# Function to copy file with status
function Copy-Doc {
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
        Write-Host "  ⚠️  $Source (not found, skipping)" -ForegroundColor Yellow
        return $false
    }
}

$copied = 0
$skipped = 0

# ========================================
# 1. ARCHITECTURE
# ========================================
Write-Host "📐 1. Organizing Architecture Documentation..." -ForegroundColor Cyan
$archFiles = @(
    "BOME_CONTEXT_STANDARD.md",
    "BOME_BRAIDS_SUMMARY.md",
    "BOME_NAMING_CONVENTIONS.md",
    "BRAIDS_INDEX.md",
    "ARCHITECTURAL_REFINEMENT_PLAN.md",
    "ARCHITECTURE_COMPARISON.md",
    "MASTER_BRAID_IMPLEMENTATION_PLAN.md",
    "PROJECT_STRUCTURE.md",
    "SHARED_SERVICES_IMPLEMENTATION_STATUS.md"
)
foreach ($file in $archFiles) {
    if (Copy-Doc $file "CONTEXT/1-ARCHITECTURE/$file") { $copied++ } else { $skipped++ }
}

# ========================================
# 2. DATABASE
# ========================================
Write-Host "`n🗄️  2. Organizing Database Documentation..." -ForegroundColor Cyan
$dbFiles = @(
    "DATABASE_SCHEMA.md",
    "POSTGRESQL_MIGRATION_SUMMARY.md",
    "POSTGRESQL_MIGRATION.md",
    "DEPARTMENT_ROLES_MIGRATION.md",
    "RBAC_ASSESSMENT.md"
)
foreach ($file in $dbFiles) {
    if (Copy-Doc $file "CONTEXT/2-DATABASE/$file") { $copied++ } else { $skipped++ }
}

# ========================================
# 3. TESTING
# ========================================
Write-Host "`n🧪 3. Organizing Testing Documentation..." -ForegroundColor Cyan
$testFiles = @(
    "BRAID_COMBING_STANDARD.md",
    "BRAID_COMB_CHECKLIST.md",
    "BRAID_COMBING_METHODOLOGY.md",
    "TESTING_STANDARD_COMPLETE.md",
    "TESTING_CHECKLIST.md",
    "QUICK_TEST_REFERENCE.md",
    "SPLIT_END_NAMING_CONVENTION.md"
)
foreach ($file in $testFiles) {
    if (Copy-Doc $file "CONTEXT/3-TESTING/$file") { $copied++ } else { $skipped++ }
}

# ========================================
# 4. FRONTEND
# ========================================
Write-Host "`n🎨 4. Organizing Frontend Documentation..." -ForegroundColor Cyan
$frontendFiles = @(
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
    "ADMIN_SIDEBAR_COLLAPSE_IMPLEMENTATION.md"
)
foreach ($file in $frontendFiles) {
    if (Copy-Doc $file "CONTEXT/4-FRONTEND/$file") { $copied++ } else { $skipped++ }
}

# ========================================
# 5. DEPLOYMENT
# ========================================
Write-Host "`n🚀 5. Organizing Deployment Documentation..." -ForegroundColor Cyan
$deployFiles = @(
    "DEPLOYMENT_GUIDE.md",
    "DEPLOYMENT_QUICK_REFERENCE.md",
    "DEPLOYMENT_SUMMARY.md",
    "DOCKER_README.md",
    "GIT_WORKFLOW.md"
)
foreach ($file in $deployFiles) {
    if (Copy-Doc $file "CONTEXT/5-DEPLOYMENT/$file") { $copied++ } else { $skipped++ }
}

# Deployment scripts
if (Copy-Doc "START_SERVERS.ps1" "CONTEXT/5-DEPLOYMENT/scripts/START_SERVERS.ps1") { $copied++ } else { $skipped++ }
if (Copy-Doc "start-dev-fullstack.ps1" "CONTEXT/5-DEPLOYMENT/scripts/start-dev-fullstack.ps1") { $copied++ } else { $skipped++ }

# ========================================
# 6. MIGRATIONS (By Feature)
# ========================================
Write-Host "`n🔄 6. Organizing Migration Documentation..." -ForegroundColor Cyan

# Stripe
Write-Host "  Stripe migrations..."
$stripeFiles = @(
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
    "STRIPE_SUBS_TAB_ANALYSIS.md"
)
foreach ($file in $stripeFiles) {
    if (Copy-Doc $file "CONTEXT/6-MIGRATIONS/stripe/$file") { $copied++ } else { $skipped++ }
}

# Videos
Write-Host "  Video migrations..."
$videoFiles = @(
    "VIDEOS_MIGRATION_COMPLETE.md",
    "VIDEOS_MIGRATION_FINAL_SUMMARY.md",
    "VIDEOS_COMPLETE_INTEGRATION.md",
    "VIDEOS_30_ENDPOINTS_LIST.md",
    "VIDEOS_STRAND_ANALYSIS.md",
    "VIDEOS_INTEGRATION_STATUS.md",
    "VIDEOS_FRONTEND_INTEGRATION_ANALYSIS.md",
    "BUNNY_STATUS_MAPPING.md"
)
foreach ($file in $videoFiles) {
    if (Copy-Doc $file "CONTEXT/6-MIGRATIONS/videos/$file") { $copied++ } else { $skipped++ }
}

# YouTube
Write-Host "  YouTube migrations..."
$youtubeFiles = @(
    "YOUTUBE_MIGRATION_COMPLETE.md",
    "YOUTUBE_IMPLEMENTATION_SUMMARY.md",
    "YOUTUBE_PHASES_1-3_COMPLETE.md",
    "YOUTUBE_SETUP.md",
    "YOUTUBE_STRAND_ANALYSIS.md",
    "YOUTUBE_BACKEND_INTEGRATION.md",
    "YOUTUBE_DATABASE_MIGRATION.md"
)
foreach ($file in $youtubeFiles) {
    if (Copy-Doc $file "CONTEXT/6-MIGRATIONS/youtube/$file") { $copied++ } else { $skipped++ }
}

# Subscriptions
Write-Host "  Subscription migrations..."
$subFiles = @(
    "SUBSCRIBERS_MIGRATION_COMPLETE.md",
    "SUBSCRIBERS_SUBSCRIPTIONS_MIGRATION.md",
    "SUBSCRIPTION_PLANS_OFFERS_COMPLETE.md",
    "SUBSCRIPTION_SYSTEM_ARCHITECTURE.md",
    "SUBSCRIPTION_SYSTEM_IMPLEMENTATION.md"
)
foreach ($file in $subFiles) {
    if (Copy-Doc $file "CONTEXT/6-MIGRATIONS/subscriptions/$file") { $copied++ } else { $skipped++ }
}

# Creator Payouts
Write-Host "  Creator Payout migrations..."
$payoutFiles = @(
    "CREATOR_PAYOUT_MIGRATION_GUIDE.md",
    "CREATOR_PAYOUTS_500_ERROR_FIX.md",
    "CREATOR_PAYOUTS_NAVIGATION_ADDED.md",
    "PHASE_7_COMPREHENSIVE_PLAN.md",
    "PHASE_7_SQL_COMPLETE.md",
    "PHASE_7B_CREATOR_PAYOUTS_COMPLETE.md",
    "PHASE_7B_QUICK_START.md"
)
foreach ($file in $payoutFiles) {
    if (Copy-Doc $file "CONTEXT/6-MIGRATIONS/creator-payouts/$file") { $copied++ } else { $skipped++ }
}

# Admin
Write-Host "  Admin migrations..."
$adminFiles = @(
    "ADMIN_ROUTES_MIGRATION_PLAN.md",
    "ADMIN_ROUTES_STATUS.md",
    "ADMIN_DASHBOARD_CLEANUP_PLAN.md",
    "ADMIN_DASHBOARD_SESSION_STATUS.md",
    "ADMIN_PANEL_README.md"
)
foreach ($file in $adminFiles) {
    if (Copy-Doc $file "CONTEXT/6-MIGRATIONS/admin/$file") { $copied++ } else { $skipped++ }
}

# Braid migrations
Write-Host "  Braid structure migrations..."
$braidMigFiles = @(
    "BRAID_MIGRATION_COMPLETE.md",
    "BRAID_MIGRATION_STATUS.md",
    "FINAL_MIGRATION_STATUS.md",
    "IMPORT_MAPPING.md",
    "IMPORT_PATH_FIX.md"
)
foreach ($file in $braidMigFiles) {
    if (Copy-Doc $file "CONTEXT/6-MIGRATIONS/braids/$file") { $copied++ } else { $skipped++ }
}

# ========================================
# 7. PHASES (Milestones)
# ========================================
Write-Host "`n🎯 7. Organizing Phase/Milestone Documentation..." -ForegroundColor Cyan
$phaseFiles = @(
    "50_PERCENT_MILESTONE.md",
    "90-PERCENT-STATUS.md",
    "99-PERCENT-STATUS.md",
    "100-PERCENT-FINAL.md",
    "🎊_100_PERCENT_COMPLETE.md",
    "100_PERCENT_COMPLETE_FINAL_REPORT.md",
    "🎊_TODAY_ACHIEVEMENTS.md",
    "TODAYS_ACCOMPLISHMENTS.md",
    "PHASE_6_COMPLETE.md"
)
foreach ($file in $phaseFiles) {
    if (Copy-Doc $file "CONTEXT/7-PHASES/$file") { $copied++ } else { $skipped++ }
}

# ========================================
# 8. STATUS (Reports)
# ========================================
Write-Host "`n📊 8. Organizing Status Reports..." -ForegroundColor Cyan
$statusFiles = @(
    "PRODUCTION_READINESS_REPORT.md",
    "FINAL_STATUS_REPORT.md",
    "BUGS_FIXED_SUMMARY.md",
    "ERROR_HANDLING_IMPROVEMENTS.md",
    "PROJECT_TASK_LIST.md",
    "FRONTEND_BACKEND_FIXES.md",
    "ROUTING_FIXES_COMPLETE.md",
    "REVERT_STREAMING_FIXES.md"
)
foreach ($file in $statusFiles) {
    if (Copy-Doc $file "CONTEXT/8-STATUS/$file") { $copied++ } else { $skipped++ }
}

# ========================================
# 9. BRAIDS (Individual braid docs)
# ========================================
Write-Host "`n🧬 9. Organizing Individual Braid Documentation..." -ForegroundColor Cyan

# Authentication
$authFiles = @(
    "AUTH_BRAID_COMPLETE_REPORT.md",
    "AUTHENTICATION_BRAID_COMPLETE.md",
    "AUTHENTICATION_BRAID_COMBING.md",
    "AUTHENTICATION_DEBUG_PLAN.md",
    "AUTHENTICATION_IMPLEMENTATION.md"
)
foreach ($file in $authFiles) {
    if (Copy-Doc $file "CONTEXT/9-BRAIDS/authentication/$file") { $copied++ } else { $skipped++ }
}

# Subscription
$subBraidFiles = @(
    "SUBSCRIPTION_BILLING_BRAID_COMPLETE.md",
    "SUBSCRIPTION_BRAID_COMPLETE.md"
)
foreach ($file in $subBraidFiles) {
    if (Copy-Doc $file "CONTEXT/9-BRAIDS/subscription/$file") { $copied++ } else { $skipped++ }
}

# Other braids
if (Copy-Doc "USER_MANAGEMENT_BRAID_COMPLETE.md" "CONTEXT/9-BRAIDS/user-management/USER_MANAGEMENT_BRAID_COMPLETE.md") { $copied++ } else { $skipped++ }
if (Copy-Doc "CONTENT_MANAGEMENT_BRAID_COMPLETE.md" "CONTEXT/9-BRAIDS/content/CONTENT_MANAGEMENT_BRAID_COMPLETE.md") { $copied++ } else { $skipped++ }
if (Copy-Doc "COMMUNICATION_BRAID_COMPLETE.md" "CONTEXT/9-BRAIDS/communication/COMMUNICATION_BRAID_COMPLETE.md") { $copied++ } else { $skipped++ }
if (Copy-Doc "ADVERTISEMENT_SYSTEM_BRAID_COMPLETE.md" "CONTEXT/9-BRAIDS/advertisement/ADVERTISEMENT_SYSTEM_BRAID_COMPLETE.md") { $copied++ } else { $skipped++ }

# Split-End Trackers
Write-Host "  Split-end documentation..."
$splitEndFiles = @(
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
    "SPLIT_ENDS_REPAIRED.md"
)
foreach ($file in $splitEndFiles) {
    if ($file -like "*REPAIR*") {
        $braid = if ($file -match "_([A-Za-z]+)Braid_") { $matches[1].ToLower() } else { "general" }
        if (Copy-Doc $file "CONTEXT/9-BRAIDS/$braid/split-ends/$file") { $copied++ } else { $skipped++ }
    } elseif ($file -like "*TRACKER*") {
        $braid = if ($file -match "_([A-Za-z]+)Braid\.md") { $matches[1].ToLower() } else { "general" }
        if (Copy-Doc $file "CONTEXT/9-BRAIDS/$braid/split-ends/$file") { $copied++ } else { $skipped++ }
    } else {
        if (Copy-Doc $file "CONTEXT/9-BRAIDS/$file") { $copied++ } else { $skipped++ }
    }
}

# ========================================
# 10. FEATURES
# ========================================
Write-Host "`n✨ 10. Organizing Feature Documentation..." -ForegroundColor Cyan
$featureFiles = @(
    "PASSWORD_CHANGE_FEATURE.md",
    "WEBSOCKET_REALTIME_COMPLETE.md",
    "STREAMING_OPTIMIZATION_SUMMARY.md",
    "STREAMING_SYSTEM_OPTIMIZATION_GUIDE.md",
    "ADVERTISER_WORKFLOW_TEST.md",
    "BRAND_EXTRACTION_STATUS.md",
    "BRAND_INFRASTRUCTURE_COMPLETE.md",
    "TOGGLE_FIX_COMPLETE.md",
    "MISSION_3_FULLSTACK_INTEGRATION.md"
)
foreach ($file in $featureFiles) {
    if (Copy-Doc $file "CONTEXT/10-FEATURES/$file") { $copied++ } else { $skipped++ }
}

# ========================================
# 11. GUIDES
# ========================================
Write-Host "`n📖 11. Organizing User Guides..." -ForegroundColor Cyan
$guideFiles = @(
    "TEST_ACCOUNTS.md",
    "TEST_USERS_GUIDE.md",
    "Taskfile.md"
)
foreach ($file in $guideFiles) {
    if (Copy-Doc $file "CONTEXT/11-GUIDES/$file") { $copied++ } else { $skipped++ }
}

# ========================================
# SUMMARY
# ========================================
Write-Host "`n" -NoNewline
Write-Host "========================================" -ForegroundColor Green
Write-Host "✅ ORGANIZATION COMPLETE!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "📊 Summary:" -ForegroundColor Cyan
Write-Host "  ✅ Files copied: $copied" -ForegroundColor Green
Write-Host "  ⚠️  Files skipped: $skipped" -ForegroundColor Yellow
Write-Host ""
Write-Host "📁 New Structure:" -ForegroundColor Cyan
Write-Host "  CONTEXT/" -ForegroundColor White
Write-Host "  ├── 1-ARCHITECTURE/      ($($archFiles.Count) files)" -ForegroundColor Gray
Write-Host "  ├── 2-DATABASE/          ($($dbFiles.Count) files)" -ForegroundColor Gray
Write-Host "  ├── 3-TESTING/           ($($testFiles.Count) files)" -ForegroundColor Gray
Write-Host "  ├── 4-FRONTEND/          ($($frontendFiles.Count) files)" -ForegroundColor Gray
Write-Host "  ├── 5-DEPLOYMENT/        ($($deployFiles.Count + 2) files)" -ForegroundColor Gray
Write-Host "  ├── 6-MIGRATIONS/        (Organized by feature)" -ForegroundColor Gray
Write-Host "  ├── 7-PHASES/            ($($phaseFiles.Count) files)" -ForegroundColor Gray
Write-Host "  ├── 8-STATUS/            ($($statusFiles.Count) files)" -ForegroundColor Gray
Write-Host "  ├── 9-BRAIDS/            (Organized by braid)" -ForegroundColor Gray
Write-Host "  ├── 10-FEATURES/         ($($featureFiles.Count) files)" -ForegroundColor Gray
Write-Host "  └── 11-GUIDES/           ($($guideFiles.Count) files)" -ForegroundColor Gray
Write-Host ""
Write-Host "🎉 Your documentation is now beautifully organized!" -ForegroundColor Green
Write-Host "👀 Check CONTEXT/README.md for the master index" -ForegroundColor Cyan
Write-Host ""

