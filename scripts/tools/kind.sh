#!/bin/bash

BIN_DIR="$HOME/bin"

# Download kind if it doesn't exist
if [ -f "$BIN_DIR/kind" ]; then
  echo "kind already exists, skipping download."

  read -p "Do you want to redownload kind? (y/N) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Skipping kind download."
    exit 0
  fi
fi

echo "Downloading kind..."
curl -fSLo "$BIN_DIR/kind" "https://github.com/kubernetes-sigs/kind/releases/download/v0.31.0/kind-linux-amd64"
chmod +x "$BIN_DIR/kind"

$BIN_DIR/kind --version

echo "kind downloaded successfully. To enable bash completion, run the following command:"
echo "echo 'source <(kind completion bash)' >> ~/.bashrc"
