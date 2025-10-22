# BOME Project Structure

**Comprehensive Project Organization and Architecture Guide**

## Overview

BOME (Book of Mormon Evidences) is a modern, full-stack streaming platform designed for scalability, maintainability, and performance. The project follows a modular architecture with clear separation of concerns and modern development practices.

### Core Components
- **Backend**: Go API server with microservices architecture
- **Frontend**: SvelteKit web application with custom design system
- **Database**: PostgreSQL with Redis caching layer
- **CDN**: Bunny.net for video streaming and static assets
- **Infrastructure**: Containerized deployment on Digital Ocean
- **Documentation**: Comprehensive technical and user documentation

---

## 🏗️ Directory Structure

```
BOME/
├── backend/                 # Go backend API server
│   ├── cmd/                # Application entry points
│   ├── internal/           # Private application code
│   │   ├── api/           # API handlers and routes
│   │   ├── auth/          # Authentication logic
│   │   ├── database/      # Database models and migrations
│   │   ├── middleware/    # HTTP middleware
│   │   ├── services/      # Business logic services
│   │   └── utils/         # Utility functions
│   ├── pkg/               # Public packages
│   ├── configs/           # Configuration files
│   ├── scripts/           # Build and deployment scripts
│   ├── tests/             # Test files
│   ├── go.mod             # Go module file
│   └── go.sum             # Go dependencies checksum
│
├── frontend/               # Svelte web application
│   ├── src/
│   │   ├── lib/           # Shared components and utilities
│   │   │   ├── components/ # Reusable UI components
│   │   │   ├── stores/     # Svelte stores
│   │   │   ├── utils/      # Utility functions
│   │   │   └── api/        # API client functions
│   │   ├── routes/         # SvelteKit routes
│   │   │   ├── auth/       # Authentication pages
│   │   │   ├── videos/     # Video pages
│   │   │   ├── profile/    # User profile pages
│   │   │   └── admin/      # Admin pages
│   │   ├── app.css         # Global styles with Tailwind
│   │   ├── app.html        # HTML template
│   │   └── app.d.ts        # TypeScript declarations
│   ├── static/             # Static assets
│   ├── package.json        # Node.js dependencies
│   ├── tailwind.config.js  # Tailwind CSS configuration
│   ├── svelte.config.js    # SvelteKit configuration
│   └── vite.config.ts      # Vite configuration
│
├── admin-dashboard/         # Admin management interface
│   ├── src/
│   │   ├── components/     # Admin UI components
│   │   ├── pages/          # Admin pages
│   │   ├── services/       # Admin API services
│   │   └── utils/          # Admin utilities
│   ├── package.json        # Admin dependencies
│   └── README.md           # Admin documentation
│
├── roku-app/               # Roku streaming application
│   ├── components/         # Roku UI components
│   ├── source/             # BrightScript source code
│   ├── images/             # App images and assets
│   ├── manifest            # Roku app manifest
│   └── README.md           # Roku app documentation
│
├── docs/                   # Project documentation
│   ├── api/                # API documentation
│   ├── deployment/         # Deployment guides
│   ├── development/        # Development guides
│   └── user/               # User documentation
│
├── scripts/                # Build and deployment scripts
│   ├── build.sh            # Build script
│   ├── deploy.sh           # Deployment script
│   └── setup.sh            # Setup script
│
├── configs/                # Configuration files
│   ├── docker/             # Docker configurations
│   ├── nginx/              # Nginx configurations
│   └── ssl/                # SSL certificates
│
├── .github/                # GitHub workflows
│   └── workflows/          # CI/CD workflows
│
├── PROJECT_TASK_LIST.md    # Project task tracking
├── GIT_WORKFLOW.md         # Git workflow documentation
├── PROJECT_STRUCTURE.md    # This file
├── README.md               # Project overview
└── .gitignore              # Git ignore rules
```

## Technology Stack

