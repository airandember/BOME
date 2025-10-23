# BOME Backend Production Startup Script
# This script ensures reliable server startup by managing Go module cache properly

param(
    [string]$Environment = "production",
    [int]$Port = 8080,
    [switch]$Debug = $false
)

# Set script error handling
$ErrorActionPreference = "Stop"

# Colors for output
$Green = "`e[32m"
$Red = "`e[31m"
$Yellow = "`e[33m"
$Blue = "`e[34m"
$Reset = "`e[0m"

function Write-ColorOutput {
    param([string]$Message, [string]$Color = $Reset)
    Write-Host "${Color}${Message}${Reset}"
}

function Test-ServerHealth {
    param([int]$Port, [int]$TimeoutSeconds = 30)
    
    $startTime = Get-Date
    do {
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:$Port/health" -TimeoutSec 5 -UseBasicParsing
            if ($response.StatusCode -eq 200) {
                return $true
            }
        }
        catch {
            # Server not ready yet, continue waiting
        }
        
        Start-Sleep -Seconds 2
        $elapsed = (Get-Date) - $startTime
    } while ($elapsed.TotalSeconds -lt $TimeoutSeconds)
    
    return $false
}

try {
    Write-ColorOutput "🚀 Starting BOME Backend Server..." $Blue
    
    # Get script directory
    $ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
    Set-Location $ScriptDir
    
    # Set production-ready Go module cache location
    $LocalGoModCache = Join-Path $ScriptDir ".gomodcache"
    $env:GOMODCACHE = $LocalGoModCache
    
    Write-ColorOutput "📁 Go module cache: $LocalGoModCache" $Yellow
    
    # Ensure .gomodcache directory exists
    if (-not (Test-Path $LocalGoModCache)) {
        Write-ColorOutput "📂 Creating local Go module cache directory..." $Yellow
        New-Item -ItemType Directory -Path $LocalGoModCache -Force | Out-Null
    }
    
    # Clean and rebuild modules for production reliability
    Write-ColorOutput "🧹 Cleaning Go module cache..." $Yellow
    go clean -modcache 2>$null
    
    Write-ColorOutput "📦 Downloading Go modules..." $Yellow
    go mod download
    
    Write-ColorOutput "🔧 Tidying Go modules..." $Yellow
    go mod tidy
    
    # Verify Go installation and modules
    Write-ColorOutput "✅ Verifying Go installation..." $Green
    $goVersion = go version
    Write-ColorOutput "   $goVersion" $Green
    
    # Set production environment variables
    $env:GO_ENV = $Environment
    $env:PORT = $Port.ToString()
    
    if ($Debug) {
        Write-ColorOutput "🐛 Debug mode enabled" $Yellow
        $env:DEBUG = "true"
    } else {
        Write-ColorOutput "🏭 Production mode enabled" $Green
        $env:DEBUG = "false"
    }
    
    # Start the server
    Write-ColorOutput "🎯 Starting server on port $Port..." $Blue
    
    if ($Debug) {
        # Debug mode - run in foreground
        Write-ColorOutput "🐛 Running in debug mode (foreground)" $Yellow
        go run main.go
    } else {
        # Production mode - run in background
        Write-ColorOutput "🏭 Running in production mode (background)" $Green
        
        # Start server in background
        $job = Start-Job -ScriptBlock {
            param($ScriptDir, $Port)
            Set-Location $ScriptDir
            $env:GOMODCACHE = Join-Path $ScriptDir ".gomodcache"
            $env:GO_ENV = "production"
            $env:PORT = $Port.ToString()
            go run main.go
        } -ArgumentList $ScriptDir, $Port
        
        # Wait for server to start
        Write-ColorOutput "⏳ Waiting for server to start..." $Yellow
        
        if (Test-ServerHealth -Port $Port -TimeoutSeconds 60) {
            Write-ColorOutput "✅ Server started successfully!" $Green
            Write-ColorOutput "🌐 Server running at: http://localhost:$Port" $Green
            Write-ColorOutput "💚 Health check: http://localhost:$Port/health" $Green
            Write-ColorOutput "📊 Analytics: http://localhost:$Port/api/v1/admin/analytics" $Green
            
            # Keep script running and monitor the job
            Write-ColorOutput "📋 Monitoring server process..." $Blue
            Write-ColorOutput "   Press Ctrl+C to stop the server" $Yellow
            
            try {
                while ($job.State -eq "Running") {
                    Start-Sleep -Seconds 5
                    
                    # Check if server is still responding
                    if (-not (Test-ServerHealth -Port $Port -TimeoutSeconds 5)) {
                        Write-ColorOutput "❌ Server health check failed!" $Red
                        break
                    }
                }
            }
            catch {
                Write-ColorOutput "🛑 Server stopped by user" $Yellow
            }
            finally {
                # Clean up
                Write-ColorOutput "🧹 Stopping server..." $Yellow
                Stop-Job $job -PassThru | Remove-Job
                Write-ColorOutput "✅ Server stopped" $Green
            }
        } else {
            Write-ColorOutput "❌ Server failed to start within timeout!" $Red
            Write-ColorOutput "📋 Job output:" $Yellow
            Receive-Job $job
            Stop-Job $job -PassThru | Remove-Job
            exit 1
        }
    }
}
catch {
    Write-ColorOutput "❌ Error starting server: $($_.Exception.Message)" $Red
    Write-ColorOutput "📋 Full error details:" $Yellow
    Write-Host $_.Exception
    exit 1
}
