#!/bin/bash

BIN_DIR="$HOME/bin"

# Download cilium if it doesn't exist
if [ -f "$BIN_DIR/cilium" ]; then
  echo "cilium already exists, skipping download."

  exit 0
fi

echo "Downloading cilium..."
curl -fSL https://github.com/cilium/cilium-cli/releases/download/v0.16.22/cilium-linux-amd64.tar.gz | tar -xz -C "$BIN_DIR" cilium
chmod +x "$BIN_DIR/cilium"

$BIN_DIR/cilium version

echo "cilium downloaded successfully. To enable bash completion, run the following command:"
echo "echo 'source <(cilium completion bash)' >> ~/.bashrc"
