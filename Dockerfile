# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o k8s-image-policy-monitor

# Final stage
FROM alpine:3.19

WORKDIR /app

# Add non-root user
RUN adduser -D -u 1000 appuser
USER appuser

# Copy binary from builder
COPY --from=builder /app/k8s-image-policy-monitor .

# Expose metrics port
EXPOSE 2112

ENTRYPOINT ["./k8s-image-policy-monitor"]
