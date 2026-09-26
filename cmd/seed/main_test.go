package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/config"
	"github.com/soumajitgh/mobicode/internal/store"
)

func TestSeedCreatesInitialAccountOnce(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "mobicode.db")
	cfg := &config.Config{Environment: "development", Database: config.DatabaseConfig{Path: path}}
	var output bytes.Buffer

	if err := seed(ctx, cfg, &output); err != nil {
		t.Fatal(err)
	}
	password, found := strings.CutPrefix(strings.Split(strings.TrimSpace(output.String()), "\n")[1], "Password (shown once): ")
	if !found || password == "" {
		t.Fatalf("seed did not print a generated password: %q", output.String())
	}

	persistence, err := store.Open(ctx, store.Config{SQLitePath: path})
	if err != nil {
		t.Fatal(err)
	}
	service := &auth.Service{Users: persistence.Users}
	if _, err := service.Login(ctx, developmentEmail, password); err != nil {
		t.Fatalf("seeded account cannot log in: %v", err)
	}
	if err := persistence.Close(); err != nil {
		t.Fatal(err)
	}

	output.Reset()
	if err := seed(ctx, cfg, &output); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "Password") || !strings.Contains(output.String(), "seed skipped") {
		t.Fatalf("repeat seed should not replace the account: %q", output.String())
	}
}

func TestSeedRejectsProductionWithoutOpeningDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mobicode.db")
	cfg := &config.Config{Environment: "production", Database: config.DatabaseConfig{Path: path}}
	if err := seed(context.Background(), cfg, &bytes.Buffer{}); err == nil {
		t.Fatal("production seed should fail")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("production seed touched database: %v", err)
	}
}
