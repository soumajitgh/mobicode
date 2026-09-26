package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexedwards/scs/v2"
	"gorm.io/gorm/logger"

	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/store"
	"github.com/soumajitgh/mobicode/internal/store/model"

	"go.uber.org/zap"

	appgraphql "github.com/soumajitgh/mobicode/internal/graphql"
	"github.com/soumajitgh/mobicode/internal/health"
	"github.com/soumajitgh/mobicode/internal/web/handlers"
)

func testRouter(t *testing.T, enablePlayground bool) http.Handler {
	t.Helper()
	persistence, err := store.Open(context.Background(), store.Config{SQLitePath: filepath.Join(t.TempDir(), "router.db"), GORMLogLevel: logger.Silent})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = persistence.Close() })
	_ = persistence.Users.Create(context.Background(), &model.User{Email: "test@example.com", PasswordHash: "hash123456789", SessionVersion: 1})
	sessions := scs.New()
	sessions.Store = persistence.Sessions
	browserAuth := handlers.NewAuth(&auth.Service{Users: persistence.Users, RecoveryToken: "a-long-recovery-token-of-at-least-32-bytes"}, sessions, persistence.Users)
	onboarding := handlers.NewOnboarding(browserAuth, persistence.Users, sessions)
	healthService := &health.Service{}
	return NewRouter(
		&appgraphql.Resolver{HealthService: healthService},
		&handlers.Handler{HealthService: healthService},
		enablePlayground,
		false,
		zap.NewNop(),
		browserAuth,
		onboarding,
	)
}

func TestHealthRoute(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	testRouter(t, false).ServeHTTP(recorder, request)

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

	testRouter(t, false).ServeHTTP(recorder, request)

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
			testRouter(t, false).ServeHTTP(recorder, request)
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
		{path: "/", contentType: "", contains: ""},
		{path: "/partials/status", contentType: "", contains: ""},
		{path: "/assets/css/app.css", contentType: "text/css", contains: ".bg-background"},
		{path: "/assets/js/htmx.min.js", contentType: "text/javascript", contains: "htmx"},
	} {
		t.Run(test.path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			recorder := httptest.NewRecorder()
			testRouter(t, false).ServeHTTP(recorder, request)
			want := http.StatusOK
			if test.contentType == "" {
				want = http.StatusSeeOther
			}
			if recorder.Code != want {
				t.Fatalf("status = %d, want %d", recorder.Code, want)
			}
			if want == http.StatusSeeOther {
				if !strings.HasPrefix(recorder.Header().Get("Location"), "/auth/login?next=") {
					t.Fatal("missing login redirect")
				}
				return
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

func TestAuthRoutesAreRateLimitedSeparately(t *testing.T) {
	router := testRouter(t, false)
	for _, path := range []string{"/auth/login", "/auth/register", "/auth/reset-password"} {
		for attempt := 1; attempt <= 11; attempt++ {
			request := httptest.NewRequest(http.MethodPost, path, nil)
			request.RemoteAddr = "192.0.2.1:1234"
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			want := http.StatusForbidden // Missing CSRF token, if the request reaches auth.
			if attempt == 11 {
				want = http.StatusTooManyRequests
			}
			if recorder.Code != want {
				t.Fatalf("%s attempt %d: status = %d, want %d", path, attempt, recorder.Code, want)
			}
		}
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
			testRouter(t, test.enabled).ServeHTTP(recorder, request)
			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
		})
	}
}
