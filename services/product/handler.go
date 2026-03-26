package product

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("product-service")

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "ListProducts")
	defer span.End()

	products := h.store.List()
	span.SetAttributes(attribute.Int("product.count", len(products)))
	slog.InfoContext(ctx, "listing products", "count", len(products))
	writeJSON(w, http.StatusOK, products)
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "GetProduct")
	defer span.End()

	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing product id"})
		return
	}
	span.SetAttributes(attribute.String("product.id", id))

	product, err := h.store.Get(id)
	if err != nil {
		slog.WarnContext(ctx, "product not found", "id", id)
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	slog.InfoContext(ctx, "product viewed", "id", id, "name", product.Name)
	writeJSON(w, http.StatusOK, product)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
