#!/bin/bash

#!/usr/bin/env bash

###################
# Prometheus Stack
###################

set -e

echo "> [1/10] PROMETHEUS STACK PART"

echo "> [2/10] Add prometheus stack repo"
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts || true
helm repo update

echo "> [3/10] Install prometheus stack"
helm install kube-prometheus-stack \
  --create-namespace \
  --namespace kube-prometheus-stack \
  prometheus-community/kube-prometheus-stack \
  --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage=4Gi \
  --set grafana.persistence.enabled=true \
  --set grafana.persistence.type=pvc \
  --set grafana.persistence.size=2Gi
  #--set "grafana.adminPassword=awesomepassword"

echo "> [5/10] Add grafana repo"
helm repo add grafana https://grafana.github.io/helm-charts || true
helm repo update

echo "> [8/10] Install loki-stack"
helm install loki \
  --namespace kube-prometheus-stack \
  grafana/loki-stack

echo "> [9/10] Install tempo"
helm install tempo \
  --namespace kube-prometheus-stack \
  grafana/tempo \
  --set persistence.enabled=true \
  --set persistence.size=5Gi

echo "> [10/10] Add grafana.kube.com"
cat <<EOF | kubectl -n kube-prometheus-stack apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: grafana-kube
spec:
  parentRefs:
  - name: kube
    namespace: kube-gateway
  hostnames:
  - "grafana.kube.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - name: kube-prometheus-stack-grafana
      port: 80
      namespace: kube-prometheus-stack
EOF
