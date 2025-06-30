# BOME Database Verification Script
# This script verifies the PostgreSQL database setup

param(
    [string]$DatabaseName = "bome_db",
    [string]$Username = "bome_admin",
    [string]$Password = "AdminBOME",
    [string]$DbHost = "localhost",
    [int]$Port = 5432
)

Write-Host "=== BOME Database Verification ===" -ForegroundColor Green
Write-Host ""

# Function to execute SQL command
function Invoke-PostgreSQL {
    param(
        [string]$Command,
        [string]$Database = $DatabaseName
    )
    
    $env:PGPASSWORD = $Password
    $result = psql -h $DbHost -p $Port -U $Username -d $Database -c $Command 2>&1
    $env:PGPASSWORD = ""
    
    return $result
}

Write-Host "Step 1: Testing database connection..." -ForegroundColor Yellow
try {
    $testQuery = "SELECT version();"
    $result = Invoke-PostgreSQL -Command $testQuery
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Database connection successful" -ForegroundColor Green
        Write-Host "  PostgreSQL version: $result" -ForegroundColor Green
    } else {
        Write-Host "✗ Database connection failed" -ForegroundColor Red
        Write-Host "  Error: $result" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "✗ Error testing connection: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Step 2: Checking table structure..." -ForegroundColor Yellow

$tablesQuery = @"
SELECT 
    table_name,
    (SELECT COUNT(*) FROM information_schema.columns WHERE table_name = t.table_name) as column_count
FROM information_schema.tables t
WHERE table_schema = 'public'
ORDER BY table_name;
"@

$tables = Invoke-PostgreSQL -Command $tablesQuery

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Database tables found:" -ForegroundColor Green
    Write-Host $tables -ForegroundColor White
} else {
    Write-Host "✗ Error checking tables: $tables" -ForegroundColor Red
}

Write-Host ""
Write-Host "Step 3: Checking table counts..." -ForegroundColor Yellow

$countsQuery = @"
SELECT 
    'users' as table_name, COUNT(*) as count FROM users
UNION ALL
SELECT 'videos', COUNT(*) FROM videos
UNION ALL
SELECT 'subscriptions', COUNT(*) FROM subscriptions
UNION ALL
SELECT 'comments', COUNT(*) FROM comments
UNION ALL
SELECT 'likes', COUNT(*) FROM likes
UNION ALL
SELECT 'favorites', COUNT(*) FROM favorites
UNION ALL
SELECT 'user_activity', COUNT(*) FROM user_activity
UNION ALL
SELECT 'admin_logs', COUNT(*) FROM admin_logs
UNION ALL
SELECT 'advertiser_accounts', COUNT(*) FROM advertiser_accounts
UNION ALL
SELECT 'ad_campaigns', COUNT(*) FROM ad_campaigns
UNION ALL
SELECT 'advertisements', COUNT(*) FROM advertisements
UNION ALL
SELECT 'ad_placements', COUNT(*) FROM ad_placements
UNION ALL
SELECT 'ad_schedules', COUNT(*) FROM ad_schedules
UNION ALL
SELECT 'ad_analytics', COUNT(*) FROM ad_analytics
UNION ALL
SELECT 'ad_clicks', COUNT(*) FROM ad_clicks
UNION ALL
SELECT 'ad_impressions', COUNT(*) FROM ad_impressions
UNION ALL
SELECT 'ad_billing', COUNT(*) FROM ad_billing
UNION ALL
SELECT 'ad_audit_log', COUNT(*) FROM ad_audit_log
ORDER BY table_name;
"@

$counts = Invoke-PostgreSQL -Command $countsQuery

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Table record counts:" -ForegroundColor Green
    Write-Host $counts -ForegroundColor White
} else {
    Write-Host "✗ Error checking table counts: $counts" -ForegroundColor Red
}

Write-Host ""
Write-Host "Step 4: Checking indexes..." -ForegroundColor Yellow

$indexesQuery = @"
SELECT 
    indexname,
    tablename,
    indexdef
FROM pg_indexes
WHERE schemaname = 'public'
ORDER BY tablename, indexname;
"@

$indexes = Invoke-PostgreSQL -Command $indexesQuery

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Database indexes found:" -ForegroundColor Green
    Write-Host $indexes -ForegroundColor White
} else {
    Write-Host "✗ Error checking indexes: $indexes" -ForegroundColor Red
}

Write-Host ""
Write-Host "Step 5: Testing sample queries..." -ForegroundColor Yellow

# Test a few sample queries
$sampleQueries = @(
    "SELECT COUNT(*) as user_count FROM users WHERE role = 'admin';",
    "SELECT COUNT(*) as active_ads FROM advertisements WHERE status = 'active';",
    "SELECT COUNT(*) as active_placements FROM ad_placements WHERE is_active = true;"
)

foreach ($query in $sampleQueries) {
    $result = Invoke-PostgreSQL -Command $query
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Query successful: $result" -ForegroundColor Green
    } else {
        Write-Host "✗ Query failed: $result" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "=== Verification Complete ===" -ForegroundColor Green
Write-Host ""
Write-Host "Database connection string:" -ForegroundColor Cyan
Write-Host "postgresql://$Username`:$Password@$DbHost`:$Port/$DatabaseName" -ForegroundColor White
Write-Host ""
Write-Host "If all tests passed, your database is ready for use!" -ForegroundColor Green
Write-Host "" 