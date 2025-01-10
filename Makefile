# Variables
DB_URL=postgres://user:password@localhost:5432/orders?sslmode=disable
MIGRATIONS_DIR=./migrations
GOOSE_CMD=goose
DOCKER_COMPOSE_CMD=docker compose

# Targets
.PHONY: all build run clean migrate-up migrate-down migrate-create

# Build the Go application
build:
	go build -o app ./cmd

# Run the Go application
run:
	go run ./cmd/main.go

# Clean up generated files
clean:
	rm -f app

# Apply all "up" migrations
migrate-up:
	$(GOOSE_CMD) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

# Rollback the most recent migration
migrate-down:
	$(GOOSE_CMD) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down

# Create a new migration
# Usage: make migrate-create name=<migration_name>
migrate-create:
ifndef name
	$(error name is required. Usage: make migrate-create name=<migration_name>)
endif
	$(GOOSE_CMD) create -dir $(MIGRATIONS_DIR) $(name) sql

# Run everything: build, migrate, and start
all: migrate-up build run


# Start services
up:
	$(DOCKER_COMPOSE_CMD) -f docker-compose.yaml up -d

# Stop services
down:
	$(DOCKER_COMPOSE_CMD) -f docker-compose.yaml down