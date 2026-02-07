#!/bin/bash

echo "> Add cert-manager repo"
helm repo add jetstack https://charts.jetstack.io || true
helm repo update

echo "> Install cert-manager"
helm install cert-manager jetstack/cert-manager \
  --create-namespace \
  --namespace cert-manager \
  --set config.apiVersion="controller.config.cert-manager.io/v1alpha1" \
  --set config.kind="ControllerConfiguration" \
  --set crds.enabled=true \
  --set config.enableGatewayAPI=true

kubectl create namespace kube-gateway

echo "> Create CA Cluster Issuer"
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: self-signed
  namespace: kube-gateway
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: ca
  namespace: kube-gateway
spec:
  isCA: true
  privateKey:
    algorithm: ECDSA
    size: 256
  secretName: ca
  commonName: ca
  issuerRef:
    name: self-signed
    kind: Issuer
---
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: ca-issuer
  namespace: kube-gateway
spec:
  ca:
    secretName: ca
EOF

echo "> Add *.kube.com address gateway"
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: kube
  namespace: kube-gateway
  annotations:
    cert-manager.io/issuer: ca-issuer
spec:
  gatewayClassName: cilium
  listeners:
    - name: kube-subdomain
      hostname: "*.kube.com"
      port: 443
      protocol: HTTPS
      allowedRoutes:
        namespaces:
          from: All
      tls:
        mode: Terminate
        certificateRefs:
          - name: kube-com-tls
    - name: kube
      hostname: kube.com
      port: 443
      protocol: HTTPS
      allowedRoutes:
        namespaces:
          from: All
      tls:
        mode: Terminate
        certificateRefs:
          - name: kube-com-tls
EOF

echo "> Add proxy address gateway"
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: kube-proxy
  namespace: kube-gateway
spec:
  gatewayClassName: cilium
  listeners:
    - name: kube-proxy
      hostname: "proxy"
      port: 80
      protocol: HTTP
      allowedRoutes:
        namespaces:
          from: All
EOF
