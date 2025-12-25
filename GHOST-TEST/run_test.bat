@echo off
echo ========================================
echo   STRIPE GHOST PRODUCT TEST
echo ========================================
echo.

REM Check if Go is installed
where go >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo ERROR: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    pause
    exit /b 1
)

cd /d "%~dp0"

echo Running ghost test...
echo.

REM Run the Go program
go run ghost_test.go

echo.
echo Test complete!
pause

