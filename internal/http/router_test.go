package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appgraphql "github.com/soumajitgh/mobicode/internal/graphql"
	"github.com/soumajitgh/mobicode/internal/health"
	"github.com/soumajitgh/mobicode/internal/web/handlers"
)

func testRouter(enablePlayground bool) http.Handler {
	healthService := &health.Service{}
	return NewRouter(
		&appgraphql.Resolver{HealthService: healthService},
		&handlers.Handler{HealthService: healthService},
		enablePlayground,
		false,
	)
}

func TestHealthRoute(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	testRouter(false).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if recorder.Body.String() != "ok\n" {
		t.Fatalf("body = %q, want %q", recorder.Body.String(), "ok\n")
	}
}

func TestGraphQLHealth(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/mobile/graphql", bytes.NewBufferString(`{"query":"{ health { status } }"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	testRouter(false).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var response struct {
		Data struct {
			Health struct {
				Status string `json:"status"`
			} `json:"health"`
		} `json:"data"`
		Errors json.RawMessage `json:"errors"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.Health.Status != "ok" || len(response.Errors) > 0 {
		t.Fatalf("unexpected GraphQL response: %s", recorder.Body.String())
	}
}

func TestGraphQLTransports(t *testing.T) {
	for _, test := range []struct {
		method     string
		wantStatus int
	}{
		{method: http.MethodOptions, wantStatus: http.StatusOK},
		{method: http.MethodGet, wantStatus: http.StatusMethodNotAllowed},
	} {
		t.Run(test.method, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/mobile/graphql", nil)
			recorder := httptest.NewRecorder()
			testRouter(false).ServeHTTP(recorder, request)
			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, test.wantStatus, recorder.Body.String())
			}
		})
	}
}

func TestWebRoutes(t *testing.T) {
	for _, test := range []struct {
		path        string
		contentType string
		contains    string
	}{
		{path: "/", contentType: "text/html", contains: "hx-get=\"/partials/status\""},
		{path: "/partials/status", contentType: "text/html", contains: "Server status: ok"},
		{path: "/assets/css/app.css", contentType: "text/css", contains: ".bg-background"},
		{path: "/assets/js/htmx.min.js", contentType: "text/javascript", contains: "htmx"},
	} {
		t.Run(test.path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			recorder := httptest.NewRecorder()
			testRouter(false).ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
			}
			if !strings.HasPrefix(recorder.Header().Get("Content-Type"), test.contentType) {
				t.Fatalf("content type = %q, want prefix %q", recorder.Header().Get("Content-Type"), test.contentType)
			}
			if !strings.Contains(recorder.Body.String(), test.contains) {
				t.Fatalf("response for %s is missing %q", test.path, test.contains)
			}
		})
	}
}

func TestPlaygroundLocalOnly(t *testing.T) {
	for _, test := range []struct {
		name       string
		enabled    bool
		remoteAddr string
		host       string
		wantStatus int
	}{
		{name: "disabled", remoteAddr: "127.0.0.1:1234", wantStatus: http.StatusNotFound},
		{name: "remote", enabled: true, remoteAddr: "192.0.2.1:1234", wantStatus: http.StatusForbidden},
		{name: "public host", enabled: true, remoteAddr: "127.0.0.1:1234", host: "example.com", wantStatus: http.StatusForbidden},
		{name: "local", enabled: true, remoteAddr: "127.0.0.1:1234", wantStatus: http.StatusOK},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/mobile/graphql/playground", nil)
			request.Host = "localhost:8080"
			if test.host != "" {
				request.Host = test.host
			}
			request.RemoteAddr = test.remoteAddr
			recorder := httptest.NewRecorder()
			testRouter(test.enabled).ServeHTTP(recorder, request)
			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
		})
	}
}
