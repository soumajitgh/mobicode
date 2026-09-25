package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/soumajitgh/mobicode/internal/app"
	"github.com/soumajitgh/mobicode/internal/store"
	"gorm.io/gorm/logger"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("load .env: %w", err)
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
			log.Printf("store shutdown: %v", err)
		}
	}()

	port := os.Getenv("MOBICODE_SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           app.New(persistence),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown: %v", err)
		}
	}()

	log.Printf("listening on %s", addr)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP: %w", err)
	}
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
