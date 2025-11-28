# Quick Analytics Fix Script
# Run this from: S:\AirEmber\BOME\BOME

Write-Host "🔧 Fixing Video Analytics..." -ForegroundColor Cyan

# Step 1: Run Migration
Write-Host "`n📦 Step 1: Running database migration..." -ForegroundColor Yellow
Set-Location backend
psql -d bome_db -f migrations/067_fix_watch_history_constraints.sql

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Migration completed successfully!" -ForegroundColor Green
} else {
    Write-Host "❌ Migration failed! Check the error above." -ForegroundColor Red
    exit 1
}

# Step 2: Restart Backend
Write-Host "`n🚀 Step 2: Backend ready to restart..." -ForegroundColor Yellow
Write-Host "   Stop your current backend (Ctrl+C)" -ForegroundColor White
Write-Host "   Then run: go run main.go" -ForegroundColor White

# Step 3: Frontend
Write-Host "`n🌐 Step 3: Refresh your browser..." -ForegroundColor Yellow
Write-Host "   Press Ctrl+Shift+R (hard refresh)" -ForegroundColor White

# Step 4: Test
Write-Host "`n🎯 Step 4: Test analytics..." -ForegroundColor Yellow
Write-Host "   Watch a video for 30+ seconds" -ForegroundColor White
Write-Host "   Check backend logs - should see:" -ForegroundColor White
Write-Host "     ✅ [BUFFER←REDIS] Event pushed to Redis successfully" -ForegroundColor Green

Write-Host "`n✨ Ready to go! Restart backend and test." -ForegroundColor Cyan

