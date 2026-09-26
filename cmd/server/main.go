package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/soumajitgh/mobicode/internal/app"
	"github.com/soumajitgh/mobicode/internal/config"
	"github.com/soumajitgh/mobicode/internal/store"
	apputils "github.com/soumajitgh/mobicode/internal/utils"
)

func main() {
	os.Exit(start())
}

func start() int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		return 1
	}
	log, err := apputils.New(cfg.Environment)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer func() { _ = log.Sync() }()
	if err := run(cfg, log); err != nil {
		log.Error("server failed", zap.Error(err))
		return 1
	}
	return 0
}

func run(cfg *config.Config, log *zap.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	persistence, err := store.Open(ctx, store.Config{
		SQLitePath: cfg.Database.Path,
		LogLevel:   cfg.Database.LogLevel,
	})
	if err != nil {
		return fmt.Errorf("initialize store: %w", err)
	}
	defer func() {
		if err := persistence.Close(); err != nil {
			log.Error("store shutdown failed", zap.Error(err))
		}
	}()

	addr := fmt.Sprintf(":%d", cfg.Server.Port)

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           app.New(cfg, persistence, log),
		ReadHeaderTimeout: 5 * time.Second,
	}

	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		log.Info("server shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Error("server shutdown failed", zap.Error(err))
		}
	}()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen HTTP: %w", err)
	}
	log.Info("server started", zap.String("address", addr))
	if err := httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP: %w", err)
	}
	if ctx.Err() != nil {
		<-shutdownDone
	}
	log.Info("server stopped")
	return nil
}
