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
	// Session store (redis url, e.g. redis://localhost:6379)
	SessionStoreURL string
	// Session TTL hours
	SessionTTLHours int
}

type AppLockConfig struct {
	PINHash  string // Argon2id PHC encoded hash of the master PIN
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
		SessionStoreURL: getEnv("SESSION_STORE_URL", "redis://localhost:6379"),
		SessionTTLHours: getEnvInt("SESSION_TTL_HOURS", 24),
	}

	if cfg.Database.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD environment variable is required")
	}
	cfg.AppLock.PINHash = os.Getenv("APP_PIN_HASH")
	cfg.AppLock.DemoMode = getEnvBool("DEMO_MODE", false)

	// Validate PIN hash is set (if not in DEMO_MODE)
	if !cfg.AppLock.DemoMode && cfg.AppLock.PINHash == "" {
		return nil, errors.New("APP_PIN_HASH must be set in environment (or enable DEMO_MODE=true)")
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

func getEnvInt(key string, defaultValue int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	var v int
	_, _ = fmt.Sscanf(val, "%d", &v)
	if v == 0 {
		return defaultValue
	}
	return v
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
