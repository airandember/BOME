# 🚀 BOME Webapp Deployment Guide

**Book of Mormon Evidence Hub - Production Deployment Instructions**

---

## 📋 Table of Contents

1. [Prerequisites](#prerequisites)
2. [System Requirements](#system-requirements)
3. [Pre-Deployment Setup](#pre-deployment-setup)
4. [Configuration](#configuration)
5. [Deployment Steps](#deployment-steps)
6. [Post-Deployment Verification](#post-deployment-verification)
7. [Monitoring & Maintenance](#monitoring--maintenance)
8. [Troubleshooting](#troubleshooting)
9. [Scaling & Performance](#scaling--performance)

---

## 🔧 Prerequisites

### Required Software

- **Docker** (v20.10+) & **Docker Compose** (v2.0+)
- **Git** (v2.30+)
- **Node.js** (v18+) - for local development
- **Go** (v1.24+) - for local development
- **PostgreSQL Client** (v15+) - for database management

### System Access

- **Server Access**: SSH access to production server
- **Domain**: Configured domain name pointing to server
- **SSL Certificate**: Valid SSL certificate for HTTPS
- **External Services**: API keys for required services

### Required API Keys & Credentials

- **Bunny.net**: Storage Zone, API Key, Stream API Key
- **Stripe**: Secret Key, Publishable Key, Webhook Secret
- **Digital Ocean Spaces**: Access Key, Secret Key
- **SendGrid**: API Key (for email notifications)
- **Database**: Secure password for PostgreSQL

---

## 💻 System Requirements

### Minimum Production Requirements

- **CPU**: 4 vCPUs
- **Memory**: 8GB RAM
- **Storage**: 100GB SSD
- **Network**: 1Gbps connection
- **OS**: Ubuntu 22.04 LTS or CentOS 8+

### Recommended Production Requirements

- **CPU**: 8 vCPUs
- **Memory**: 16GB RAM
- **Storage**: 200GB SSD
- **Network**: 10Gbps connection
- **Load Balancer**: NGINX reverse proxy

### Horizontal Scaling Setup

- **3x Backend Instances**: 2 vCPUs, 4GB RAM each
- **Database Master**: 4 vCPUs, 8GB RAM
- **Redis Cache**: 2 vCPUs, 4GB RAM
- **Load Balancer**: 2 vCPUs, 2GB RAM

---

## 🛠️ Pre-Deployment Setup

### 1. Server Preparation

```bash
# Update system packages
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Verify installation
docker --version
docker-compose --version
```

### 2. Project Setup

```bash
# Clone the repository
git clone <your-repository-url>
cd BOME

# Create necessary directories
mkdir -p configs/nginx configs/postgres logs/nginx data/postgres data/redis

# Set proper permissions
sudo chown -R $USER:$USER .
chmod +x deploy.sh
```

### 3. SSL Certificate Setup

```bash
# Install Certbot for Let's Encrypt
sudo apt install certbot python3-certbot-nginx -y

# Generate SSL certificate (replace with your domain)
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com

# Verify certificate
sudo certbot certificates
```

---

## ⚙️ Configuration

### 1. Environment Variables

Create `.env` file in the root directory:

```bash
# Copy example environment file
cp .env.example .env

# Edit environment variables
nano .env
```

**Required Environment Variables:**

```env
# Database Configuration
DB_PASSWORD=your_super_secure_database_password_here
DB_HOST=postgres
DB_PORT=5432
DB_NAME=bome_db
DB_USER=bome_user
DB_SSL_MODE=disable

# JWT Configuration (CRITICAL - Generate secure keys)
JWT_SECRET=your-256-bit-secret-key-change-in-production-immediately
JWT_REFRESH_SECRET=your-different-256-bit-refresh-secret-key

# Bunny.net Video Streaming
BUNNY_STORAGE_ZONE=your-bunny-storage-zone
BUNNY_API_KEY=your-bunny-api-key
BUNNY_STREAM_API_KEY=your-bunny-stream-api-key
BUNNY_STREAM_LIBRARY_ID=your-stream-library-id
BUNNY_PULL_ZONE=your-pull-zone-name
BUNNY_REGION=de

# Stripe Payment Processing
STRIPE_SECRET_KEY=sk_live_your_stripe_secret_key
STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret
PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_live_your_stripe_publishable_key

# Digital Ocean Spaces
DO_SPACES_KEY=your-spaces-access-key
DO_SPACES_SECRET=your-spaces-secret-key
DO_SPACES_ENDPOINT=nyc3.digitaloceanspaces.com
DO_SPACES_BUCKET=your-bucket-name

# SendGrid Email Service
SENDGRID_API_KEY=SG.your_sendgrid_api_key
SENDGRID_FROM_EMAIL=noreply@yourdomain.com
SENDGRID_FROM_NAME=BOME Team

# Application URLs
PUBLIC_API_URL=https://yourdomain.com/api
PUBLIC_ADMIN_API_URL=https://yourdomain.com/api/admin
PUBLIC_APP_URL=https://yourdomain.com

# Application Information
PUBLIC_APP_NAME=BOME
PUBLIC_APP_VERSION=1.0.0
PUBLIC_BUNNY_CDN_URL=https://your-bunny-cdn.b-cdn.net
PUBLIC_GA_TRACKING_ID=G-XXXXXXXXXX
```

### 2. Database Configuration

Verify PostgreSQL configuration in `configs/postgres/postgresql.conf`:

```conf
# Connection Settings
max_connections = 300
shared_buffers = 512MB
effective_cache_size = 2GB
work_mem = 8MB
maintenance_work_mem = 128MB

# Performance Settings
checkpoint_completion_target = 0.9
random_page_cost = 1.1
effective_io_concurrency = 200
max_worker_processes = 8
max_parallel_workers_per_gather = 4
max_parallel_workers = 8
```

### 3. NGINX Configuration

Verify load balancer configuration in `configs/nginx/nginx.conf`:

```nginx
# Backend upstream with load balancing
upstream backend_api {
    least_conn;
    server backend-1:8080 max_fails=3 fail_timeout=30s weight=1;
    server backend-2:8080 max_fails=3 fail_timeout=30s weight=1;
    server backend-3:8080 max_fails=3 fail_timeout=30s weight=1;
    
    # Health check
    keepalive 32;
    keepalive_requests 100;
    keepalive_timeout 60s;
}
```

### 4. Security Configuration

```bash
# Set secure file permissions
chmod 600 .env
chmod 644 configs/nginx/nginx.conf
chmod 644 configs/postgres/postgresql.conf

# Create SSL directory
sudo mkdir -p /etc/nginx/ssl
sudo chmod 700 /etc/nginx/ssl
```

---

## 🚀 Deployment Steps

### Step 1: Pre-Deployment Checks

```bash
# Verify Docker is running
docker info

# Check available disk space
df -h

# Verify environment file
cat .env | grep -v "PASSWORD\|SECRET\|KEY"

# Test database connection (optional)
docker run --rm postgres:15-alpine psql -h your-db-host -U bome_user -d bome_db -c "SELECT 1;"
```

### Step 2: Build and Deploy

```bash
# Make deployment script executable
chmod +x deploy.sh

# Run deployment
./deploy.sh
```

**Manual Deployment (if script fails):**

```bash
# Stop existing containers
docker-compose down

# Remove old images
docker system prune -f

# Build images
docker-compose build --no-cache

# Start services
docker-compose up -d

# Wait for services to initialize
sleep 60

# Check service status
docker-compose ps
```

### Step 3: Database Initialization

```bash
# Check database logs
docker-compose logs postgres

# Run database migrations (if needed)
docker-compose exec backend-1 ./main migrate

# Verify database tables
docker-compose exec postgres psql -U bome_user -d bome_db -c "\dt"
```

### Step 4: SSL and Domain Setup

```bash
# Update NGINX configuration for SSL
sudo nano /etc/nginx/sites-available/bome

# Test NGINX configuration
sudo nginx -t

# Reload NGINX
sudo systemctl reload nginx

# Verify SSL certificate
curl -I https://yourdomain.com
```

---

## ✅ Post-Deployment Verification

### 1. Service Health Checks

```bash
# Check all containers are running
docker-compose ps

# Verify service health
curl -f http://localhost/health
curl -f http://localhost/api/health

# Check logs for errors
docker-compose logs --tail=50 backend-1
docker-compose logs --tail=50 frontend
docker-compose logs --tail=50 postgres
```

### 2. Load Balancer Testing

```bash
# Test load balancing
for i in {1..10}; do
    curl -s http://localhost/api/health | grep -o "backend-[0-9]"
done

# Check NGINX status
curl http://localhost:8080/nginx_status
```

### 3. Database Connectivity

```bash
# Test database connection
docker-compose exec postgres psql -U bome_user -d bome_db -c "SELECT COUNT(*) FROM users;"

# Check database performance
docker-compose exec postgres psql -U bome_user -d bome_db -c "SELECT * FROM pg_stat_activity;"
```

### 4. Application Testing

```bash
# Test API endpoints
curl -X GET http://localhost/api/v1/videos
curl -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"testpass"}'

# Test frontend
curl -I http://localhost/
```

### 5. Performance Verification

```bash
# Monitor resource usage
docker stats

# Check response times
curl -w "@curl-format.txt" -o /dev/null -s http://localhost/api/v1/videos

# Test concurrent connections
ab -n 100 -c 10 http://localhost/api/health
```

---

## 📊 Monitoring & Maintenance

### 1. Log Management

```bash
# View real-time logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f backend-1
docker-compose logs -f nginx

# Log rotation setup
sudo nano /etc/logrotate.d/docker-compose
```

### 2. Database Maintenance

```bash
# Database backup
docker-compose exec postgres pg_dump -U bome_user bome_db > backup_$(date +%Y%m%d).sql

# Database optimization
docker-compose exec postgres psql -U bome_user -d bome_db -c "VACUUM ANALYZE;"

# Check database size
docker-compose exec postgres psql -U bome_user -d bome_db -c "SELECT pg_size_pretty(pg_database_size('bome_db'));"
```

### 3. Performance Monitoring

```bash
# Monitor system resources
htop
iostat -x 1
free -h

# Check Docker container stats
docker stats --no-stream

# Monitor database connections
docker-compose exec postgres psql -U bome_user -d bome_db -c "SELECT count(*) FROM pg_stat_activity;"
```

### 4. Security Updates

```bash
# Update Docker images
docker-compose pull
docker-compose up -d

# Update system packages
sudo apt update && sudo apt upgrade -y

# Check for security vulnerabilities
docker scan bome-backend
```

---

## 🔧 Troubleshooting

### Common Issues

**1. Container Won't Start**
```bash
# Check logs
docker-compose logs [service-name]

# Verify environment variables
docker-compose config

# Check port conflicts
netstat -tulpn | grep :8080
```

**2. Database Connection Issues**
```bash
# Check database status
docker-compose exec postgres pg_isready -U bome_user

# Reset database password
docker-compose exec postgres psql -U postgres -c "ALTER USER bome_user PASSWORD 'new_password';"

# Check database logs
docker-compose logs postgres
```

**3. Load Balancer Issues**
```bash
# Test NGINX configuration
docker-compose exec nginx nginx -t

# Check upstream servers
curl http://localhost:8080/nginx_status

# Restart NGINX
docker-compose restart nginx
```

**4. SSL Certificate Issues**
```bash
# Renew SSL certificate
sudo certbot renew

# Check certificate expiry
sudo certbot certificates

# Test SSL configuration
openssl s_client -connect yourdomain.com:443
```

### Emergency Procedures

**1. Rollback Deployment**
```bash
# Stop current deployment
docker-compose down

# Restore from backup
docker-compose -f docker-compose.backup.yml up -d

# Restore database
docker-compose exec postgres psql -U bome_user -d bome_db < backup_latest.sql
```

**2. Scale Services**
```bash
# Scale backend instances
docker-compose up -d --scale backend=5

# Scale specific service
docker-compose up -d --scale redis=2
```

---

## 📈 Scaling & Performance

### Horizontal Scaling

**Add More Backend Instances:**
```bash
# Update docker-compose.yml to add backend-4, backend-5
# Update nginx.conf to include new backends
# Redeploy
docker-compose up -d --scale backend=5
```

**Database Scaling:**
```bash
# Add read replicas
# Configure master-slave replication
# Update application to use read replicas
```

### Performance Optimization

**1. Database Optimization**
```sql
-- Create indexes for frequently queried columns
CREATE INDEX CONCURRENTLY idx_videos_status_created ON videos(status, created_at);
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);

-- Analyze query performance
EXPLAIN ANALYZE SELECT * FROM videos WHERE status = 'ready';
```

**2. Cache Configuration**
```bash
# Configure Redis for caching
docker-compose exec redis redis-cli CONFIG SET maxmemory 2gb
docker-compose exec redis redis-cli CONFIG SET maxmemory-policy allkeys-lru
```

**3. CDN Integration**
```bash
# Configure Bunny.net CDN
# Update frontend to use CDN URLs
# Enable static asset caching
```

---

## 🔐 Security Checklist

- [ ] **Environment Variables**: All secrets properly configured
- [ ] **Database**: Strong password, limited access
- [ ] **SSL Certificate**: Valid and properly configured
- [ ] **Firewall**: Only necessary ports open
- [ ] **Updates**: All packages and images updated
- [ ] **Backup**: Regular database backups configured
- [ ] **Monitoring**: Error logging and alerting setup
- [ ] **Access Control**: Limited SSH access
- [ ] **API Security**: Rate limiting and authentication enabled

---

## 📞 Support & Maintenance

### Regular Maintenance Tasks

**Daily:**
- Check service health
- Monitor logs for errors
- Verify backup completion

**Weekly:**
- Update Docker images
- Database maintenance
- Security updates

**Monthly:**
- SSL certificate renewal
- Performance review
- Capacity planning

### Emergency Contacts

- **System Administrator**: [Your Contact]
- **Database Administrator**: [Your Contact]
- **Security Team**: [Your Contact]

---

## 📚 Additional Resources

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [NGINX Configuration Guide](https://nginx.org/en/docs/)
- [PostgreSQL Performance Tuning](https://www.postgresql.org/docs/current/performance-tips.html)
- [Let's Encrypt SSL Setup](https://letsencrypt.org/getting-started/)

---

**Last Updated**: $(date)
**Version**: 1.0.0
**Status**: Production Ready ✅ 