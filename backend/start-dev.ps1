# Development Startup Script
# Loads environment variables and starts backend

Write-Host "`n🚀 BOME Backend - Development Mode`n" -ForegroundColor Cyan

# Load environment variables from .env.development
if (Test-Path ".env.development") {
    Write-Host "📄 Loading environment variables..." -ForegroundColor Yellow
    Get-Content ".env.development" | ForEach-Object {
        if ($_ -notmatch '^\s*#' -and $_ -match '=') {
            $parts = $_ -split '=', 2
            $key = $parts[0].Trim()
            $value = $parts[1].Trim()
            [Environment]::SetEnvironmentVariable($key, $value, "Process")
            Write-Host "   ✓ Set $key" -ForegroundColor Gray
        }
    }
    Write-Host "✅ Environment variables loaded!`n" -ForegroundColor Green
} else {
    Write-Host "⚠️  No .env.development file found`n" -ForegroundColor Yellow
}

# Kill existing backend processes
taskkill /F /IM bome-backend.exe 2>$null | Out-Null
taskkill /F /IM go.exe 2>$null | Out-Null

Write-Host "🔨 Building backend..." -ForegroundColor Yellow
go build -o bome-backend.exe

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Build successful!`n" -ForegroundColor Green
    Write-Host "🚀 Starting backend server...`n" -ForegroundColor Cyan
    .\bome-backend.exe
} else {
    Write-Host "❌ Build failed!`n" -ForegroundColor Red
    exit 1
}

