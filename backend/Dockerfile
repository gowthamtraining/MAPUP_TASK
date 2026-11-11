# Multi-stage Dockerfile for MapUp Geofence Backend
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /app/bin/geofence-server \
    ./cmd/server

# Final minimal production image
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata curl

# Copy binary from builder
COPY --from=builder /app/bin/geofence-server /app/geofence-server

# Non-root user for security
RUN adduser -D -g '' appuser
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/geofence-server"]
