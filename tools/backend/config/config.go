package config

import "os"

// Config holds all application configuration.
type Config struct {
	DBType        string // "postgres", "sqlite", "mysql"
	DSN           string // connection string
	ServerPort    string
	AdminPassword string
	JWTSecret     string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		DBType:        getEnv("DB_TYPE", "postgres"),
		DSN:           getEnv("DSN", "host=/var/run/postgresql user=vtc password=vtc123456 dbname=vtc sslmode=disable"),
		ServerPort:    getEnv("SERVER_PORT", "8003"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "mima123123"),
		JWTSecret:     getEnv("JWT_SECRET", "change-me-in-production"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
