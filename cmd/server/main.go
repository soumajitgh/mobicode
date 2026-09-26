package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"

	"github.com/soumajitgh/mobicode/internal/app"
	"github.com/soumajitgh/mobicode/internal/store"
	apputils "github.com/soumajitgh/mobicode/internal/utils"
)

func main() {
	os.Exit(start())
}

func start() int {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(os.Stderr, "load .env: %v\n", err)
		return 1
	}
	log, err := apputils.New(os.Getenv("MOBICODE_SERVER_ENV"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer func() { _ = log.Sync() }()
	if err := run(log); err != nil {
		log.Error("server failed", zap.Error(err))
		return 1
	}
	return 0
}

func run(log *zap.Logger) error {
	secret := os.Getenv("MOBICODE_SERVER_SECRET_TOKEN")
	if len(secret) < 32 || strings.Contains(secret, "replace-with-") || strings.Trim(secret, string(secret[0])) == "" {
		return fmt.Errorf("MOBICODE_SERVER_SECRET_TOKEN must be at least 32 bytes")
	}
	logLevel, err := parseGORMLogLevel(os.Getenv("MOBICODE_SERVER_DB_LOG_LEVEL"))
	if err != nil {
		return err
	}
	path := os.Getenv("MOBICODE_SERVER_DB_PATH")
	if path == "" {
		path = "mobicode.db"
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	persistence, err := store.Open(ctx, store.Config{SQLitePath: path, GORMLogLevel: logLevel})
	if err != nil {
		return fmt.Errorf("initialize store: %w", err)
	}
	defer func() {
		if err := persistence.Close(); err != nil {
			log.Error("store shutdown failed", zap.Error(err))
		}
	}()

	port := os.Getenv("MOBICODE_SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           app.New(persistence, log),
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

func parseGORMLogLevel(value string) (logger.LogLevel, error) {
	switch strings.ToLower(value) {
	case "", "warn":
		return logger.Warn, nil
	case "silent":
		return logger.Silent, nil
	case "error":
		return logger.Error, nil
	case "info":
		return logger.Info, nil
	default:
		return 0, fmt.Errorf("invalid MOBICODE_SERVER_DB_LOG_LEVEL %q", value)
	}
}
