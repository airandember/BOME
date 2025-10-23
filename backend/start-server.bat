@echo off
REM BOME Backend Production Startup Script (Windows Batch)
REM This script ensures reliable server startup by managing Go module cache properly

setlocal enabledelayedexpansion

REM Set default values
set ENVIRONMENT=production
set PORT=8080
set DEBUG=false

REM Parse command line arguments
:parse_args
if "%~1"=="" goto start_server
if "%~1"=="--debug" set DEBUG=true
if "%~1"=="--port" (
    set PORT=%~2
    shift
)
if "%~1"=="--env" (
    set ENVIRONMENT=%~2
    shift
)
shift
goto parse_args

:start_server
echo 🚀 Starting BOME Backend Server...

REM Get script directory
set SCRIPT_DIR=%~dp0
cd /d "%SCRIPT_DIR%"

REM Set production-ready Go module cache location
set LOCAL_GOMODCACHE=%SCRIPT_DIR%.gomodcache
set GOMODCACHE=%LOCAL_GOMODCACHE%

echo 📁 Go module cache: %LOCAL_GOMODCACHE%

REM Ensure .gomodcache directory exists
if not exist "%LOCAL_GOMODCACHE%" (
    echo 📂 Creating local Go module cache directory...
    mkdir "%LOCAL_GOMODCACHE%"
)

REM Clean and rebuild modules for production reliability
echo 🧹 Cleaning Go module cache...
go clean -modcache >nul 2>&1

echo 📦 Downloading Go modules...
go mod download
if errorlevel 1 (
    echo ❌ Failed to download Go modules
    exit /b 1
)

echo 🔧 Tidying Go modules...
go mod tidy
if errorlevel 1 (
    echo ❌ Failed to tidy Go modules
    exit /b 1
)

REM Verify Go installation
echo ✅ Verifying Go installation...
go version
if errorlevel 1 (
    echo ❌ Go is not installed or not in PATH
    exit /b 1
)

REM Set production environment variables
set GO_ENV=%ENVIRONMENT%
set PORT=%PORT%

if "%DEBUG%"=="true" (
    echo 🐛 Debug mode enabled
    set DEBUG=true
) else (
    echo 🏭 Production mode enabled
    set DEBUG=false
)

REM Start the server
echo 🎯 Starting server on port %PORT%...

if "%DEBUG%"=="true" (
    echo 🐛 Running in debug mode (foreground)
    go run main.go
) else (
    echo 🏭 Running in production mode
    go run main.go
)

if errorlevel 1 (
    echo ❌ Server failed to start
    exit /b 1
)

echo ✅ Server stopped
