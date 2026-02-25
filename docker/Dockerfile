# Build stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy local workspace dependencies
COPY core /core
COPY framework /framework
COPY plugins/tenant /plugins/tenant
COPY plugins/ou /plugins/ou

# Copy examples go mod files
COPY examples/go.mod examples/go.sum ./

# Rewrite replace directives for Docker build paths
RUN go mod edit \
    -replace=github.com/leeforge/core=/core \
    -replace=github.com/leeforge/framework=/framework \
    -replace=github.com/leeforge/plugins/tenant=/plugins/tenant \
    -replace=github.com/leeforge/plugins/ou=/plugins/ou

# Download dependencies
RUN go mod download

# Copy examples source code
COPY examples/ .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o leeforge-examples ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/leeforge-examples .

# Copy configs
COPY examples/configs ./configs

# Create non-root user
RUN addgroup -g 1000 leeforge && \
    adduser -D -u 1000 -G leeforge leeforge && \
    chown -R leeforge:leeforge /app

USER leeforge

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./leeforge-examples"]
