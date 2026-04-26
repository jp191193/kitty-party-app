// Package config handles application configuration loaded from environment variables.
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration.
type Config struct {
	AppName string
	AppEnv      string
	Port        string
	DatabaseURL string
}

// Load reads the .env file (if present) and returns a populated Config.
// In production, environment variables should be set directly in the host.
func Load() (*Config, error) {
	// Intentionally ignore error – .env file is optional in production.
	_ = godotenv.Load()

	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "postgres")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := getEnv("DB_NAME", "kitty_party")

	cfg := &Config{
		AppName: getEnv("APP_NAME", "kitty-party-app"),
		AppEnv:  getEnv("APP_ENV", "development"),
		Port:    getEnv("APP_PORT", "8080"),
		DatabaseURL: fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", 
			dbUser, dbPass, dbHost, dbPort, dbName),
	}

	if cfg.Port == "" {
		return nil, fmt.Errorf("APP_PORT must not be empty")
	}

	return cfg, nil
}

// getEnv returns the value of the environment variable named by key,
// or fallback if the variable is not set.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
