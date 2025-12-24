package router

import (
	"api-gateway/internal/handler"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *handler.Handler, orderHost, paymentHost string) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(corsMiddleware)

	// Маршруты
	r.HandleFunc("/orders", h.ProxyHandler(orderHost))
	r.HandleFunc("/orders/*", h.ProxyHandler(orderHost))

	r.HandleFunc("/accounts", h.ProxyHandler(paymentHost))
	r.HandleFunc("/accounts/*", h.ProxyHandler(paymentHost))

	return r
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-Id, Accept, Origin")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
