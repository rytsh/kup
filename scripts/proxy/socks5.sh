#!/usr/bin/env bash

export KIND_EXPERIMENTAL_PROVIDER=${KIND_EXPERIMENTAL_PROVIDER:-docker}

sudo ip route add 10.0.10.0/24 via $(${KIND_EXPERIMENTAL_PROVIDER} inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' kup-control-plane)

if ${KIND_EXPERIMENTAL_PROVIDER} ps | grep -q kube-proxy; then
  ${KIND_EXPERIMENTAL_PROVIDER} restart kube-proxy
else
  ${KIND_EXPERIMENTAL_PROVIDER} run -d --restart=always \
  --name kube-proxy \
  --network kind \
  -p 1080:1080 \
  -v ${PWD}/scripts/proxy/kube-proxy.yaml:/turna.yaml \
  ghcr.io/rakunlabs/turna:v0.8.21
fi
