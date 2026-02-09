#!/bin/bash

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
