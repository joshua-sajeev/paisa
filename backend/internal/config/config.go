// Package config provides typed application configuration loaded from
// environment variables.
package config

import (
	"errors"
	"fmt"
	"os"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig

	AppLock  AppLockConfig
	PINHash  string // Argon2 hash of the master PIN
	DemoMode bool   // Bypass authentication if true
}
type AppLockConfig struct {
	PINHash  string // Argon2 hash of the master PIN
	DemoMode bool   // Bypass authentication if true
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host string
	Port string
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "paisa"),
			Password: os.Getenv("DB_PASSWORD"),
			Database: getEnv("DB_NAME", "paisa"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}

	if cfg.Database.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD environment variable is required")
	}
	cfg.AppLock.PINHash = os.Getenv("APP_PIN_HASH")
	cfg.AppLock.DemoMode = getEnvBool("DEMO_MODE", false)

	// Validate PIN hash is set (if not in DEMO_MODE)
	if !cfg.AppLock.DemoMode && cfg.AppLock.PINHash == "" {
		return nil, errors.New("APP_PIN_HASH must be set in .env (or enable DEMO_MODE=true)")
	}
	return cfg, nil
}

// getEnv reads an environment variable with a default fallback
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetDatabaseURL returns PostgreSQL connection string
func (c *DatabaseConfig) GetDatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.SSLMode,
	)
}

func getEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val == "true" || val == "1" || val == "yes"
}
