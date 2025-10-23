# Comb Advertisement System Braid

Write-Host ""
Write-Host "ADVERTISEMENT SYSTEM BRAID - COMPLETE COMBING" -ForegroundColor Red
Write-Host "==============================================" -ForegroundColor Red
Write-Host ""

# Check if advertisement directory exists
if (-not (Test-Path "advertisement")) {
    Write-Host "[INFO] No dedicated advertisement/ directory" -ForegroundColor Yellow
    Write-Host "[INFO] Ad system might be in admin or other packages" -ForegroundColor Gray
    Write-Host ""
}

# STRAND 1: Advertiser Onboarding
Write-Host "STRAND 1: Advertiser Onboarding" -ForegroundColor Yellow
Write-Host "--------------------------------" -ForegroundColor Yellow

$hasAdvertiser = Select-String -Path "**/*.go" -Pattern "func.*Advertiser|AdvertiserAccount" -Quiet 2>$null
if ($hasAdvertiser) { 
    Write-Host "  [OK] Advertiser functions found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Advertiser management missing" -ForegroundColor Yellow 
}

Write-Host ""

# STRAND 2: Campaign Management
Write-Host "STRAND 2: Campaign Creation and Management" -ForegroundColor Yellow
Write-Host "-------------------------------------------" -ForegroundColor Yellow

$hasCampaign = Select-String -Path "**/*.go" -Pattern "func.*Campaign|AdCampaign" -Quiet 2>$null
if ($hasCampaign) { 
    Write-Host "  [OK] Campaign functions found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Campaign management missing" -ForegroundColor Red 
}

Write-Host ""

# STRAND 3: Ad Placement
Write-Host "STRAND 3: Advertisement Placement" -ForegroundColor Yellow
Write-Host "---------------------------------" -ForegroundColor Yellow

$hasPlacement = Select-String -Path "**/*.go" -Pattern "func.*Placement|AdPlacement" -Quiet 2>$null
if ($hasPlacement) { 
    Write-Host "  [OK] Ad placement functions found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Ad placement missing" -ForegroundColor Red 
}

Write-Host ""

# STRAND 4: Ad Analytics
Write-Host "STRAND 4: Advertisement Analytics" -ForegroundColor Yellow
Write-Host "---------------------------------" -ForegroundColor Yellow

$hasAdAnalytics = Select-String -Path "**/*.go" -Pattern "ad.*analytics|AdAnalytics|ad.*performance" -Quiet 2>$null
if ($hasAdAnalytics) { 
    Write-Host "  [OK] Ad analytics found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Ad analytics missing" -ForegroundColor Yellow 
}

Write-Host ""

# STRAND 5: Ad Billing
Write-Host "STRAND 5: Billing and Revenue Management" -ForegroundColor Yellow
Write-Host "----------------------------------------" -ForegroundColor Yellow

$hasAdBilling = Select-String -Path "**/*.go" -Pattern "ad.*billing|ad.*revenue" -Quiet 2>$null
if ($hasAdBilling) { 
    Write-Host "  [OK] Ad billing found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Ad billing missing" -ForegroundColor Yellow 
}

Write-Host ""

# Check admin package for advertisement management
Write-Host "CHECKING ADMIN PACKAGE:" -ForegroundColor Cyan
if (Test-Path "admin/handlers/admin-routes.go") {
    $hasAdminAds = Select-String -Path "admin/handlers/admin-routes.go" -Pattern "advertisement|ad_campaign|advertiser" -Quiet
    if ($hasAdminAds) {
        Write-Host "  [OK] Advertisement management found in admin" -ForegroundColor Green
    } else {
        Write-Host "  [INFO] No ad management in admin-routes.go" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "ALL 5 STRANDS COMBED" -ForegroundColor Green
Write-Host ""

