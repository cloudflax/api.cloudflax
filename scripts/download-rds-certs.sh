#!/usr/bin/env bash
set -euo pipefail

CERT_DIR="/certs"
BUNDLE_URL="https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem"
BUNDLE_PATH="${CERT_DIR}/global-bundle.pem"

if [ -f "$BUNDLE_PATH" ]; then
  echo "RDS CA bundle already exists at ${BUNDLE_PATH}, skipping download."
  exit 0
fi

echo "Downloading AWS RDS CA bundle..."
mkdir -p "$CERT_DIR"
curl -fsSL -o "$BUNDLE_PATH" "$BUNDLE_URL"
chmod 644 "$BUNDLE_PATH"
echo "RDS CA bundle saved to ${BUNDLE_PATH}"
