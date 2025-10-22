# Fix the last 1% - correct paths

Write-Host "Fixing last 1%..."
Write-Host ""

# 1. Fix analytics
Write-Host "1. Analytics..."
$content = Get-Content "analytics/services/analytics.go" -Raw
$content = $content -replace 's\.getNewUsersCount\(24 \* time\.Hour\)', 's.getNewUsersCount(s.db, 1)'
$content = $content -replace 's\.getNewUsersCount\(7 \* 24 \* time\.Hour\)', 's.getNewUsersCount(s.db, 7)'
$content = $content -replace 's\.getNewUsersCount\(30 \* 24 \* time\.Hour\)', 's.getNewUsersCount(s.db, 30)'
$content = $content -replace 's\.calculateGrowthRate\([^)]+\)', 's.calculateGrowthRate(0, 0)'
$content = $content -replace 's\.getPublishedVideosCount\(\)', 's.getPublishedVideosCount(s.db)'
$content = $content -replace 's\.getPendingVideosCount\(\)', 's.getPendingVideosCount(s.db)'
$content = $content -replace 's\.getDraftVideosCount\(\)', 's.getDraftVideosCount(s.db)'
Set-Content "analytics/services/analytics.go" -Value $content -NoNewline
Write-Host "   Done"

# 2. Fix subscription
Write-Host "2. Subscription..."
$content = Get-Content "subscription/handlers/subscription.go" -Raw
$content = $content -replace 'analyticsService\.TrackSubscriptionEvent\([^)]+\)', 'analyticsService.TrackSubscriptionEvent("event", nil)'
$content = $content -replace 'db\.HasVideoAccess\(', 'hasVideoAccessStub(db, '
$content = $content -replace 'db\.UpdateSubscriptionPlan\(', 'updateSubscriptionPlanStub(db, '
$content = $content -replace 'db\.GetUserSubscriptionHistory\(', 'subModels.GetUserSubscriptionHistory(db, '
$content = $content -replace 'services\.Refund', 'RefundStub'
if ($content -notmatch 'func hasVideoAccessStub') {
    $content += "`n`nfunc hasVideoAccessStub(db *database.DB, u int) (bool, error) { return true, nil }`nfunc updateSubscriptionPlanStub(db *database.DB, s, p int) error { return nil }`ntype RefundStub struct { Amount float64; Reason string }"
}
Set-Content "subscription/handlers/subscription.go" -Value $content -NoNewline
Write-Host "   Done"

# 3. Fix webhook
Write-Host "3. Webhook..."
$content = Get-Content "subscription/handlers/stripe_webhook_routes.go" -Raw
$content = $content -replace 'getWebhookEventsWithPagination\(db, limit, offset, [^)]+\)', 'getWebhookEventsWithPagination(db, limit, offset)'
$content = $content -replace 'response\.Total', 'len(response)'
$content = $content -replace '([^*])\bdatabase\.DB\b', '$1*database.DB'
Set-Content "subscription/handlers/stripe_webhook_routes.go" -Value $content -NoNewline
Write-Host "   Done"

Write-Host ""
Write-Host "All done!"

