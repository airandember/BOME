#!/bin/bash

# Create Admin Account Script for BOME
# This script creates a test admin account for development

set -e

echo "🔧 Creating test admin account for BOME..."

# Set environment variables for development
export ENVIRONMENT=development
export DEBUG=true
export SERVER_PORT=8080
export SERVER_HOST=0.0.0.0
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=bome_streaming
export DB_USER=bome_user
export DB_PASSWORD=dev_password
export DB_SSL_MODE=disable
export JWT_SECRET=dev-jwt-secret-key-change-in-production
export JWT_EXPIRY=24h
export JWT_REFRESH_EXPIRY=168h
export CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:5174,http://localhost:4173
export CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
export CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-Requested-With
export ADMIN_EMAIL=admin@bome.test
export ADMIN_PASSWORD=Admin123!
export ADMIN_SECRET_KEY=dev-admin-secret-key
export BCRYPT_COST=10
export SESSION_SECRET=dev-session-secret-key
export CSRF_SECRET=dev-csrf-secret-key

# Change to backend directory
cd backend

# Build and run the create-admin command
echo "📦 Building admin creation tool..."
go build -o create-admin cmd/create-admin/main.go

echo "🚀 Running admin creation..."
./create-admin

echo ""
echo "🎉 Admin account creation completed!"
echo ""
echo "📋 Login Credentials:"
echo "   Email: admin@bome.test"
echo "   Password: Admin123!"
echo ""
echo "🌐 Access the admin dashboard at:"
echo "   http://localhost:5174/admin"
echo ""
echo "⚠️  IMPORTANT: These are test credentials for development only!"
echo "   Change them in production!"

# Clean up
rm -f create-admin

echo ""
echo "✅ Script completed successfully!" 