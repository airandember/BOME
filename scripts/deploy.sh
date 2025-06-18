#!/bin/bash

# BOME Platform Deployment Script
# This script deploys the entire BOME streaming platform

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
ENVIRONMENT=${1:-production}
DOMAIN=${2:-bome-streaming.com}

echo -e "${GREEN}🚀 Starting BOME Platform Deployment${NC}"
echo -e "${YELLOW}Environment: ${ENVIRONMENT}${NC}"
echo -e "${YELLOW}Domain: ${DOMAIN}${NC}"

# Function to print status
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Check if Docker and Docker Compose are installed
check_dependencies() {
    print_status "Checking dependencies..."
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed"
        exit 1
    fi
    
    print_status "Dependencies check passed"
}

# Load environment variables
load_env() {
    print_status "Loading environment variables..."
    
    if [ ! -f .env ]; then
        print_error ".env file not found"
        exit 1
    fi
    
    source .env
    print_status "Environment variables loaded"
}

# Build and deploy services
deploy_services() {
    print_status "Building and deploying services..."
    
    # Stop existing containers
    docker-compose down
    
    # Build images
    docker-compose build --no-cache
    
    # Start services
    docker-compose up -d
    
    print_status "Services deployed successfully"
}

# Setup SSL certificates
setup_ssl() {
    print_status "Setting up SSL certificates..."
    
    # Check if certbot is available
    if command -v certbot &> /dev/null; then
        # Get SSL certificates
        certbot certonly --standalone \
            -d ${DOMAIN} \
            -d www.${DOMAIN} \
            -d api.${DOMAIN} \
            -d admin.${DOMAIN} \
            --non-interactive --agree-tos --email admin@${DOMAIN}
        
        print_status "SSL certificates obtained"
    else
        print_warning "Certbot not found. SSL certificates must be set up manually."
    fi
}

# Setup database
setup_database() {
    print_status "Setting up database..."
    
    # Wait for PostgreSQL to be ready
    echo "Waiting for PostgreSQL to be ready..."
    sleep 30
    
    # Run database migrations
    docker-compose exec backend ./migrate up
    
    print_status "Database setup completed"
}

# Health checks
health_check() {
    print_status "Performing health checks..."
    
    # Check if services are running
    services=("bome-backend" "bome-frontend" "bome-admin" "bome-postgres" "bome-redis")
    
    for service in "${services[@]}"; do
        if docker ps | grep -q $service; then
            print_status "$service is running"
        else
            print_error "$service is not running"
            exit 1
        fi
    done
    
    print_status "All services are healthy"
}

# Main deployment process
main() {
    echo -e "${GREEN}Starting deployment process...${NC}"
    
    check_dependencies
    load_env
    deploy_services
    setup_database
    health_check
    
    if [ "$ENVIRONMENT" = "production" ]; then
        setup_ssl
    fi
    
    echo -e "${GREEN}🎉 BOME Platform deployment completed successfully!${NC}"
    echo -e "${YELLOW}Frontend: https://${DOMAIN}${NC}"
    echo -e "${YELLOW}API: https://api.${DOMAIN}${NC}"
    echo -e "${YELLOW}Admin: https://admin.${DOMAIN}${NC}"
}

# Run main function
main "$@" 