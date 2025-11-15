FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build bot
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/bot ./cmd/bot

# Build server
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/server ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

WORKDIR /app

# Copy binaries
COPY --from=builder /app/bin/bot /app/bin/bot
COPY --from=builder /app/bin/server /app/bin/server

# Copy config
COPY config/config.yaml /app/config/config.yaml

# Create directories
RUN mkdir -p /app/data /app/storage

EXPOSE 8080

# Default to running server, can override with docker-compose
CMD ["/app/bin/server"]

