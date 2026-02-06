#!/bin/sh
# Hook pre-commit: ejecuta lint antes de permitir el commit
set -e

echo "Running lint..."
make lint
echo "Lint OK"
exit 0