### Backend (Go)
- **Framework**: Gin or Echo for HTTP server
- **Database**: PostgreSQL with GORM
- **Cache**: Redis
- **Authentication**: JWT tokens
- **File Storage**: Digital Ocean Spaces
- **Video Streaming**: Bunny.net CDN
- **Payments**: Stripe integration

### Frontend (Svelte)
- **Framework**: SvelteKit
- **Styling**: Tailwind CSS with neumorphic design
- **State Management**: Svelte stores
- **HTTP Client**: Fetch API or Axios
- **Build Tool**: Vite
- **Testing**: Playwright

### Admin Dashboard
- **Framework**: SvelteKit or React
- **UI Library**: Tailwind CSS
- **Charts**: Chart.js or D3.js
- **Tables**: TanStack Table
- **Forms**: React Hook Form or similar

### Roku App
- **Language**: BrightScript
- **UI Framework**: SceneGraph
- **Video Player**: Roku Video Player
- **Authentication**: Device-based auth
- **Analytics**: Roku Analytics

## Development Workflow

### 1. Backend Development
```bash
cd backend
go mod tidy
go run cmd/server/main.go
```

### 2. Frontend Development
```bash
cd frontend
npm install
npm run dev
```

### 3. Admin Dashboard Development
```bash
cd admin-dashboard
npm install
npm run dev
```

### 4. Roku App Development
- Use Roku VS Code extension
- Test on Roku device or simulator
- Submit to Roku app store

## Environment Variables

### Backend (.env)
```env
# Database
DATABASE_URL=postgresql://user:password@localhost:5432/bome_db

# Redis
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=your-jwt-secret

# Bunny.net
BUNNY_STORAGE_ZONE=your-storage-zone
BUNNY_API_KEY=your-api-key

# Stripe
STRIPE_SECRET_KEY=your-stripe-secret
STRIPE_WEBHOOK_SECRET=your-webhook-secret

# Digital Ocean
DO_SPACES_KEY=your-spaces-key
DO_SPACES_SECRET=your-spaces-secret
DO_SPACES_ENDPOINT=your-spaces-endpoint

# Email
SMTP_HOST=your-smtp-host
SMTP_PORT=587
SMTP_USER=your-smtp-user
SMTP_PASS=your-smtp-password
```

### Frontend (.env)
```env
# API
PUBLIC_API_URL=http://localhost:8080/api

# Stripe
PUBLIC_STRIPE_PUBLISHABLE_KEY=your-stripe-publishable-key

# Analytics
PUBLIC_GA_TRACKING_ID=your-ga-tracking-id
```

## Deployment

### Production Environment
- **Server**: Digital Ocean Droplet
- **Database**: Digital Ocean Managed PostgreSQL
- **Cache**: Digital Ocean Managed Redis
- **Storage**: Digital Ocean Spaces
- **CDN**: Bunny.net
- **SSL**: Let's Encrypt
- **Reverse Proxy**: Nginx

### CI/CD Pipeline
- **Source Control**: GitHub
- **Build**: GitHub Actions
- **Testing**: Automated tests
- **Deployment**: Automated deployment to Digital Ocean
- **Monitoring**: Application monitoring and logging

## Security Considerations

### Backend Security
- JWT token authentication
- Rate limiting
- Input validation and sanitization
- CORS configuration
- HTTPS enforcement
- SQL injection prevention
- XSS protection

### Frontend Security
- Content Security Policy (CSP)
- HTTPS enforcement
- Input validation
- XSS protection
- CSRF protection

### Data Protection
- Data encryption at rest
- Data encryption in transit
- GDPR compliance
- Privacy policy
- Terms of service

## Performance Optimization

### Backend Optimization
- Database query optimization
- Redis caching
- CDN for static assets
- API response compression
- Connection pooling

### Frontend Optimization
- Code splitting
- Lazy loading
- Image optimization
- CSS/JS minification
- Service worker for caching

### Video Streaming Optimization
- Adaptive bitrate streaming
- Multiple quality levels
- CDN distribution
- Video compression
- Thumbnail generation 