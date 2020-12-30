SHELL := /bin/bash

export PROJECT = service-template

# ==============================================================================
# Building containers

all: api metrics

api:
	docker build \
		-f dockerfile.service-api \
		-t service-api-amd64:1.0 \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

metrics:
	docker build \
		-f dockerfile.metrics \
		-t metrics-amd64:1.0 \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

# ==============================================================================
# Running with docker-compose

run: up seed

up:
	docker-compose up --detach --remove-orphans

down:
	docker-compose down --remove-orphans

logs:
	docker-compose logs -f

# ==============================================================================
# Running with k8s (using Kind)

kind-up:
	kind create cluster --image kindest/node:v1.20.0 --name service-template-cluster --config infra/k8s/dev/kind-config.yaml

kind-down:
	kind delete cluster --name service-template-cluster

kind-load:
	kind load docker-image service-api-amd64:1.0 --name service-template-cluster
	kind load docker-image metrics-amd64:1.0 --name service-template-cluster

kind-services:
	kustomize build infra/k8s/dev | kubectl apply -f -

kind-update-api: api
	kind load docker-image service-api-amd64:1.0 --name service-template-cluster
	kubectl delete pods -lapp=service-api

kind-update-metrics: metrics
	kind load docker-image metrics-amd64:1.0 --name service-template-cluster
	kubectl delete pods -lapp=service-api

kind-logs:
	kubectl logs -lapp=service-api --all-containers=true -f

kind-status:
	kubectl get nodes
	kubectl get pods --watch

kind-status-full:
	kubectl describe pod -lapp=service-api

kind-shell:
	kubectl exec -it $(shell kubectl get pods | grep service-api | cut -c1-26) --container app -- /bin/sh

kind-delete:
	kustomize build infra/k8s/dev | kubectl delete -f -

# ==============================================================================
# Administration

migrate:
	go run app/service-admin/main.go migrate

seed: migrate
	go run app/service-admin/main.go seed

# ==============================================================================
# Running tests within the local computer

test:
	go test ./... -count=1
	staticcheck ./...

coverprofile:
	go test -v -coverpkg=./... -coverprofile=cover.out ./...
	go tool cover -html cover.out
	go tool cover -func cover.out

# ==============================================================================
# Modules support

deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

tidy:
	go mod tidy
	go mod vendor
