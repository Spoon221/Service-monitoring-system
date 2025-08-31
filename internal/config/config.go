package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	LogLevel    string
	CheckInterval int // интервал проверки в секундах
}

func Load() (*Config, error) {
	// Загружаем .env файл если он существует
	godotenv.Load()

	port := getEnv("PORT", "8080")
	databaseURL := getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/service_monitor?sslmode=disable")
	logLevel := getEnv("LOG_LEVEL", "info")
	checkInterval, _ := strconv.Atoi(getEnv("CHECK_INTERVAL", "30"))

	return &Config{
		Port:          port,
		DatabaseURL:   databaseURL,
		LogLevel:      logLevel,
		CheckInterval: checkInterval,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
