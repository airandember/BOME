# Comb Subscription Creation Strand

Write-Host ""
Write-Host "COMBING: Subscription Creation Strand" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "Checking handler functions..." -ForegroundColor Yellow
$hasCreateSubscription = Select-String -Path "subscription/handlers/subscription.go" -Pattern "func.*CreateSubscription|func.*SubscriptionCheckout" -Quiet
if ($hasCreateSubscription) { 
    Write-Host "  [OK] Subscription creation handler found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Subscription creation handler missing" -ForegroundColor Red 
}

Write-Host ""
Write-Host "Checking Stripe integration..." -ForegroundColor Yellow
$hasStripeCheckout = Select-String -Path "subscription/services/stripe.go" -Pattern "func.*CreateCheckoutSession|func.*CreateSubscription" -Quiet
if ($hasStripeCheckout) { 
    Write-Host "  [OK] Stripe checkout functions found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Stripe checkout functions missing" -ForegroundColor Red 
}

Write-Host ""
Write-Host "Checking model functions..." -ForegroundColor Yellow
$hasSubModel = Test-Path "subscription/models/subscription.go"
if ($hasSubModel) { 
    Write-Host "  [OK] Subscription model exists" -ForegroundColor Green
    $hasCreateFunc = Select-String -Path "subscription/models/subscription.go" -Pattern "func.*CreateSubscription" -Quiet
    if ($hasCreateFunc) {
        Write-Host "  [OK] CreateSubscription model function found" -ForegroundColor Green
    } else {
        Write-Host "  [SPLIT-END] CreateSubscription model function missing" -ForegroundColor Red
    }
} else { 
    Write-Host "  [SPLIT-END] Subscription model file missing" -ForegroundColor Red 
}

Write-Host ""
Write-Host "Checking database schema..." -ForegroundColor Yellow
if (Test-Path "../../backend_original/migrations/*subscription*.sql") {
    Write-Host "  [OK] Subscription table migrations exist" -ForegroundColor Green
} else {
    Write-Host "  [SPLIT-END] Subscription migrations missing" -ForegroundColor Red
}

Write-Host ""
Write-Host "Checking frontend..." -ForegroundColor Yellow
if (Test-Path "../../frontend/src/routes/subscription/+page.svelte") {
    Write-Host "  [OK] Subscription page exists" -ForegroundColor Green
} else {
    Write-Host "  [SPLIT-END] Subscription page missing" -ForegroundColor Red
}

if (Test-Path "../../frontend/src/routes/checkout/+page.svelte") {
    Write-Host "  [OK] Checkout page exists" -ForegroundColor Green
} else {
    Write-Host "  [SPLIT-END] Checkout page missing" -ForegroundColor Red
}

Write-Host ""
Write-Host "Strand 1: Subscription Creation - Combing Complete!" -ForegroundColor Green
Write-Host ""

