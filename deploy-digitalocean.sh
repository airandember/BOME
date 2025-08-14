#!/bin/bash

# BOME Digital Ocean Deployment Script
# This script helps deploy the BOME application to a Digital Ocean droplet

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Check if running on Digital Ocean
check_environment() {
    if [ -f /etc/digitalocean ]; then
        print_success "Running on Digital Ocean droplet"
    else
        print_warning "Not running on Digital Ocean - some features may not work"
    fi
}

# Install Docker if not present
install_docker() {
    if ! command -v docker &> /dev/null; then
        print_status "Installing Docker..."
        curl -fsSL https://get.docker.com -o get-docker.sh
        sh get-docker.sh
        sudo usermod -aG docker $USER
        print_success "Docker installed successfully"
        print_warning "Please log out and back in for group changes to take effect"
    else
        print_success "Docker already installed"
    fi
}

# Install Docker Compose if not present
install_docker_compose() {
    if ! command -v docker-compose &> /dev/null; then
        print_status "Installing Docker Compose..."
        sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
        sudo chmod +x /usr/local/bin/docker-compose
        print_success "Docker Compose installed successfully"
    else
        print_success "Docker Compose already installed"
    fi
}

# Create production environment file
setup_environment() {
    if [ ! -f .env ]; then
        print_status "Creating production environment file..."
        cp env.example .env
        print_warning "Please edit .env with your production values:"
        print_warning "  - Set strong passwords for DB_PASSWORD, JWT_SECRET, etc."
        print_warning "  - Configure Stripe, Bunny.net, and other service keys"
        print_warning "  - Set ADMIN_PASSWORD to a secure value"
    else
        print_success "Environment file already exists"
    fi
}

# Build and start services
deploy_services() {
    print_status "Building and starting services..."
    
    # Stop any existing services
    docker-compose -f docker-compose.prod.yml down 2>/dev/null || true
    
    # Build and start
    docker-compose -f docker-compose.prod.yml up -d --build
    
    print_success "Services deployed successfully!"
}

# Check service health
check_health() {
    print_status "Checking service health..."
    
    # Wait for services to start
    sleep 30
    
    # Check PostgreSQL
    if docker-compose -f docker-compose.prod.yml exec -T postgres pg_isready -U bome_user -d bome > /dev/null 2>&1; then
        print_success "PostgreSQL: Healthy"
    else
        print_error "PostgreSQL: Unhealthy"
        return 1
    fi
    
    # Check Backend
    local port=$(docker-compose -f docker-compose.prod.yml exec -T backend printenv PORT 2>/dev/null || echo "8080")
    if curl -s "http://localhost:$port/health" > /dev/null 2>&1; then
        print_success "Backend: Healthy"
    else
        print_error "Backend: Unhealthy"
        return 1
    fi
    
    print_success "All services are healthy!"
}

# Setup firewall
setup_firewall() {
    print_status "Setting up firewall..."
    
    # Allow SSH
    sudo ufw allow ssh
    
    # Allow HTTP/HTTPS (if using nginx)
    sudo ufw allow 80
    sudo ufw allow 443
    
    # Allow the port Digital Ocean assigns to your backend
    # This will be dynamic, so we'll enable it after deployment
    
    # Enable firewall
    sudo ufw --force enable
    
    print_success "Firewall configured"
}

# Setup automatic backups
setup_backups() {
    print_status "Setting up automatic database backups..."
    
    # Create backup directory
    sudo mkdir -p /opt/bome/backups
    sudo chown $USER:$USER /opt/bome/backups
    
    # Create backup script
    cat > /opt/bome/backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/opt/bome/backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/bome_backup_$DATE.sql"

# Create backup
docker-compose -f /opt/bome/docker-compose.prod.yml exec -T postgres pg_dump -U bome_user bome > "$BACKUP_FILE"

# Compress backup
gzip "$BACKUP_FILE"

# Keep only last 7 days of backups
find "$BACKUP_DIR" -name "*.sql.gz" -mtime +7 -delete

echo "Backup completed: $BACKUP_FILE.gz"
EOF
    
    chmod +x /opt/bome/backup.sh
    
    # Add to crontab (daily at 2 AM)
    (crontab -l 2>/dev/null; echo "0 2 * * * /opt/bome/backup.sh") | crontab -
    
    print_success "Automatic backups configured"
}

# Setup monitoring
setup_monitoring() {
    print_status "Setting up basic monitoring..."
    
    # Create monitoring script
    cat > /opt/bome/monitor.sh << 'EOF'
#!/bin/bash
LOG_FILE="/opt/bome/monitor.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

# Check PostgreSQL
if ! docker-compose -f /opt/bome/docker-compose.prod.yml exec -T postgres pg_isready -U bome_user -d bome > /dev/null 2>&1; then
    echo "[$DATE] ERROR: PostgreSQL is down" >> "$LOG_FILE"
fi

# Check Backend
PORT=$(docker-compose -f /opt/bome/docker-compose.prod.yml exec -T backend printenv PORT 2>/dev/null || echo "8080")
if ! curl -s "http://localhost:$PORT/health" > /dev/null 2>&1; then
    echo "[$DATE] ERROR: Backend is down" >> "$LOG_FILE"
fi

# Check disk space
DISK_USAGE=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
if [ "$DISK_USAGE" -gt 80 ]; then
    echo "[$DATE] WARNING: Disk usage is ${DISK_USAGE}%" >> "$LOG_FILE"
fi

# Check memory
MEM_USAGE=$(free | grep Mem | awk '{printf("%.0f", $3/$2 * 100.0)}')
if [ "$MEM_USAGE" -gt 80 ]; then
    echo "[$DATE] WARNING: Memory usage is ${MEM_USAGE}%" >> "$LOG_FILE"
fi
EOF
    
    chmod +x /opt/bome/monitor.sh
    
    # Add to crontab (every 5 minutes)
    (crontab -l 2>/dev/null; echo "*/5 * * * * /opt/bome/monitor.sh") | crontab -
    
    print_success "Monitoring configured"
}

# Main deployment function
main() {
    print_status "Starting BOME deployment to Digital Ocean..."
    
    check_environment
    install_docker
    install_docker_compose
    setup_environment
    deploy_services
    check_health
    setup_firewall
    setup_backups
    setup_monitoring
    
    print_success "Deployment completed successfully!"
    print_status "Next steps:"
    print_status "1. Edit .env with your production values"
    print_status "2. Restart services: docker-compose -f docker-compose.prod.yml restart"
    print_status "3. Check logs: docker-compose -f docker-compose.prod.yml logs -f"
    print_status "4. Access your application at the URL Digital Ocean provides"
}

# Run main function
main "$@"
