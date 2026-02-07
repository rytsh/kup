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
  --set grafana.persistence.size=2Gi \
  --set 'grafana.additionalDataSources[0].name=Loki' \
  --set 'grafana.additionalDataSources[0].type=loki' \
  --set 'grafana.additionalDataSources[0].url=http://loki:3100' \
  --set 'grafana.additionalDataSources[0].access=proxy' \
  --set 'grafana.additionalDataSources[0].isDefault=false' \
  --set 'grafana.additionalDataSources[1].name=Tempo' \
  --set 'grafana.additionalDataSources[1].type=tempo' \
  --set 'grafana.additionalDataSources[1].url=http://tempo:3200' \
  --set 'grafana.additionalDataSources[1].access=proxy' \
  --set 'grafana.additionalDataSources[1].isDefault=false' \
  --set 'grafana.additionalDataSources[1].jsonData.tracesToLogsV2.datasourceUid=Loki' \
  --set 'grafana.additionalDataSources[1].jsonData.nodeGraph.enabled=true'
  #--set "grafana.adminPassword=awesomepassword"

echo "> [5/10] Add grafana repo"
helm repo add grafana https://grafana.github.io/helm-charts || true
helm repo add grafana-community https://grafana-community.github.io/helm-charts || true
helm repo update

echo "> [8/10] Install loki-stack"
helm install loki \
  --namespace kube-prometheus-stack \
  grafana/loki \
  --set deploymentMode=SingleBinary \
  --set loki.commonConfig.replication_factor=1 \
  --set loki.storage.type=filesystem \
  --set loki.schemaConfig.configs[0].from="2024-01-01" \
  --set loki.schemaConfig.configs[0].store=tsdb \
  --set loki.schemaConfig.configs[0].object_store=filesystem \
  --set loki.schemaConfig.configs[0].schema=v13 \
  --set loki.schemaConfig.configs[0].index.prefix=loki_index_ \
  --set loki.schemaConfig.configs[0].index.period=24h \
  --set loki.auth_enabled=false \
  --set singleBinary.replicas=1 \
  --set read.replicas=0 \
  --set write.replicas=0 \
  --set backend.replicas=0 \
  --set chunksCache.enabled=false \
  --set resultsCache.enabled=false

echo "> [9/10] Install promtail"
helm install promtail \
  --namespace kube-prometheus-stack \
  grafana/promtail \
  --set config.clients[0].url=http://loki:3100/loki/api/v1/push

echo "> [10/10] Install tempo"
helm install tempo \
  --namespace kube-prometheus-stack \
  grafana-community/tempo \
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
