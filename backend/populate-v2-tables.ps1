# Populate V2 Tables - One-Click Script
# Purpose: Sync Stripe data and link users to customers

Write-Host "🚀 Starting V2 Table Population..." -ForegroundColor Cyan
Write-Host ""

# Step 1: Run migration to fix table schema
Write-Host "📊 Step 1: Checking database schema..." -ForegroundColor Yellow
Write-Host "⚠️  Please run migration 050_1_alter_user_stripe_customers_v2.sql first if you haven't!" -ForegroundColor Yellow
Write-Host ""
Read-Host "Press Enter after running the migration (or if already done)"

# Step 2: Sync Stripe data
Write-Host ""
Write-Host "📦 Step 2: Syncing Stripe data to v2 tables..." -ForegroundColor Yellow
Write-Host "   (This will sync products, prices, customers, and subscriptions)" -ForegroundColor Gray
Write-Host ""

Push-Location cmd/stripe-sync
& .\stripe-sync.exe
$syncResult = $LASTEXITCODE
Pop-Location

if ($syncResult -ne 0) {
    Write-Host "❌ Stripe sync failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "✅ Stripe sync complete!" -ForegroundColor Green

# Step 3: Link users to customers
Write-Host ""
Write-Host "🔗 Step 3: Linking users to Stripe customers..." -ForegroundColor Yellow
Write-Host "   (This will match users by email to their Stripe customers)" -ForegroundColor Gray
Write-Host ""

Push-Location cmd/customer-linking
& .\customer-linking.exe --link-all --pretty
$linkResult = $LASTEXITCODE
Pop-Location

if ($linkResult -ne 0) {
    Write-Host "❌ Customer linking failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "✅ Customer linking complete!" -ForegroundColor Green

# Summary
Write-Host ""
Write-Host "═══════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "🎉 V2 Tables Population Complete!" -ForegroundColor Green
Write-Host "═══════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""
Write-Host "✅ Next steps:" -ForegroundColor Yellow
Write-Host "   1. Reload /user/subscriptions in your browser" -ForegroundColor Gray
Write-Host "   2. Check that user 4826 now sees their subscription" -ForegroundColor Gray
Write-Host "   3. Test with other users" -ForegroundColor Gray
Write-Host ""
Write-Host "📊 To verify a specific user:" -ForegroundColor Yellow
Write-Host "   curl http://localhost:8080/api/v1/admin/customer-linking/user/4826 \" -ForegroundColor Gray
Write-Host "     -H 'Authorization: Bearer YOUR_TOKEN'" -ForegroundColor Gray
Write-Host ""

