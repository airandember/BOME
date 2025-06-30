# Test PostgreSQL connection for BOME backend
Write-Host "Testing PostgreSQL connection..." -ForegroundColor Green

# Set environment variables
$env:PGPASSWORD = "AdminBOME"

# Test connection
try {
    $result = psql -U bome_admin -d bome_db -h localhost -p 5432 -c "SELECT version();" 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ PostgreSQL connection successful!" -ForegroundColor Green
        Write-Host "Database: bome_db" -ForegroundColor Cyan
        Write-Host "User: bome_admin" -ForegroundColor Cyan
        Write-Host "Host: localhost:5432" -ForegroundColor Cyan
    } else {
        Write-Host "❌ PostgreSQL connection failed!" -ForegroundColor Red
        Write-Host "Error: $result" -ForegroundColor Red
        Write-Host ""
        Write-Host "Please make sure:" -ForegroundColor Yellow
        Write-Host "1. PostgreSQL is running" -ForegroundColor Yellow
        Write-Host "2. User 'bome_admin' exists" -ForegroundColor Yellow
        Write-Host "3. Database 'bome_db' exists" -ForegroundColor Yellow
        Write-Host "4. Password is correct" -ForegroundColor Yellow
    }
} catch {
    Write-Host "❌ Error testing connection: $_" -ForegroundColor Red
}

# Clear password from environment
Remove-Item Env:PGPASSWORD -ErrorAction SilentlyContinue 