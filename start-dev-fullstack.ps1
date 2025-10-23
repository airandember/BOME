# ============================================================================
# 🚀 BOME Full-Stack Development Startup
# ============================================================================
# Starts both backend (Go) and frontend (SvelteKit) in development mode
# ============================================================================

$ErrorActionPreference = "Stop"

Write-Host "`n╔════════════════════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                                                                                ║" -ForegroundColor Cyan
Write-Host "║                     🧬 BOME FULL-STACK DEV STARTUP                             ║" -ForegroundColor Cyan
Write-Host "║                                                                                ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════════════════════════╝`n" -ForegroundColor Cyan

# Check if backend directory exists
if (!(Test-Path "backend")) {
    Write-Host "❌ ERROR: backend directory not found" -ForegroundColor Red
    Write-Host "   Please run this script from the project root directory`n" -ForegroundColor Yellow
    exit 1
}

# Check if frontend directory exists
if (!(Test-Path "frontend")) {
    Write-Host "❌ ERROR: frontend directory not found" -ForegroundColor Red
    Write-Host "   Please run this script from the project root directory`n" -ForegroundColor Yellow
    exit 1
}

# Check if frontend .env.local exists
if (!(Test-Path "frontend/.env.local")) {
    Write-Host "⚠️  WARNING: frontend/.env.local not found" -ForegroundColor Yellow
    Write-Host "   Creating default configuration...`n" -ForegroundColor Yellow
    
    $envContent = @"
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_WS_URL=ws://localhost:8080/ws
VITE_APP_NAME=BOME
VITE_APP_VERSION=1.0.0
VITE_ENVIRONMENT=development
VITE_ENABLE_ANALYTICS=false
VITE_ENABLE_DEBUG=true
"@
    $envContent | Out-File -FilePath "frontend/.env.local" -Encoding UTF8
    Write-Host "✅ Created frontend/.env.local`n" -ForegroundColor Green
}

Write-Host "📋 Starting BOME Full-Stack Development Environment...`n" -ForegroundColor Cyan

# Step 1: Start Backend
Write-Host "🔙 Step 1: Starting Backend (Go + Gin)..." -ForegroundColor Yellow
Write-Host "   Port: 8080" -ForegroundColor Gray
Write-Host "   Starting in separate window...`n" -ForegroundColor Gray

Start-Process powershell -ArgumentList `
    "-NoExit", `
    "-Command", `
    "Write-Host '🔙 BOME Backend' -ForegroundColor Cyan; cd '$PWD\backend'; if (Test-Path 'start-dev.ps1') { .\start-dev.ps1 } else { go run main.go }"

Write-Host "✅ Backend starting in separate window" -ForegroundColor Green
Write-Host "   Check the new window for backend logs`n" -ForegroundColor Gray

# Step 2: Wait for backend to be ready
Write-Host "⏳ Step 2: Waiting for backend to start..." -ForegroundColor Yellow
Write-Host "   Testing connection to http://localhost:8080/health`n" -ForegroundColor Gray

$maxAttempts = 30
$attempt = 0
$backendReady = $false

while ($attempt -lt $maxAttempts -and !$backendReady) {
    $attempt++
    Start-Sleep -Seconds 1
    
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET -TimeoutSec 2 -ErrorAction Stop
        $backendReady = $true
        Write-Host "✅ Backend is ready! ($attempt seconds)" -ForegroundColor Green
    } catch {
        if ($attempt % 5 -eq 0) {
            Write-Host "   Still waiting... ($attempt/$maxAttempts)" -ForegroundColor Gray
        }
    }
}

if (!$backendReady) {
    Write-Host "`n⚠️  WARNING: Backend did not respond within $maxAttempts seconds" -ForegroundColor Yellow
    Write-Host "   Continuing anyway... Check the backend window for errors`n" -ForegroundColor Yellow
} else {
    Write-Host ""
}

# Step 3: Check if node_modules exists
if (!(Test-Path "frontend/node_modules")) {
    Write-Host "📦 Step 3: Installing frontend dependencies..." -ForegroundColor Yellow
    Write-Host "   This may take a few minutes on first run...`n" -ForegroundColor Gray
    
    try {
        cd frontend
        npm install
        cd ..
        Write-Host "`n✅ Dependencies installed`n" -ForegroundColor Green
    } catch {
        Write-Host "`n⚠️  WARNING: npm install had issues" -ForegroundColor Yellow
        Write-Host "   Continuing anyway...`n" -ForegroundColor Yellow
    }
}

# Step 4: Start Frontend
Write-Host "🎨 Step 4: Starting Frontend (SvelteKit + Vite)..." -ForegroundColor Yellow
Write-Host "   Port: Usually 5173 (Vite will assign)" -ForegroundColor Gray
Write-Host "   Opening in this window...`n" -ForegroundColor Gray

Write-Host "╔════════════════════════════════════════════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║                                                                                ║" -ForegroundColor Green
Write-Host "║                   ✅ FULL-STACK DEV ENVIRONMENT READY                           ║" -ForegroundColor Green
Write-Host "║                                                                                ║" -ForegroundColor Green
Write-Host "╠════════════════════════════════════════════════════════════════════════════════╣" -ForegroundColor Green
Write-Host "║  🔙 Backend:  http://localhost:8080                                            ║" -ForegroundColor White
Write-Host "║  🎨 Frontend: http://localhost:5173 (starting below...)                        ║" -ForegroundColor White
Write-Host "║                                                                                ║" -ForegroundColor Green
Write-Host "║  📋 API Docs: http://localhost:8080/api/v1/*                                   ║" -ForegroundColor Cyan
Write-Host "║  🔍 Health:   http://localhost:8080/health                                     ║" -ForegroundColor Cyan
Write-Host "║                                                                                ║" -ForegroundColor Green
Write-Host "║  Press Ctrl+C to stop frontend (backend runs in separate window)              ║" -ForegroundColor Yellow
Write-Host "║                                                                                ║" -ForegroundColor Green
Write-Host "╚════════════════════════════════════════════════════════════════════════════════╝`n" -ForegroundColor Green

# Start frontend in this window
cd frontend
npm run dev

