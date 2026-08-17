package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/soumajitgh/mobicode/internal/config"
)

func TestNewConfiguresHTTPServerBoundary(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) })
	cfg := config.Config{Server: config.ServerConfig{
		Host:              "127.0.0.1",
		Port:              9090,
		ReadHeaderTimeout: time.Second,
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      3 * time.Second,
		IdleTimeout:       4 * time.Second,
		MaxHeaderBytes:    2048,
	}}
	server := New(cfg, handler)

	if server.Addr != "127.0.0.1:9090" {
		t.Fatalf("server address = %q, want 127.0.0.1:9090", server.Addr)
	}
	response := httptest.NewRecorder()
	server.Handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusTeapot {
		t.Fatalf("handler status = %d, want %d", response.Code, http.StatusTeapot)
	}
	if server.ReadHeaderTimeout != time.Second || server.ReadTimeout != 2*time.Second || server.WriteTimeout != 3*time.Second || server.IdleTimeout != 4*time.Second {
		t.Fatalf("server timeouts = %#v", server)
	}
	if server.MaxHeaderBytes != 2048 {
		t.Fatalf("MaxHeaderBytes = %d, want 2048", server.MaxHeaderBytes)
	}
}
