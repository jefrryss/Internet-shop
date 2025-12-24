package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"payment-service/internal/domain/account"

	"github.com/go-chi/chi/v5"
)

type AccountHandler struct {
	repo account.Repository
}

func NewAccountHandler(r account.Repository) *AccountHandler {
	return &AccountHandler{repo: r}
}

func (h *AccountHandler) Register(r chi.Router) {
	r.Post("/", h.Create)
	r.Post("/deposit", h.Deposit)
	r.Get("/balance", h.Balance)
}

// Вспомогательная функция для получения ID пользователя
func userID(r *http.Request) (int64, error) {
	uidStr := r.Header.Get("X-User-Id")
	if uidStr == "" {
		return 0, account.ErrAccountNotFound
	}
	return strconv.ParseInt(uidStr, 10, 64)
}

func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, err := userID(r)
	if err != nil {
		http.Error(w, "Некорректный или отсутствующий X-User-Id", http.StatusBadRequest)
		return
	}

	err = h.repo.Create(r.Context(), account.NewAccount(uid))
	if err != nil {
		http.Error(w, "Счёт уже существует или ошибка БД", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *AccountHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	uid, err := userID(r)
	if err != nil {
		http.Error(w, "Некорректный X-User-Id", http.StatusBadRequest)
		return
	}

	var req struct {
		Amount int64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректное тело запроса", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, "Сумма должна быть больше нуля", http.StatusBadRequest)
		return
	}

	err = h.repo.Deposit(r.Context(), uid, req.Amount)
	if err != nil {
		if err == account.ErrAccountNotFound {
			http.Error(w, "Счёт не найден", http.StatusNotFound)
			return
		}
		http.Error(w, "Ошибка обновления баланса", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AccountHandler) Balance(w http.ResponseWriter, r *http.Request) {
	uid, err := userID(r)
	if err != nil {
		http.Error(w, "Некорректный X-User-Id", http.StatusBadRequest)
		return
	}

	acc, err := h.repo.GetByUserID(r.Context(), uid)
	if err != nil {

		if err == account.ErrAccountNotFound {
			http.Error(w, "Счёт не найден", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{
		"balance": acc.Balance,
	})
}
