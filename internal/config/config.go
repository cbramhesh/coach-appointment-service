package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	DSN  string
}

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func LoadConfig() *Config {
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		getEnvOrDefault("DB_USER", "root"),
		getEnvOrDefault("DB_PASSWORD", ""),
		getEnvOrDefault("DB_HOST", "localhost"),
		getEnvOrDefault("DB_PORT", "3306"),
		getEnvOrDefault("DB_NAME", "coach_booking"),
	)

	return &Config{
		Port: port,
		DSN:  dsn,
	}
}
