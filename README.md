# Kup

Kubernetes installation guide in WSL.

## Prerequisites

- WSL 2
- Windows Terminal
- A Linux distribution installed in WSL
- Kernel installed `https://github.com/Locietta/xanmod-kernel-WSL2` check `./scripts/kernel/install.sh` for installation instructions.

## Installation

All tools are installed in the `~/bin` directory. Make sure to add it to your PATH.

```sh
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
```

```sh
./scripts/tools/kind.sh
./scripts/tools/kubectl.sh
./scripts/tools/cilium.sh
./scripts/tools/helm.sh
```

Kind install a Kubernetes cluster with Cilium as the CNI plugin.

```sh
kind create cluster --config=configs/kind.yaml
# kubeconfig file ~/.kube/config.
```

Install registry, cilium network and metric server in the cluster.

```sh
# Add registry for kind cluster
./scripts/kube/registry.sh
# Install cilium network in the cluster.
./scripts/kube/cilium.sh
# Install metric server in the cluster.
./scripts/kube/metrics-server.sh
# Add cert manager in the cluster and install *.kube.com
./scripts/kube/cert-manager.sh
# Add prometheus and grafana in the cluster.
./scripts/kube/prometheus.sh
```

Get CA certificate for cert-manager to trust the cluster.

```sh
kubectl -n kube-gateway get secrets ca -o jsonpath='{.data.tls\.crt}' | base64 -d > ./tmp/ca.crt
# chrome://certificate-manager/localcerts/usercerts
```

## IDE

Lens IDE is a popular Kubernetes IDE that provides a graphical interface for managing Kubernetes clusters.
> https://freelensapp.github.io/

Headlamp is web-based Kubernetes IDE download with `helm` and install in the cluster.

```sh
./scripts/kube/headlamp.sh
```

## Access cluster

Use `socks5` proxy to access the cluster.

```sh
# check 'kubectl get gateways.gateway.networking.k8s.io -n kube-gateway' for IP address of the gateway (kube).
./scripts/proxy/socks5.sh
```

In browser add extension `FoxyProxy` and set proxy for `*.kube.com` to `socks5://localhost:1080`.  
Also add proxy pattern `*://*.kube.com/`
