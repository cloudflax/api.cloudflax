.PHONY: build run test test-verbose test-cover lint db-certs clean-cache db-reset

# Terminal colors
GREEN  := $(shell tput -Txterm setaf 2)
RED    := $(shell tput -Txterm setaf 1)
YELLOW := $(shell tput -Txterm setaf 3)
RESET  := $(shell tput -Txterm sgr0)

# ... (tus otras reglas build, run, etc)

test:
	@chmod +x scripts/test_runner.sh
	@./scripts/test_runner.sh

clean-cache:
	@echo "$(YELLOW)Cleaning test cache...$(RESET)"
	@go clean -testcache

# Development only: truncates all app tables. Never run against production.
# Credentials are fetched from AWS Secrets Manager (AWS_SECRET_NAME).
db-reset:
	@echo "$(YELLOW)Resetting development database tables...$(RESET)"
	@APP_ENV=development go run ./cmd/db-reset
	@echo "$(GREEN)Done.$(RESET)"