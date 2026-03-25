#!/usr/bin/env bash
set -euo pipefail

echo "=== Observability Test App - Demo Scenario ==="
echo ""
echo "This script demonstrates progressively enabling chaos features."
echo "Run each service in separate terminals, then run this script."
echo ""

GATEWAY_URL="${GATEWAY_URL:-http://localhost:8080}"

echo "Step 1: Verify baseline (healthy requests)"
echo "---"
for i in $(seq 1 5); do
  curl -s "${GATEWAY_URL}/products" > /dev/null && echo "  Request ${i}: OK" || echo "  Request ${i}: FAIL"
done
echo ""

echo "Step 2: Enable error injection"
echo "  Restart services with: CHAOS_ERROR_ROUTES=\"/products:0.5,/orders:0.3\""
echo "  Press Enter when ready..."
read -r
for i in $(seq 1 10); do
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${GATEWAY_URL}/products")
  echo "  Request ${i}: HTTP ${STATUS}"
done
echo ""

echo "Step 3: Enable latency injection"
echo "  Restart services with: CHAOS_LATENCY_ROUTES=\"/products:500ms,/orders:1s\""
echo "  Press Enter when ready..."
read -r
for i in $(seq 1 5); do
  START=$(date +%s%N)
  curl -s "${GATEWAY_URL}/products" > /dev/null
  END=$(date +%s%N)
  DURATION=$(( (END - START) / 1000000 ))
  echo "  Request ${i}: ${DURATION}ms"
done
echo ""

echo "Step 4: Enable log volume"
echo "  Restart services with: CHAOS_LOG_VOLUME_ENABLED=true CHAOS_LOG_RATE_PER_SEC=10"
echo "  Watch the service logs for generated entries."
echo ""

echo "Step 5: Enable CPU/memory load"
echo "  Restart services with: CHAOS_CPU_LOAD_ENABLED=true CHAOS_CPU_LOAD_PERCENT=50"
echo "  Monitor CPU usage in your observability backend."
echo ""

echo "Done! Check your observability backend for traces, metrics, and logs."
