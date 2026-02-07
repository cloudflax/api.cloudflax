.PHONY: build run test lint db-certs

# Generar certificados SSL para PostgreSQL (requerido antes del primer docker-compose up)
db-certs:
	@sh scripts/generate-db-certs.sh

# Compilar la aplicación
build:
	go build -o ./tmp/main ./cmd/api

# Ejecutar la aplicación (requiere variables de entorno)
run: build
	./tmp/main

# Ejecutar tests
test:
	go test -v ./...

# Ejecutar tests con cobertura (genera coverage.html)
test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Cobertura: $(shell go tool cover -func=coverage.out | grep total | awk '{print $$3}')"

# Ejecutar linter
lint:
	golangci-lint run --config .golangci.yml ./...
