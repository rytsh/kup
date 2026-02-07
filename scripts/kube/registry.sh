#!/bin/bash

# Before to start container if any authentication needed pass it with AUTH_REGISTRIES environment variable.
# -e AUTH_REGISTRIES="auth.docker.io:dockerhub_username:dockerhub_password your.own.registry:username:password" \
# Always set REGISTRIES environment variable and add what need, if not it gets error when download images.

export KIND_EXPERIMENTAL_PROVIDER=${KIND_EXPERIMENTAL_PROVIDER:-docker}

# check registry container if not exists
if ${KIND_EXPERIMENTAL_PROVIDER} ps -a --format '{{.Names}}' | grep -q docker_registry_proxy; then
  echo "> Registry container already exists"
else
  echo "> Creating registry container"
  sudo install -d -m 0777 /opt/registry

  ${KIND_EXPERIMENTAL_PROVIDER} run -d --restart=always --name docker_registry_proxy \
    --hostname docker-registry-proxy --network kind \
    -p 0.0.0.0:3128:3128 \
    -v /opt/registry/docker_mirror_cache:/docker_mirror_cache \
    -v /opt/registry/docker_mirror_certs:/ca \
    -e REGISTRIES="registry.k8s.io k8s.gcr.io gcr.io ghcr.io quay.io docker.io" \
    rpardini/docker-registry-proxy:0.6.5
  echo "> Waiting for registry container to start 10s"
  sleep 10
fi

echo "> Setting up container proxy"
KIND_NAME=kind
SETUP_URL=http://docker-registry-proxy:3128/setup/systemd
pids=""
for NODE in $(kind get nodes --name "$KIND_NAME"); do
  ${KIND_EXPERIMENTAL_PROVIDER} exec "$NODE" sh -c "\
      curl $SETUP_URL \
      | sed s/docker\.service/containerd\.service/g \
      | sed '/Environment/ s/$/ \"NO_PROXY=127.0.0.0\/8,10.0.0.0\/8,172.16.0.0\/12,192.168.0.0\/16\"/' \
      | bash" & pids="$pids $!" # Configure every node in background
done
wait $pids # Wait for all configurations to end
