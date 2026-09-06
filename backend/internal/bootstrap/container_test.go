package bootstrap_test

import (
	"context"
	"errors"
	"testing"

	"github.com/joshu-sajeev/paisa/internal/bootstrap"
	"github.com/joshu-sajeev/paisa/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: "8080",
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "paisa",
			Password: "password",
			Database: "paisa_test",
			SSLMode:  "disable",
		},
		AppLock: config.AppLockConfig{
			PINHash: "test-hash",
		},
		SessionTTLMinutes: 10,
		DemoMode:          false,
	}
}

func TestNew_NilConfig(t *testing.T) {
	container, err := bootstrap.New(
		context.Background(),
		nil,
	)

	require.Error(t, err)
	assert.Nil(t, container)
	assert.EqualError(t, err, "config is nil")
}

func TestNew_ContextCancelled(t *testing.T) {
	cfg := setupTestConfig()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	container, err := bootstrap.New(ctx, cfg)

	require.Error(t, err)
	assert.Nil(t, container)
}

func TestNew_InvalidDatabase(t *testing.T) {
	cfg := setupTestConfig()

	cfg.Database.Host = "invalid-host-that-does-not-exist-12345.local"

	container, err := bootstrap.New(
		context.Background(),
		cfg,
	)

	require.Error(t, err)
	assert.Nil(t, container)
	assert.Contains(
		t,
		err.Error(),
		"failed to initialize database",
	)
}

func TestNew_WithDatabase(t *testing.T) {
	cfg := setupTestConfig()

	container, err := bootstrap.New(
		context.Background(),
		cfg,
	)

	if errors.Is(err, context.Canceled) {
		t.Skip("database unavailable")
	}

	if err != nil {
		t.Skipf("database unavailable: %v", err)
	}

	require.NotNil(t, container)
	defer container.Close()

	assert.NotNil(t, container.AccountHandler)
	assert.NotNil(t, container.AuthHandler)
	assert.NotNil(t, container.SessionStore)
	assert.NotNil(t, container.Logger())
}

func TestContainer_Logger(t *testing.T) {
	cfg := setupTestConfig()

	container, err := bootstrap.New(
		context.Background(),
		cfg,
	)
	if err != nil {
		t.Skipf("database unavailable: %v", err)
	}

	defer container.Close()

	logger := container.Logger()

	require.NotNil(t, logger)

	assert.NotPanics(t, func() {
		logger.Info("test message")
		logger.Warn("test warning")
		logger.Error("test error")
	})
}

func TestContainer_Close(t *testing.T) {
	cfg := setupTestConfig()

	container, err := bootstrap.New(
		context.Background(),
		cfg,
	)
	if err != nil {
		t.Skipf("database unavailable: %v", err)
	}

	require.NotNil(t, container)

	assert.NotPanics(t, func() {
		container.Close()
	})

	// Close should be safe to call more than once.
	assert.NotPanics(t, func() {
		container.Close()
	})
}

func TestContainer_SessionStore(t *testing.T) {
	cfg := setupTestConfig()

	container, err := bootstrap.New(
		context.Background(),
		cfg,
	)
	if err != nil {
		t.Skipf("database unavailable: %v", err)
	}

	defer container.Close()

	require.NotNil(t, container.SessionStore)

	ctx := context.Background()

	_, err = container.SessionStore.Get(
		ctx,
		"non-existent-session",
	)

	assert.Error(t, err)
}

func TestNew_MultipleContainers(t *testing.T) {
	cfg := setupTestConfig()

	ctx := context.Background()

	container1, err := bootstrap.New(ctx, cfg)
	if err != nil {
		t.Skipf("database unavailable: %v", err)
	}
	defer container1.Close()

	container2, err := bootstrap.New(ctx, cfg)
	if err != nil {
		t.Skipf("database unavailable: %v", err)
	}
	defer container2.Close()

	require.NotNil(t, container1)
	require.NotNil(t, container2)

	assert.NotSame(t, container1, container2)

	assert.NotSame(
		t,
		container1.SessionStore,
		container2.SessionStore,
	)

	assert.NotSame(
		t,
		container1.AccountHandler,
		container2.AccountHandler,
	)

	assert.NotSame(
		t,
		container1.AuthHandler,
		container2.AuthHandler,
	)
}
