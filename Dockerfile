# Build stage
FROM golang:1.25.1-alpine AS builder

# Install git and ca-certificates (needed for go modules)
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mikrotik-exporter .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh mikrotik

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/mikrotik-exporter .

# Copy default config
COPY --from=builder /app/config.yaml .

# Change ownership to non-root user
RUN chown -R mikrotik:mikrotik /app

# Switch to non-root user
USER mikrotik

# Expose port
EXPOSE 9261

# Set default environment variables
ENV LISTEN_ADDR=0.0.0.0
ENV LISTEN_PORT=9261
ENV CONFIG_FILE=./config.yaml

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9261/ || exit 1

# Run the binary
CMD ["./mikrotik-exporter"]
