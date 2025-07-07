#!/bin/bash

# BOME Database Setup Script
# This script sets up the PostgreSQL database using the unified migration system

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 BOME Database Setup${NC}"
echo "================================"

# Check if PostgreSQL is running
if ! pg_isready -q; then
    echo -e "${RED}❌ PostgreSQL is not running. Please start PostgreSQL first.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ PostgreSQL is running${NC}"

# Load environment variables
if [ -f .env ]; then
    echo -e "${YELLOW}📄 Loading environment variables from .env${NC}"
    export $(cat .env | grep -v '^#' | xargs)
else
    echo -e "${YELLOW}⚠️  No .env file found, using system environment variables${NC}"
fi

# Set default values if not provided
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-bome_admin}
DB_PASSWORD=${DB_PASSWORD:-AdminBOME}
DB_NAME=${DB_NAME:-bome_db}
DB_SSL_MODE=${DB_SSL_MODE:-disable}

echo -e "${YELLOW}📊 Database Configuration:${NC}"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  User: $DB_USER"
echo "  Database: $DB_NAME"
echo "  SSL Mode: $DB_SSL_MODE"

# Create database and user if they don't exist
echo -e "${YELLOW}🔧 Setting up database and user...${NC}"

# Connect as postgres superuser to create database and user
psql -h $DB_HOST -p $DB_PORT -U postgres -d postgres -c "
DO \$\$
BEGIN
    -- Create user if it doesn't exist
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '$DB_USER') THEN
        CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';
        RAISE NOTICE 'User $DB_USER created successfully';
    ELSE
        RAISE NOTICE 'User $DB_USER already exists';
    END IF;
    
    -- Create database if it doesn't exist
    IF NOT EXISTS (SELECT FROM pg_database WHERE datname = '$DB_NAME') THEN
        CREATE DATABASE $DB_NAME OWNER $DB_USER;
        RAISE NOTICE 'Database $DB_NAME created successfully';
    ELSE
        RAISE NOTICE 'Database $DB_NAME already exists';
    END IF;
    
    -- Grant privileges
    GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;
    ALTER USER $DB_USER CREATEDB;
END
\$\$;
" 2>/dev/null || {
    echo -e "${YELLOW}⚠️  Could not connect as postgres superuser. Database and user may already exist.${NC}"
}

echo -e "${GREEN}✅ Database setup completed${NC}"

# Run the backend to execute migrations
echo -e "${YELLOW}🔄 Running database migrations...${NC}"

# Build the backend
echo -e "${YELLOW}🔨 Building backend...${NC}"
go build -o bome-backend .

# Run migrations
echo -e "${YELLOW}📦 Executing migrations...${NC}"
./bome-backend &
BACKEND_PID=$!

# Wait a moment for the backend to start and run migrations
sleep 5

# Stop the backend
kill $BACKEND_PID 2>/dev/null || true

echo -e "${GREEN}✅ Database setup and migrations completed successfully!${NC}"
echo ""
echo -e "${GREEN}🎉 Your BOME database is ready to use!${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "  1. Start the backend: go run main.go"
echo "  2. Start the frontend: cd ../frontend && npm run dev"
echo ""
echo -e "${GREEN}Happy coding! 🚀${NC}" 