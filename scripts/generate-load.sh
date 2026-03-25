#!/usr/bin/env bash
set -euo pipefail

GATEWAY_URL="${GATEWAY_URL:-http://localhost:8080}"
INTERVAL="${INTERVAL:-1}"

echo "Generating load against ${GATEWAY_URL} every ${INTERVAL}s"
echo "Press Ctrl+C to stop"

while true; do
  # List products
  curl -s "${GATEWAY_URL}/products" > /dev/null && echo "[$(date +%T)] GET /products - OK" || echo "[$(date +%T)] GET /products - FAIL"

  # Get specific product
  curl -s "${GATEWAY_URL}/products/prod-1" > /dev/null && echo "[$(date +%T)] GET /products/prod-1 - OK" || echo "[$(date +%T)] GET /products/prod-1 - FAIL"

  # Create order
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${GATEWAY_URL}/orders" \
    -H "Content-Type: application/json" \
    -d '{"items":[{"product_id":"prod-1","quantity":2},{"product_id":"prod-3","quantity":1}]}')
  echo "[$(date +%T)] POST /orders - ${STATUS}"

  # List orders
  curl -s "${GATEWAY_URL}/orders" > /dev/null && echo "[$(date +%T)] GET /orders - OK" || echo "[$(date +%T)] GET /orders - FAIL"

  sleep "${INTERVAL}"
done
