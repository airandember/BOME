# 📁 COMPLETE CONTEXT FOLDER ORGANIZATION SCRIPT
# Organizes ALL documentation including newly discovered files

Write-Host "📚 Starting COMPLETE CONTEXT organization..." -ForegroundColor Cyan
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

# Function to move file to archive
function Archive-Doc {
    param($Source, $Destination)
    
    if (Test-Path $Source) {
        $destDir = Split-Path -Parent $Destination
        if (-not (Test-Path $destDir)) {
            New-Item -ItemType Directory -Path $destDir -Force | Out-Null
        }
        Copy-Item $Source $Destination -Force
        Write-Host "  📦 $Source → $Destination (archived)" -ForegroundColor Blue
        return $true
    } else {
        return $false
    }
}

$copiedCount = 0
$archivedCount = 0
$skippedCount = 0

# ===== 1. ARCHITECTURE =====
Write-Host "📁 1-ARCHITECTURE..." -ForegroundColor Yellow
if (Copy-Doc "BOME_CONTEXT_STANDARD.md" "CONTEXT/1-ARCHITECTURE/BOME_CONTEXT_STANDARD.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "BOME_BRAIDS_SUMMARY.md" "CONTEXT/1-ARCHITECTURE/BOME_BRAIDS_SUMMARY.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "BOME_NAMING_CONVENTIONS.md" "CONTEXT/1-ARCHITECTURE/BOME_NAMING_CONVENTIONS.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "ARCHITECTURAL_REFINEMENT_PLAN.md" "CONTEXT/1-ARCHITECTURE/ARCHITECTURAL_REFINEMENT_PLAN.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "ARCHITECTURE_COMPARISON.md" "CONTEXT/1-ARCHITECTURE/ARCHITECTURE_COMPARISON.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "AUTHENTICATION_IMPLEMENTATION.md" "CONTEXT/1-ARCHITECTURE/AUTHENTICATION_IMPLEMENTATION.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "BRAIDS_INDEX.md" "CONTEXT/1-ARCHITECTURE/BRAIDS_INDEX.md") { $copiedCount++ } else { $skippedCount++ }

# ===== 2. DATABASE =====
Write-Host "📁 2-DATABASE..." -ForegroundColor Yellow
if (Copy-Doc "DATABASE_SCHEMA.md" "CONTEXT/2-DATABASE/DATABASE_SCHEMA.md") { $copiedCount++ } else { $skippedCount++ }

