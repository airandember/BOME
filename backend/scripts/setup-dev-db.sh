#!/bin/bash

# BOME Development Database Setup Script
# This script helps set up PostgreSQL for development

echo "🐘 BOME Development Database Setup"
echo "=================================="

# Check if PostgreSQL is installed
if ! command -v psql &> /dev/null; then
    echo "❌ PostgreSQL is not installed or not in PATH"
    echo "Please install PostgreSQL first:"
    echo "  - macOS: brew install postgresql"
    echo "  - Ubuntu: sudo apt-get install postgresql postgresql-contrib"
    echo "  - Windows: Download from https://www.postgresql.org/download/windows/"
    exit 1
fi

# Check if PostgreSQL service is running
if ! pg_isready -q; then
    echo "❌ PostgreSQL service is not running"
    echo "Please start PostgreSQL service:"
    echo "  - macOS: brew services start postgresql"
    echo "  - Ubuntu: sudo systemctl start postgresql"
    echo "  - Windows: Start PostgreSQL service from Services"
    exit 1
fi

echo "✅ PostgreSQL is running"

# Database configuration
DB_NAME="bome_streaming"
DB_USER="bome_user"
DB_PASSWORD="bome_password_dev"

echo "📝 Setting up database: $DB_NAME"
echo "👤 User: $DB_USER"

# Create database user if it doesn't exist
echo "Creating database user..."
psql -d postgres -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';" 2>/dev/null || echo "User already exists"

# Create database if it doesn't exist
echo "Creating database..."
createdb -U postgres $DB_NAME 2>/dev/null || echo "Database already exists"

# Grant privileges
echo "Granting privileges..."
psql -d postgres -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"

# Create .env file with database configuration
echo "Creating .env file..."
cat > .env << EOF
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
EOF

echo "✅ Database setup complete!"
echo ""
echo "📋 Next steps:"
echo "1. Start the backend: go run main.go"
echo "2. The backend will automatically run migrations"
echo "3. Test the admin users endpoint: http://localhost:8080/api/v1/admin/users"
echo ""
echo "🔧 Database connection details:"
echo "  Host: localhost"
echo "  Port: 5432"
echo "  Database: $DB_NAME"
echo "  User: $DB_USER"
echo "  Password: $DB_PASSWORD"
echo ""
echo "⚠️  Remember to change passwords and secrets in production!" 