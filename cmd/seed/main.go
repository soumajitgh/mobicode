package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/config"
	"github.com/soumajitgh/mobicode/internal/store"
)

const developmentEmail = "dev@mobicode.test"

func main() {
	if err := run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "seed server: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, output io.Writer) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	return seed(ctx, cfg, output)
}

func seed(ctx context.Context, cfg *config.Config, output io.Writer) error {
	if cfg.Environment != "development" {
		return fmt.Errorf("seeding is available only in development")
	}

	persistence, err := store.Open(ctx, store.Config{
		SQLitePath: cfg.Database.Path,
		LogLevel:   cfg.Database.LogLevel,
	})
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer persistence.Close()

	count, err := persistence.Users.Count(ctx)
	if err != nil {
		return fmt.Errorf("count users: %w", err)
	}
	if count > 0 {
		fmt.Fprintln(output, "Server already has an account; seed skipped.")
		return nil
	}

	passwordBytes := make([]byte, 24)
	if _, err := rand.Read(passwordBytes); err != nil {
		return fmt.Errorf("generate password: %w", err)
	}
	password := base64.RawURLEncoding.EncodeToString(passwordBytes)
	if _, err := (&auth.Service{Users: persistence.Users}).CreateInitialUser(ctx, developmentEmail, password); err != nil {
		return fmt.Errorf("create development account: %w", err)
	}

	fmt.Fprintf(output, "Created development account: %s\nPassword (shown once): %s\n", developmentEmail, password)
	return nil
}
