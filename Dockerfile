FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the API server application for production
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/api/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy configuration
COPY --from=builder /app/configs ./configs/

# Create non-root user
RUN adduser -D -s /bin/sh appuser
USER appuser

# Expose port (if needed for future web interface)
EXPOSE 8080

# Run the application with production config
CMD ["./main", "-port=8080", "-config=configs/production.yaml"]