# ===== 3. TESTING =====
Write-Host "📁 3-TESTING..." -ForegroundColor Yellow
if (Copy-Doc "BRAID_COMBING_STANDARD.md" "CONTEXT/3-TESTING/BRAID_COMBING_STANDARD.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "BRAID_COMB_CHECKLIST.md" "CONTEXT/3-TESTING/BRAID_COMB_CHECKLIST.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "TESTING_STANDARD_COMPLETE.md" "CONTEXT/3-TESTING/TESTING_STANDARD_COMPLETE.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "BRAID_COMBING_METHODOLOGY.md" "CONTEXT/3-TESTING/BRAID_COMBING_METHODOLOGY.md") { $copiedCount++ } else { $skippedCount++ }

# ===== 4. FRONTEND =====
Write-Host "📁 4-FRONTEND..." -ForegroundColor Yellow
if (Copy-Doc "SVELTE5_REACTIVITY_GUIDE.md" "CONTEXT/4-FRONTEND/SVELTE5_REACTIVITY_GUIDE.md") { $copiedCount++ } else { $skippedCount++ }

# ===== 5. DEPLOYMENT =====
Write-Host "📁 5-DEPLOYMENT..." -ForegroundColor Yellow
if (Copy-Doc "PRODUCTION_READINESS_REPORT.md" "CONTEXT/5-DEPLOYMENT/PRODUCTION_READINESS_REPORT.md") { $copiedCount++ } else { $skippedCount++ }

# ===== 6. MIGRATIONS - Videos =====
Write-Host "📁 6-MIGRATIONS/videos/..." -ForegroundColor Yellow
if (Copy-Doc "VIDEOS_STRAND_ANALYSIS.md" "CONTEXT/6-MIGRATIONS/videos/VIDEOS_STRAND_ANALYSIS.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "VIDEOS_MIGRATION_COMPLETE.md" "CONTEXT/6-MIGRATIONS/videos/VIDEOS_MIGRATION_COMPLETE.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "VIDEOS_30_ENDPOINTS_LIST.md" "CONTEXT/6-MIGRATIONS/videos/VIDEOS_30_ENDPOINTS_LIST.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "TAGS_CATEGORIES_STRAND_ASSESSMENT.md" "CONTEXT/6-MIGRATIONS/videos/TAGS_CATEGORIES_STRAND_ASSESSMENT.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "VIDEOS_FRONTEND_INTEGRATION_ANALYSIS.md" "CONTEXT/6-MIGRATIONS/videos/VIDEOS_FRONTEND_INTEGRATION_ANALYSIS.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "VIDEOS_INTEGRATION_STATUS.md" "CONTEXT/6-MIGRATIONS/videos/VIDEOS_INTEGRATION_STATUS.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "VIDEOS_COMPLETE_INTEGRATION.md" "CONTEXT/6-MIGRATIONS/videos/VIDEOS_COMPLETE_INTEGRATION.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "VIDEOS_MIGRATION_FINAL_SUMMARY.md" "CONTEXT/6-MIGRATIONS/videos/VIDEOS_MIGRATION_FINAL_SUMMARY.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "BUNNY_STATUS_MAPPING.md" "CONTEXT/6-MIGRATIONS/videos/BUNNY_STATUS_MAPPING.md") { $copiedCount++ } else { $skippedCount++ }

# ===== 6. MIGRATIONS - YouTube =====
Write-Host "📁 6-MIGRATIONS/youtube/..." -ForegroundColor Yellow
if (Copy-Doc "YOUTUBE_STRAND_ANALYSIS.md" "CONTEXT/6-MIGRATIONS/youtube/YOUTUBE_STRAND_ANALYSIS.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "YOUTUBE_MIGRATION_COMPLETE.md" "CONTEXT/6-MIGRATIONS/youtube/YOUTUBE_MIGRATION_COMPLETE.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "YOUTUBE_DATABASE_MIGRATION.md" "CONTEXT/6-MIGRATIONS/youtube/YOUTUBE_DATABASE_MIGRATION.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "YOUTUBE_PHASES_1-3_COMPLETE.md" "CONTEXT/6-MIGRATIONS/youtube/YOUTUBE_PHASES_1-3_COMPLETE.md") { $copiedCount++ } else { $skippedCount++ }

# ===== 6. MIGRATIONS - Stripe =====
Write-Host "📁 6-MIGRATIONS/stripe/..." -ForegroundColor Yellow
if (Copy-Doc "STRIPE_SYNC_MIGRATION_PLAN.md" "CONTEXT/6-MIGRATIONS/stripe/STRIPE_SYNC_MIGRATION_PLAN.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "PHASE_6_COMPLETE.md" "CONTEXT/6-MIGRATIONS/stripe/PHASE_6_COMPLETE.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "STRIPE_QUERY_OPTIMIZATION_PLAN.md" "CONTEXT/6-MIGRATIONS/stripe/STRIPE_QUERY_OPTIMIZATION_PLAN.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "STRIPE_GHOSTS_MIGRATION_GUIDE.md" "CONTEXT/6-MIGRATIONS/stripe/STRIPE_GHOSTS_MIGRATION_GUIDE.md") { $copiedCount++ } else { $skippedCount++ }

# ===== 6. MIGRATIONS - Creator Payouts =====
Write-Host "📁 6-MIGRATIONS/creator-payouts/..." -ForegroundColor Yellow
if (Copy-Doc "PHASE_7_COMPREHENSIVE_PLAN.md" "CONTEXT/6-MIGRATIONS/creator-payouts/PHASE_7_COMPREHENSIVE_PLAN.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "PHASE_7B_CREATOR_PAYOUTS_COMPLETE.md" "CONTEXT/6-MIGRATIONS/creator-payouts/PHASE_7B_CREATOR_PAYOUTS_COMPLETE.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "CREATOR_PAYOUT_MIGRATION_GUIDE.md" "CONTEXT/6-MIGRATIONS/creator-payouts/CREATOR_PAYOUT_MIGRATION_GUIDE.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "PHASE_7_SQL_COMPLETE.md" "CONTEXT/6-MIGRATIONS/creator-payouts/PHASE_7_SQL_COMPLETE.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "PHASE_7B_QUICK_START.md" "CONTEXT/6-MIGRATIONS/creator-payouts/PHASE_7B_QUICK_START.md") { $copiedCount++ } else { $skippedCount++ }

# ===== 6. MIGRATIONS - Subscribers =====
Write-Host "📁 6-MIGRATIONS/subscribers/..." -ForegroundColor Yellow
if (Copy-Doc "SUBSCRIBERS_SUBSCRIPTIONS_MIGRATION.md" "CONTEXT/6-MIGRATIONS/subscribers/SUBSCRIBERS_SUBSCRIPTIONS_MIGRATION.md") { $copiedCount++ } else { $skippedCount++ }

# ===== 6. MIGRATIONS - Admin =====
Write-Host "📁 6-MIGRATIONS/admin/..." -ForegroundColor Yellow
if (Copy-Doc "ADMIN_DASHBOARD_CLEANUP_PLAN.md" "CONTEXT/6-MIGRATIONS/admin/ADMIN_DASHBOARD_CLEANUP_PLAN.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "ADMIN_ROUTES_MIGRATION_PLAN.md" "CONTEXT/6-MIGRATIONS/admin/ADMIN_ROUTES_MIGRATION_PLAN.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "ADMIN_ROUTES_STATUS.md" "CONTEXT/6-MIGRATIONS/admin/ADMIN_ROUTES_STATUS.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "ADMIN_PANEL_README.md" "CONTEXT/6-MIGRATIONS/admin/ADMIN_PANEL_README.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "ADMIN_SIDEBAR_COLLAPSE_IMPLEMENTATION.md" "CONTEXT/6-MIGRATIONS/admin/ADMIN_SIDEBAR_COLLAPSE_IMPLEMENTATION.md") { $copiedCount++ } else { $skippedCount++ }

# ===== 6. MIGRATIONS - General =====
Write-Host "📁 6-MIGRATIONS/ (general)..." -ForegroundColor Yellow
if (Copy-Doc "BRAID_MIGRATION_COMPLETE.md" "CONTEXT/6-MIGRATIONS/BRAID_MIGRATION_COMPLETE.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "BRAID_MIGRATION_STATUS.md" "CONTEXT/6-MIGRATIONS/BRAID_MIGRATION_STATUS.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "BUGS_FIXED_SUMMARY.md" "CONTEXT/6-MIGRATIONS/BUGS_FIXED_SUMMARY.md") { $copiedCount++ } else { $skippedCount++ }

# ===== 7. PHASES =====
Write-Host "📁 7-PHASES..." -ForegroundColor Yellow
if (Copy-Doc "ADMIN_ROUTES_MIGRATION_PLAN.md" "CONTEXT/7-PHASES/ADMIN_ROUTES_MIGRATION_PLAN.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "AUTHENTICATION_BRAID_COMPLETE.md" "CONTEXT/7-PHASES/AUTHENTICATION_BRAID_COMPLETE.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "COMMUNICATION_BRAID_COMPLETE.md" "CONTEXT/7-PHASES/COMMUNICATION_BRAID_COMPLETE.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "CONTENT_MANAGEMENT_BRAID_COMPLETE.md" "CONTEXT/7-PHASES/CONTENT_MANAGEMENT_BRAID_COMPLETE.md") { $copiedCount++ } else { $skippedCount++ }
if (Copy-Doc "ADVERTISEMENT_SYSTEM_BRAID_COMPLETE.md" "CONTEXT/7-PHASES/ADVERTISEMENT_SYSTEM_BRAID_COMPLETE.md") { $copiedCount++ } else { $skippedCount++ }

# ===== 8. STATUS =====
Write-Host "📁 8-STATUS..." -ForegroundColor Yellow
if (Copy-Doc "TODAYS_ACCOMPLISHMENTS.md" "CONTEXT/8-STATUS/TODAYS_ACCOMPLISHMENTS.md") { $copiedCount++ } else { $skippedCount++ }

# ===== ARCHIVE - Milestones =====
Write-Host "📦 ARCHIVE/milestones/..." -ForegroundColor Blue
if (Archive-Doc "50_PERCENT_MILESTONE.md" "ARCHIVE/milestones/50_PERCENT_MILESTONE.md") { $archivedCount++ }
if (Archive-Doc "90-PERCENT-STATUS.md" "ARCHIVE/milestones/90-PERCENT-STATUS.md") { $archivedCount++ }
if (Archive-Doc "99-PERCENT-STATUS.md" "ARCHIVE/milestones/99-PERCENT-STATUS.md") { $archivedCount++ }
if (Archive-Doc "100-PERCENT-FINAL.md" "ARCHIVE/milestones/100-PERCENT-FINAL.md") { $archivedCount++ }
if (Archive-Doc "100_PERCENT_COMPLETE_FINAL_REPORT.md" "ARCHIVE/milestones/100_PERCENT_COMPLETE_FINAL_REPORT.md") { $archivedCount++ }
if (Archive-Doc "🎊_100_PERCENT_COMPLETE.md" "ARCHIVE/milestones/100_PERCENT_COMPLETE.md") { $archivedCount++ }

# ===== ARCHIVE - Session Status =====
Write-Host "📦 ARCHIVE/session-status/..." -ForegroundColor Blue
if (Archive-Doc "🎊_TODAY_ACHIEVEMENTS.md" "ARCHIVE/session-status/TODAY_ACHIEVEMENTS.md") { $archivedCount++ }
if (Archive-Doc "ADMIN_DASHBOARD_SESSION_STATUS.md" "ARCHIVE/session-status/ADMIN_DASHBOARD_SESSION_STATUS.md") { $archivedCount++ }

# ===== ARCHIVE - Troubleshooting =====
Write-Host "📦 ARCHIVE/troubleshooting/..." -ForegroundColor Blue
if (Archive-Doc "AUTH_BRAID_COMPLETE_REPORT.md" "ARCHIVE/troubleshooting/AUTH_BRAID_COMPLETE_REPORT.md") { $archivedCount++ }
if (Archive-Doc "AUTHENTICATION_BRAID_COMBING.md" "ARCHIVE/troubleshooting/AUTHENTICATION_BRAID_COMBING.md") { $archivedCount++ }
if (Archive-Doc "AUTHENTICATION_DEBUG_PLAN.md" "ARCHIVE/troubleshooting/AUTHENTICATION_DEBUG_PLAN.md") { $archivedCount++ }
if (Archive-Doc "CREATOR_PAYOUTS_500_ERROR_FIX.md" "ARCHIVE/troubleshooting/CREATOR_PAYOUTS_500_ERROR_FIX.md") { $archivedCount++ }
if (Archive-Doc "CREATOR_PAYOUTS_NAVIGATION_ADDED.md" "ARCHIVE/troubleshooting/CREATOR_PAYOUTS_NAVIGATION_ADDED.md") { $archivedCount++ }
if (Archive-Doc "ADVERTISER_WORKFLOW_TEST.md" "ARCHIVE/troubleshooting/ADVERTISER_WORKFLOW_TEST.md") { $archivedCount++ }

# ===== ARCHIVE - Brand Work =====
Write-Host "📦 ARCHIVE/brand-work/..." -ForegroundColor Blue
if (Archive-Doc "BRAND_EXTRACTION_STATUS.md" "ARCHIVE/brand-work/BRAND_EXTRACTION_STATUS.md") { $archivedCount++ }
if (Archive-Doc "BRAND_INFRASTRUCTURE_COMPLETE.md" "ARCHIVE/brand-work/BRAND_INFRASTRUCTURE_COMPLETE.md") { $archivedCount++ }

Write-Host ""
Write-Host "═══════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "✨ COMPLETE CONTEXT ORGANIZATION DONE! ✨" -ForegroundColor Green
Write-Host "═══════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""
Write-Host "📊 Summary:" -ForegroundColor White
Write-Host "   ✅ Files copied to CONTEXT: $copiedCount" -ForegroundColor Green
Write-Host "   📦 Files archived: $archivedCount" -ForegroundColor Blue
if ($skippedCount -gt 0) {
    Write-Host "   ⚠️  Files skipped: $skippedCount" -ForegroundColor Yellow
}
Write-Host ""
Write-Host "📁 Documentation structure:" -ForegroundColor White
Write-Host "   CONTEXT/1-ARCHITECTURE/     - Platform architecture (7 files)" -ForegroundColor Cyan
Write-Host "   CONTEXT/2-DATABASE/          - Database schema" -ForegroundColor Cyan
Write-Host "   CONTEXT/3-TESTING/           - Testing methodology (4 files)" -ForegroundColor Cyan
Write-Host "   CONTEXT/4-FRONTEND/          - Frontend patterns" -ForegroundColor Cyan
Write-Host "   CONTEXT/5-DEPLOYMENT/        - Production deployment" -ForegroundColor Cyan
Write-Host "   CONTEXT/6-MIGRATIONS/        - Feature migrations (40+ files)" -ForegroundColor Cyan
Write-Host "   CONTEXT/7-PHASES/            - Phase summaries (5 files)" -ForegroundColor Cyan
Write-Host "   CONTEXT/8-STATUS/            - Current status" -ForegroundColor Cyan
Write-Host ""
Write-Host "📦 Archive structure:" -ForegroundColor White
Write-Host "   ARCHIVE/milestones/          - Historical milestones (6 files)" -ForegroundColor Blue
Write-Host "   ARCHIVE/session-status/      - Session logs (2 files)" -ForegroundColor Blue
Write-Host "   ARCHIVE/troubleshooting/     - Fixed issues (6 files)" -ForegroundColor Blue
Write-Host "   ARCHIVE/brand-work/          - Brand extraction (2 files)" -ForegroundColor Blue
Write-Host ""
Write-Host "📖 Start with: CONTEXT/README.md" -ForegroundColor Yellow
Write-Host ""

