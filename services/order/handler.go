package order

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/erikeverts/observability-test-app/internal/model"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("order-service")

type Handler struct {
	store           OrderStore
	productClient   *ProductClient
	inventoryClient *InventoryClient
}

func NewHandler(store OrderStore, productClient *ProductClient, inventoryClient *InventoryClient) *Handler {
	return &Handler{
		store:           store,
		productClient:   productClient,
		inventoryClient: inventoryClient,
	}
}

func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "ListOrders")
	defer span.End()

	orders, err := h.store.List(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list orders", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
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

	order, err := h.store.Get(ctx, id)
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

		_, reserveSpan := tracer.Start(ctx, "ReserveInventory")
		reserveSpan.SetAttributes(attribute.String("product.id", item.ProductID), attribute.Int("quantity", item.Quantity))

		_, err = h.inventoryClient.Reserve(ctx, item.ProductID, item.Quantity)
		if err != nil {
			reserveSpan.End()
			slog.WarnContext(ctx, "inventory reservation failed", "product_id", item.ProductID, "error", err)
			writeJSON(w, http.StatusConflict, map[string]string{"error": fmt.Sprintf("inventory reservation failed: %v", err)})
			return
		}
		reserveSpan.End()
	}

	order := &model.Order{
		ID:        "ord-" + uuid.NewString(),
		Items:     req.Items,
		Status:    model.OrderStatusConfirmed,
		Total:     total,
		CreatedAt: time.Now(),
	}

	if err := h.store.Save(ctx, order); err != nil {
		slog.ErrorContext(ctx, "failed to save order", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
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
