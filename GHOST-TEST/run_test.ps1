# Ghost Product Test Runner
# Run this script to test if your known ghost products are truly deleted or just archived

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  STRIPE GHOST PRODUCT TEST" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if Go is installed
$goInstalled = Get-Command go -ErrorAction SilentlyContinue
if (-not $goInstalled) {
    Write-Host "ERROR: Go is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please install Go from https://golang.org/dl/" -ForegroundColor Yellow
    exit 1
}

# Navigate to script directory
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptDir

Write-Host "Running ghost test..." -ForegroundColor Green
Write-Host ""

# Run the Go program
# You can set STRIPE_SECRET_KEY environment variable or enter it when prompted
go run ghost_test.go

Write-Host ""
Write-Host "Test complete!" -ForegroundColor Green

