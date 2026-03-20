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
    newTag: v0.1.3

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


cat <<EOF | kubectl apply -n pika -f -
apiVersion: external-secrets.io/v1
kind: SecretStore
metadata:
  name: pika
spec:
  provider:
    webhook:
      url: "http://pika.pika.svc:9090/data/{{ .remoteRef.key }}"
      result:
        jsonPath: "$"
      headers:
        Accept: application/octet-stream
EOF

echo "> Pika installed successfully"
