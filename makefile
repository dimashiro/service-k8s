SHELL := /bin/bash

VERSION := 1.0

all: retail

retail:
	docker build \
		-f build/docker/dockerfile \
		-t retail-api-amd64:$(VERSION) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

KIND_CLUSTER := starter-cluster
kind-up:
	kind create cluster \
		--image kindest/node:v1.21.10@sha256:84709f09756ba4f863769bdcabe5edafc2ada72d3c8c44d6515fc581b66b029c \
		--name $(KIND_CLUSTER) \
		--config deploy/k8s/kind/kind-config.yaml
	kubectl config set-context --current --namespace=retail-api-system
	
kind-down:
	kind delete cluster --name $(KIND_CLUSTER)

kind-load:
	cd deploy/k8s/kind/retail-api-pod; kustomize edit set image retail-api-image=retail-api-amd64:$(VERSION)
	kind load docker-image retail-api-amd64:$(VERSION) --name $(KIND_CLUSTER)

kind-apply:
	kustomize build deploy/k8s/kind/retail-api-pod | kubectl apply -f -

kind-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

kind-status-retail-api:
	kubectl get pods -o wide --watch

kind-logs:
	kubectl logs -l app=retail-api --all-containers=true -f --tail=100

kind-restart:
	kubectl rollout restart deployment retail-api-pod

kind-update: all kind-load kind-restart

kind-update-apply: all kind-load kind-apply

kind-describe:
	kubectl describe pod -l app=retail-api

tidy:
	go mod tidy
	go mod vendor
