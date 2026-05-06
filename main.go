package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gentle-ai/gentle-ai/internal/config"
	"github.com/gentle-ai/gentle-ai/internal/server"
)

// Version is set at build time via ldflags
var Version = "dev"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration from environment and config files
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	log.Printf("starting gentle-ai %s", Version)
	log.Printf("listening on %s", cfg.ServerAddr)

	// Initialize and start the HTTP server
	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	// Run server in a goroutine so we can handle shutdown signals
	errCh := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errCh <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Wait for interrupt signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Printf("received signal %s, shutting down gracefully...", sig)
	case err := <-errCh:
		log.Printf("server encountered an error: %v", err)
	}

	// Graceful shutdown with timeout
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("error during shutdown: %v", err)
		os.Exit(1)
	}

	log.Println("server stopped")
}
