#!/bin/sh
# Instala el hook pre-commit para ejecutar lint antes de cada commit
set -e

HOOK_SRC="scripts/pre-commit.sh"
HOOK_DEST=".git/hooks/pre-commit"

if [ -f "$HOOK_SRC" ] && [ -d ".git" ]; then
  cp "$HOOK_SRC" "$HOOK_DEST"
  chmod +x "$HOOK_DEST"
  echo "Pre-commit hook instalado en $HOOK_DEST"
fi
