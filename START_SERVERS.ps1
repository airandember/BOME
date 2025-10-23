# 🚀 BOME Creator Payouts - Quick Start Script
# This script starts both backend and frontend servers

Write-Host "🚀 Starting BOME Creator Payouts Servers..." -ForegroundColor Cyan
Write-Host ""

# Check if we're in the right directory
if (-not (Test-Path "backend") -or -not (Test-Path "frontend")) {
    Write-Host "❌ Error: Must be run from BOME root directory (S:\AirEmber\BOME\BOME)" -ForegroundColor Red
    Write-Host "   Current directory: $(Get-Location)" -ForegroundColor Yellow
    exit 1
}

# Start Backend
Write-Host "📦 Starting Backend Server..." -ForegroundColor Yellow
if (Test-Path "backend/bin/bome-backend-phase7.exe") {
    Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd backend; Write-Host '🔥 Backend Server Starting...' -ForegroundColor Green; ./bin/bome-backend-phase7.exe"
    Write-Host "✅ Backend server starting on http://localhost:8080" -ForegroundColor Green
} else {
    Write-Host "❌ Backend executable not found! Please build first:" -ForegroundColor Red
    Write-Host "   cd backend && go build -o bin/bome-backend-phase7.exe ." -ForegroundColor Yellow
    exit 1
}

Start-Sleep -Seconds 2

# Start Frontend
Write-Host "🎨 Starting Frontend Server..." -ForegroundColor Yellow
if (Test-Path "frontend/package.json") {
    Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd frontend; Write-Host '🎨 Frontend Server Starting...' -ForegroundColor Green; npm run dev"
    Write-Host "✅ Frontend server starting on http://localhost:5173" -ForegroundColor Green
} else {
    Write-Host "❌ Frontend not found!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "═══════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "✨ SERVERS STARTED SUCCESSFULLY! ✨" -ForegroundColor Green
Write-Host "═══════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""
Write-Host "📍 URLs:" -ForegroundColor White
Write-Host "   Backend:  http://localhost:8080" -ForegroundColor Yellow
Write-Host "   Frontend: http://localhost:5173" -ForegroundColor Yellow
Write-Host ""
Write-Host "💰 Creator Payouts Dashboard:" -ForegroundColor White
Write-Host "   http://localhost:5173/admin/streaming/creator-payouts" -ForegroundColor Cyan
Write-Host ""
Write-Host "Press Ctrl+C in each terminal to stop the servers" -ForegroundColor Gray
Write-Host ""

