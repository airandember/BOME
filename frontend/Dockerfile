# Build stage
FROM node:18-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Production stage
FROM node:18-alpine AS runner

# Create non-root user
RUN addgroup -g 1001 -S bome && \
    adduser -u 1001 -S bome -G bome

# Set working directory
WORKDIR /app

# Copy built application from builder stage
COPY --from=builder --chown=bome:bome /app/build ./build
COPY --from=builder --chown=bome:bome /app/package.json ./package.json

# Install only production dependencies
RUN npm ci --only=production && npm cache clean --force

# Switch to non-root user
USER bome

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1

# Start the application
CMD ["npm", "start"] 