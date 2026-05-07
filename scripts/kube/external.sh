#!/bin/bash

echo "> Install external-secrets"

helm repo add external-secrets https://charts.external-secrets.io

helm install external-secrets \
   external-secrets/external-secrets \
    -n external-secrets \
    --create-namespace \
    --wait \
    --timeout 10m

echo "> Wait for external-secrets webhook to be ready"
kubectl -n external-secrets rollout status deployment/external-secrets-webhook --timeout=5m
kubectl -n external-secrets wait --for=condition=Available deployment/external-secrets-webhook --timeout=5m

echo "> external-secrets installed"
