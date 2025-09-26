# Build stage
FROM node:18-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./


# Install ALL dependencies (including dev dependencies needed for build)
RUN npm ci

# Copy source code
COPY . .

# Build the application with production search index
RUN npm run build

# Debug: Check what was actually built
RUN echo "=== BUILD DIRECTORY CONTENTS ===" && ls -la build/ && echo "=== END BUILD CONTENTS ==="

# Production stage
FROM node:18-alpine AS runner

# Install runtime dependencies
RUN apk add --no-cache dumb-init

# Create non-root user
RUN addgroup -g 1001 -S bome && \
    adduser -u 1001 -S bome -G bome

# Set working directory
WORKDIR /app

# Copy built application from builder stage
COPY --from=builder --chown=bome:bome /app/build ./build
COPY --from=builder --chown=bome:bome /app/package.json ./package.json
COPY --from=builder --chown=bome:bome /app/package-lock.json ./package-lock.json

# Debug: Check what was copied
RUN echo "=== COPIED BUILD CONTENTS ===" && ls -la build/ && echo "=== END COPIED CONTENTS ==="


# Install only production dependencies
RUN npm ci --only=production && npm cache clean --force

# Change ownership
RUN chown -R bome:bome /app

# Switch to non-root user
USER bome

# Expose port
EXPOSE 3000

# Set environment variables
ENV NODE_ENV=production
ENV PORT=3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1

# Start the application
ENTRYPOINT ["dumb-init", "--"]

CMD ["node", "build/index.js"] 

