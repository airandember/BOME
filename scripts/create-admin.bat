@echo off
REM Create Admin Account Script for BOME (Windows)
REM This script creates a test admin account for development

echo 🔧 Creating test admin account for BOME...

REM Set environment variables for development
set ENVIRONMENT=development
set DEBUG=true
set SERVER_PORT=8080
set SERVER_HOST=0.0.0.0
set DB_HOST=localhost
set DB_PORT=5432
set DB_NAME=bome_streaming
set DB_USER=bome_user
set DB_PASSWORD=dev_password
set DB_SSL_MODE=disable
set JWT_SECRET=dev-jwt-secret-key-change-in-production
set JWT_EXPIRY=24h
set JWT_REFRESH_EXPIRY=168h
set CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:5174,http://localhost:4173
set CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
set CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-Requested-With
set ADMIN_EMAIL=admin@bome.test
set ADMIN_PASSWORD=Admin123!
set ADMIN_SECRET_KEY=dev-admin-secret-key
set BCRYPT_COST=10
set SESSION_SECRET=dev-session-secret-key
set CSRF_SECRET=dev-csrf-secret-key

REM Change to backend directory
cd backend

REM Build and run the create-admin command
echo 📦 Building admin creation tool...
go build -o create-admin.exe cmd/create-admin/main.go

echo 🚀 Running admin creation...
create-admin.exe

echo.
echo 🎉 Admin account creation completed!
echo.
echo 📋 Login Credentials:
echo    Email: admin@bome.test
echo    Password: Admin123!
echo.
echo 🌐 Access the admin dashboard at:
echo    http://localhost:5174/admin
echo.
echo ⚠️  IMPORTANT: These are test credentials for development only!
echo    Change them in production!

REM Clean up
del create-admin.exe

echo.
echo ✅ Script completed successfully!
pause 