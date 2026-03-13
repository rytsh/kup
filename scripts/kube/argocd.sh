#!/bin/bash

echo "> Install ArgoCD"

kubectl create namespace argocd || true
kubectl apply -n argocd --server-side --force-conflicts -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Disable TLS on argocd-server (TLS terminated at gateway)
kubectl -n argocd patch configmap argocd-cmd-params-cm --type merge -p '{"data":{"server.insecure":"true"}}'
kubectl -n argocd rollout restart deployment argocd-server

cat <<EOF | kubectl -n argocd apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: argocd
spec:
  parentRefs:
  - name: kube
    namespace: kube-gateway
  hostnames:
  - "argocd.kube.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - name: argocd-server
      port: 80
      namespace: argocd
EOF


echo "> ArgoCD installed successfully"
