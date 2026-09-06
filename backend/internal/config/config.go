// Package config provides typed application configuration loaded from
// environment variables.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

const defaultSessionTTLMinutes = 10

// Config holds all application configuration.
type Config struct {
	Server            ServerConfig
	Database          DatabaseConfig
	AppLock           AppLockConfig
	SessionTTLMinutes int
	DemoMode          bool
}

// AppLockConfig holds application authentication configuration.
type AppLockConfig struct {
	PINHash string
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host string
	Port string
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	demoMode, err := strconv.ParseBool(getEnv("DEMO_MODE", "false"))
	if err != nil {
		return nil, errors.New("DEMO_MODE must be true or false")
	}
	cfg := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		DemoMode: demoMode,
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "paisa"),
			Password: os.Getenv("DB_PASSWORD"),
			Database: getEnv("DB_NAME", "paisa"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},

		AppLock: AppLockConfig{
			PINHash: os.Getenv("APP_PIN_HASH"),
		},

		SessionTTLMinutes: getEnvInt(
			"SESSION_TTL_MINUTES",
			defaultSessionTTLMinutes,
		),
	}

	if cfg.Database.Password == "" {
		return nil, errors.New("DB_PASSWORD environment variable is required")
	}

	if cfg.AppLock.PINHash == "" {
		return nil, errors.New("APP_PIN_HASH environment variable is required")
	}

	return cfg, nil
}

// getEnv reads an environment variable with a default fallback.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	v, err := strconv.Atoi(value)
	if err != nil || v <= 0 {
		return defaultValue
	}

	return v
}

// ConnectionURL returns the PostgreSQL connection URL.
// The returned URL contains the database password and must never be logged.
func (d *DatabaseConfig) ConnectionURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Database,
		d.SSLMode,
	)
}
