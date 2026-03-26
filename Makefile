.PHONY: build test run-gateway run-product run-order run-inventory run-dashboard helm-lint clean

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

run-inventory:
	OTEL_SERVICE_NAME=inventory SERVICE_PORT=8084 INVENTORY_DATA_DIR=./data go run ./cmd/inventory

run-dashboard:
	SERVICE_PORT=8083 go run ./cmd/dashboard

helm-lint:
	helm lint deploy/helm/observability-test-app

clean:
	go clean ./...
