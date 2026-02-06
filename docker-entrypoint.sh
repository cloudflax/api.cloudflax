#!/bin/sh
set -e

# Sincronizar go.mod y go.sum antes de iniciar (evita errores al reconstruir)
cd /app && go mod tidy

exec "$@"
