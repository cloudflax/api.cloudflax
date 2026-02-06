# Usar imagen Debian (glibc) - Alpine/musl causa fallos con Cursor Server
FROM golang:1.25-bookworm

WORKDIR /app

# Instalar dependencias del sistema
RUN apt-get update && apt-get install -y --no-install-recommends git \
    && rm -rf /var/lib/apt/lists/*

# Instalar herramientas de desarrollo
RUN go install github.com/air-verse/air@latest \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Copiar archivos de dependencias
COPY go.mod ./

# Descargar dependencias
RUN go mod download

# Copiar código fuente
COPY . .

# Entrypoint que sincroniza go.mod/go.sum antes de iniciar (evita errores al reconstruir)
COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]

EXPOSE 3000

# Por defecto ejecutar con air (hot reload)
CMD ["air", "-c", ".air.toml"]
