#!/bin/sh
set -e

cd /app && go mod tidy

if [ -f /app/scripts/download-rds-certs.sh ]; then
  /app/scripts/download-rds-certs.sh
fi

exec "$@"
