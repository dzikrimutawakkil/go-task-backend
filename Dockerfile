# Multi-stage Dockerfile for GoTask Backend
# Optimized for production with minimal image size and security best practices

# ============================================
# Stage 1: Build
# ============================================
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go.mod and go.sum first for better Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o gotask-backend \
    main.go

# Generate Swagger docs (if swag is available)
RUN which swag && swag init -g main.go -o docs/generated || true

# ============================================
# Stage 2: Production
# ============================================
FROM alpine:3.19 AS production

# Install runtime dependencies only
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN adduser -D -g '' appuser

WORKDIR /app

# Set ownership
RUN chown -R appuser:appuser /app

# Copy binary from builder stage
COPY --from=builder /build/gotask-backend .

# Copy .env.example as template (actual .env should be mounted via docker-compose or secrets)
COPY --from=builder /build/.env.example .

# Switch to non-root user
USER appuser

# Expose the port
EXPOSE 8080

# Healthcheck for container orchestration
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run as non-root user
ENV GIN_MODE=release
ENV PORT=8080

ENTRYPOINT ["./gotask-backend"]
CMD ["--env", "production"]

# ============================================
# Stage 3: Development (optional)
# ============================================
FROM builder AS development

WORKDIR /app

# Copy source code for hot reload (use with volume mounts)
COPY . .

# Expose debug port
EXPOSE 8080 40000

CMD ["air", "-c", ".air.toml"]