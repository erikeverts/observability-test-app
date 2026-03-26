# observability-test-app

An OpenTelemetry observability test/demo application with 4 Go microservices, configurable chaos/fault injection, and Kubernetes deployment via Helm.

## Architecture

```
Client -> API Gateway (:8080) -> Order Service (:8082) -> Inventory Service (:8084) -> reserve stock
                               -> Product Service (:8081)                           -> Product Service (:8081)
                               -> Inventory Service (:8084)

Chaos Dashboard (:8083) -> controls chaos on all services via /admin/chaos API
```

- **API Gateway** - Proxies requests to backend services
- **Product Service** - In-memory product catalog (5 seeded products)
- **Order Service** - Creates orders, calls Product Service and Inventory Service to validate items, reserve stock, and calculate totals
- **Inventory Service** - Tracks stock levels per product with an append-only ledger on disk (StatefulSet with PVC)
- **Chaos Dashboard** - Web UI for controlling chaos/fault injection across all services at runtime

All services export traces, metrics, and logs via OTLP (gRPC or HTTP). Supports Grafana Cloud native OTLP export.

## Quick Start

### Run locally

```bash
# Terminal 1: Product Service
OTEL_SERVICE_NAME=product SERVICE_PORT=8081 go run ./cmd/product

# Terminal 2: Inventory Service
OTEL_SERVICE_NAME=inventory SERVICE_PORT=8084 INVENTORY_DATA_DIR=./data go run ./cmd/inventory

# Terminal 3: Order Service
OTEL_SERVICE_NAME=order SERVICE_PORT=8082 go run ./cmd/order

# Terminal 4: API Gateway
OTEL_SERVICE_NAME=gateway SERVICE_PORT=8080 go run ./cmd/gateway

# Terminal 5 (optional): Chaos Dashboard
SERVICE_PORT=8083 go run ./cmd/dashboard
```

Or use the Makefile:
```bash
make run-product    # in terminal 1
make run-inventory  # in terminal 2
make run-order      # in terminal 3
make run-gateway    # in terminal 4
make run-dashboard  # in terminal 5 (optional)
```

### Test endpoints

```bash
# List products
curl http://localhost:8080/products

# Get a product
curl http://localhost:8080/products/prod-1

# Create an order
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"items":[{"product_id":"prod-1","quantity":2}]}'

# List orders
curl http://localhost:8080/orders

# Check inventory
curl http://localhost:8080/inventory

# Check stock for a product
curl http://localhost:8080/inventory/prod-1

# Health check
curl http://localhost:8080/health
```

## Configuration

All configuration is via environment variables:

| Variable | Default | Description |
|---|---|---|
| `OTEL_SERVICE_NAME` | `unknown` | Service identity |
| `SERVICE_PORT` | `8080` | HTTP listen port |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `localhost:4317` | OTLP collector endpoint |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | `grpc` | `grpc` or `http/protobuf` |
| `OTEL_EXPORTER_OTLP_INSECURE` | `true` | Use insecure (plaintext) connection |
| `OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY` | `delta` | Metric temporality: `delta` or `cumulative` (use `cumulative` for Prometheus) |
| `GRAFANA_OTLP_ENDPOINT` | _(empty)_ | Grafana Cloud OTLP endpoint (overrides generic) |
| `GRAFANA_API_TOKEN` | _(empty)_ | Grafana Cloud API token |
| `PRODUCT_SERVICE_URL` | `http://localhost:8081` | Product service base URL |
| `ORDER_SERVICE_URL` | `http://localhost:8082` | Order service base URL |
| `INVENTORY_SERVICE_URL` | `http://localhost:8084` | Inventory service base URL |
| `INVENTORY_DATA_DIR` | `/data` | Inventory ledger data directory |

### Chaos Configuration

