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
    newTag: latest
EOF
kubectl apply -k "$TMPDIR" -n pika && rm -rf "$TMPDIR"

cat <<EOF | kubectl apply -n pika -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: pika
spec:
  parentRefs:
    - name: kube
      namespace: kube-gateway
  hostnames:
    - "pika.kube.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /
      backendRefs:
        - name: pika
          port: 8080
EOF

echo "> Wait for external-secrets webhook to be ready"
kubectl -n external-secrets rollout status deployment/external-secrets-webhook --timeout=5m
kubectl -n external-secrets wait --for=condition=Available deployment/external-secrets-webhook --timeout=5m

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
