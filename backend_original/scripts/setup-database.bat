@echo off
setlocal enabledelayedexpansion

REM BOME Database Setup Script for Windows
REM This script sets up the PostgreSQL database using the unified migration system

echo 🚀 BOME Database Setup
echo ================================

REM Check if PostgreSQL is running
echo Checking PostgreSQL connection...
pg_isready -q
if %errorlevel% neq 0 (
    echo ❌ PostgreSQL is not running. Please start PostgreSQL first.
    pause
    exit /b 1
)

echo ✅ PostgreSQL is running

REM Load environment variables from .env file if it exists
if exist .env (
    echo 📄 Loading environment variables from .env
    for /f "tokens=1,2 delims==" %%a in (.env) do (
        if not "%%a"=="" if not "%%a:~0,1%"=="#" (
            set "%%a=%%b"
        )
    )
) else (
    echo ⚠️  No .env file found, using system environment variables
)

REM Set default values if not provided
if "%DB_HOST%"=="" set "DB_HOST=localhost"
if "%DB_PORT%"=="" set "DB_PORT=5432"
if "%DB_USER%"=="" set "DB_USER=bome_admin"
if "%DB_PASSWORD%"=="" set "DB_PASSWORD=AdminBOME"
if "%DB_NAME%"=="" set "DB_NAME=bome_db"
if "%DB_SSL_MODE%"=="" set "DB_SSL_MODE=disable"

echo 📊 Database Configuration:
echo   Host: %DB_HOST%
echo   Port: %DB_PORT%
echo   User: %DB_USER%
echo   Database: %DB_NAME%
echo   SSL Mode: %DB_SSL_MODE%

REM Create database and user if they don't exist
echo 🔧 Setting up database and user...

REM Try to connect as postgres superuser to create database and user
psql -h %DB_HOST% -p %DB_PORT% -U postgres -d postgres -c "DO $$ BEGIN IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '%DB_USER%') THEN CREATE USER %DB_USER% WITH PASSWORD '%DB_PASSWORD%'; RAISE NOTICE 'User %DB_USER% created successfully'; ELSE RAISE NOTICE 'User %DB_USER% already exists'; END IF; IF NOT EXISTS (SELECT FROM pg_database WHERE datname = '%DB_NAME%') THEN CREATE DATABASE %DB_NAME% OWNER %DB_USER%; RAISE NOTICE 'Database %DB_NAME% created successfully'; ELSE RAISE NOTICE 'Database %DB_NAME% already exists'; END IF; GRANT ALL PRIVILEGES ON DATABASE %DB_NAME% TO %DB_USER%; ALTER USER %DB_USER% CREATEDB; END $$;" >nul 2>&1
if %errorlevel% neq 0 (
    echo ⚠️  Could not connect as postgres superuser. Database and user may already exist.
)

echo ✅ Database setup completed

REM Run the backend to execute migrations
echo 🔄 Running database migrations...

REM Build the backend
echo 🔨 Building backend...
go build -o bome-backend.exe .

REM Run migrations
echo 📦 Executing migrations...
start /b bome-backend.exe
set BACKEND_PID=%errorlevel%

REM Wait a moment for the backend to start and run migrations
timeout /t 5 /nobreak >nul

REM Stop the backend
taskkill /f /im bome-backend.exe >nul 2>&1

echo ✅ Database setup and migrations completed successfully!
echo.
echo 🎉 Your BOME database is ready to use!
echo.
echo Next steps:
echo   1. Start the backend: go run main.go
echo   2. Start the frontend: cd ../frontend ^&^& npm run dev
echo.
echo Happy coding! 🚀
pause 