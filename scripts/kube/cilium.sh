#!/bin/bash

echo "> Installing CRDs"
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.1/experimental-install.yaml

echo "> Add helm Cilium repo"
# add helm repo if not already added
helm repo add cilium https://helm.cilium.io/ || true
helm repo update

echo "> Install Cilium"
helm install cilium oci://quay.io/cilium/charts/cilium  --version 1.19.0 \
  --namespace kube-system \
  --set ipam.mode=kubernetes \
  --set socketLB.enabled=true \
  --set bpf.tproxy=true \
  --set bpf.masquerade=true \
  --set image.pullPolicy=IfNotPresent \
  --set gatewayAPI.enabled=true \
  --set k8sServiceHost=kind-control-plane \
  --set k8sServicePort=6443 \
  --set l7Proxy=true \
  --set kubeProxyReplacement=true \
  --set hubble.relay.enabled=true \
  --set hubble.ui.enabled=true \
  --set operator.replicas=1

echo "> Wait cilium is ready"
cilium status --wait

echo "> Set load balancer IPs"
cat <<EOF | kubectl apply -f -
apiVersion: "cilium.io/v2"
kind: CiliumLoadBalancerIPPool
metadata:
  name: "pool"
spec:
  blocks:
  - start: "10.0.10.1"
    stop: "10.0.10.100"
EOF

echo "> Cilium installed successfully"
echo "> Check IP Pools 'kubectl get ippools'"
