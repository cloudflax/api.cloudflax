.PHONY: build run test test-verbose test-cover lint db-certs clean-cache localstack-ses-emails localstack-ses-verify-identity

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

# LocalStack SES: list sent emails (requires LocalStack running with SES).
# From host use default. From devcontainer: make localstack-ses-emails LOCALSTACK_ENDPOINT=http://host.docker.internal:4566
LOCALSTACK_ENDPOINT ?= http://localhost:4566
localstack-ses-emails:
	@echo "$(GREEN)Fetching sent emails from LocalStack SES...$(RESET)"
	@curl -s "$(LOCALSTACK_ENDPOINT)/_aws/ses" | jq . 2>/dev/null || curl -s "$(LOCALSTACK_ENDPOINT)/_aws/ses"

# Verify the sender identity in LocalStack (required before SendEmail works).
# Usage: make localstack-ses-verify-identity [EMAIL=jose.guerrero@cloudflax.com] [LOCALSTACK_ENDPOINT=http://localhost:4566]
EMAIL ?= $(shell grep SES_FROM_ADDRESS .env 2>/dev/null | cut -d= -f2 || echo "jose.guerrero@cloudflax.com")
localstack-ses-verify-identity:
	@echo "$(GREEN)Verifying SES identity: $(EMAIL)$(RESET)"
	@AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test aws --endpoint-url="$(LOCALSTACK_ENDPOINT)" sesv2 create-email-identity --email-identity "$(EMAIL)" --region us-east-1 2>/dev/null && echo "$(GREEN)Identity verified.$(RESET)" || echo "$(YELLOW)Identity may already exist or aws CLI not in PATH.$(RESET)"