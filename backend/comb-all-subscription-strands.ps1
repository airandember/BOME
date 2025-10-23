# Comb All Subscription Strands

Write-Host ""
Write-Host "╔════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║  SUBSCRIPTION BRAID - COMPLETE COMBING        ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# STRAND 2: Stripe Webhook Handling
Write-Host "STRAND 2: Stripe Webhook Handling" -ForegroundColor Yellow
Write-Host "-----------------------------------" -ForegroundColor Yellow

$hasWebhookHandler = Test-Path "subscription/handlers/stripe_webhook_routes.go"
if ($hasWebhookHandler) { 
    Write-Host "  [OK] Webhook routes file exists" -ForegroundColor Green
    
    # Check for known compilation errors
    Write-Host "  [CHECK] Looking for known issues..." -ForegroundColor Cyan
    $hasGetWebhookEvents = Select-String -Path "subscription/handlers/stripe_webhook_routes.go" -Pattern "getWebhookEventsWithPagination" -Quiet
    if ($hasGetWebhookEvents) {
        Write-Host "  [ISSUE] getWebhookEventsWithPagination call found (has compilation error)" -ForegroundColor Red
    }
} else { 
    Write-Host "  [SPLIT-END] Webhook routes file missing" -ForegroundColor Red 
}

$hasWebhookService = Select-String -Path "subscription/services/*.go" -Pattern "func.*HandleWebhook|func.*ProcessWebhook" -Quiet
if ($hasWebhookService) { 
    Write-Host "  [OK] Webhook processing functions found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Webhook processing missing" -ForegroundColor Red 
}

Write-Host ""

# STRAND 3: Subscription Management
Write-Host "STRAND 3: Subscription Management" -ForegroundColor Yellow
Write-Host "----------------------------------" -ForegroundColor Yellow

$hasUpdateSub = Select-String -Path "subscription/handlers/subscription.go" -Pattern "func.*Update|func.*Cancel|func.*Reactivate" -Quiet
if ($hasUpdateSub) { 
    Write-Host "  [OK] Subscription management handlers found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Subscription management handlers missing" -ForegroundColor Red 
}

# Check for unused variables (known compilation issues)
Write-Host "  [CHECK] Looking for unused variables..." -ForegroundColor Cyan
$unusedVars = Select-String -Path "subscription/handlers/subscription.go" -Pattern "declared and not used"
if ($unusedVars) {
    Write-Host "  [ISSUE] Found unused variables (compilation errors)" -ForegroundColor Red
}

Write-Host ""

# STRAND 4: Billing & Invoicing
Write-Host "STRAND 4: Billing and Invoicing" -ForegroundColor Yellow
Write-Host "--------------------------------" -ForegroundColor Yellow

$hasInvoice = Select-String -Path "subscription/services/*.go" -Pattern "func.*Invoice|func.*Billing" -Quiet
if ($hasInvoice) { 
    Write-Host "  [OK] Invoice/billing functions found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Invoice/billing functions missing" -ForegroundColor Yellow
}

Write-Host ""

# STRAND 5: Subscription Plans
Write-Host "STRAND 5: Subscription Plans" -ForegroundColor Yellow
Write-Host "----------------------------" -ForegroundColor Yellow

$hasPlanModel = Test-Path "subscription/models/*plan*.go"
if ($hasPlanModel) { 
    Write-Host "  [OK] Plan model files exist" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Plan model files missing" -ForegroundColor Red 
}

$hasPlanHandler = Select-String -Path "subscription/handlers/*.go" -Pattern "func.*Plan|func.*GetPlans" -Quiet
if ($hasPlanHandler) { 
    Write-Host "  [OK] Plan handlers found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Plan handlers missing" -ForegroundColor Red 
}

Write-Host ""
Write-Host "ALL 5 STRANDS COMBED" -ForegroundColor Green
Write-Host ""

