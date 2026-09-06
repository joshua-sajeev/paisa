package config_test

import (
	"os"
	"testing"

	"github.com/joshu-sajeev/paisa/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setRequiredEnv(t *testing.T) {
	t.Helper()

	t.Setenv("DB_PASSWORD", "password")
	t.Setenv("APP_PIN_HASH", "hash")
}

func TestLoad_RequiredEnvVars(t *testing.T) {
	tests := []struct {
		name        string
		dbPassword  string
		pinHash     string
		wantErrText string
	}{
		{
			name:        "missing database password",
			pinHash:     "some-hash",
			wantErrText: "DB_PASSWORD",
		},
		{
			name:        "missing PIN hash",
			dbPassword:  "password",
			wantErrText: "APP_PIN_HASH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dbPassword == "" {
				t.Setenv("DB_PASSWORD", "")
			} else {
				t.Setenv("DB_PASSWORD", tt.dbPassword)
			}

			if tt.pinHash == "" {
				t.Setenv("APP_PIN_HASH", "")
			} else {
				t.Setenv("APP_PIN_HASH", tt.pinHash)
			}

			cfg, err := config.Load()

			require.Error(t, err)
			assert.Nil(t, cfg)
			assert.Contains(t, err.Error(), tt.wantErrText)
		})
	}
}

func TestLoad_Server(t *testing.T) {
	tests := []struct {
		name string
		host string
		port string
	}{
		{
			name: "defaults",
			host: "localhost",
			port: "8080",
		},
		{
			name: "custom values",
			host: "0.0.0.0",
			port: "3000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setRequiredEnv(t)

			if tt.name == "defaults" {
				t.Setenv("SERVER_HOST", "")
				t.Setenv("SERVER_PORT", "")
			} else {
				t.Setenv("SERVER_HOST", tt.host)
				t.Setenv("SERVER_PORT", tt.port)
			}

			cfg, err := config.Load()

			require.NoError(t, err)
			require.NotNil(t, cfg)

			assert.Equal(t, tt.host, cfg.Server.Host)
			assert.Equal(t, tt.port, cfg.Server.Port)
		})
	}
}

func TestLoad_Database(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     string
		user     string
		database string
		sslMode  string
	}{
		{
			name:     "defaults",
			host:     "localhost",
			port:     "5432",
			user:     "paisa",
			database: "paisa",
			sslMode:  "disable",
		},
		{
			name:     "custom values",
			host:     "db.example.com",
			port:     "5433",
			user:     "admin",
			database: "finance",
			sslMode:  "require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setRequiredEnv(t)

			if tt.name == "defaults" {
				t.Setenv("DB_HOST", "")
				t.Setenv("DB_PORT", "")
				t.Setenv("DB_USER", "")
				t.Setenv("DB_NAME", "")
				t.Setenv("DB_SSLMODE", "")
			} else {
				t.Setenv("DB_HOST", tt.host)
				t.Setenv("DB_PORT", tt.port)
				t.Setenv("DB_USER", tt.user)
				t.Setenv("DB_NAME", tt.database)
				t.Setenv("DB_SSLMODE", tt.sslMode)
			}

			cfg, err := config.Load()

			require.NoError(t, err)
			require.NotNil(t, cfg)

			assert.Equal(t, tt.host, cfg.Database.Host)
			assert.Equal(t, tt.port, cfg.Database.Port)
			assert.Equal(t, tt.user, cfg.Database.User)
			assert.Equal(t, tt.database, cfg.Database.Database)
			assert.Equal(t, tt.sslMode, cfg.Database.SSLMode)
		})
	}
}

func TestLoad_SessionTTL(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantValue int
	}{
		{
			name:      "default",
			value:     "",
			wantValue: 10,
		},
		{
			name:      "custom value",
			value:     "30",
			wantValue: 30,
		},
		{
			name:      "invalid value",
			value:     "not-a-number",
			wantValue: 10,
		},
		{
			name:      "zero value",
			value:     "0",
			wantValue: 10,
		},
		{
			name:      "negative value",
			value:     "-5",
			wantValue: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setRequiredEnv(t)
			t.Setenv("SESSION_TTL_MINUTES", tt.value)

			cfg, err := config.Load()

			require.NoError(t, err)
			require.NotNil(t, cfg)

			assert.Equal(t, tt.wantValue, cfg.SessionTTLMinutes)
		})
	}
}

func TestLoad_DemoMode(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantDemo  bool
		wantError bool
	}{
		{
			name:     "defaults to false",
			value:    "",
			wantDemo: false,
		},
		{
			name:     "enabled",
			value:    "true",
			wantDemo: true,
		},
		{
			name:     "disabled",
			value:    "false",
			wantDemo: false,
		},
		{
			name:      "invalid value",
			value:     "invalid",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setRequiredEnv(t)
			t.Setenv("DEMO_MODE", tt.value)

			cfg, err := config.Load()

			if tt.wantError {
				require.Error(t, err)
				assert.Nil(t, cfg)
				assert.Contains(
					t,
					err.Error(),
					"DEMO_MODE must be true or false",
				)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)

			assert.Equal(t, tt.wantDemo, cfg.DemoMode)
		})
	}
}

func TestLoad_WithRequiredVars(t *testing.T) {
	setRequiredEnv(t)

	cfg, err := config.Load()

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "password", cfg.Database.Password)
	assert.Equal(t, "hash", cfg.AppLock.PINHash)
}

func TestConnectionURL(t *testing.T) {
	tests := []struct {
		name     string
		database config.DatabaseConfig
		want     string
	}{
		{
			name: "standard connection",
			database: config.DatabaseConfig{
				User:     "testuser",
				Password: "testpass",
				Host:     "localhost",
				Port:     "5432",
				Database: "testdb",
				SSLMode:  "disable",
			},
			want: "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable",
		},
		{
			name: "SSL connection",
			database: config.DatabaseConfig{
				User:     "user",
				Password: "pass",
				Host:     "db.example.com",
				Port:     "5433",
				Database: "proddb",
				SSLMode:  "require",
			},
			want: "postgres://user:pass@db.example.com:5433/proddb?sslmode=require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.database.ConnectionURL())
		})
	}
}

func TestConnectionURL_WithSpecialCharacters(t *testing.T) {
	db := &config.DatabaseConfig{
		User:     "user@domain",
		Password: "p@ss:word",
		Host:     "db.local",
		Port:     "5432",
		Database: "mydb",
		SSLMode:  "disable",
	}

	url := db.ConnectionURL()

	assert.Contains(t, url, "postgres://")
	assert.Contains(t, url, "p@ss:word")
	assert.Contains(t, url, "?sslmode=disable")
}

func TestLoad_EnvironmentIsolation(t *testing.T) {
	// Make sure tests don't accidentally depend on the developer's
	// environment for these values.
	for _, key := range []string{
		"DB_PASSWORD",
		"APP_PIN_HASH",
		"SERVER_HOST",
		"SERVER_PORT",
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_NAME",
		"DB_SSLMODE",
		"SESSION_TTL_MINUTES",
		"DEMO_MODE",
	} {
		t.Setenv(key, "")
	}

	setRequiredEnv(t)

	cfg, err := config.Load()

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "5432", cfg.Database.Port)
	assert.Equal(t, "paisa", cfg.Database.User)
	assert.Equal(t, "paisa", cfg.Database.Database)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, 10, cfg.SessionTTLMinutes)
	assert.False(t, cfg.DemoMode)
}

func TestEnvironmentVariableExists(t *testing.T) {
	_, ok := os.LookupEnv("PATH")
	assert.True(t, ok)
}
