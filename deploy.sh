#!/bin/bash

echo "🚀 Deploying BOME with Horizontal Scaling..."

# Stop existing containers
echo "Stopping existing containers..."
docker-compose down

# Remove old images
echo "Cleaning up old images..."
docker system prune -f

# Build and start services
echo "Building and starting services..."
docker-compose up -d --build

# Wait for services to be ready
echo "Waiting for services to initialize..."
sleep 30

# Check service health
echo "Checking service health..."
docker-compose ps

# Test load balancer
echo "Testing load balancer..."
for i in {1..10}; do
    curl -s http://localhost/health
    echo " - Request $i"
done

echo "✅ Deployment complete!"
echo "🌐 Frontend: http://localhost"
echo "🔧 Admin: http://localhost/admin"
echo "📊 API: http://localhost/api"
echo "📈 Nginx Status: http://localhost:8080/nginx_status" 