.PHONY: build run test lint db-certs

# Generar certificados SSL para PostgreSQL (requerido antes del primer docker-compose up)
db-certs:
	@sh scripts/generate-db-certs.sh

# Compilar la aplicación
build:
	go build -buildvcs=false -o ./tmp/main ./cmd/api

# Ejecutar la aplicación (requiere variables de entorno)
run: build
	./tmp/main

# Ejecutar tests
test:
	go test -v ./...

# Ejecutar linter
lint:
	golangci-lint run --config .golangci.yml ./...
