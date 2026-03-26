# observability-test-app

An OpenTelemetry observability test/demo application with 3 Go microservices, configurable chaos/fault injection, and Kubernetes deployment via Helm.

## Architecture

```
Client -> API Gateway (:8080) -> Order Service (:8082) -> Product Service (:8081)
                               -> Product Service (:8081)
```

- **API Gateway** - Proxies requests to backend services
- **Product Service** - In-memory product catalog (5 seeded products)
- **Order Service** - Creates orders, calls Product Service to validate items and calculate totals

All services export traces, metrics, and logs via OTLP (gRPC or HTTP). Supports Grafana Cloud native OTLP export.

## Quick Start

### Run locally

```bash
# Terminal 1: Product Service
OTEL_SERVICE_NAME=product SERVICE_PORT=8081 go run ./cmd/product

# Terminal 2: Order Service
OTEL_SERVICE_NAME=order SERVICE_PORT=8082 go run ./cmd/order

# Terminal 3: API Gateway
OTEL_SERVICE_NAME=gateway SERVICE_PORT=8080 go run ./cmd/gateway
```

Or use the Makefile:
```bash
make run-product  # in terminal 1
make run-order    # in terminal 2
make run-gateway  # in terminal 3
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
| `GRAFANA_OTLP_ENDPOINT` | _(empty)_ | Grafana Cloud OTLP endpoint (overrides generic) |
| `GRAFANA_API_TOKEN` | _(empty)_ | Grafana Cloud API token |
| `PRODUCT_SERVICE_URL` | `http://localhost:8081` | Product service base URL |
| `ORDER_SERVICE_URL` | `http://localhost:8082` | Order service base URL |

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

## Telemetry

- **Traces**: Full distributed tracing across all 3 services with W3C TraceContext propagation. Custom spans on business logic operations.
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
make build        # Build all binaries
make test         # Run tests
make docker-build # Build Docker images
make helm-lint    # Lint Helm chart
make clean        # Clean build artifacts
```
