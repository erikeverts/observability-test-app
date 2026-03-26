package order

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/erikeverts/observability-test-app/internal/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("order-service")

type Handler struct {
	store         *Store
	productClient *ProductClient
	orderCounter  int
}

func NewHandler(store *Store, productClient *ProductClient) *Handler {
	return &Handler{
		store:         store,
		productClient: productClient,
	}
}

func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "ListOrders")
	defer span.End()

	orders := h.store.List()
	span.SetAttributes(attribute.Int("order.count", len(orders)))
	slog.InfoContext(ctx, "listing orders", "count", len(orders))
	writeJSON(w, http.StatusOK, orders)
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "GetOrder")
	defer span.End()

	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing order id"})
		return
	}
	span.SetAttributes(attribute.String("order.id", id))

	order, err := h.store.Get(id)
	if err != nil {
		slog.WarnContext(ctx, "order not found", "id", id)
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	slog.InfoContext(ctx, "order retrieved", "id", id, "status", string(order.Status))
	writeJSON(w, http.StatusOK, order)
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "CreateOrder")
	defer span.End()

	var req model.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if len(req.Items) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "order must have at least one item"})
		return
	}

	span.SetAttributes(attribute.Int("order.item_count", len(req.Items)))

	var total float64
	for _, item := range req.Items {
		_, productSpan := tracer.Start(ctx, "LookupProduct")
		productSpan.SetAttributes(attribute.String("product.id", item.ProductID))

		product, err := h.productClient.GetProduct(ctx, item.ProductID)
		if err != nil {
			productSpan.End()
			slog.ErrorContext(ctx, "product lookup failed", "product_id", item.ProductID, "error", err)
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("product lookup failed: %v", err)})
			return
		}
		total += product.Price * float64(item.Quantity)
		productSpan.End()
	}

	h.orderCounter++
	order := &model.Order{
		ID:        fmt.Sprintf("order-%d", h.orderCounter),
		Items:     req.Items,
		Status:    model.OrderStatusConfirmed,
		Total:     total,
		CreatedAt: time.Now(),
	}

	h.store.Save(order)
	span.SetAttributes(
		attribute.String("order.id", order.ID),
		attribute.Float64("order.total", order.Total),
	)
	slog.InfoContext(ctx, "order created", "id", order.ID, "total", order.Total, "items", len(order.Items))
	writeJSON(w, http.StatusCreated, order)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
