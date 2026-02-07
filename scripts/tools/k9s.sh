#!/bin/bash

set -e

BIN_DIR="$HOME/bin"

echo "> Installing k9s"

# Download k9s if it doesn't exist
if [ -f "$BIN_DIR/k9s" ]; then
  echo "> k9s already exists, skipping download."

  exit 0
fi

curl -fSL https://github.com/derailed/k9s/releases/download/v0.50.18/k9s_Linux_amd64.tar.gz | tar -xz -C "$BIN_DIR" k9s
chmod +x "$BIN_DIR/k9s"

echo "> k9s downloaded successfully."
