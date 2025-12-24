package main

import (
	"api-gateway/internal/config"
	"api-gateway/internal/handler"
	"api-gateway/internal/router"
	"api-gateway/internal/service"
	"log"
	"net/http"
)

func main() {
	//Загрузка конфига
	cfg := config.Load()

	//Инициализация сервиса
	svc := service.NewGatewayService()

	//Инициализация хендлера
	h := handler.NewHandler(svc)

	// Инициализация роутера
	r := router.NewRouter(h, cfg.OrderURL, cfg.PaymentURL)

	//Запуск сервера
	log.Printf("API Gateway работает на порте: %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Ошибка старта api-gateway: %v", err)
	}
}
