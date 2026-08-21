package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
