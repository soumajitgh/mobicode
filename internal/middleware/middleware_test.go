package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/soumajitgh/mobicode/internal/requestctx"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func TestRequestContextPropagatesChiRequestID(t *testing.T) {
	var requestID string
	handler := chimiddleware.RequestID(RequestContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID = requestctx.RequestID(r.Context())
	})))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set(requestIDHeader, "request-123")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if requestID != "request-123" || response.Header().Get(requestIDHeader) != "request-123" {
		t.Fatalf("context ID = %q, header ID = %q", requestID, response.Header().Get(requestIDHeader))
	}
}

func TestSecurityHeaders(t *testing.T) {
	handler := SecurityHeaders()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if response.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatal("response has no frame protection")
	}
}

func TestBodyLimitBoundsReads(t *testing.T) {
	const maxBytes = 4
	var readError error
	handler := BodyLimit(maxBytes)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, readError = io.ReadAll(r.Body)
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", strings.NewReader("12345")))

	var maxBytesError *http.MaxBytesError
	if !errors.As(readError, &maxBytesError) {
		t.Fatalf("read error = %v, want *http.MaxBytesError", readError)
	}
}

func TestCORSAllowsOnlyConfiguredOrigins(t *testing.T) {
	handler := CORS([]string{"https://app.example.com"})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))

	t.Run("allowed preflight", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodOptions, "/graphql", nil)
		request.Header.Set("Origin", "https://app.example.com")
		request.Header.Set("Access-Control-Request-Method", http.MethodPost)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
		}
		if got := response.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
			t.Fatalf("allowed origin = %q", got)
		}
	})

	t.Run("denied origin", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/graphql", nil)
		request.Header.Set("Origin", "https://attacker.example.com")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
		}
	})

	t.Run("same origin", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "http://api.example.com/graphql", nil)
		request.Header.Set("Origin", "http://api.example.com")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusAccepted {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusAccepted)
		}
	})
}
