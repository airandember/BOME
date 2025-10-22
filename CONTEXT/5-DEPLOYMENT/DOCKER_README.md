# BOME Docker Setup

This document explains how to run the BOME application using Docker Compose.

## Prerequisites

- Docker
- Docker Compose
- Git

## Quick Start

1. **Clone the repository**
   ```bash
   git clone <your-repo-url>
   cd BOME
   ```

2. **Set up environment variables**
   ```bash
   cp env.example .env
   # Edit .env with your actual values
   ```

3. **Start the services**
   ```bash
   docker-compose up -d
   ```

4. **Check service status**
   ```bash
   docker-compose ps
   ```

5. **View logs**
   ```bash
   # All services
   docker-compose logs -f
   
   # Specific service
   docker-compose logs -f backend
   docker-compose logs -f postgres
   docker-compose logs -f frontend
   ```

## Services

### PostgreSQL Database
- **Port**: 5432
- **Database**: bome
- **User**: bome_user
- **Password**: Set via DB_PASSWORD environment variable
- **Data Persistence**: Stored in `postgres_data` volume

### Backend API
- **Port**: 8080
- **Health Check**: http://localhost:8080/health
- **API Base**: http://localhost:8080/api/v1
- **Dependencies**: PostgreSQL

### Frontend
- **Port**: 3000
- **Health Check**: http://localhost:3000
- **Dependencies**: Backend API

## Environment Variables

Copy `env.example` to `.env` and configure:

- **Database**: Connection details for PostgreSQL
- **JWT**: Secret keys for authentication
- **Stripe**: API keys for payment processing
- **Bunny.net**: CDN and storage configuration
- **Digital Ocean Spaces**: Object storage configuration
- **SendGrid**: Email service configuration
- **Admin**: Default admin account credentials

## Database Management

### Run Migrations
```bash
# Connect to PostgreSQL container
docker-compose exec postgres psql -U bome_user -d bome

# Or run migrations from backend
docker-compose exec backend ./main --migrate
```

### Backup Database
```bash
docker-compose exec postgres pg_dump -U bome_user bome > backup.sql
```

### Restore Database
```bash
docker-compose exec -T postgres psql -U bome_user -d bome < backup.sql
```

### Reset Database
```bash
docker-compose down -v
docker-compose up -d postgres
# Wait for PostgreSQL to be ready, then run migrations
```

## Development

### Hot Reload
For development, you can mount source code as volumes:

```bash
# Backend hot reload
docker-compose up -d postgres
go run backend/main.go

# Frontend hot reload
cd frontend
npm run dev
```

### Debug Mode
```bash
# Enable debug logging
LOG_LEVEL=debug docker-compose up
```

## Production Deployment

### Digital Ocean App Platform
1. Set source directories:
   - Backend: `backend`
   - Frontend: `frontend`

2. Configure environment variables in Digital Ocean dashboard

3. Deploy from your GitHub main branch

### Self-Hosted VPS
1. Copy files to your server
2. Set up environment variables
3. Run `docker-compose up -d`
4. Configure reverse proxy (nginx) if needed

## Troubleshooting

### Common Issues

**PostgreSQL connection refused**
- Check if PostgreSQL container is running: `docker-compose ps`
- Verify environment variables in `.env`
- Check logs: `docker-compose logs postgres`

**Backend won't start**
- Ensure PostgreSQL is healthy before starting backend
- Check environment variables
- Verify port 8080 is available

**Frontend build fails**
- Check Node.js version compatibility
- Verify all dependencies are installed
- Check for syntax errors in Svelte components

### Logs and Debugging
```bash
# View all logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f backend

# Access container shell
docker-compose exec backend sh
docker-compose exec postgres psql -U bome_user -d bome
```

### Health Checks
- **PostgreSQL**: `docker-compose exec postgres pg_isready -U bome_user -d bome`
- **Backend**: `curl http://localhost:8080/health`
- **Frontend**: `curl http://localhost:3000`

## Security Notes

- Change default passwords in production
- Use strong JWT secrets
- Enable SSL/TLS in production
- Restrict database access to application only
- Regularly update Docker images
- Monitor logs for suspicious activity

## Performance Tuning

### PostgreSQL
- Adjust `shared_buffers` and `work_mem` based on available RAM
- Enable connection pooling for high traffic
- Regular VACUUM and ANALYZE

### Backend
- Enable response compression
- Implement caching strategies
- Monitor memory usage

### Frontend
- Enable gzip compression
- Implement CDN caching
- Optimize bundle size

## Support

For issues or questions:
1. Check the logs first
2. Verify environment variables
3. Check service health status
4. Review this documentation
5. Create an issue in the repository
