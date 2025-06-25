# BOME Code Documentation

**Development Guidelines, Patterns, and Best Practices**

Complete reference for developers working on the BOME streaming platform.

---

## 📋 Table of Contents

- [Development Environment](#development-environment)
- [Code Architecture](#code-architecture)
- [Frontend Development](#frontend-development)
- [Backend Development](#backend-development)
- [Database Patterns](#database-patterns)
- [API Development](#api-development)
- [Testing Guidelines](#testing-guidelines)
- [Code Style Guide](#code-style-guide)
- [Performance Guidelines](#performance-guidelines)
- [Security Best Practices](#security-best-practices)
- [Debugging and Troubleshooting](#debugging-and-troubleshooting)
- [Contribution Guidelines](#contribution-guidelines)

---

## Development Environment

### 🚀 Quick Setup

```bash
# 1. Clone and navigate to project
git clone https://github.com/your-org/BOME.git
cd BOME

# 2. Backend setup
cd backend
cp env.example .env
# Edit .env with your configuration
go mod tidy
go run main.go

# 3. Frontend setup (new terminal)
cd ../frontend
npm install
npm run dev

# 4. Access the application
# Frontend: http://localhost:5173
# Backend API: http://localhost:8080
```

### 📋 Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| **Go** | 1.21+ | Backend development |
| **Node.js** | 18+ | Frontend development |
| **PostgreSQL** | 15+ | Database (optional for dev) |
| **Redis** | 7+ | Caching (optional for dev) |
| **Git** | 2.40+ | Version control |

### 🔧 Development Tools

#### Recommended VS Code Extensions
```json
{
  "recommendations": [
    "golang.go",
    "svelte.svelte-vscode",
    "bradlc.vscode-tailwindcss",
    "ms-vscode.vscode-typescript-next",
    "esbenp.prettier-vscode",
    "ms-vscode.vscode-json"
  ]
}
```

#### Go Tools Setup
```bash
# Essential Go tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/air-verse/air@latest # Hot reload
```

---

## Code Architecture

### ��️ Project Structure 