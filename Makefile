.PHONY: build run test test-verbose test-cover lint db-certs clean-cache

# Terminal colors
GREEN  := $(shell tput -Txterm setaf 2)
RED    := $(shell tput -Txterm setaf 1)
YELLOW := $(shell tput -Txterm setaf 3)
RESET  := $(shell tput -Txterm sgr0)

db-certs:
	@echo "$(YELLOW)Generating SSL certificates...$(RESET)"
	@sh scripts/generate-db-certs.sh

build:
	@echo "$(YELLOW)Compiling...$(RESET)"
	@go build -o ./tmp/main ./cmd/api

run: build
	@echo "$(GREEN)Starting application...$(RESET)"
	@./tmp/main

test:
	@bash -c '\
	out=$$(go test -v ./... 2>&1); ec=$$?; \
	echo "$$out" | awk " \
	/^--- PASS:/ { \
		count++; \
		names[count] = \$$3; \
		raw = substr(\$$4, 2, length(\$$4)-2); \
		sub(/s/, \"\", raw); \
		times[count] = sprintf(\"%.2fs\", raw); \
		pkg_sum += raw; \
	} \
	/^ok[[:space:]]+/ { \
		pkg = \$$2; \
		is_cached = (\$$0 ~ /\(cached\)/); \
		if (is_cached) { \
			final_time = sprintf(\"%.2fs\", pkg_sum); \
		} else { \
			raw_pkg = \$$3; sub(/s/, \"\", raw_pkg); \
			final_time = sprintf(\"%.2fs\", raw_pkg); \
		} \
		printf \"- ok %-60s [%s]\n\", pkg, final_time; \
		suffix = is_cached ? \"(cached)\" : \"\"; \
		for (i = 1; i <= count; i++) { \
			printf \"    - ok %-56s [%s] %s\n\", names[i], times[i], suffix; \
		} \
		delete names; delete times; count = 0; pkg_sum = 0; \
		next; \
	} \
	/^FAIL/ { print \"$(RED)\" \$$0 \"$(RESET)\" }"; \
	echo ""; \
	if [ $$ec -eq 0 ]; then \
		printf "$(GREEN)✓ All tests passed$(RESET)\n"; \
	else \
		printf "$(RED)✗ Some tests failed$(RESET)\n"; \
	fi; \
	exit $$ec'

test-verbose:
	go test -v ./...

test-cover:
	@echo "$(YELLOW)Calculating coverage...$(RESET)"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Total coverage: $$(go tool cover -func=coverage.out | grep total | awk '{print $$3}')$(RESET)"

lint:
	@echo "$(YELLOW)Running linter...$(RESET)"
	@golangci-lint run --config .golangci.yml ./...

clean-cache:
	@echo "$(YELLOW)Cleaning test cache...$(RESET)"
	@go clean -testcache