package config

import "os"

type Config struct {
	Port       string
	OrderURL   string
	PaymentURL string
}

func Load() *Config {
	return &Config{
		Port:       getEnv("PORT", "8000"),
		OrderURL:   getEnv("ORDERS_SERVICE_URL", "http://orders-service:8081"),
		PaymentURL: getEnv("PAYMENTS_SERVICE_URL", "http://payments-service:8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
