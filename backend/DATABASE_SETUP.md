# BOME PostgreSQL Database Setup

This document provides comprehensive instructions for setting up the PostgreSQL database for the BOME (Book of Mormon Evidences) application.

## Prerequisites

1. **PostgreSQL Installation**
   - Download and install PostgreSQL from: https://www.postgresql.org/download/windows/
   - Ensure `psql` command-line tool is available in your PATH
   - Default installation includes the `postgres` superuser

2. **Go Environment**
   - Ensure Go is installed and configured
   - Required Go modules are already defined in `go.mod`

## Quick Setup (Automated)

### Option 1: PowerShell Script (Recommended)

1. **Run the setup script:**
   ```powershell
   cd backend
   .\setup-postgres.ps1
   ```

2. **Customize parameters (optional):**
   ```powershell
   .\setup-postgres.ps1 -DatabaseName "bome_prod" -Username "bome_user" -Password "SecurePassword123"
   ```

### Option 2: Manual Setup

If you prefer manual setup or the script fails:

1. **Create database user:**
   ```sql
   -- Connect as postgres superuser
   psql -U postgres
   
   -- Create user and database
   CREATE USER bome_admin WITH PASSWORD 'AdminBOME';
   CREATE DATABASE bome_db OWNER bome_admin;
   GRANT ALL PRIVILEGES ON DATABASE bome_db TO bome_admin;
   ALTER USER bome_admin CREATEDB;
   ```

2. **Run the schema setup:**
   ```bash
   psql -U bome_admin -d bome_db -f setup-database.sql
   ```

## Database Schema Overview

### Core Tables

| Table | Purpose | Key Fields |
|-------|---------|------------|
| `users` | User accounts and authentication | email, password_hash, role, email_verified |
| `videos` | Video content management | title, bunny_video_id, status, category |
| `subscriptions` | User subscription management | stripe_subscription_id, status, billing_period |
| `comments` | Video comments system | video_id, user_id, content, parent_id |
| `likes` | Video likes tracking | user_id, video_id |
| `favorites` | User favorites tracking | user_id, video_id |
| `user_activity` | User activity logging | user_id, activity_type, activity_data |
| `admin_logs` | Administrative action logging | admin_user_id, action, target_type |

### Advertising Tables

| Table | Purpose | Key Fields |
|-------|---------|------------|
| `advertiser_accounts` | Business advertiser accounts | company_name, business_email, status |
| `ad_campaigns` | Advertising campaigns | name, budget, status, billing_type |
| `advertisements` | Individual ads | title, ad_type, status, campaign_id |
| `ad_placements` | Ad placement locations | location, ad_type, base_rate |
| `ad_schedules` | Ad scheduling | start_date, end_date, days_of_week |
| `ad_analytics` | Ad performance metrics | impressions, clicks, revenue |
| `ad_clicks` | Individual click tracking | ad_id, user_id, ip_address |
| `ad_impressions` | Individual impression tracking | ad_id, user_id, view_duration |
| `ad_billing` | Billing records | campaign_id, amount, status |
| `ad_audit_log` | Advertising audit trail | entity_type, action, actor_id |

## Verification

### Run Verification Script

```powershell
.\verify-database.ps1
```

### Manual Verification

1. **Test connection:**
   ```bash
   psql -U bome_admin -d bome_db -c "SELECT version();"
   ```

2. **Check tables:**
   ```sql
   SELECT table_name, table_type 
   FROM information_schema.tables 
   WHERE table_schema = 'public' 
   ORDER BY table_name;
   ```

3. **Check table counts:**
   ```sql
   SELECT 'users' as table_name, COUNT(*) as count FROM users
   UNION ALL
   SELECT 'videos', COUNT(*) FROM videos
   UNION ALL
   SELECT 'ad_placements', COUNT(*) FROM ad_placements;
   ```

## Environment Configuration

The setup script creates a `.env` file with the following structure:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=bome_db
DB_USER=bome_admin
DB_PASSWORD=AdminBOME

# JWT Configuration
JWT_SECRET=]d;SHv1;EL70)-l}ajibeNIKL>j$}:WD
JWT_REFRESH_SECRET=)KeV)cH8NStoq!4%6)xXt7MK7&)Xq*rX

# Server Configuration
PORT=8080
ENVIRONMENT=development

# Email Configuration (update with your email service)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# Stripe Configuration (update with your Stripe keys)
STRIPE_SECRET_KEY=sk_test_your_stripe_secret_key
STRIPE_PUBLISHABLE_KEY=pk_test_your_stripe_publishable_key
STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret

# Bunny.net Configuration (update with your Bunny.net credentials)
BUNNY_STORAGE_ZONE=your-storage-zone
BUNNY_API_KEY=your-api-key
BUNNY_REGION=de

# YouTube API Configuration (update with your YouTube API key)
YOUTUBE_API_KEY=your-youtube-api-key
```

## Performance Optimization

### Indexes

The schema includes comprehensive indexes for optimal performance:

- **User queries**: email, role, email_verified
- **Video queries**: status, category, created_at, bunny_video_id
- **Subscription queries**: user_id, status, stripe_subscription_id
- **Advertising queries**: campaign_id, status, dates, analytics

### Connection Pooling

The Go application uses connection pooling for efficient database connections.

## Security Considerations

1. **Password Security**: Use strong passwords for database users
2. **Network Security**: Configure PostgreSQL to only accept connections from trusted hosts
3. **SSL/TLS**: Enable SSL connections for production environments
4. **Regular Backups**: Implement automated database backups
5. **Access Control**: Limit database user permissions to minimum required

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Ensure PostgreSQL service is running
   - Check if port 5432 is open
   - Verify host and port configuration

2. **Authentication Failed**
   - Check username and password
   - Verify pg_hba.conf configuration
   - Ensure user has proper permissions

3. **Permission Denied**
   - Run setup as postgres superuser
   - Grant necessary permissions to bome_admin user

4. **Schema Creation Failed**
   - Check if database exists
   - Verify user has CREATE privileges
   - Review PostgreSQL error logs

### Useful Commands

```bash
# Check PostgreSQL service status
sc query postgresql-x64-15

# Start PostgreSQL service
net start postgresql-x64-15

# Connect to database
psql -U bome_admin -d bome_db

# List all databases
\l

# List all tables
\dt

# Describe table structure
\d table_name

# Check PostgreSQL logs
tail -f /var/log/postgresql/postgresql-15-main.log
```

## Production Deployment

For production environments:

1. **Use dedicated database server**
2. **Implement connection pooling (e.g., PgBouncer)**
3. **Set up automated backups**
4. **Configure monitoring and alerting**
5. **Use SSL/TLS connections**
6. **Implement proper firewall rules**
7. **Regular security updates**

## Support

If you encounter issues:

1. Check the troubleshooting section above
2. Review PostgreSQL error logs
3. Verify all prerequisites are met
4. Test with the verification script
5. Check the application logs for database-related errors

## Next Steps

After successful database setup:

1. Update the `.env` file with your actual API keys
2. Run the backend application: `go run main.go`
3. Test the API endpoints
4. Create an admin user: `go run cmd/create-admin/main.go`
5. Start the frontend development server 