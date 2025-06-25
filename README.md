# BOME - Book of Mormon Evidence Hub
**A Comprehensive Streaming Platform for Book of Mormon Research and Education**

[![License: All Rights Reserved](https://img.shields.io/badge/License-All%20Rights%20Reserved-red.svg)](./LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24.3-blue.svg)](https://golang.org/)
[![Svelte](https://img.shields.io/badge/Svelte-5.0-orange.svg)](https://svelte.dev/)
[![Status](https://img.shields.io/badge/Status-Development-green.svg)](https://github.com/yourusername/bome)

## 🎯 Project Overview

BOME (Book of Mormon Evidences) is a sophisticated full-stack streaming platform designed as a comprehensive hub for Book of Mormon research, education, and community engagement. The platform features three integrated subsites providing diverse content and services for scholars, researchers, and the general public.

### 🏛️ Three Core Subsites

#### 📚 Articles Hub (`/articles`)
- **Purpose**: Comprehensive research articles and scholarly content
- **Features**: 18 curated articles, 8 categories, 25 research tags, 6 expert authors
- **Content**: Archaeological evidence, linguistic analysis, historical research, scientific studies
- **Status**: ✅ **COMPLETE** - Fully functional with search, filtering, and author profiles

#### 🎥 Streaming Platform (`/videos` & `/youtube`)
- **Purpose**: Educational video content and live streaming
- **Features**: Bunny.net CDN integration, YouTube channel integration, advanced video player
- **Content**: 25+ videos with HLS streaming, categories, comments, analytics
- **Status**: ✅ **COMPLETE** - Production-ready with seamless API transition path

#### 🎪 Events & Tours (`/events`)
- **Purpose**: Educational events, conferences, and guided tours
- **Features**: Event registration, ticketing, venue management, speaker coordination
- **Content**: Academic conferences, site tours, workshops, lectures
- **Status**: 🔄 **PLANNED** - UI complete, backend integration pending

### 🎯 Target Audience
- **Academic Researchers**: Scholars studying Book of Mormon historicity
- **Educational Institutions**: Universities, seminaries, and religious schools
- **General Public**: Individuals interested in Book of Mormon evidence and research
- **Content Creators**: Researchers and educators contributing original content

## 🛠️ Technology Stack

### Frontend Architecture
- **Framework**: SvelteKit 2.16.0 with TypeScript
- **Styling**: Custom CSS design system with glass morphism effects
- **State Management**: Svelte stores with intelligent caching
- **Build Tool**: Vite 6.2.6 with optimized bundling
- **Testing**: Vitest 3.2.4 with comprehensive test coverage

### Backend Architecture
- **Language**: Go 1.24.3 with modern patterns
- **Framework**: Gin web framework with middleware
- **Database**: SQLite (development) / PostgreSQL (production)
- **Caching**: Redis with intelligent cache management
- **Authentication**: JWT with role-based access control

### Infrastructure & Services
- **Video Streaming**: Bunny.net CDN with HLS support
- **Payments**: Stripe integration for subscriptions
- **Cloud Storage**: Digital Ocean Spaces for backups
- **Email Service**: SendGrid for notifications
- **Analytics**: Custom analytics system with real-time tracking
- **Deployment**: Docker containers with Nginx reverse proxy

### Third-Party Integrations
- **YouTube Data API v3**: Production-ready integration path
- **Stripe Payments**: Subscription management and billing
- **Bunny.net**: Video streaming and CDN services
- **Digital Ocean**: Cloud infrastructure and storage

## 📁 Project Structure

```
BOME/
├── frontend/                    # Svelte frontend application
│   ├── src/
│   │   ├── routes/
│   │   │   ├── articles/       # Articles subsite (COMPLETE)
│   │   │   ├── videos/         # Streaming subsite (COMPLETE)
│   │   │   ├── youtube/        # YouTube integration (COMPLETE)
│   │   │   ├── events/         # Events subsite (UI COMPLETE)
│   │   │   ├── admin/          # Admin dashboard (95% complete)
│   │   │   └── dashboard/      # User dashboard
│   │   ├── lib/
│   │   │   ├── components/     # Reusable UI components
│   │   │   ├── services/       # API services and integrations
│   │   │   ├── stores/         # State management
│   │   │   └── types/          # TypeScript definitions
│   │   └── app.css            # Custom design system
│   └── package.json
├── backend/                     # Go backend application
│   ├── internal/
│   │   ├── routes/             # API route handlers
│   │   ├── services/           # Business logic services
│   │   ├── database/           # Database models and migrations
│   │   ├── middleware/         # HTTP middleware
│   │   ├── config/             # Configuration management
│   │   └── MOCK_DATA/          # Development mock data
│   ├── go.mod                  # Go dependencies
│   └── main.go                 # Application entry point
├── docs/                        # Comprehensive documentation
├── deployment/                  # Docker and deployment configs
├── scripts/                     # Build and utility scripts
└── docker-compose.yml          # Multi-service orchestration
```

## 🚀 Quick Start

### Prerequisites
- **Go 1.24.3+** for backend development
- **Node.js 18+** for frontend development
- **Docker & Docker Compose** for containerized deployment
- **Git** for version control

### Development Setup

1. **Clone the Repository**
   ```bash
   git clone https://github.com/yourusername/bome.git
   cd bome
   ```

2. **Backend Setup**
   ```bash
   cd backend
   cp env.example .env          # Configure environment variables
   go mod download              # Install Go dependencies
   go run main.go              # Start backend server (port 8080)
   ```

3. **Frontend Setup**
   ```bash
   cd frontend
   npm install                  # Install Node dependencies
   npm run dev                 # Start development server (port 5173)
   ```

4. **Access the Application**
   - **Frontend**: http://localhost:5173
   - **Backend API**: http://localhost:8080
   - **Admin Dashboard**: http://localhost:5173/admin

### Production Deployment

1. **Docker Compose (Recommended)**
   ```bash
   cp .env.example .env         # Configure production environment
   docker-compose up -d         # Start all services
   ```

2. **Manual Deployment**
   ```bash
   # Backend
   cd backend && go build -o bome-backend
   ./bome-backend

   # Frontend
   cd frontend && npm run build
   # Serve dist/ with your preferred web server
   ```

## 🎨 Design System

BOME features a modern, custom CSS design system with:

- **Glass Morphism Effects**: Subtle transparency and backdrop blur
- **Neumorphic Elements**: Soft shadows and depth
- **Responsive Grid Layouts**: Mobile-first design approach
- **Custom CSS Properties**: Consistent theming and spacing
- **Smooth Animations**: 0.4s cubic-bezier transitions
- **Accessibility**: WCAG 2.1 AA compliance

### Color Palette
- **Primary**: Glass morphism with transparency
- **Secondary**: Subtle accent colors
- **Success**: #43E97B (Green)
- **Warning**: #FFAB00 (Amber)
- **Error**: #FF5630 (Red)

## 📊 Current Development Status

### Overall Completion: **85%**

#### ✅ Completed Systems
- **Articles Subsite**: 100% complete with 18 articles, search, filtering
- **Streaming Platform**: 100% complete with Bunny.net integration
- **YouTube Integration**: 100% complete with production-ready API path
- **Admin Dashboard**: 95% complete with comprehensive management
- **Analytics System**: 95% complete with real-time tracking
- **Role Management**: 100% complete with 18 roles and permissions
- **Advertisement System**: 95% complete with campaign management

#### 🔄 In Progress
- **Events Subsite**: UI complete, backend integration pending
- **API Integration**: Replacing mock data with live endpoints
- **Analytics Optimization**: Performance enhancements

#### 📋 Planned Features
- **Roku App**: Cross-platform streaming application
- **Mobile Apps**: iOS and Android applications
- **Advanced Search**: AI-powered content discovery
- **Community Features**: User forums and discussions

## 🔧 Key Features

### Content Management
- **Rich Text Editor**: Advanced article creation and editing
- **Video Upload**: Direct integration with Bunny.net CDN
- **Media Library**: Centralized asset management
- **SEO Optimization**: Meta tags and structured data

### User Experience
- **Responsive Design**: Optimized for all devices
- **Progressive Web App**: Offline functionality
- **Real-time Updates**: WebSocket integration
- **Advanced Search**: Multi-faceted content discovery

### Administration
- **Role-Based Access**: 18 predefined roles with granular permissions
- **Analytics Dashboard**: Comprehensive usage and performance metrics
- **Content Moderation**: Review and approval workflows
- **System Monitoring**: Health checks and performance tracking

### Security & Performance
- **JWT Authentication**: Secure token-based authentication
- **Rate Limiting**: API protection and abuse prevention
- **Caching Strategy**: Multi-layer caching for optimal performance
- **Data Encryption**: Secure data storage and transmission

## 📚 Documentation

### Technical Documentation
- **[API Reference](./docs/api/README.md)** - Complete REST API documentation
- **[Architecture Guide](./docs/architecture/README.md)** - System design and patterns
- **[Deployment Guide](./docs/deployment/README.md)** - Production deployment instructions
- **[Development Guide](./docs/development/README.md)** - Development setup and guidelines

### User Documentation
- **[User Manual](./docs/user/README.md)** - Complete platform usage guide
- **[Admin Guide](./docs/admin/README.md)** - Administrative dashboard manual
- **[Content Creator Guide](./docs/creator/README.md)** - Content creation and management

## 🤝 Contributing

We welcome contributions from the community! Please see our [Contributing Guidelines](./CONTRIBUTING.md) for details on:

- Code style and standards
- Pull request process
- Issue reporting
- Development workflow

### Development Workflow
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project and all its contents are proprietary and confidential. All rights are reserved by the copyright holder.

**Copyright © 2024 BOME Development Team. All Rights Reserved.**

No part of this software, including but not limited to source code, documentation, assets, or any other materials, may be:
- Used, copied, modified, or distributed without explicit written permission
- Reverse engineered, decompiled, or disassembled
- Used for commercial or non-commercial purposes
- Incorporated into other projects or derivative works

For licensing inquiries, please contact: licensing@bome.example.com

## 🔗 Links

- **Live Demo**: [https://bome.example.com](https://bome.example.com)
- **Documentation**: [https://docs.bome.example.com](https://docs.bome.example.com)
- **API Reference**: [https://api.bome.example.com/docs](https://api.bome.example.com/docs)
- **Issue Tracker**: [GitHub Issues](https://github.com/yourusername/bome/issues)

## 📞 Support

For support and questions:

- **Email**: support@bome.example.com
- **Documentation**: Check our comprehensive docs
- **GitHub Issues**: Report bugs and feature requests
- **Community**: Join our discussion forums

---

**BOME** - Advancing Book of Mormon research through technology and scholarship.

*Last Updated: December 2024 | Version: 1.0.0 | Maintained by: BOME Development Team*