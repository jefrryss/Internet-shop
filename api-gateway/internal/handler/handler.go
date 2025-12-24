package handler

import (
	"api-gateway/internal/service"
	"io"
	"log"
	"net/http"
)

type Handler struct {
	svc *service.GatewayService
}

func NewHandler(svc *service.GatewayService) *Handler {
	return &Handler{svc: svc}
}

// ProxyHandler создает функцию-обработчик, которая пересылает запрос на Host.
func (h *Handler) ProxyHandler(targetHost string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetURL := targetHost + r.URL.Path

		log.Printf("Proxying: %s %s -> %s", r.Method, r.URL.Path, targetURL)

		resp, err := h.svc.ProxyRequest(r.Method, targetURL, r.Body, r.Header)
		if err != nil {
			log.Printf("Gateway Error: %v", err)
			http.Error(w, "Service Unavailable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}
