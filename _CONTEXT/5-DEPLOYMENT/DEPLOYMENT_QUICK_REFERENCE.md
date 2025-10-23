# ⚡ BOME Deployment Quick Reference

**Essential commands for BOME webapp deployment and maintenance**

---

## 🚀 Quick Deploy Commands

```bash
# Full deployment
./deploy.sh

# Manual deployment
docker-compose down && docker-compose up -d --build

# Check status
docker-compose ps

# View logs
docker-compose logs -f
```

---

## 🔧 Essential Commands

### Container Management
```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# Restart specific service
docker-compose restart backend-1

# Scale backend instances
docker-compose up -d --scale backend=5

# Remove all containers and volumes
docker-compose down -v
```

### Health Checks
```bash
# Service health
curl http://localhost/health

# Load balancer status
curl http://localhost:8080/nginx_status

# Database connection
docker-compose exec postgres pg_isready -U bome_user

# Check all containers
docker-compose ps
```

### Log Management
```bash
# View all logs
docker-compose logs -f

# Specific service logs
docker-compose logs -f backend-1
docker-compose logs -f nginx
docker-compose logs -f postgres

# Last 50 lines
docker-compose logs --tail=50 backend-1
```

---

## 🐛 Troubleshooting

### Container Issues
```bash
# Check container status
docker-compose ps

# Inspect container
docker inspect bome-backend-1

# Execute command in container
docker-compose exec backend-1 /bin/sh

# Check container resources
docker stats
```

### Database Issues
```bash
# Connect to database
docker-compose exec postgres psql -U bome_user -d bome_db

# Check database size
docker-compose exec postgres psql -U bome_user -d bome_db -c "SELECT pg_size_pretty(pg_database_size('bome_db'));"

# Check active connections
docker-compose exec postgres psql -U bome_user -d bome_db -c "SELECT count(*) FROM pg_stat_activity;"

# Reset database password
docker-compose exec postgres psql -U postgres -c "ALTER USER bome_user PASSWORD 'new_password';"
```

### Network Issues
```bash
# Check port usage
netstat -tulpn | grep :8080

# Test connectivity
curl -I http://localhost/
curl -I http://localhost/api/health

# Check Docker networks
docker network ls
docker network inspect bome_bome-network
```

---

## 📊 Monitoring Commands

### Performance Monitoring
```bash
# System resources
htop
free -h
df -h

# Container stats
docker stats --no-stream

# Database performance
docker-compose exec postgres psql -U bome_user -d bome_db -c "SELECT * FROM pg_stat_activity;"
```

### Load Testing
```bash
# Simple load test
ab -n 100 -c 10 http://localhost/api/health

# Test load balancing
for i in {1..10}; do curl -s http://localhost/api/health; done
```

---

## 🔒 Security Commands

### SSL Certificate
```bash
# Check SSL certificate
sudo certbot certificates

# Renew certificate
sudo certbot renew

# Test SSL
openssl s_client -connect yourdomain.com:443
```

### Updates
```bash
# Update Docker images
docker-compose pull

# Update system packages
sudo apt update && sudo apt upgrade -y

# Security scan
docker scan bome-backend
```

---

## 💾 Backup & Recovery

### Database Backup
```bash
# Create backup
docker-compose exec postgres pg_dump -U bome_user bome_db > backup_$(date +%Y%m%d).sql

# Restore backup
docker-compose exec postgres psql -U bome_user -d bome_db < backup_20240101.sql

# Automated backup script
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
docker-compose exec postgres pg_dump -U bome_user bome_db > "backups/backup_${DATE}.sql"
```

### Volume Backup
```bash
# Backup Docker volumes
docker run --rm -v bome_postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/postgres_backup.tar.gz /data

# Restore Docker volumes
docker run --rm -v bome_postgres_data:/data -v $(pwd):/backup alpine tar xzf /backup/postgres_backup.tar.gz -C /
```

---

## 🚨 Emergency Procedures

### Service Recovery
```bash
# Restart all services
docker-compose restart

# Force recreate containers
docker-compose up -d --force-recreate

# Rollback to previous version
git checkout HEAD~1
docker-compose up -d --build
```

### Database Recovery
```bash
# Stop services
docker-compose stop

# Restore from backup
docker-compose exec postgres psql -U bome_user -d bome_db < backup_latest.sql

# Start services
docker-compose start
```

---

## 📈 Scaling Commands

### Add Backend Instances
```bash
# Scale to 5 backend instances
docker-compose up -d --scale backend=5

# Update nginx.conf to include new backends
# server backend-4:8080 max_fails=3 fail_timeout=30s weight=1;
# server backend-5:8080 max_fails=3 fail_timeout=30s weight=1;

# Restart nginx
docker-compose restart nginx
```

### Database Scaling
```bash
# Add read replica (manual setup required)
# Update application configuration
# Restart backend services
docker-compose restart backend-1 backend-2 backend-3
```

---

## 🔍 Debugging Commands

### Application Debugging
```bash
# Check environment variables
docker-compose exec backend-1 env

# Test API endpoints
curl -X GET http://localhost/api/v1/videos
curl -X POST http://localhost/api/v1/auth/login -H "Content-Type: application/json" -d '{"email":"test@example.com","password":"testpass"}'

# Check configuration
docker-compose config
```

### Network Debugging
```bash
# Test internal connectivity
docker-compose exec backend-1 ping postgres
docker-compose exec backend-1 ping redis

# Check DNS resolution
docker-compose exec backend-1 nslookup postgres
```

---

## 📋 Pre-Deployment Checklist

- [ ] `.env` file configured with all required variables
- [ ] SSL certificate valid and configured
- [ ] Database backup completed
- [ ] Docker and Docker Compose installed
- [ ] Sufficient disk space available
- [ ] All required ports available
- [ ] External services (Bunny.net, Stripe) configured

---

## 🆘 Common Error Solutions

### "Port already in use"
```bash
# Find process using port
sudo netstat -tulpn | grep :8080
sudo lsof -i :8080

# Kill process
sudo kill -9 <PID>
```

### "Container health check failed"
```bash
# Check container logs
docker-compose logs backend-1

# Check health endpoint
curl http://localhost:8080/health

# Restart container
docker-compose restart backend-1
```

### "Database connection refused"
```bash
# Check database status
docker-compose exec postgres pg_isready

# Check database logs
docker-compose logs postgres

# Restart database
docker-compose restart postgres
```

### "SSL certificate expired"
```bash
# Renew certificate
sudo certbot renew

# Check certificate status
sudo certbot certificates

# Restart nginx
sudo systemctl restart nginx
```

---

## 📞 Emergency Contacts

- **System Administrator**: [Your Contact]
- **Database Administrator**: [Your Contact]
- **Security Team**: [Your Contact]

---

**Last Updated**: $(date)
**Version**: 1.0.0 