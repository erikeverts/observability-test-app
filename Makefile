.PHONY: build test run-gateway run-product run-order docker-build docker-push helm-lint clean

REGISTRY ?= ghcr.io/erikeverts/observability-test-app
TAG ?= latest

build:
	go build ./...

test:
	go test ./...

run-gateway:
	OTEL_SERVICE_NAME=gateway SERVICE_PORT=8080 go run ./cmd/gateway

run-product:
	OTEL_SERVICE_NAME=product SERVICE_PORT=8081 go run ./cmd/product

run-order:
	OTEL_SERVICE_NAME=order SERVICE_PORT=8082 go run ./cmd/order

docker-build:
	docker build -f build/Dockerfile.gateway -t $(REGISTRY)/gateway:$(TAG) .
	docker build -f build/Dockerfile.product -t $(REGISTRY)/product:$(TAG) .
	docker build -f build/Dockerfile.order -t $(REGISTRY)/order:$(TAG) .

docker-push:
	docker push $(REGISTRY)/gateway:$(TAG)
	docker push $(REGISTRY)/product:$(TAG)
	docker push $(REGISTRY)/order:$(TAG)

helm-lint:
	helm lint deploy/helm/observability-test-app

clean:
	rm -f gateway product order
	go clean ./...
