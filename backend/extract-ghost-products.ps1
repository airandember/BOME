# Extract Ghost Products from Backend Logs
# Purpose: Parse all "GHOST BLOCKED" messages and create a comprehensive report

Write-Host "🔍 Extracting Ghost Products from Backend Logs..." -ForegroundColor Cyan
Write-Host ""

# Find all ghost product references
$ghostProducts = @{}
$ghostSubscriptions = @()

# Read from console output or log file
Write-Host "📋 Parsing logs for GHOST BLOCKED messages..." -ForegroundColor Yellow

# Sample data from your logs (you can expand this)
$logLines = @(
    "👻 GHOST BLOCKED: Subscription sub_I4NzA0pt1W6zE6 references ghost product prod_HjYKGcWGP9r4EC - REJECTED",
    "👻 GHOST BLOCKED: Subscription sub_I4NRvQ7ipOAv6u references ghost product prod_HjYKGcWGP9r4EC - REJECTED",
    "👻 GHOST BLOCKED: Subscription sub_I4Mw5n6bF2TXFo references ghost product prod_FvNAlEGGL452nN - REJECTED",
    "👻 GHOST BLOCKED: Subscription sub_I4MlWzCVdbM3Xm references ghost product prod_HjYKGcWGP9r4EC - REJECTED",
    "👻 GHOST BLOCKED: Subscription sub_I4MBSu7cnwHgYz references ghost product prod_HjYKGcWGP9r4EC - REJECTED",
    "👻 GHOST BLOCKED: Subscription sub_I4M3OOhOBvSDBj references ghost product prod_HjYKGcWGP9r4EC - REJECTED",
    "👻 GHOST BLOCKED: Subscription sub_I4LLecS0F7utEO references ghost product prod_FvNAlEGGL452nN - REJECTED",
    "👻 GHOST BLOCKED: Subscription sub_I4BZzxF3Jy5ypK references ghost product prod_HjYKGcWGP9r4EC - REJECTED",
    "👻 GHOST BLOCKED: Subscription sub_I4BYcp19EJqf1t references ghost product prod_HjYKGcWGP9r4EC - REJECTED",
    "👻 GHOST BLOCKED: Subscription sub_I47sMx05Yi1wsV references ghost product prod_HjYKGcWGP9r4EC - REJECTED",
    "👻 GHOST BLOCKED: Subscription sub_I46Mwz4iNse7ye references ghost product prod_HjYKGcWGP9r4EC - REJECTED",
    "👻 GHOST BLOCKED: Subscription sub_I469HoMvsnFaa6 references ghost product prod_HjYKGcWGP9r4EC - REJECTED"
)

foreach ($line in $logLines) {
    if ($line -match "sub_([A-Za-z0-9]+).*prod_([A-Za-z0-9]+)") {
        $subID = "sub_$($matches[1])"
        $prodID = "prod_$($matches[2])"
        
        # Track product
        if (-not $ghostProducts.ContainsKey($prodID)) {
            $ghostProducts[$prodID] = @()
        }
        $ghostProducts[$prodID] += $subID
        
        # Track subscription
        $ghostSubscriptions += @{
            subscription_id = $subID
            product_id = $prodID
        }
    }
}

Write-Host "✅ Found $($ghostProducts.Count) unique ghost products" -ForegroundColor Green
Write-Host "✅ Found $($ghostSubscriptions.Count) ghost subscriptions" -ForegroundColor Green
Write-Host ""

# Generate Report
$report = @"
# 👻 Ghost Products Report
Generated: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

## 📊 Summary
- Unique Ghost Products: $($ghostProducts.Count)
- Total Ghost Subscriptions: $($ghostSubscriptions.Count)

## 🔍 Ghost Products Breakdown

"@

foreach ($prodID in $ghostProducts.Keys | Sort-Object) {
    $subs = $ghostProducts[$prodID]
    $report += @"

### Product: $prodID
- Blocked Subscriptions: $($subs.Count)
- Subscription IDs:
"@
    foreach ($sub in $subs) {
        $report += "`n  - $sub"
    }
    $report += "`n"
}

$report += @"

## 📋 All Ghost Subscriptions (CSV Format)

subscription_id,product_id
"@

foreach ($sub in $ghostSubscriptions) {
    $report += "`n$($sub.subscription_id),$($sub.product_id)"
}

$report += @"


## 🎯 Recommended SQL to Create Placeholders

-- Insert placeholder products for historical data
INSERT INTO stripe_products_v2 
(stripe_id, name, description, active, video_approved, stripe_created_at, stripe_updated_at)
VALUES
"@

$first = $true
foreach ($prodID in $ghostProducts.Keys | Sort-Object) {
    if (-not $first) { $report += "," }
    $first = $false
    $report += "`n('$prodID', 'Legacy Product (Historical)', 'Historical product from 2020-2021 - discontinued', false, false, '2020-01-01', NOW())"
}

$report += ";"

$report += @"


## 🔧 Recommended Code Changes

Remove these from ghost blocklist in:
- backend/internal/services/stripe_sync.go
- backend/services/payment/stripe/stripe_sync.go
- backend/subscription/services/stripe_sync.go

```go
// REMOVE THESE LINES:
ghostProducts := map[string]bool{
"@

foreach ($prodID in $ghostProducts.Keys | Sort-Object) {
    $report += "`n    `"$prodID`": true,  // Historical product - now has placeholder"
}

$report += @"

}
```

## ✅ Next Steps

1. Review this report with team
2. Decide: Preserve historical data or keep blocking?
3. If preserving:
   - Run SQL to create placeholders
   - Update code to remove from blocklist
   - Re-run Stripe sync
4. If keeping blocklist:
   - Document decision
   - Accept loss of pre-2024 data

---
**Note:** All these are canceled subscriptions from 2020-2021.
They don't affect current operations but are valuable for historical reporting.
"@

# Save report
$report | Out-File -FilePath "ghost-products-report.txt" -Encoding UTF8

Write-Host "═══════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "✅ Report Generated!" -ForegroundColor Green
Write-Host "═══════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""
Write-Host "📄 Report saved to: ghost-products-report.txt" -ForegroundColor Yellow
Write-Host ""
Write-Host "📊 Quick Stats:" -ForegroundColor Cyan
foreach ($prodID in $ghostProducts.Keys | Sort-Object) {
    Write-Host "   $prodID : $($ghostProducts[$prodID].Count) subscriptions" -ForegroundColor White
}
Write-Host ""
Write-Host "💡 Next: Review report and decide on historical data handling" -ForegroundColor Yellow

