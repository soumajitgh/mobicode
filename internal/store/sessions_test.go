package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/gorm/logger"
)

func TestSessionPersistenceAcrossOpen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions.db")
	config := Config{SQLitePath: path, GORMLogLevel: logger.Silent}
	first, err := Open(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Sessions.Commit("token", []byte("session-data"), time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := Open(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := second.Close(); err != nil {
			t.Errorf("close store: %v", err)
		}
	})
	data, found, err := second.Sessions.Find("token")
	if err != nil || !found || string(data) != "session-data" {
		t.Fatalf("find after reopen: %q %t %v", data, found, err)
	}
	if err := second.Sessions.Commit("expired", []byte("old"), time.Now().Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	if _, found, err := second.Sessions.Find("expired"); err != nil || found {
		t.Fatalf("expired session: %t %v", found, err)
	}
}
