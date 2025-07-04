# Multi-stage Dockerfile for Argus System Monitor
# --- Go backend build stage ---
FROM golang:1.23-alpine AS go-builder

# Install dependencies for building
RUN apk --no-cache add ca-certificates git

WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/argus/main.go

# --- Node.js frontend build stage ---
FROM node:18-alpine AS node-builder

WORKDIR /app/web/argus-react

# Copy package files
COPY web/argus-react/package*.json ./

# Install dependencies (include devDependencies for build)
RUN npm ci

# Copy source code
COPY web/argus-react/ ./

# Build frontend (skip if not ready yet)
RUN if [ -f "vite.config.ts" ]; then npm run build; fi

# --- Final runtime stage ---
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -s /bin/sh argus

WORKDIR /home/argus

# Copy built application
COPY --from=go-builder /app/main .

# Copy built React app to release directory
COPY --from=node-builder /app/web/argus-react/dist ./web/release/

# Copy configuration template (user should mount config.yaml in production)
COPY config.example.yaml ./config.yaml

# Create necessary directories
RUN mkdir -p .argus/alerts .argus/tasks && \
    chown -R argus:argus . && \
    chmod +x main

# Switch to non-root user
USER argus

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"] 