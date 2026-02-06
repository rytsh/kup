#!/bin/bash

BIN_DIR="$HOME/bin"

# Download kubectl if it doesn't exist
if [ -f "$BIN_DIR/kubectl" ]; then
  echo "kubectl already exists, skipping download."
  exit 0
fi

echo "Downloading kubectl..."

curl -fSLo "$BIN_DIR/kubectl" "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x "$BIN_DIR/kubectl"

$BIN_DIR/kubectl version --client

echo "kubectl downloaded successfully. To enable bash completion, run the following command:"
echo "echo 'source <(kubectl completion bash)' >> ~/.bashrc"
