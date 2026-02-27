#!/bin/sh
set -e

cd /app && go mod tidy

exec "$@"
