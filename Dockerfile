FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies for CGO
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go mod files first
COPY go.mod go.sum ./

# Verify module and download dependencies
RUN go mod verify && go mod download

# Copy all source code (copy everything except what's in .dockerignore)
COPY . .

# Verify files are copied correctly and module is recognized
RUN ls -la && \
    ls -la cmd/ && \
    ls -la internal/ && \
    cat go.mod && \
    go list -m

# Build server only
RUN CGO_ENABLED=1 go build -v -o bin/server ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

WORKDIR /app

# Copy binary
COPY --from=builder /app/bin/server /app/bin/server

# Create directories
RUN mkdir -p /app/data /app/storage

EXPOSE 8080

# Use PORT environment variable (Railway provides this)
ENV PORT=8080
CMD ["./bin/server"]

