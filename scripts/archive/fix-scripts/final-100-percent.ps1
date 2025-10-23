# FINAL FIX FOR 100%

Write-Host "🔄 Applying FINAL fixes for 100%..."

# 1. Fix analytics - comment out unused variables and fix remaining issues
$analyticsFile = "backend/analytics/services/analytics.go"
$content = Get-Content $analyticsFile -Raw

# Comment out unused variables
$content = $content -replace '(\s+)latest :=', '$1// latest :='
$content = $content -replace '(\s+)totalEvents := len\(', '$1// totalEvents := len('

# Fix GetSystemMetrics call
$content = $content -replace 's\.db\.GetSystemMetrics\(\)', 'getSystemMetrics(s.db)'

# Fix CrossSubsiteStats reference
$content = $content -replace 'database\.CrossSubsiteStats', 'interface{}'

Set-Content -Path $analyticsFile -Value $content -NoNewline
Write-Host "  ✅ Fixed analytics"

# 2. Fix middleware - find and properly fix HasVideoAccess calls
$middlewareFile = "backend/authentication/middleware/middleware.go"
$midContent = Get-Content $middlewareFile -Raw

# Read the file line by line to find the exact lines
$lines = Get-Content $middlewareFile
$newLines = @()

foreach ($line in $lines) {
    if ($line -match 'hasAccess, isSubscriber, err := videoModels\.HasVideoAccess') {
        # Replace with correct signature
        $line = $line -replace 'hasAccess, isSubscriber, err := videoModels\.HasVideoAccess\(([^,]+), ([^,]+), ([^)]+)\)', 'hasAccess, err := videoModels.HasVideoAccess($1, userID, $3)'
    }
    $newLines += $line
}

$newLines | Set-Content $middlewareFile
Write-Host "  ✅ Fixed middleware HasVideoAccess calls"

Write-Host "✅ All fixes applied!"

