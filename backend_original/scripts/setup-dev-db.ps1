# BOME Development Database Setup Script (PowerShell)
# This script helps set up PostgreSQL for development on Windows

Write-Host "🐘 BOME Development Database Setup" -ForegroundColor Green
Write-Host "==================================" -ForegroundColor Green

# Check if PostgreSQL is installed
try {
    $psqlPath = Get-Command psql -ErrorAction Stop
    Write-Host "✅ PostgreSQL found at: $($psqlPath.Source)" -ForegroundColor Green
} catch {
    Write-Host "❌ PostgreSQL is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please install PostgreSQL first:" -ForegroundColor Yellow
    Write-Host "  - Download from https://www.postgresql.org/download/windows/" -ForegroundColor Yellow
    Write-Host "  - Add PostgreSQL bin directory to your PATH" -ForegroundColor Yellow
    exit 1
}

# Check if PostgreSQL service is running
try {
    $pgService = Get-Service -Name "postgresql*" -ErrorAction Stop
    if ($pgService.Status -eq "Running") {
        Write-Host "✅ PostgreSQL service is running" -ForegroundColor Green
    } else {
        Write-Host "❌ PostgreSQL service is not running" -ForegroundColor Red
        Write-Host "Please start PostgreSQL service from Services or run:" -ForegroundColor Yellow
        Write-Host "  Start-Service postgresql*" -ForegroundColor Yellow
        exit 1
    }
} catch {
    Write-Host "❌ PostgreSQL service not found" -ForegroundColor Red
    Write-Host "Please ensure PostgreSQL is properly installed" -ForegroundColor Yellow
    exit 1
}

# Database configuration
$DB_NAME = "bome_streaming"
$DB_USER = "bome_user"
$DB_PASSWORD = "bome_password_dev"

Write-Host "📝 Setting up database: $DB_NAME" -ForegroundColor Cyan
Write-Host "👤 User: $DB_USER" -ForegroundColor Cyan

# Create database user if it doesn't exist
Write-Host "Creating database user..." -ForegroundColor Yellow
try {
    psql -d postgres -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';" 2>$null
    Write-Host "✅ User created successfully" -ForegroundColor Green
} catch {
    Write-Host "ℹ️  User already exists" -ForegroundColor Blue
}

# Create database if it doesn't exist
Write-Host "Creating database..." -ForegroundColor Yellow
try {
    createdb -U postgres $DB_NAME 2>$null
    Write-Host "✅ Database created successfully" -ForegroundColor Green
} catch {
    Write-Host "ℹ️  Database already exists" -ForegroundColor Blue
}

# Grant privileges
Write-Host "Granting privileges..." -ForegroundColor Yellow
try {
    psql -d postgres -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;" 2>$null
    Write-Host "✅ Privileges granted" -ForegroundColor Green
} catch {
    Write-Host "⚠️  Warning: Could not grant privileges" -ForegroundColor Yellow
}

# Create .env file with database configuration
Write-Host "Creating .env file..." -ForegroundColor Yellow
$envContent = @"
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
ENVIRONMENT=development
DEBUG=true

# PostgreSQL Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=$DB_NAME
DB_USER=$DB_USER
DB_PASSWORD=$DB_PASSWORD
DB_SSL_MODE=disable

# JWT Configuration (CHANGE IN PRODUCTION)
JWT_SECRET=dev-jwt-secret-change-in-production
JWT_REFRESH_SECRET=dev-refresh-secret-change-in-production
JWT_EXPIRY=4h
JWT_REFRESH_EXPIRY=168h
PUBLIC_APP_URL=http://localhost:5173

# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:4173
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-Requested-With

# Rate Limiting Configuration
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m
RATE_LIMIT_BURST=200

# Logging Configuration
LOG_LEVEL=debug
LOG_FORMAT=json

# Admin Configuration (CHANGE IN PRODUCTION)
ADMIN_EMAIL=admin@bome.test
ADMIN_PASSWORD=admin123
ADMIN_SECRET_KEY=dev-admin-secret-change-in-production
"@

$envContent | Out-File -FilePath ".env" -Encoding UTF8
Write-Host "✅ .env file created" -ForegroundColor Green

Write-Host ""
Write-Host "✅ Database setup complete!" -ForegroundColor Green
Write-Host ""
Write-Host "📋 Next steps:" -ForegroundColor Cyan
Write-Host "1. Start the backend: go run main.go" -ForegroundColor White
Write-Host "2. The backend will automatically run migrations" -ForegroundColor White
Write-Host "3. Test the admin users endpoint: http://localhost:8080/api/v1/admin/users" -ForegroundColor White
Write-Host ""
Write-Host "🔧 Database connection details:" -ForegroundColor Cyan
Write-Host "  Host: localhost" -ForegroundColor White
Write-Host "  Port: 5432" -ForegroundColor White
Write-Host "  Database: $DB_NAME" -ForegroundColor White
Write-Host "  User: $DB_USER" -ForegroundColor White
Write-Host "  Password: $DB_PASSWORD" -ForegroundColor White
Write-Host ""
Write-Host "⚠️  Remember to change passwords and secrets in production!" -ForegroundColor Red 