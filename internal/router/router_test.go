package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/soumajitgh/mobicode/internal/config"
	"github.com/soumajitgh/mobicode/internal/requestctx"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func TestRouterOwnsHTTPContract(t *testing.T) {
	var graphqlRequestID string
	var graphqlClientIP string
	graphqlHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		graphqlRequestID = requestctx.RequestID(r.Context())
		graphqlClientIP = chimiddleware.GetClientIP(r.Context())
		w.WriteHeader(http.StatusAccepted)
	})
	router := New(
		config.Config{Environment: "development", Server: config.ServerConfig{MaxBodyBytes: 1024}},
		nil,
		graphqlHandler,
	)

	t.Run("posts GraphQL through request context", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/graphql", nil)
		request.RemoteAddr = "192.0.2.1:4321"
		request.Header.Set("X-Request-ID", "request-123")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		if response.Code != http.StatusAccepted {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusAccepted)
		}
		if graphqlRequestID != "request-123" {
			t.Fatalf("request ID = %q, want request-123", graphqlRequestID)
		}
		if graphqlClientIP != "192.0.2.1" {
			t.Fatalf("client IP = %q, want 192.0.2.1", graphqlClientIP)
		}
		if got := response.Header().Get("X-Request-ID"); got != "request-123" {
			t.Fatalf("response request ID = %q, want request-123", got)
		}
		if got := response.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Fatalf("X-Content-Type-Options = %q, want nosniff", got)
		}
	})

	t.Run("keeps GraphQL endpoint POST-only", func(t *testing.T) {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/graphql", nil))
		if response.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("serves liveness outside GraphQL", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/health/live", nil)
		request.Header.Set("X-Request-ID", "health-123")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
		}
		var body healthResponse
		if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if body.Data == nil || body.Data.Status != "ok" || body.RequestID != "health-123" {
			t.Fatalf("response = %#v", body)
		}
	})

	t.Run("exposes playground only on its development route", func(t *testing.T) {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/playground", nil))
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
		}
	})
}

func TestRouterHidesPlaygroundOutsideDevelopment(t *testing.T) {
	router := New(
		config.Config{Environment: "production", Server: config.ServerConfig{MaxBodyBytes: 1024}},
		nil,
		http.NotFoundHandler(),
	)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/playground", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestRouterRecoversGraphQLPanic(t *testing.T) {
	router := New(
		config.Config{Environment: "production", Server: config.ServerConfig{MaxBodyBytes: 1024}},
		nil,
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") }),
	)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/graphql", nil))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
}
