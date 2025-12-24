package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"order-service/internal/domain/order"
	"order-service/internal/infrastructure/persistence"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	repo *persistence.Repository
}

func NewHandler(repo *persistence.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Register(r chi.Router) {
	r.Post("/", h.CreateOrder)
	r.Get("/", h.ListOrders)
	r.Get("/{id}", h.GetOrder)
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-Id")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)
	if userID == 0 {
		http.Error(w, "X-User-Id header required", http.StatusUnauthorized)
		return
	}

	var req struct {
		Amount int64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	orderID := uuid.New().String()

	newOrder := &order.Order{
		ID:     orderID,
		UserID: userID,
		Amount: req.Amount,
		Status: order.StatusPending,
	}

	event := order.PaymentRequestEvent{
		MessageID: uuid.New().String(),
		OrderID:   orderID,
		UserID:    userID,
		Amount:    req.Amount,
	}

	err := h.repo.CreateOrderWithOutbox(r.Context(), newOrder, event)
	if err != nil {
		http.Error(w, "Не получилось создать заказ", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"order_id": orderID,
		"status":   "PENDING",
	})
}

func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-Id")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	orders, err := h.repo.GetByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ord, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ord)
}
