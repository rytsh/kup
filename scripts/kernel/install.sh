#!/bin/bash

# Download kernel and edit wsl config to use it

KERNEL_DIR="/mnt/c/tools/kernel"

# Create tools/kernel directory if it doesn't exist
if [ ! -d "$KERNEL_DIR" ]; then
  echo "Creating $KERNEL_DIR directory..."
  mkdir -p "$KERNEL_DIR"
fi

# Download kernel if it doesn't exist
if [ ! -f ${KERNEL_DIR}/bzImage-x64v3 ]; then
  echo "Downloading kernel..."
  curl -fSL https://github.com/Locietta/xanmod-kernel-WSL2/releases/download/6.12.68-locietta-WSL2-xanmod1.1-lts/bzImage-x64v3 -o ${KERNEL_DIR}/bzImage-x64v3
  curl -fSL https://github.com/Locietta/xanmod-kernel-WSL2/releases/download/6.12.68-locietta-WSL2-xanmod1.1-lts/bzImage-x64v3-addons.vhdx -o ${KERNEL_DIR}/bzImage-x64v3-addons.vhdx
fi

if read -p "Kernel downloaded successfully. Do you want to edit your wsl config to use it? (y/N) " -n 1 -r; then
  echo
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    read -p "Enter your Windows username [$USER]: " WIN_USER
    WIN_USER=${WIN_USER:-$USER}
    echo "Editing wsl config..."
  else
    echo "Skipping wsl config edit."
    exit 0
  fi
fi

cat << EOF > /mnt/c/Users/$WIN_USER/.wslconfig
[wsl2]
kernel = C:\\\\tools\\\\kernel\\\\bzImage-x64v3
kernelModules = C:\\\\tools\\\\kernel\\\\bzImage-x64v3-addons.vhdx
kernelCommandLine = cgroup_no_v1=all
EOF

echo "wsl config edited successfully."
echo "use 'wsl.exe --shutdown' and check 'uname -a' to verify the new kernel is being used."
