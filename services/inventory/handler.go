package inventory

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("inventory-service")

type Handler struct {
	store InventoryStore
}

func NewHandler(store InventoryStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) ListStock(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "ListStock")
	defer span.End()

	stock, err := h.store.ListStock(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list inventory", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	span.SetAttributes(attribute.Int("inventory.count", len(stock)))
	slog.InfoContext(ctx, "listing inventory", "count", len(stock))
	writeJSON(w, http.StatusOK, stock)
}

func (h *Handler) GetStock(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "GetStock")
	defer span.End()

	id := r.PathValue("id")
	span.SetAttributes(attribute.String("product.id", id))

	qty, ok := h.store.GetStock(ctx, id)
	if !ok {
		slog.WarnContext(ctx, "product not in inventory", "id", id)
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "product not found in inventory"})
		return
	}
	slog.InfoContext(ctx, "stock retrieved", "product_id", id, "quantity", qty)
	writeJSON(w, http.StatusOK, StockEntry{ProductID: id, Quantity: qty})
}

func (h *Handler) ReserveStock(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "ReserveStock")
	defer span.End()

	var req struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	span.SetAttributes(
		attribute.String("product.id", req.ProductID),
		attribute.Int("reserve.quantity", req.Quantity),
	)

	remaining, err := h.store.Reserve(ctx, req.ProductID, req.Quantity)
	if err != nil {
		slog.WarnContext(ctx, "reservation failed", "product_id", req.ProductID, "quantity", req.Quantity, "error", err)
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}

	slog.InfoContext(ctx, "stock reserved", "product_id", req.ProductID, "reserved", req.Quantity, "remaining", remaining)
	writeJSON(w, http.StatusOK, map[string]any{
		"product_id": req.ProductID,
		"reserved":   req.Quantity,
		"remaining":  remaining,
	})
}

func (h *Handler) DiskUsage(w http.ResponseWriter, r *http.Request) {
	bytes, err := h.store.DiskUsage()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"bytes": bytes})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
