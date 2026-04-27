# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o convoy ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates and postgresql-client for migrations
RUN apk --no-cache add ca-certificates postgresql-client curl

# Install golang-migrate
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/convoy .

# Copy migrations
COPY migrations ./migrations

# Copy entrypoint script
COPY scripts/docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Expose port
EXPOSE 8080

# Run the entrypoint script
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["./convoy"]
