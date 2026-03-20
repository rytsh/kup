.DEFAULT_GOAL := help

.PHONY: create
create: ## Initialize the project
	@echo "Initializing project..."
	kind create cluster --config=configs/kind.yaml
	./scripts/kube/registry.sh
	./scripts/kube/cilium.sh
	./scripts/kube/metrics-server.sh
	./scripts/kube/cert-manager.sh
	./scripts/kube/gateway.sh
	./scripts/kube/argocd.sh

.PHONY: registry
registry: ## Add local registry
	docker rm -f docker_registry_proxy || true
	./scripts/kube/registry.sh

.PHONE: delete
delete: ## Delete the cluster
	@echo "Deleting cluster..."
	kind delete cluster -n kup

.PHONY: ca
ca: ## Get CA file in the ./tmp/ca.crt; chrome://certificate-manager/localcerts/usercerts
	kubectl -n kube-gateway get secrets ca -o jsonpath='{.data.tls\.crt}' | base64 -d > ./tmp/ca.crt

.PHONY: prometheus
prometheus: ## Add metrics/trace/logging
	./scripts/kube/prometheus.sh

.PHONY: pika
pika: ## Add pika configuration
	./scripts/kube/pika.sh

.PHONY: socks5
socks5: ## Add socks5 configuration
	./scripts/proxy/socks5.sh

.PHONY: help
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
