@echo off
setlocal enabledelayedexpansion

REM BOME PostgreSQL Setup Script for Windows
REM This script helps set up PostgreSQL database for the BOME project

echo.
echo 🚀 BOME PostgreSQL Setup (Windows)
echo ===================================
echo.

REM Default values
set DB_NAME=bome_streaming
set DB_USER=bome_user
set DB_PASSWORD=
set DB_HOST=localhost
set DB_PORT=5432

REM Check if PostgreSQL is installed
where psql >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ PostgreSQL client not found. Please install PostgreSQL first.
    echo   Download from: https://www.postgresql.org/download/windows/
    echo   Or use chocolatey: choco install postgresql
    exit /b 1
)
echo ✅ PostgreSQL client found

REM Check if PostgreSQL server is running
pg_isready -h %DB_HOST% -p %DB_PORT% >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ PostgreSQL server is not running on %DB_HOST%:%DB_PORT%
    echo Please start PostgreSQL server first.
    pause
    exit /b 1
)
echo ✅ PostgreSQL server is running

REM Generate random password if not provided
if "%DB_PASSWORD%"=="" (
    REM Generate a simple random password (Windows compatible)
    set /a password_num=%random% * 99999 / 32768
    set DB_PASSWORD=BomePass!password_num!
    echo ⚠️ Generated password: !DB_PASSWORD!
)

echo.
echo Creating PostgreSQL database and user...

REM Create user and database
psql -h %DB_HOST% -p %DB_PORT% -U postgres -c "CREATE USER %DB_USER% WITH PASSWORD '%DB_PASSWORD%';" 2>nul
if %errorlevel% equ 0 echo ✅ Created user: %DB_USER%

psql -h %DB_HOST% -p %DB_PORT% -U postgres -c "CREATE DATABASE %DB_NAME% OWNER %DB_USER%;" 2>nul
if %errorlevel% equ 0 echo ✅ Created database: %DB_NAME%

psql -h %DB_HOST% -p %DB_PORT% -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE %DB_NAME% TO %DB_USER%;" 2>nul
psql -h %DB_HOST% -p %DB_PORT% -U postgres -c "ALTER USER %DB_USER% CREATEDB;" 2>nul
echo ✅ Granted privileges to %DB_USER%

echo.
echo Creating .env file...

REM Create .env file
(
echo # BOME Backend Environment Configuration
echo # Generated on %date% %time%
echo.
echo # Server Configuration
echo SERVER_PORT=8080
echo SERVER_HOST=0.0.0.0
echo ENVIRONMENT=development
echo DEBUG=true
echo.
echo # PostgreSQL Database Configuration
echo DB_HOST=%DB_HOST%
echo DB_PORT=%DB_PORT%
echo DB_NAME=%DB_NAME%
echo DB_USER=%DB_USER%
echo DB_PASSWORD=%DB_PASSWORD%
echo DB_SSL_MODE=disable
echo.
echo # Redis Configuration
echo REDIS_HOST=localhost
echo REDIS_PORT=6379
echo REDIS_PASSWORD=
echo REDIS_DB=0
echo.
echo # JWT Configuration
echo JWT_SECRET=your-super-secret-jwt-key-change-in-production
echo JWT_EXPIRY=24h
echo JWT_REFRESH_EXPIRY=168h
echo.
echo # CORS Configuration
echo CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:4173
echo CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
echo CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-Requested-With
echo.
echo # Rate Limiting Configuration
echo RATE_LIMIT_REQUESTS=100
echo RATE_LIMIT_WINDOW=1m
echo RATE_LIMIT_BURST=200
echo.
echo # Security Configuration
echo BCRYPT_COST=12
echo SESSION_SECRET=your-session-secret-key
echo CSRF_SECRET=your-csrf-secret-key
echo.
echo # Admin Configuration
echo ADMIN_EMAIL=admin@bookofmormonevidence.org
echo ADMIN_PASSWORD=change_this_in_production
echo ADMIN_SECRET_KEY=your-admin-secret-key
echo.
echo # Add your third-party service keys below:
echo # BUNNY_STORAGE_ZONE=
echo # BUNNY_API_KEY=
echo # STRIPE_SECRET_KEY=
echo # DO_SPACES_KEY=
echo # SENDGRID_API_KEY=
) > .env

echo ✅ Created .env file with database configuration

echo.
echo Testing database connection...
set PGPASSWORD=%DB_PASSWORD%
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -c "SELECT version();" >nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ Database connection test successful
) else (
    echo ❌ Database connection test failed
    pause
    exit /b 1
)

echo.
echo 🎉 PostgreSQL setup completed successfully!
echo.
echo Database Details:
echo   Host: %DB_HOST%
echo   Port: %DB_PORT%
echo   Database: %DB_NAME%
echo   User: %DB_USER%
echo   Password: %DB_PASSWORD%
echo.
echo Next steps:
echo 1. Build the backend: go build -o bome-backend.exe main.go
echo 2. Start the server: bome-backend.exe
echo 3. The database migrations will run automatically
echo.
echo ⚠️ IMPORTANT: Change the admin password in .env before production!
echo.
pause 