# BOME Infrastructure Setup Guide

## Overview
This guide covers the complete infrastructure setup for the BOME streaming platform, including Digital Ocean, domain configuration, SSL certificates, and deployment automation.

## 1. Digital Ocean Setup

### 1.1 Create Digital Ocean Account
- Sign up at [Digital Ocean](https://digitalocean.com)
- Add payment method
- Verify account

### 1.2 Create Droplet
```bash
# Recommended specifications for BOME streaming platform
- **Size**: Basic Plan - $12/month (2GB RAM, 1 CPU, 50GB SSD)
- **Region**: Choose closest to your target audience
- **Image**: Ubuntu 22.04 LTS
- **Authentication**: SSH Key (recommended) or Password
- **Hostname**: bome-streaming-server
- **Tags**: production, streaming, api
```

### 1.3 Initial Server Setup
```bash
# Connect to your droplet
ssh root@your-server-ip

# Update system
apt update && apt upgrade -y

# Create non-root user
adduser bome
usermod -aG sudo bome

# Set up SSH key authentication
mkdir -p /home/bome/.ssh
cp ~/.ssh/authorized_keys /home/bome/.ssh/
chown -R bome:bome /home/bome/.ssh
chmod 700 /home/bome/.ssh
chmod 600 /home/bome/.ssh/authorized_keys

# Disable root SSH login
sed -i 's/PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config
systemctl restart sshd
```

## 2. Domain and DNS Configuration

### 2.1 Domain Registration
- Register domain (e.g., `bome-streaming.com`)
- Recommended registrars: Namecheap, Cloudflare, or Digital Ocean

### 2.2 DNS Configuration
```bash
# A Records
@           -> Your server IP
www         -> Your server IP
api         -> Your server IP
admin       -> Your server IP

# CNAME Records
*.bome-streaming.com -> bome-streaming.com
```

### 2.3 Cloudflare Setup (Optional but Recommended)
- Add domain to Cloudflare
- Enable proxy (orange cloud)
- Configure SSL/TLS encryption mode: Full (strict)
- Enable Always Use HTTPS
- Configure Page Rules for caching

## 3. SSL Certificate Setup

### 3.1 Install Certbot
```bash
# Install Certbot
apt install certbot python3-certbot-nginx -y

# Get SSL certificate
certbot --nginx -d bome-streaming.com -d www.bome-streaming.com -d api.bome-streaming.com -d admin.bome-streaming.com

# Auto-renewal
crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

## 4. Nginx Configuration

### 4.1 Install Nginx
```bash
apt install nginx -y
systemctl enable nginx
systemctl start nginx
```

### 4.2 Main Site Configuration
```nginx
# /etc/nginx/sites-available/bome-streaming.com
server {
    listen 80;
    server_name bome-streaming.com www.bome-streaming.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name bome-streaming.com www.bome-streaming.com;

    ssl_certificate /etc/letsencrypt/live/bome-streaming.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/bome-streaming.com/privkey.pem;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;

    # Frontend
    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

### 4.3 API Configuration
```nginx
# /etc/nginx/sites-available/api.bome-streaming.com
server {
    listen 80;
    server_name api.bome-streaming.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.bome-streaming.com;

    ssl_certificate /etc/letsencrypt/live/api.bome-streaming.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.bome-streaming.com/privkey.pem;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;

    # API
    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

### 4.4 Admin Dashboard Configuration
```nginx
# /etc/nginx/sites-available/admin.bome-streaming.com
server {
    listen 80;
    server_name admin.bome-streaming.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name admin.bome-streaming.com;

    ssl_certificate /etc/letsencrypt/live/admin.bome-streaming.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/admin.bome-streaming.com/privkey.pem;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;

    # Admin Dashboard
    location / {
        proxy_pass http://localhost:3001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

## 5. Database Setup

### 5.1 PostgreSQL Installation
```bash
# Install PostgreSQL
apt install postgresql postgresql-contrib -y

# Start and enable PostgreSQL
systemctl start postgresql
systemctl enable postgresql

# Create database and user
sudo -u postgres psql
CREATE DATABASE bome_db;
CREATE USER bome_user WITH PASSWORD 'your-secure-password';
GRANT ALL PRIVILEGES ON DATABASE bome_db TO bome_user;
\q
```

### 5.2 Redis Installation
```bash
# Install Redis
apt install redis-server -y

# Configure Redis
sed -i 's/bind 127.0.0.1/bind 127.0.0.1/' /etc/redis/redis.conf
systemctl restart redis-server
systemctl enable redis-server
```

## 6. Firewall Configuration

### 6.1 UFW Setup
```bash
# Install UFW
apt install ufw -y

# Default policies
ufw default deny incoming
ufw default allow outgoing

# Allow SSH
ufw allow ssh

# Allow HTTP and HTTPS
ufw allow 80
ufw allow 443

# Enable firewall
ufw enable
```

## 7. Monitoring and Logging

### 7.1 Install Monitoring Tools
```bash
# Install htop for system monitoring
apt install htop -y

# Install logrotate
apt install logrotate -y

# Install fail2ban for security
apt install fail2ban -y
```

### 7.2 Fail2ban Configuration
```bash
# Configure fail2ban
cp /etc/fail2ban/jail.conf /etc/fail2ban/jail.local

# Edit jail.local to add custom rules
nano /etc/fail2ban/jail.local

# Restart fail2ban
systemctl restart fail2ban
```

## 8. Backup Strategy

### 8.1 Automated Backups
```bash
# Create backup script
nano /home/bome/backup.sh

#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/home/bome/backups"

# Create backup directory
mkdir -p $BACKUP_DIR

# Database backup
pg_dump bome_db > $BACKUP_DIR/db_backup_$DATE.sql

# Application backup
tar -czf $BACKUP_DIR/app_backup_$DATE.tar.gz /home/bome/app

# Keep only last 7 days of backups
find $BACKUP_DIR -name "*.sql" -mtime +7 -delete
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete
```

### 8.2 Setup Cron Job
```bash
# Add to crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * /home/bome/backup.sh
```

## 9. Performance Optimization

### 9.1 Nginx Optimization
```nginx
# Add to nginx.conf
worker_processes auto;
worker_connections 1024;
keepalive_timeout 65;
gzip on;
gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;
```

### 9.2 System Optimization
```bash
# Optimize system settings
echo 'vm.swappiness=10' >> /etc/sysctl.conf
echo 'net.core.somaxconn=65536' >> /etc/sysctl.conf
sysctl -p
```

## 10. Security Hardening

### 10.1 SSH Security
```bash
# Change SSH port (optional)
sed -i 's/#Port 22/Port 2222/' /etc/ssh/sshd_config

# Disable password authentication
sed -i 's/PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config

# Restart SSH
systemctl restart sshd
```

### 10.2 Regular Security Updates
```bash
# Setup automatic security updates
apt install unattended-upgrades -y
dpkg-reconfigure -plow unattended-upgrades
```

## Next Steps
1. Deploy the application code
2. Set up CI/CD pipeline
3. Configure monitoring and alerting
4. Set up load balancing (if needed)
5. Implement CDN configuration

---

*Remember to replace placeholder values with actual configuration data!* 