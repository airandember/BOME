#!/bin/bash

# BOME Docker Management Script
# Usage: ./docker-scripts.sh [command]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
}

# Function to check if docker-compose is available
check_compose() {
    if ! command -v docker-compose &> /dev/null; then
        print_error "docker-compose is not installed. Please install it and try again."
        exit 1
    fi
}

# Function to start all services
start() {
    print_status "Starting BOME services..."
    check_docker
    check_compose
    
    if [ ! -f .env ]; then
        print_warning ".env file not found. Creating from template..."
        if [ -f env.example ]; then
            cp env.example .env
            print_warning "Please edit .env file with your configuration values"
        else
            print_error "env.example not found. Please create .env file manually."
            exit 1
        fi
    fi
    
    docker-compose up -d
    print_success "Services started successfully!"
    
    # Wait for services to be ready
    print_status "Waiting for services to be ready..."
    sleep 10
    
    # Check service status
    status
}

# Function to stop all services
stop() {
    print_status "Stopping BOME services..."
    docker-compose down
    print_success "Services stopped successfully!"
}

# Function to restart all services
restart() {
    print_status "Restarting BOME services..."
    docker-compose restart
    print_success "Services restarted successfully!"
}

# Function to show service status
status() {
    print_status "Service Status:"
    docker-compose ps
    
    echo ""
    print_status "Health Checks:"
    
    # Check PostgreSQL
    if docker-compose ps postgres | grep -q "Up"; then
        if docker-compose exec -T postgres pg_isready -U bome_user -d bome > /dev/null 2>&1; then
            print_success "PostgreSQL: Healthy"
        else
            print_warning "PostgreSQL: Starting up..."
        fi
    else
        print_error "PostgreSQL: Not running"
    fi
    
    # Check Backend
    if docker-compose ps backend | grep -q "Up"; then
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            print_success "Backend: Healthy"
        else
            print_warning "Backend: Starting up..."
        fi
    else
        print_error "Backend: Not running"
    fi
    
    # Check Frontend
    if docker-compose ps frontend | grep -q "Up"; then
        if curl -s http://localhost:3000 > /dev/null 2>&1; then
            print_success "Frontend: Healthy"
        else
            print_warning "Frontend: Starting up..."
        fi
    else
        print_error "Frontend: Not running"
    fi
}

# Function to show logs
logs() {
    local service=${1:-""}
    if [ -z "$service" ]; then
        print_status "Showing logs for all services..."
        docker-compose logs -f
    else
        print_status "Showing logs for $service..."
        docker-compose logs -f "$service"
    fi
}

# Function to rebuild services
rebuild() {
    local service=${1:-""}
    if [ -z "$service" ]; then
        print_status "Rebuilding all services..."
        docker-compose down
        docker-compose build --no-cache
        docker-compose up -d
    else
        print_status "Rebuilding $service..."
        docker-compose build --no-cache "$service"
        docker-compose up -d "$service"
    fi
    print_success "Rebuild completed!"
}

# Function to reset database
reset_db() {
    print_warning "This will delete all data in the database. Are you sure? (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        print_status "Resetting database..."
        docker-compose down -v
        docker-compose up -d postgres
        print_status "Waiting for PostgreSQL to be ready..."
        sleep 15
        print_success "Database reset completed!"
        print_status "You may need to run migrations manually."
    else
        print_status "Database reset cancelled."
    fi
}

# Function to backup database
backup_db() {
    local filename="bome_backup_$(date +%Y%m%d_%H%M%S).sql"
    print_status "Creating database backup: $filename"
    
    if docker-compose exec -T postgres pg_dump -U bome_user bome > "$filename"; then
        print_success "Backup created successfully: $filename"
    else
        print_error "Backup failed!"
        exit 1
    fi
}

# Function to restore database
restore_db() {
    local filename=${1:-""}
    if [ -z "$filename" ]; then
        print_error "Please specify a backup file to restore"
        echo "Usage: $0 restore-db <backup_file>"
        exit 1
    fi
    
    if [ ! -f "$filename" ]; then
        print_error "Backup file not found: $filename"
        exit 1
    fi
    
    print_warning "This will overwrite the current database. Are you sure? (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        print_status "Restoring database from $filename..."
        if docker-compose exec -T postgres psql -U bome_user -d bome < "$filename"; then
            print_success "Database restored successfully!"
        else
            print_error "Database restore failed!"
            exit 1
        fi
    else
        print_status "Database restore cancelled."
    fi
}

# Function to show help
show_help() {
    echo "BOME Docker Management Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  start       Start all services"
    echo "  stop        Stop all services"
    echo "  restart     Restart all services"
    echo "  status      Show service status and health"
    echo "  logs [svc]  Show logs (all services or specific service)"
    echo "  rebuild [svc] Rebuild services (all or specific)"
    echo "  reset-db    Reset database (delete all data)"
    echo "  backup-db   Create database backup"
    echo "  restore-db <file> Restore database from backup"
    echo "  help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 start"
    echo "  $0 logs backend"
    echo "  $0 rebuild frontend"
    echo "  $0 backup-db"
    echo "  $0 restore-db backup.sql"
}

# Main script logic
case "${1:-help}" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    status)
        status
        ;;
    logs)
        logs "$2"
        ;;
    rebuild)
        rebuild "$2"
        ;;
    reset-db)
        reset_db
        ;;
    backup-db)
        backup_db
        ;;
    restore-db)
        restore_db "$2"
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        print_error "Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
