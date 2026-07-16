package config

import (
	"fmt"
	"os"
)

// Config holds every piece of runtime configuration the service needs.
type Config struct {
	DBUser string
	DBPass string
	DBHost string
	DBPort string
	DBName string

	ServerPort string
}

// Load reads configuration from environment variables, falling back to
// sane local-development defaults when a variable isn't set.
func Load() Config {
	return Config{
		DBUser:     getEnv("DB_USER", "root"),
		DBPass:     getEnv("DB_PASS", ""),
		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBName:     getEnv("DB_NAME", "article"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}
}

// DSN builds a go-sql-driver/mysql compatible data source name.
func (c Config) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
