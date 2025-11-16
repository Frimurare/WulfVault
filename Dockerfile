# WulfVault Dockerfile
# Multi-stage build for efficient final image

# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o wulfvault ./cmd/server

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/wulfvault .

# Create directories
RUN mkdir -p /data /uploads

# Expose port
EXPOSE 8080

# Set environment variables
ENV DATA_DIR=/data
ENV UPLOADS_DIR=/uploads
ENV PORT=8080

# Run the application
CMD ["./wulfvault"]
