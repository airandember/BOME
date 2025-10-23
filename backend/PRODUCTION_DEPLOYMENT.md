# BOME Backend Production Deployment Guide

## 🚀 Production-Ready Server Startup

This guide ensures your BOME backend runs reliably in production without Go module cache corruption issues.

## 📋 Prerequisites

- Go 1.21+ installed
- PowerShell 5.1+ (Windows) or Bash (Linux/Mac)
- Network access to database and external services

## 🛠️ Quick Start

### Windows (PowerShell)
```powershell
# Navigate to backend directory
cd backend

# Start server in production mode
.\start-server.ps1

# Start server in debug mode
.\start-server.ps1 --debug

# Start server on custom port
.\start-server.ps1 --port 3000
```

### Windows (Command Prompt)
```cmd
# Navigate to backend directory
cd backend

# Start server in production mode
start-server.bat

# Start server in debug mode
start-server.bat --debug

# Start server on custom port
start-server.bat --port 3000
```

### Linux/Mac (Bash)
```bash
# Navigate to backend directory
cd backend

# Make script executable
chmod +x start-server.sh

# Start server in production mode
./start-server.sh

# Start server in debug mode
./start-server.sh --debug
```

## 🔧 Manual Production Setup

If you prefer manual setup or need to troubleshoot:

### 1. Set Local Go Module Cache
```powershell
# Windows PowerShell
$env:GOMODCACHE = "S:\AirEmber\BOME\BOME\backend\.gomodcache"

# Windows Command Prompt
set GOMODCACHE=S:\AirEmber\BOME\BOME\backend\.gomodcache

# Linux/Mac
export GOMODCACHE=/path/to/your/project/backend/.gomodcache
```

### 2. Clean and Rebuild Modules
```bash
# Clean module cache
go clean -modcache

# Download modules
go mod download

# Tidy modules
go mod tidy
```

### 3. Set Production Environment
```powershell
# Windows PowerShell
$env:GO_ENV = "production"
$env:PORT = "8080"
$env:DEBUG = "false"

# Windows Command Prompt
set GO_ENV=production
set PORT=8080
set DEBUG=false

# Linux/Mac
export GO_ENV=production
export PORT=8080
export DEBUG=false
```

### 4. Start Server
```bash
go run main.go
```

## 🏭 Production Environment Variables

Create a `.env` file in the backend directory:

```env
# Server Configuration
GO_ENV=production
PORT=8080
DEBUG=false

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

# Stripe Configuration
STRIPE_SECRET_KEY=sk_live_your_stripe_secret_key
STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret

# Bunny.net Configuration
BUNNY_STREAM_LIBRARY_ID=your_library_id
BUNNY_STREAM_API_KEY=your_api_key
BUNNY_STORAGE_ZONE=your_storage_zone
BUNNY_API_KEY=your_api_key
BUNNY_PULL_ZONE=your_pull_zone

# Email Configuration
SMTP_HOST=smtp.your-provider.com
SMTP_PORT=587
SMTP_USER=your_email@domain.com
SMTP_PASSWORD=your_email_password
```

## 🔍 Health Checks

The server provides several health check endpoints:

- **Basic Health**: `GET /health`
- **Database Health**: `GET /api/v1/admin/health`
- **Analytics Health**: `GET /api/v1/admin/analytics`

## 📊 Monitoring

### Server Status
```bash
# Check if server is running
curl http://localhost:8080/health

# Check analytics endpoint
curl http://localhost:8080/api/v1/admin/analytics
```

### Process Monitoring
```powershell
# Windows - Check if process is running
Get-Process | Where-Object {$_.ProcessName -eq "go"}

# Windows - Check port usage
netstat -an | findstr :8080

# Linux/Mac - Check if process is running
ps aux | grep "go run main.go"

# Linux/Mac - Check port usage
netstat -tulpn | grep :8080
```

## 🚨 Troubleshooting

### Go Module Cache Corruption
If you encounter module cache corruption:

1. **Use the startup script** (recommended):
   ```powershell
   .\start-server.ps1
   ```

2. **Manual fix**:
   ```bash
   # Clean everything
   go clean -modcache
   go clean -cache
   
   # Set local cache
   $env:GOMODCACHE = "S:\AirEmber\BOME\BOME\backend\.gomodcache"
   
   # Rebuild
   go mod download
   go mod tidy
   ```

### Server Won't Start
1. Check if port is already in use:
   ```bash
   netstat -an | findstr :8080
   ```

2. Check Go installation:
   ```bash
   go version
   ```

3. Check database connection:
   ```bash
   # Test database connectivity
   psql -h localhost -p 5432 -U bome_user -d bome_production
   ```

### Performance Issues
1. Monitor server resources:
   ```bash
   # Windows
   Get-Process | Where-Object {$_.ProcessName -eq "go"} | Select-Object CPU, WorkingSet
   
   # Linux/Mac
   top -p $(pgrep -f "go run main.go")
   ```

2. Check database performance:
   ```sql
   -- Check active connections
   SELECT count(*) FROM pg_stat_activity;
   
   -- Check slow queries
   SELECT query, mean_time, calls FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;
   ```

## 🔒 Security Considerations

1. **Environment Variables**: Never commit `.env` files to version control
2. **Database**: Use strong passwords and limit access
3. **API Keys**: Rotate keys regularly
4. **HTTPS**: Use SSL/TLS in production
5. **Firewall**: Restrict access to necessary ports only

## 📈 Scaling

### Horizontal Scaling
- Use load balancer (nginx, HAProxy)
- Multiple server instances
- Database connection pooling
- Redis for session storage

### Vertical Scaling
- Increase server resources
- Optimize database queries
- Use CDN for static assets
- Implement caching strategies

## 🚀 Deployment Checklist

- [ ] Go module cache configured locally
- [ ] Environment variables set
- [ ] Database connection tested
- [ ] Health checks passing
- [ ] SSL certificates configured
- [ ] Firewall rules configured
- [ ] Monitoring setup
- [ ] Backup strategy implemented
- [ ] Log rotation configured
- [ ] Error alerting setup

## 📞 Support

If you encounter issues:

1. Check the logs: `tail -f logs/server.log`
2. Run health checks: `curl http://localhost:8080/health`
3. Verify environment variables
4. Check database connectivity
5. Review this guide for troubleshooting steps

---

**Remember**: The startup scripts handle Go module cache management automatically, preventing the corruption issues you've experienced. Always use them for production deployments!
