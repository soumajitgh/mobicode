package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/soumajitgh/mobicode/internal/config"
)

func TestNewRouterMiddleware(t *testing.T) {
	r := NewRouter()
	if r == nil {
		t.Fatalf("expected non-nil router")
	}

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("healthz status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestLogDevelopmentUserID(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logDevelopmentUserID(zap.New(core), &config.Config{Env: "development", DevUserID: "dev-user-123"})

	entries := logs.FilterMessage("development user ID configured for bearer authentication").All()
	if len(entries) != 1 {
		t.Fatalf("development user ID log entries = %d, want 1", len(entries))
	}
	if got := entries[0].ContextMap()["dev_user_id"]; got != "dev-user-123" {
		t.Fatalf("logged dev_user_id = %v, want %q", got, "dev-user-123")
	}
}

func TestLogDevelopmentUserIDSkipsNonDevelopment(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logDevelopmentUserID(zap.New(core), &config.Config{Env: "production", DevUserID: "dev-user-123"})

	if got := logs.FilterMessage("development user ID configured for bearer authentication").Len(); got != 0 {
		t.Fatalf("development user ID log entries = %d, want 0", got)
	}
}
