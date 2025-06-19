# BOME Deployment Guide

Complete deployment guide for the Book of Mormon Evidences (BOME) streaming platform.

## 📋 Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Infrastructure Setup](#infrastructure-setup)
- [Environment Configuration](#environment-configuration)
- [Database Setup](#database-setup)
- [Application Deployment](#application-deployment)
- [SSL Configuration](#ssl-configuration)
- [Monitoring Setup](#monitoring-setup)
- [Backup Configuration](#backup-configuration)
- [CI/CD Pipeline](#cicd-pipeline)
- [Production Checklist](#production-checklist)
- [Troubleshooting](#troubleshooting)

## Overview

The BOME platform is deployed using a containerized architecture on Digital Ocean infrastructure with the following components:

### Architecture Overview
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │    │   Web Server    │    │   API Server    │
│     (Nginx)     │────│    (Svelte)     │────│      (Go)       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   PostgreSQL    │    │      Redis      │    │   Bunny.net     │
│    Database     │    │     Cache       │    │   Video CDN     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Technology Stack
- **Frontend**: Svelte/SvelteKit (Node.js)
- **Backend**: Go API server
- **Database**: PostgreSQL 14+
- **Cache**: Redis 6+
- **Web Server**: Nginx
- **Containers**: Docker & Docker Compose
- **Infrastructure**: Digital Ocean Droplets
- **CDN**: Bunny.net for video streaming
- **Monitoring**: Prometheus, Grafana, ELK Stack

## Prerequisites

### System Requirements
- **OS**: Ubuntu 20.04 LTS or newer
- **RAM**: Minimum 8GB (16GB recommended)
- **CPU**: 4 cores minimum (8 cores recommended)
- **Storage**: 100GB SSD minimum (500GB recommended)
- **Network**: 1Gbps connection

### Required Software
- Docker 20.10+
- Docker Compose 2.0+
- Git
- Node.js 18+
- Go 1.19+
- PostgreSQL 14+
- Redis 6+
- Nginx

### Required Accounts & Services
- Digital Ocean account
- Domain name and DNS management
- SSL certificate (Let's Encrypt)
- Bunny.net account for video CDN
- Stripe account for payments
- Email service (SendGrid/AWS SES)

## Infrastructure Setup

### 1. Digital Ocean Droplet Creation

```bash
# Create droplet using Digital Ocean CLI
doctl compute droplet create bome-production \
  --image ubuntu-20-04-x64 \
  --size s-4vcpu-8gb \
  --region nyc3 \
  --ssh-keys your-ssh-key-id \
  --enable-monitoring \
  --enable-backups
```

### 2. Initial Server Setup

```bash
# Connect to server
ssh root@your-server-ip

# Update system
apt update && apt upgrade -y

# Install required packages
apt install -y curl wget git unzip software-properties-common

# Create application user
useradd -m -s /bin/bash bome
usermod -aG sudo bome
su - bome
```

### 3. Docker Installation

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Add user to docker group
sudo usermod -aG docker bome

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.12.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify installation
docker --version
docker-compose --version
```

### 4. Firewall Configuration

```bash
# Configure UFW firewall
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw --force enable
```

## Environment Configuration

### 1. Environment Variables

Create `.env` file in the project root:

```bash
# Application Configuration
NODE_ENV=production
PORT=3000
API_PORT=8080

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=bome_production
DB_USER=bome_user
DB_PASSWORD=your_secure_password

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password

# JWT Configuration
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRES_IN=24h

# Stripe Configuration
STRIPE_PUBLIC_KEY=pk_live_your_stripe_public_key
STRIPE_SECRET_KEY=sk_live_your_stripe_secret_key
STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret

# Bunny.net Configuration
BUNNY_API_KEY=your_bunny_api_key
BUNNY_STORAGE_ZONE=your_storage_zone
BUNNY_CDN_URL=https://your-cdn-url.b-cdn.net

# Email Configuration
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASSWORD=your_sendgrid_api_key
FROM_EMAIL=noreply@bome.com

# Application URLs
FRONTEND_URL=https://bome.com
API_URL=https://api.bome.com

# Security
CORS_ORIGIN=https://bome.com
RATE_LIMIT_WINDOW=3600
RATE_LIMIT_MAX=1000

# Monitoring
PROMETHEUS_PORT=9090
GRAFANA_PORT=3001
```

### 2. Docker Compose Configuration

Create `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  # PostgreSQL Database
  postgres:
    image: postgres:14-alpine
    container_name: bome-postgres
    environment:
      POSTGRES_DB: ${DB_NAME}
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init-db.sql
    ports:
      - "5432:5432"
    restart: unless-stopped
    networks:
      - bome-network

  # Redis Cache
  redis:
    image: redis:6-alpine
    container_name: bome-redis
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    restart: unless-stopped
    networks:
      - bome-network

  # Go API Server
  api:
    build:
      context: ./backend
      dockerfile: Dockerfile.prod
    container_name: bome-api
    environment:
      - NODE_ENV=${NODE_ENV}
      - DB_HOST=postgres
      - DB_PORT=${DB_PORT}
      - DB_NAME=${DB_NAME}
      - DB_USER=${DB_USER}
      - DB_PASSWORD=${DB_PASSWORD}
      - REDIS_HOST=redis
      - REDIS_PORT=${REDIS_PORT}
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - JWT_SECRET=${JWT_SECRET}
      - STRIPE_SECRET_KEY=${STRIPE_SECRET_KEY}
      - BUNNY_API_KEY=${BUNNY_API_KEY}
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
    restart: unless-stopped
    networks:
      - bome-network

  # Svelte Frontend
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile.prod
    container_name: bome-frontend
    environment:
      - NODE_ENV=${NODE_ENV}
      - API_URL=${API_URL}
      - STRIPE_PUBLIC_KEY=${STRIPE_PUBLIC_KEY}
    ports:
      - "3000:3000"
    depends_on:
      - api
    restart: unless-stopped
    networks:
      - bome-network

  # Nginx Reverse Proxy
  nginx:
    image: nginx:alpine
    container_name: bome-nginx
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
      - ./logs/nginx:/var/log/nginx
    ports:
      - "80:80"
      - "443:443"
    depends_on:
      - frontend
      - api
    restart: unless-stopped
    networks:
      - bome-network

  # Prometheus Monitoring
  prometheus:
    image: prom/prometheus:latest
    container_name: bome-prometheus
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    ports:
      - "9090:9090"
    restart: unless-stopped
    networks:
      - bome-network

  # Grafana Dashboard
  grafana:
    image: grafana/grafana:latest
    container_name: bome-grafana
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana
    ports:
      - "3001:3000"
    restart: unless-stopped
    networks:
      - bome-network

volumes:
  postgres_data:
  redis_data:
  prometheus_data:
  grafana_data:

networks:
  bome-network:
    driver: bridge
```

## Database Setup

### 1. Database Initialization

```sql
-- scripts/init-db.sql
CREATE DATABASE bome_production;
CREATE USER bome_user WITH ENCRYPTED PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE bome_production TO bome_user;

-- Create extensions
\c bome_production;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";
```

### 2. Run Database Migrations

```bash
# Run migrations
cd backend
go run cmd/migrate/main.go up

# Verify migrations
go run cmd/migrate/main.go status
```

### 3. Database Backup Script

```bash
#!/bin/bash
# scripts/backup-db.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/var/backups/bome"
DB_NAME="bome_production"

mkdir -p $BACKUP_DIR

# Create database backup
pg_dump -h localhost -U bome_user -d $DB_NAME > $BACKUP_DIR/bome_db_$DATE.sql

# Compress backup
gzip $BACKUP_DIR/bome_db_$DATE.sql

# Remove backups older than 30 days
find $BACKUP_DIR -name "*.sql.gz" -mtime +30 -delete

echo "Database backup completed: bome_db_$DATE.sql.gz"
```

## Application Deployment

### 1. Clone Repository

```bash
# Clone the repository
git clone https://github.com/your-org/bome.git
cd bome

# Checkout production branch
git checkout production
```

### 2. Build and Deploy

```bash
# Build and start services
docker-compose -f docker-compose.prod.yml up -d --build

# Verify services are running
docker-compose -f docker-compose.prod.yml ps

# Check logs
docker-compose -f docker-compose.prod.yml logs -f
```

### 3. Frontend Build (Alternative)

```bash
# Build frontend separately if needed
cd frontend
npm ci --production
npm run build

# Copy build files to nginx
sudo cp -r build/* /var/www/bome/
```

### 4. Backend Build (Alternative)

```bash
# Build Go backend
cd backend
go build -o bome-api cmd/server/main.go

# Create systemd service
sudo tee /etc/systemd/system/bome-api.service > /dev/null <<EOF
[Unit]
Description=BOME API Server
After=network.target

[Service]
Type=simple
User=bome
WorkingDirectory=/home/bome/bome/backend
ExecStart=/home/bome/bome/backend/bome-api
Restart=always
RestartSec=5
Environment=NODE_ENV=production

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable bome-api
sudo systemctl start bome-api
```

## SSL Configuration

### 1. Install Certbot

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx

# Obtain SSL certificate
sudo certbot --nginx -d bome.com -d www.bome.com -d api.bome.com

# Verify auto-renewal
sudo certbot renew --dry-run
```

### 2. Nginx Configuration

```nginx
# nginx/nginx.conf
events {
    worker_connections 1024;
}

http {
    upstream frontend {
        server frontend:3000;
    }

    upstream api {
        server api:8080;
    }

    # Redirect HTTP to HTTPS
    server {
        listen 80;
        server_name bome.com www.bome.com api.bome.com;
        return 301 https://$server_name$request_uri;
    }

    # Main website
    server {
        listen 443 ssl http2;
        server_name bome.com www.bome.com;

        ssl_certificate /etc/nginx/ssl/fullchain.pem;
        ssl_certificate_key /etc/nginx/ssl/privkey.pem;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;
        ssl_prefer_server_ciphers off;

        location / {
            proxy_pass http://frontend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }

    # API server
    server {
        listen 443 ssl http2;
        server_name api.bome.com;

        ssl_certificate /etc/nginx/ssl/fullchain.pem;
        ssl_certificate_key /etc/nginx/ssl/privkey.pem;

        location / {
            proxy_pass http://api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```

## Monitoring Setup

### 1. Prometheus Configuration

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'bome-api'
    static_configs:
      - targets: ['api:8080']
    metrics_path: /metrics

  - job_name: 'bome-frontend'
    static_configs:
      - targets: ['frontend:3000']

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres:5432']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis:6379']
```

### 2. Grafana Dashboards

```bash
# Import pre-built dashboards
curl -X POST \
  http://admin:admin@localhost:3001/api/dashboards/db \
  -H 'Content-Type: application/json' \
  -d @monitoring/dashboards/bome-dashboard.json
```

## Backup Configuration

### 1. Automated Backup Script

```bash
#!/bin/bash
# scripts/backup.sh

BACKUP_DIR="/var/backups/bome"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup directory
mkdir -p $BACKUP_DIR

# Database backup
pg_dump -h localhost -U bome_user bome_production | gzip > $BACKUP_DIR/db_$DATE.sql.gz

# Application files backup
tar -czf $BACKUP_DIR/app_$DATE.tar.gz /home/bome/bome --exclude=node_modules --exclude=.git

# Upload to Digital Ocean Spaces
s3cmd put $BACKUP_DIR/db_$DATE.sql.gz s3://bome-backups/database/
s3cmd put $BACKUP_DIR/app_$DATE.tar.gz s3://bome-backups/application/

# Clean up local backups older than 7 days
find $BACKUP_DIR -name "*.gz" -mtime +7 -delete

echo "Backup completed: $DATE"
```

### 2. Cron Job Setup

```bash
# Add to crontab
crontab -e

# Add these lines
# Daily database backup at 2 AM
0 2 * * * /home/bome/scripts/backup.sh

# Weekly full backup on Sunday at 3 AM
0 3 * * 0 /home/bome/scripts/full-backup.sh
```

## CI/CD Pipeline

### 1. GitHub Actions Workflow

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    branches: [production]

jobs:
  deploy:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18'
        
    - name: Setup Go
      uses: actions/setup-go@v3
      with:
        go-version: '1.19'
    
    - name: Run tests
      run: |
        cd frontend && npm test
        cd ../backend && go test ./...
    
    - name: Deploy to server
      uses: appleboy/ssh-action@v0.1.5
      with:
        host: ${{ secrets.HOST }}
        username: ${{ secrets.USERNAME }}
        key: ${{ secrets.SSH_KEY }}
        script: |
          cd /home/bome/bome
          git pull origin production
          docker-compose -f docker-compose.prod.yml down
          docker-compose -f docker-compose.prod.yml up -d --build
```

## Production Checklist

### Pre-Deployment
- [ ] Environment variables configured
- [ ] SSL certificates obtained
- [ ] Database migrations tested
- [ ] Backup procedures tested
- [ ] Monitoring configured
- [ ] Security hardening completed
- [ ] Performance testing completed
- [ ] Load testing completed

### Deployment
- [ ] Services deployed successfully
- [ ] Database migrations applied
- [ ] SSL certificates working
- [ ] Monitoring active
- [ ] Backups scheduled
- [ ] Health checks passing
- [ ] Performance metrics normal

### Post-Deployment
- [ ] Application functionality verified
- [ ] User authentication working
- [ ] Video streaming functional
- [ ] Payment processing working
- [ ] Email notifications working
- [ ] Admin dashboard accessible
- [ ] Monitoring alerts configured
- [ ] Documentation updated

## Troubleshooting

### Common Issues

#### 1. Container Won't Start
```bash
# Check container logs
docker-compose -f docker-compose.prod.yml logs service-name

# Check container status
docker-compose -f docker-compose.prod.yml ps

# Restart specific service
docker-compose -f docker-compose.prod.yml restart service-name
```

#### 2. Database Connection Issues
```bash
# Test database connection
docker exec -it bome-postgres psql -U bome_user -d bome_production

# Check database logs
docker-compose -f docker-compose.prod.yml logs postgres
```

#### 3. SSL Certificate Issues
```bash
# Check certificate status
sudo certbot certificates

# Renew certificates
sudo certbot renew

# Test nginx configuration
sudo nginx -t
```

#### 4. Performance Issues
```bash
# Check system resources
htop
df -h
free -h

# Check container resources
docker stats

# Check application logs
docker-compose -f docker-compose.prod.yml logs -f
```

### Emergency Procedures

#### 1. Rollback Deployment
```bash
# Stop current deployment
docker-compose -f docker-compose.prod.yml down

# Checkout previous version
git checkout previous-stable-tag

# Redeploy
docker-compose -f docker-compose.prod.yml up -d --build
```

#### 2. Database Recovery
```bash
# Restore from backup
gunzip -c /var/backups/bome/db_YYYYMMDD_HHMMSS.sql.gz | psql -U bome_user -d bome_production
```

#### 3. Service Health Check
```bash
#!/bin/bash
# scripts/health-check.sh

# Check if services are running
if ! curl -f http://localhost:3000/health; then
    echo "Frontend service down"
    docker-compose -f docker-compose.prod.yml restart frontend
fi

if ! curl -f http://localhost:8080/health; then
    echo "API service down"
    docker-compose -f docker-compose.prod.yml restart api
fi
```

## Support

For deployment support:
- **Email**: devops@bome.com
- **Documentation**: https://docs.bome.com/deployment
- **Emergency**: +1-555-BOME-911

---

**Last Updated**: December 2024  
**Version**: 1.0.0  
**Maintained By**: BOME DevOps Team 