package config

import "os"

type Config struct {
	HTTPPort     string
	DBDSN        string
	KafkaBrokers []string
	KafkaGroupID string
}

func Load() *Config {
	return &Config{
		HTTPPort:     getEnv("HTTP_PORT", "8080"),
		DBDSN:        getEnv("DB_DSN", "postgres://payments:payments@localhost:5432/payments?sslmode=disable"),
		KafkaBrokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		KafkaGroupID: getEnv("KAFKA_GROUP_ID", "payments-service"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
