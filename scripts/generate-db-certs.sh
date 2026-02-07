#!/bin/sh
# Genera certificados self-signed para PostgreSQL con SSL
# CN=db (hostname usado en la red Docker)
set -e

CERTS_DIR="postgres/certs"
mkdir -p "$CERTS_DIR"

echo "Generando certificados en $CERTS_DIR..."

# Generar clave y certificado del servidor (CN=db para Docker network)
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout "$CERTS_DIR/server.key" \
  -out "$CERTS_DIR/server.crt" \
  -subj "/CN=db" \
  -days 365

chmod 600 "$CERTS_DIR/server.key"

echo "Certificados generados: $CERTS_DIR/server.crt, $CERTS_DIR/server.key"
echo "Reinicia los contenedores: docker-compose down && docker-compose up -d"
