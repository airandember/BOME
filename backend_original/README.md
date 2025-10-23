# BOME Backend

The BOME (Book of Mormon Evidences) backend is a Go-based API server that provides a comprehensive platform for streaming religious content with advanced features including user management, subscription handling, advertising, and analytics.

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** - [Download here](https://golang.org/dl/)
- **PostgreSQL 14+** - [Download here](https://www.postgresql.org/download/)
- **Git** - [Download here](https://git-scm.com/downloads)

### Environment Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd BOME/backend
   ```

2. **Copy environment template**
   ```bash
   cp env.example .env
   ```

3. **Configure environment variables**
   Edit `.env` file with your database and service credentials:
   ```env
   # Database Configuration
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=bome_admin
   DB_PASSWORD=AdminBOME
   DB_NAME=bome_db
   DB_SSL_MODE=disable

   # Server Configuration
   SERVER_PORT=8080
   ENVIRONMENT=development

   # JWT Configuration
   JWT_SECRET=your-super-secret-jwt-key-here
   JWT_EXPIRY=24h

   # External Services (Optional for development)
   BUNNY_STREAM_API_KEY=your-bunny-stream-api-key
   STRIPE_SECRET_KEY=your-stripe-secret-key
   STRIPE_WEBHOOK_SECRET=your-stripe-webhook-secret
   ```

### Database Setup

#### Option 1: Automated Setup (Recommended)

**Linux/macOS:**
```bash
chmod +x scripts/setup-database.sh
./scripts/setup-database.sh
```

**Windows:**
```cmd
scripts\setup-database.bat
```

#### Option 2: Manual Setup

1. **Create database and user**
   ```sql
   -- Connect as postgres superuser
   CREATE USER bome_admin WITH PASSWORD 'AdminBOME';
   CREATE DATABASE bome_db OWNER bome_admin;
   GRANT ALL PRIVILEGES ON DATABASE bome_db TO bome_admin;
   ALTER USER bome_admin CREATEDB;
   ```

2. **Run the application** (migrations will run automatically)
   ```bash
   go run main.go
   ```

### Running the Application

```bash
# Development mode
go run main.go

# Production build
go build -o bome-backend .
./bome-backend
```

The server will start on `http://localhost:8080` (or the port specified in your environment).

## 🏗️ Architecture

### Database Schema

The application uses a comprehensive PostgreSQL schema with the following main components:

#### Core Tables
- **users** - User accounts and profiles
- **videos** - Video content management
- **subscriptions** - Stripe subscription handling
- **comments** - Video comments system
- **likes/favorites** - User engagement tracking

#### Session Management
- **user_sessions** - Active user sessions with device tracking
- **audit_logs** - Comprehensive security audit logging

#### Advertising System
- **advertiser_accounts** - Advertiser profiles
- **ad_campaigns** - Advertising campaigns
- **advertisements** - Individual ad creatives
- **ad_placements** - Ad placement locations
- **ad_analytics** - Ad performance tracking

#### Analytics & Monitoring
- **user_activity** - User behavior tracking
- **admin_logs** - Administrative action logging

### Migration System

The application uses a unified migration system that:

- ✅ **Automatically runs** when the application starts
- ✅ **Tracks applied migrations** in a `migrations` table
- ✅ **Handles schema evolution** safely with idempotent operations
- ✅ **Includes comprehensive indexes** for optimal performance
- ✅ **Supports rollback scenarios** through version tracking

### Key Features

#### 🔐 Authentication & Authorization
- JWT-based authentication
- Role-based access control (user, admin, super_admin)
- Session management with device tracking
- Rate limiting and security headers

#### 📺 Content Management
- Video upload and streaming via Bunny.net
- YouTube integration for external content
- Content categorization and tagging
- View tracking and analytics

#### 💳 Subscription Management
- Stripe integration for payments
- Multiple subscription tiers
- Automated billing and invoicing
- Subscription lifecycle management

#### 📊 Analytics & Reporting
- User engagement tracking
- Content performance metrics
- Advertising analytics
- Revenue reporting

#### 🎯 Advertising Platform
- Self-service advertiser portal
- Campaign management
- Ad placement optimization
- Performance tracking and billing

## 🔧 Development

### Project Structure

```
backend/
├── cmd/                    # Command-line tools
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   ├── database/         # Database models and migrations
│   ├── middleware/       # HTTP middleware
│   ├── routes/           # API route handlers
│   ├── services/         # Business logic services
│   └── data/            # Data access layer
├── migrations/           # SQL migration files
├── scripts/             # Setup and utility scripts
├── main.go              # Application entry point
└── go.mod               # Go module definition
```

### API Endpoints

#### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/logout` - User logout

#### User Management
- `GET /api/v1/users/profile` - Get user profile
- `PUT /api/v1/users/profile` - Update user profile

#### Content
- `GET /api/v1/videos` - List videos
- `GET /api/v1/videos/:id` - Get video details
- `GET /api/v1/videos/:id/comments` - Get video comments

#### Subscriptions
- `GET /api/v1/subscriptions/plans` - Get subscription plans
- `GET /api/v1/subscriptions/current` - Get current subscription
- `POST /api/v1/subscriptions` - Create subscription
- `POST /api/v1/subscriptions/checkout` - Create checkout session

#### YouTube Integration
- `GET /api/v1/youtube/videos` - List YouTube videos
- `GET /api/v1/youtube/videos/search` - Search YouTube videos
- `GET /api/v1/youtube/videos/:id` - Get YouTube video details

#### Admin (Protected)
- `GET /api/v1/admin/users` - List users
- `GET /api/v1/admin/analytics` - Get analytics data
- `GET /api/v1/admin/advertisements` - Manage advertisements

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./internal/routes -run TestLoginHandler
```

### Database Operations

#### View Applied Migrations
```sql
SELECT * FROM migrations ORDER BY applied_at;
```

#### Check Database Schema
```sql
-- List all tables
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'public' 
ORDER BY table_name;

-- Check table structure
\d+ table_name
```

#### Manual Migration (if needed)
```sql
-- Example: Add a new column
ALTER TABLE users ADD COLUMN IF NOT EXISTS new_field VARCHAR(255);
```

## 🚀 Deployment

### Docker Deployment

```bash
# Build the image
docker build -t bome-backend .

# Run the container
docker run -p 8080:8080 --env-file .env bome-backend
```

### Production Considerations

1. **Environment Variables**
   - Use strong, unique JWT secrets
   - Configure proper database credentials
   - Set up external service API keys

2. **Database**
   - Use connection pooling
   - Configure proper indexes
   - Set up regular backups

3. **Security**
   - Enable HTTPS in production
   - Configure CORS properly
   - Set up rate limiting
   - Use secure headers

4. **Monitoring**
   - Set up logging aggregation
   - Configure health checks
   - Monitor database performance

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is proprietary and confidential. See the [LICENSE](../LICENSE) file for details.

## 🆘 Support

For support and questions:
- Check the [FAQ](../docs/faq/README.md)
- Review the [API documentation](../docs/api/README.md)
- Contact the development team

---

**BOME Development Team** - Building the future of religious content streaming. 