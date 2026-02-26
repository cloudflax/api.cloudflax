.PHONY: build run test test-verbose test-cover lint db-certs clean-cache db-reset

# Terminal colors
GREEN  := $(shell tput -Txterm setaf 2)
RED    := $(shell tput -Txterm setaf 1)
YELLOW := $(shell tput -Txterm setaf 3)
RESET  := $(shell tput -Txterm sgr0)

# Database connection (development, for db-reset)
# Defaults for DevContainer: DB on host
DB_HOST ?= host.docker.internal
DB_PORT ?= 4510
DB_USER ?= postgres
DB_NAME ?= cloudflax
DB_SSL_MODE ?= disable

# ... (tus otras reglas build, run, etc)

test:
	@chmod +x scripts/test_runner.sh
	@./scripts/test_runner.sh

clean-cache:
	@echo "$(YELLOW)Cleaning test cache...$(RESET)"
	@go clean -testcache

# Development only: truncates all app tables. Never run against production.
db-reset:
	@if [ -z "$(DB_PASSWORD)" ]; then echo "$(RED)DB_PASSWORD is not set. Usage: make db-reset DB_PASSWORD=your_password$(RESET)"; exit 1; fi
	@echo "$(YELLOW)Resetting development database tables...$(RESET)"
	@APP_ENV=development DB_HOST="$(DB_HOST)" DB_PORT="$(DB_PORT)" DB_USER="$(DB_USER)" DB_PASSWORD="$(DB_PASSWORD)" DB_NAME="$(DB_NAME)" DB_SSL_MODE="$(DB_SSL_MODE)" go run ./cmd/db-reset
	@echo "$(GREEN)Done.$(RESET)"