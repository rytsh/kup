#!/bin/bash

echo "> Install Pika"

TMPDIR=$(mktemp -d)
cat <<EOF > "$TMPDIR/kustomization.yaml"
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - https://github.com/rakunlabs/pika/ci/kubernetes?ref=main

images:
  - name: ghcr.io/rakunlabs/pika
    newTag: v0.1.2

patches:
  - target:
      kind: HTTPRoute
      name: pika
    patch: |
      - op: replace
        path: /spec/hostnames/0
        value: pika.kube.com
      - op: replace
        path: /spec/parentRefs
        value:
          - name: kube
            namespace: kube-gateway
EOF
kubectl apply -k "$TMPDIR" -n pika && rm -rf "$TMPDIR"

echo "> Pika installed successfully"
