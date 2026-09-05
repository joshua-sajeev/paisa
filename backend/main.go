package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	router "github.com/joshu-sajeev/paisa/internal/adapter/http"
	"github.com/joshu-sajeev/paisa/internal/bootstrap"
	"github.com/joshu-sajeev/paisa/internal/config"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	bootstrapCtx, bootstrapCancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer bootstrapCancel()

	container, err := bootstrap.New(bootstrapCtx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize container: %v", err)
	}
	defer container.Close()

	r := router.NewRouter(&router.HandlerRegistry{
		AccountHandler: container.AccountHandler,
		AuthHandler:    container.AuthHandler,
		Config:         cfg,
		SessionStore:   container.SessionStore(),
	}, container.Logger())

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server started on %s", server.Addr)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(
		shutdownChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-shutdownChan
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}

	log.Println("Server stopped successfully")
}
