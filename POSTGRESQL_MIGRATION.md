# PostgreSQL Migration Guide

**BOME Project Database Migration: SQLite → PostgreSQL**

This guide covers the complete migration from SQLite to PostgreSQL for production readiness.

## 🎯 Overview

The BOME project has been successfully migrated from SQLite to PostgreSQL to improve:
- **Production Scalability**: PostgreSQL handles concurrent connections better
- **Data Integrity**: Better ACID compliance and constraint enforcement  
- **Performance**: Advanced indexing and query optimization
- **Features**: JSON support, advanced data types (INET, ARRAY), and more

## 📋 What Changed

### Database Engine
- **Before**: SQLite with single file database
- **After**: PostgreSQL with proper client-server architecture

### SQL Syntax Updates
- `INTEGER PRIMARY KEY AUTOINCREMENT` → `SERIAL PRIMARY KEY`
- `TEXT` → `VARCHAR(255)` or `TEXT` (for unlimited text)
- `DATETIME` → `TIMESTAMP`
- `REAL` → `DECIMAL(10,2)`
- SQLite parameters `?` → PostgreSQL parameters `$1, $2, etc.`

### Data Types Enhanced
- IP addresses: `TEXT` → `INET` 
- Arrays: `TEXT` → `TEXT[]` (native array support)
- JSON data: `TEXT` → `JSONB` (binary JSON)

## 🚀 Quick Start

### Option 1: Automated Setup (Recommended)

**Linux/macOS:**
```bash
cd backend
../scripts/setup-postgres.sh
```

**Windows:**
```cmd
cd backend
..\scripts\setup-postgres.bat
```

### Option 2: Manual Setup

1. **Install PostgreSQL**
   ```bash
   # Ubuntu/Debian
   sudo apt-get install postgresql postgresql-contrib
   
   # macOS
   brew install postgresql
   
   # Windows
   # Download from https://www.postgresql.org/download/windows/
   ```

2. **Create Database and User**
   ```sql
   sudo -u postgres psql
   CREATE USER bome_user WITH PASSWORD 'your_secure_password';
   CREATE DATABASE bome_streaming OWNER bome_user;
   GRANT ALL PRIVILEGES ON DATABASE bome_streaming TO bome_user;
   ALTER USER bome_user CREATEDB;
   \q
   ```

3. **Configure Environment**
   ```bash
   cd backend
   cp env.example .env
   # Edit .env with your database credentials
   ```

4. **Build and Run**
   ```bash
   go mod tidy
   go build -o bome-backend ./main.go
   ./bome-backend
   ```

## 📁 Files Changed

### Core Database Files
- `backend/internal/database/database.go` - Main database connection
- `backend/internal/database/advertisement.go` - Advertisement tables
- `backend/go.mod` - Dependencies updated
- `backend/main.go` - Database initialization

### Configuration Files
- `backend/env.example` - PostgreSQL configuration
- `docker-compose.yml` - Already configured for PostgreSQL
- `scripts/setup-postgres.sh` - Linux/macOS setup script
- `scripts/setup-postgres.bat` - Windows setup script

## 🔧 Database Schema

### Core Tables (PostgreSQL)
```sql
-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Videos table  
CREATE TABLE videos (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    bunny_video_id VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) DEFAULT 'processing',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Subscriptions table
CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    stripe_subscription_id VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Advertisement Tables
- `advertiser_accounts` - Advertiser account information
- `ad_campaigns` - Advertisement campaigns
- `advertisements` - Individual ads
- `ad_placements` - Ad placement locations
- `ad_schedules` - Ad scheduling
- `ad_analytics` - Performance metrics
- `ad_billing` - Billing and invoicing

## 🔄 Migration Process

### For Existing SQLite Data

If you have existing SQLite data to migrate:

1. **Export SQLite Data**
   ```bash
   sqlite3 ./data/bome.db ".dump" > sqlite_dump.sql
   ```

2. **Convert SQL Syntax**
   ```bash
   # Use a script to convert SQLite syntax to PostgreSQL
   sed -i 's/INTEGER PRIMARY KEY AUTOINCREMENT/SERIAL PRIMARY KEY/g' sqlite_dump.sql
   sed -i 's/DATETIME/TIMESTAMP/g' sqlite_dump.sql
   # Manual review recommended for complex conversions
   ```

3. **Import to PostgreSQL**
   ```bash
   psql -h localhost -U bome_user -d bome_streaming < sqlite_dump.sql
   ```

## 🐳 Docker Setup

The project already includes PostgreSQL in Docker Compose:

```bash
# Start all services including PostgreSQL
docker-compose up -d

