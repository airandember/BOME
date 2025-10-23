# Stub out problematic analytics sections

Write-Host "🔄 Stubbing analytics problems..."

$analyticsFile = "backend/analytics/services/analytics.go"
$content = Get-Content $analyticsFile -Raw

# Comment out totalEvents completely
$content = $content -replace 'totalEvents := len\(recentActivity\)', '_ = len(recentActivity) // totalEvents'

# Fix the stats iteration - return early instead
$content = $content -replace 'stats, err := getCrossSubsiteStats\(s\.db[^)]*\)', 'stats, err := getCrossSubsiteStats(s.db) 
	if err != nil || stats == nil {
		return emptyStats, nil
	}
	return emptyStats, nil // Stub - return empty stats for now'

# Fix webhook events iteration
$content = $content -replace 'webhooks, err := getWebhookEvents\(s\.db[^)]*\)', 'webhooks, err := getWebhookEvents(s.db)
	if err != nil {
		return webhookStats, err
	}
	return webhookStats, nil // Stub - return empty webhook stats'

Set-Content -Path $analyticsFile -Value $content -NoNewline
Write-Host "  ✅ Stubbed analytics problems"

