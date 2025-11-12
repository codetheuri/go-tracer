# Makefile for Go Cybersecurity AI Project

.DEFAULT_GOAL := help

# Go variables
BINARY_NAME_API := cyber-api
BINARY_NAME_MIGRATE := cyber-migrate
CMD_PATH_API := ./cmd/api
CMD_PATH_MIGRATE := ./cmd/migrate

# Build commands
.PHONY: build
build:
	@echo "Building API server..."
	@go build -o bin/$(BINARY_NAME_API) $(CMD_PATH_API)/main.go
	@echo "Building migration tool..."
	@go build -o bin/$(BINARY_NAME_MIGRATE) $(CMD_PATH_MIGRATE)/main.go
	@echo "Build complete."

# Run the API server
.PHONY: run
run: build
	@echo "Starting API server..."
	@./bin/$(BINARY_NAME_API)

# Run migrations
.PHONY: migrate
migrate: build
	@echo "Running database migrations..."
	@./bin/$(BINARY_NAME_MIGRATE) migrate

# Run database seeding
.PHONY: seed
seed: build
	@echo "Seeding database..."
	@./bin/$(BINARY_NAME_MIGRATE) seed

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	@go clean
	@rm -f bin/$(BINARY_NAME_API)
	@rm -f bin/$(BINARY_NAME_MIGRATE)

# Show help
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make build    - Build the API server and migration tool"
	@echo "  make run      - Build and run the API server"
	@echo "  make migrate  - Build and run the database migrations"
	@echo "  make seed     - Build and run the database seeder"
	@echo "  make clean    - Remove build artifacts"
