package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"golang.org/x/time/rate"

	appmiddleware "github.com/soumajitgh/mobicode/internal/http/middleware"
)

func TestRateLimitUsesDirectClientAddress(t *testing.T) {
	limiter := appmiddleware.NewKeyedRateLimiter(rate.Every(time.Hour), 1, time.Hour)
	handler := appmiddleware.RateLimit(limiter, appmiddleware.ClientIP)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	request := func(remoteAddr string) int {
		t.Helper()
		r := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		r.RemoteAddr = remoteAddr
		r.Header.Set("X-Forwarded-For", "203.0.113.10")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w.Code
	}

	if got := request("192.0.2.1:1000"); got != http.StatusNoContent {
		t.Fatalf("first request status = %d", got)
	}
	if got := request("192.0.2.1:2000"); got != http.StatusTooManyRequests {
		t.Fatalf("repeated client status = %d", got)
	}
	if got := request("192.0.2.2:1000"); got != http.StatusNoContent {
		t.Fatalf("different client status = %d", got)
	}
}

func TestLoggingRecordsRequestWithoutSecrets(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	router := chi.NewRouter()
	appmiddleware.Apply(router, zap.New(core))
	router.Get("/resource", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	request := httptest.NewRequest(http.MethodGet, "/resource?token=secret", nil)
	request.Header.Set("Authorization", "Bearer secret")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if recorded.Len() != 1 {
		t.Fatalf("log entries = %d, want 1", recorded.Len())
	}
	entry := recorded.All()[0]
	fields := entry.ContextMap()
	if fields["method"] != http.MethodGet || fields["path"] != "/resource" || fields["status"] != int64(http.StatusCreated) {
		t.Fatalf("unexpected request fields: %v", fields)
	}
	if _, ok := fields["duration"]; !ok {
		t.Fatalf("missing duration: %v", fields)
	}
	if _, leaked := fields["token"]; leaked {
		t.Fatalf("query secret leaked: %v", fields)
	}
}