| Variable | Default | Description |
|---|---|---|
| `CHAOS_ERROR_ROUTES` | _(empty)_ | Error injection, e.g. `"/products:0.5,/orders:0.3"` |
| `CHAOS_LATENCY_ROUTES` | _(empty)_ | Latency injection, e.g. `"/products:200ms,/orders:1s"` |
| `CHAOS_CPU_LOAD_ENABLED` | `false` | Enable CPU load simulation |
| `CHAOS_CPU_LOAD_PERCENT` | `0` | Target CPU utilization percentage |
| `CHAOS_MEM_LOAD_ENABLED` | `false` | Enable memory load simulation |
| `CHAOS_MEM_LOAD_MB` | `0` | Memory to allocate in MB |
| `CHAOS_LOG_VOLUME_ENABLED` | `false` | Enable log generation |
| `CHAOS_LOG_RATE_PER_SEC` | `0` | Log lines per second |
| `CHAOS_LOG_PATTERN` | `mixed` | `info`, `warn`, `error`, or `mixed` |
| `CHAOS_DISK_FILL_ENABLED` | `false` | Enable disk fill simulation |
| `CHAOS_DISK_FILL_PATH` | `/data` | Directory to write fill data to |
| `CHAOS_DISK_FILL_RATE_MB` | `1` | MB per second to write |

## Chaos Dashboard

The chaos dashboard provides a web UI for controlling chaos/fault injection across all services at runtime. No restart required — changes take effect immediately.

### Access locally

```bash
# Start the dashboard (after starting the other services)
make run-dashboard
# Open http://localhost:8083
```

### Access in Kubernetes

```bash
kubectl port-forward svc/<release-name>-dashboard 8083:8083
# Open http://localhost:8083
```

### Features

- **Per-service controls**: Configure error injection, latency injection, CPU/memory load, log volume, and disk fill for each service independently
- **Route-level granularity**: Set error rates and latency per HTTP path (e.g., `/products` at 50% error rate)
- **Presets**: One-click scenarios — Reset All, Slow Responses, Error Storm, Noisy Logs, CPU Burn, Disk Fill, Full Chaos
- **Live state**: Dashboard reads current chaos config from each service on load

### Runtime Chaos API

Each service exposes admin endpoints for programmatic control:

```bash
# Get current chaos config
curl http://localhost:8081/admin/chaos

# Set chaos config
curl -X PUT http://localhost:8081/admin/chaos \
  -H "Content-Type: application/json" \
  -d '{"error_routes":{"/products":0.5},"latency_routes":{"/products":"200ms"},"cpu_load":{"enabled":false,"percent":0},"mem_load":{"enabled":false,"mb":0},"log_volume":{"enabled":false,"rate_per_sec":0,"pattern":"mixed"}}'
```

## Telemetry

- **Traces**: Full distributed tracing across all 4 services with W3C TraceContext propagation. Custom spans on business logic operations.
- **Metrics**: Custom counters for orders created, product views, and chaos injection events. Plus automatic HTTP metrics from `otelhttp`.
- **Logs**: Structured JSON via `log/slog` with `otelslog` bridge. All logs include `trace_id` and `span_id` for correlation.

## Kubernetes Deployment

```bash
# Build Docker images
make docker-build

# Lint the Helm chart
make helm-lint

# Deploy
helm install demo deploy/helm/observability-test-app \
  --set otel.endpoint=otel-collector:4317

# Deploy with chaos enabled
helm install demo deploy/helm/observability-test-app \
  --set otel.endpoint=otel-collector:4317 \
  --set chaos.errorRoutes="/products:0.1" \
  --set chaos.latencyRoutes="/orders:500ms"

# Deploy with Grafana Cloud
helm install demo deploy/helm/observability-test-app \
  --set grafana.endpoint=otlp-gateway-prod-us-central-0.grafana.net \
  --set grafana.apiToken=<base64-encoded-instance:token>
```

## Demo Scripts

```bash
# Generate continuous load
./scripts/generate-load.sh

# Step-by-step chaos demo
./scripts/demo-scenario.sh
```

## Development

```bash
make build          # Build all binaries
make test           # Run tests
make helm-lint      # Lint Helm chart
make run-inventory  # Run inventory service locally
make run-dashboard  # Run chaos dashboard locally
make clean          # Clean build artifacts
```
