package app

import "os"

type Config struct {
	HTTPAddr     string
	PostgresDSN  string
	MaxOpenConns int32
	MaxIdleConns int32
}

func LoadConfig() Config {
	return Config{
		HTTPAddr:     getEnv("HTTP_ADDR", ":8081"),
		PostgresDSN:  getEnv("POSTGRES_DSN", "postgres://postgres@127.0.0.1:55432/payment_service?sslmode=disable"),
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
