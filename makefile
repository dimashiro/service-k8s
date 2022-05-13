SHELL := /bin/bash
VERSION := 1.0


# This will output the help for each task. thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Show this help
	@printf "\033[33m%s:\033[0m\n" 'Available commands'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[32m%-11s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

all: retail

retail: ## build retail-api image
	docker build \
		-f build/docker/dockerfile \
		-t retail-api-amd64:$(VERSION) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

KIND_CLUSTER := starter-cluster
kind-up: ## create kind cluster and set default namespace
	kind create cluster \
		--image kindest/node:v1.21.10@sha256:84709f09756ba4f863769bdcabe5edafc2ada72d3c8c44d6515fc581b66b029c \
		--name $(KIND_CLUSTER) \
		--config deploy/k8s/kind/kind-config.yaml
	kubectl config set-context --current --namespace=retail-api-system
	
kind-down: ## delete kind cluster
	kind delete cluster --name $(KIND_CLUSTER)

kind-load: ## load docker image
	cd deploy/k8s/kind/retail-api-pod; kustomize edit set image retail-api-image=retail-api-amd64:$(VERSION)
	kind load docker-image retail-api-amd64:$(VERSION) --name $(KIND_CLUSTER)

kind-apply: ## build k8s pod
	kustomize build deploy/k8s/kind/retail-api-pod | kubectl apply -f -

kind-status: ## k8s statuses
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

kind-status-retail-api: ## k8s status retail-api pod
	kubectl get pods -o wide --watch

kind-logs: ## show k8s logs
	kubectl logs -l app=retail-api --all-containers=true -f --tail=100

kind-restart: ## restart retail-api-pod
	kubectl rollout restart deployment retail-api-pod

kind-update: ## build and restart
	all kind-load kind-restart

kind-update-apply: ## build and apply
	all kind-load kind-apply

kind-describe: ## show details
	kubectl describe pod -l app=retail-api

tidy: ## go tidy + go vendor
	go mod tidy
	go mod vendor
