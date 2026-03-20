#!/bin/bash

# Before to start container if any authentication needed pass it with AUTH_REGISTRIES environment variable.
# -e AUTH_REGISTRIES="auth.docker.io:dockerhub_username:dockerhub_password your.own.registry:username:password" \
# Always set REGISTRIES environment variable and add what need, if not it gets error when download images.

export KIND_EXPERIMENTAL_PROVIDER=${KIND_EXPERIMENTAL_PROVIDER:-docker}

DOCKER_CONFIG="${DOCKER_CONFIG:-$HOME/.docker/config.json}"
DEFAULT_REGISTRIES="registry.k8s.io k8s.gcr.io gcr.io ghcr.io quay.io docker.io"

# Parse docker auth config and let user select registries
select_auth_registries() {
  if [[ ! -f "$DOCKER_CONFIG" ]]; then
    echo "> No docker config found at $DOCKER_CONFIG, skipping auth registries"
    return
  fi

  if ! command -v jq &>/dev/null; then
    echo "> jq is required to parse docker config, skipping auth registries"
    return
  fi

  # Extract registry URLs from docker config auths
  mapfile -t auth_registries < <(jq -r '.auths // {} | keys[]' "$DOCKER_CONFIG" 2>/dev/null)

  if [[ ${#auth_registries[@]} -eq 0 ]]; then
    echo "> No authenticated registries found in $DOCKER_CONFIG"
    return
  fi

  echo "> Found authenticated registries in docker config:"
  for i in "${!auth_registries[@]}"; do
    echo "  [$((i + 1))] ${auth_registries[$i]}"
  done
  echo "  [a] All"
  echo "  [n] None"

  read -rp "> Select registries to use (comma-separated numbers, 'a' for all, 'n' for none): " selection

  if [[ "$selection" == "n" ]]; then
    return
  fi

  local selected=()
  if [[ "$selection" == "a" ]]; then
    selected=("${auth_registries[@]}")
  else
    IFS=',' read -ra indices <<< "$selection"
    for idx in "${indices[@]}"; do
      idx=$(echo "$idx" | tr -d ' ')
      if [[ "$idx" =~ ^[0-9]+$ ]] && (( idx >= 1 && idx <= ${#auth_registries[@]} )); then
        selected+=("${auth_registries[$((idx - 1))]}")
      fi
    done
  fi

  if [[ ${#selected[@]} -eq 0 ]]; then
    return
  fi

  # Build AUTH_REGISTRIES string: "registry:username:password"
  local auth_entries=()
  local selected_hosts=()
  for reg in "${selected[@]}"; do
    local auth_b64
    auth_b64=$(jq -r --arg reg "$reg" '.auths[$reg].auth // empty' "$DOCKER_CONFIG" 2>/dev/null)

    if [[ -n "$auth_b64" ]]; then
      local creds
      creds=$(echo "$auth_b64" | base64 -d 2>/dev/null)
      local username="${creds%%:*}"
      local password="${creds#*:}"
      # Normalize registry hostname
      local host="$reg"
      host="${host#https://}"
      host="${host#http://}"
      host="${host%%/*}"
      auth_entries+=("${host}:${username}:${password}")
      selected_hosts+=("$host")
      echo "> Added auth for $host"
    else
      echo "> No credentials found for $reg (may use credential helper)"
    fi
  done

  if [[ ${#auth_entries[@]} -gt 0 ]]; then
    AUTH_REGISTRIES=$(printf "%s " "${auth_entries[@]}")
    AUTH_REGISTRIES="${AUTH_REGISTRIES% }"
    export AUTH_REGISTRIES

    EXTRA_REGISTRIES=$(printf "%s " "${selected_hosts[@]}")
    EXTRA_REGISTRIES="${EXTRA_REGISTRIES% }"
    export EXTRA_REGISTRIES
  fi
}

# check registry container if not exists
if ${KIND_EXPERIMENTAL_PROVIDER} ps -a --format '{{.Names}}' | grep -q docker_registry_proxy; then
  echo "> Registry container already exists"
else
  select_auth_registries

  echo "> Creating registry container"
  sudo install -d -m 0777 /opt/registry

  AUTH_ENV=()
  if [[ -n "$AUTH_REGISTRIES" ]]; then
    AUTH_ENV=(-e "AUTH_REGISTRIES=${AUTH_REGISTRIES}")
  fi

  ${KIND_EXPERIMENTAL_PROVIDER} run -d --restart=always --name docker_registry_proxy \
    --hostname docker-registry-proxy --network kind \
    -p 0.0.0.0:3128:3128 \
    -v /opt/registry/docker_mirror_cache:/docker_mirror_cache \
    -v /opt/registry/docker_mirror_certs:/ca \
    -e REGISTRIES="${DEFAULT_REGISTRIES} ${EXTRA_REGISTRIES}" \
    "${AUTH_ENV[@]}" \
    rpardini/docker-registry-proxy:0.6.5
  echo "> Waiting for registry container to start 10s"
  sleep 10
fi

echo "> Setting up container proxy"
KIND_NAME=kup
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
