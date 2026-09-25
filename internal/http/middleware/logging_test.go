package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLoggingRecordsRequestCompletion(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	router := chi.NewRouter()
	Apply(router, zap.New(core))
	router.Get("/resource", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	request := httptest.NewRequest(http.MethodGet, "/resource?token=secret", nil)
	request.Header.Set("X-Request-Id", "test-request")
	request.Header.Set("Authorization", "Bearer secret")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	if recorded.Len() != 1 {
		t.Fatalf("log entries = %d, want 1", recorded.Len())
	}
	entry := recorded.All()[0]
	if entry.Message != "request completed" {
		t.Fatalf("message = %q", entry.Message)
	}
	fields := entry.ContextMap()
	if fields["method"] != http.MethodGet || fields["path"] != "/resource" || fields["status"] != int64(http.StatusCreated) || fields["request_id"] != "test-request" {
		t.Fatalf("unexpected request fields: %v", fields)
	}
	if _, ok := fields["duration"]; !ok {
		t.Fatalf("missing duration: %v", fields)
	}
	if len(fields) != 5 {
		t.Fatalf("unexpected fields: %v", fields)
	}
}
