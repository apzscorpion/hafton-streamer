.PHONY: build build-bot build-server run-bot run-server clean test deps

# Build both bot and server
build: build-bot build-server

# Build bot
build-bot:
	@echo "Building bot..."
	@mkdir -p bin
	@CGO_ENABLED=1 go build -o bin/bot ./cmd/bot

# Build server
build-server:
	@echo "Building server..."
	@mkdir -p bin
	@CGO_ENABLED=1 go build -o bin/server ./cmd/server

# Run bot locally
run-bot:
	@go run cmd/bot/main.go

# Run server locally
run-server:
	@go run cmd/server/main.go

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf data/
	@rm -rf storage/

# Download dependencies
deps:
	@go mod download
	@go mod tidy

# Run tests
test:
	@go test ./...

# Install dependencies
install-deps:
	@go mod download
	@go mod verify

