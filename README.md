# Kup

Kubernetes installation guide in WSL.

## Prerequisites

- WSL 2
- Windows Terminal
- A Linux distribution installed in WSL
- Kernel installed `https://github.com/Locietta/xanmod-kernel-WSL2` check `./scripts/kernel/install.sh` for installation instructions.

## Installation

All tools are installed in the `~/bin` directory. Make sure to add it to your PATH.

```sh
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
```

```sh
./scripts/tools/kind.sh
./scripts/tools/kubectl.sh
./scripts/tools/cilium.sh
```
