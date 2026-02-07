#!/bin/bash

helm repo add headlamp https://kubernetes-sigs.github.io/headlamp/
helm install headlamp headlamp/headlamp --namespace kube-system

cat <<EOF | kubectl -n kube-system apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: headlamp-kube
spec:
  parentRefs:
  - name: kube
    namespace: kube-gateway
  hostnames:
  - "headlamp.kube.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - name: headlamp
      port: 80
      namespace: kube-system
EOF

echo "> Headlamp installed successfully"
