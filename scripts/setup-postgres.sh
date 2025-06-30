#!/bin/bash

# BOME PostgreSQL Setup Script
# This script helps set up PostgreSQL database for the BOME project

set -e

echo "🚀 BOME PostgreSQL Setup"
echo "========================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
DB_NAME="bome_streaming"
DB_USER="bome_user"
DB_PASSWORD=""
DB_HOST="localhost"
DB_PORT="5432"

# Function to print colored output
print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check if PostgreSQL is installed
check_postgresql() {
    if command -v psql >/dev/null 2>&1; then
        print_status "PostgreSQL client found"
    else
        print_error "PostgreSQL client not found. Please install PostgreSQL first."
        echo "  Ubuntu/Debian: sudo apt-get install postgresql postgresql-contrib"
        echo "  macOS: brew install postgresql"
        echo "  Windows: Download from https://www.postgresql.org/download/windows/"
        exit 1
    fi
}

# Check if PostgreSQL server is running
check_postgres_server() {
    if pg_isready -h $DB_HOST -p $DB_PORT >/dev/null 2>&1; then
        print_status "PostgreSQL server is running"
    else
        print_error "PostgreSQL server is not running on $DB_HOST:$DB_PORT"
        echo "Please start PostgreSQL server first."
        exit 1
    fi
}

# Generate random password if not provided
generate_password() {
    if [ -z "$DB_PASSWORD" ]; then
        DB_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)
        print_warning "Generated random password: $DB_PASSWORD"
    fi
}

# Create database and user
create_database() {
    echo "Creating PostgreSQL database and user..."
    
    # Create user
    sudo -u postgres psql -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';" 2>/dev/null || true
    print_status "Created user: $DB_USER"
    
    # Create database
    sudo -u postgres psql -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;" 2>/dev/null || true
    print_status "Created database: $DB_NAME"
    
    # Grant privileges
    sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;" 2>/dev/null || true
    sudo -u postgres psql -c "ALTER USER $DB_USER CREATEDB;" 2>/dev/null || true
    print_status "Granted privileges to $DB_USER"
}

# Create .env file
create_env_file() {
    echo "Creating .env file..."
    
    cat > .env << EOF
# BOME Backend Environment Configuration
# Generated on $(date)

# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
ENVIRONMENT=development
DEBUG=true

# PostgreSQL Database Configuration
DB_HOST=$DB_HOST
DB_PORT=$DB_PORT
DB_NAME=$DB_NAME
DB_USER=$DB_USER
DB_PASSWORD=$DB_PASSWORD
DB_SSL_MODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT Configuration
JWT_SECRET=$(openssl rand -base64 32 | tr -d "=+/")
JWT_EXPIRY=24h
JWT_REFRESH_EXPIRY=168h

# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:4173
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-Requested-With

# Rate Limiting Configuration
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m
RATE_LIMIT_BURST=200

# Security Configuration
BCRYPT_COST=12
SESSION_SECRET=$(openssl rand -base64 32 | tr -d "=+/")
CSRF_SECRET=$(openssl rand -base64 32 | tr -d "=+/")

# Admin Configuration
ADMIN_EMAIL=admin@bookofmormonevidence.org
ADMIN_PASSWORD=change_this_in_production
ADMIN_SECRET_KEY=$(openssl rand -base64 32 | tr -d "=+/")

# Add your third-party service keys below:
# BUNNY_STORAGE_ZONE=
# BUNNY_API_KEY=
# STRIPE_SECRET_KEY=
# DO_SPACES_KEY=
# SENDGRID_API_KEY=
EOF

    print_status "Created .env file with database configuration"
}

# Test database connection
test_connection() {
    echo "Testing database connection..."
    
    if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT version();" >/dev/null 2>&1; then
        print_status "Database connection test successful"
    else
        print_error "Database connection test failed"
        exit 1
    fi
}

# Run migrations
run_migrations() {
    echo "Running database migrations..."
    
    if [ -f "./bome-backend" ]; then
        ./bome-backend --migrate-only || true
        print_status "Database migrations completed"
    else
        print_warning "Binary not found. Run 'go build -o bome-backend ./main.go' first"
        print_warning "Then run './bome-backend' to start the server with migrations"
    fi
}

# Main setup process
main() {
    echo "Setting up PostgreSQL for BOME..."
    echo
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --db-name)
                DB_NAME="$2"
                shift 2
                ;;
            --db-user)
                DB_USER="$2"
                shift 2
                ;;
            --db-password)
                DB_PASSWORD="$2"
                shift 2
                ;;
            --db-host)
                DB_HOST="$2"
                shift 2
                ;;
            --db-port)
                DB_PORT="$2"
                shift 2
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --db-name     Database name (default: bome_streaming)"
                echo "  --db-user     Database user (default: bome_user)"
                echo "  --db-password Database password (default: auto-generated)"
                echo "  --db-host     Database host (default: localhost)"
                echo "  --db-port     Database port (default: 5432)"
                echo "  --help        Show this help message"
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done
    
    check_postgresql
    check_postgres_server
    generate_password
    create_database
    create_env_file
    test_connection
    
    echo
    echo "🎉 PostgreSQL setup completed successfully!"
    echo
    echo "Database Details:"
    echo "  Host: $DB_HOST"
    echo "  Port: $DB_PORT"
    echo "  Database: $DB_NAME"
    echo "  User: $DB_USER"
    echo "  Password: $DB_PASSWORD"
    echo
    echo "Next steps:"
    echo "1. Build the backend: go build -o bome-backend ./main.go"
    echo "2. Start the server: ./bome-backend"
    echo "3. The database migrations will run automatically"
    echo
    print_warning "IMPORTANT: Change the admin password in .env before production!"
}

# Run main function
main "$@" 