# View logs
docker-compose logs postgres

# Connect to database
docker-compose exec postgres psql -U bome_user -d bome_streaming
```

## 🔍 Verification

### Test Database Connection
```bash
# Using psql
psql -h localhost -U bome_user -d bome_streaming -c "SELECT version();"

# Using the application
./bome-backend
# Look for: "PostgreSQL database connection established"
```

### Check Migrations
```sql
-- View applied migrations
SELECT * FROM migrations ORDER BY applied_at;

-- Check table creation
\dt
```

## 🚨 Troubleshooting

### Common Issues

**Connection refused:**
```bash
# Check if PostgreSQL is running
sudo systemctl status postgresql
# Start if needed
sudo systemctl start postgresql
```

**Authentication failed:**
```bash
# Check pg_hba.conf for authentication settings
sudo nano /etc/postgresql/*/main/pg_hba.conf
# Ensure 'local' connections use 'md5' or 'password'
```

**Permission denied:**
```sql
-- Grant necessary permissions
GRANT ALL PRIVILEGES ON DATABASE bome_streaming TO bome_user;
GRANT ALL ON ALL TABLES IN SCHEMA public TO bome_user;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO bome_user;
```

### Performance Tuning

**Optimize PostgreSQL settings** in `/etc/postgresql/*/main/postgresql.conf`:
```ini
# Memory settings
shared_buffers = 256MB
work_mem = 4MB
maintenance_work_mem = 64MB

# Connection settings
max_connections = 100

# Checkpoint settings
checkpoint_completion_target = 0.9
wal_buffers = 16MB
```

## 📊 Performance Benefits

### Before (SQLite)
- Single-threaded database access
- File-based storage with locking issues
- Limited concurrent connections
- Basic data types only

### After (PostgreSQL)
- Multi-threaded, concurrent access
- Client-server architecture
- Advanced indexing and query optimization
- Rich data types (INET, JSONB, Arrays)
- Better constraint enforcement
- Production-ready replication and backup

## 🛡️ Security Improvements

### Database Security
- User-based access control (vs file permissions)
- Connection-level SSL/TLS support
- Row-level security capabilities
- Advanced authentication methods

### Application Security
- Prepared statements with proper parameter binding
- Better SQL injection protection
- Connection pooling with secure credentials

## 📈 Next Steps

1. **Production Deployment**
   - Set up managed PostgreSQL (AWS RDS, Digital Ocean, etc.)
   - Configure SSL connections
   - Set up automated backups

2. **Monitoring**
   - Install pg_stat_statements extension
   - Set up monitoring with tools like pgAdmin or Grafana
   - Configure log analysis

3. **Performance Optimization**
   - Analyze query performance with EXPLAIN
   - Add appropriate indexes based on usage patterns
   - Configure connection pooling (PgBouncer)

## 🎉 Conclusion

The migration to PostgreSQL significantly improves the production readiness of the BOME project by providing:

- ✅ **Scalability**: Handle multiple concurrent users
- ✅ **Reliability**: ACID compliance and data integrity
- ✅ **Performance**: Advanced query optimization
- ✅ **Features**: Rich data types and advanced SQL features
- ✅ **Production Ready**: Enterprise-grade database system

The application now uses modern database practices suitable for production deployment.

---

**Need Help?** 
- Check the troubleshooting section above
- Review PostgreSQL documentation: https://www.postgresql.org/docs/
- Open an issue in the project repository 