package integration_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"regexp"
	"testing"

	"go.uber.org/zap"

	"github.com/soumajitgh/mobicode/internal/app"
	"github.com/soumajitgh/mobicode/internal/config"
	"github.com/soumajitgh/mobicode/internal/store"
)

const (
	testOrigin        = "http://mobicode.test"
	testRecoveryToken = "a-long-recovery-token-of-at-least-32-bytes"
)

var csrfPattern = regexp.MustCompile(`name="csrf_token" value="([^"]+)"`)

type testApplication struct {
	t      *testing.T
	store  *store.Store
	client *http.Client
}

func newTestApplication(t *testing.T) *testApplication {
	t.Helper()
	cfg := &config.Config{
		Environment: "development",
		Server: config.ServerConfig{
			Port: 8080,
		},
		Database: config.DatabaseConfig{
			Path:     filepath.Join(t.TempDir(), "application.db"),
			LogLevel: config.DatabaseLogLevelSilent,
		},
		Settings: config.SettingsConfig{
			SecretToken: testRecoveryToken,
		},
	}

	persistence, err := store.Open(context.Background(), store.Config{
		SQLitePath: cfg.Database.Path,
		LogLevel:   cfg.Database.LogLevel,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := persistence.Close(); err != nil {
			t.Errorf("close store: %v", err)
		}
	})

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &testApplication{
		t:     t,
		store: persistence,
		client: &http.Client{
			Jar:           jar,
			Transport:     handlerTransport{handler: app.New(cfg, persistence, zap.NewNop())},
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
}

func (a *testApplication) get(path string) *http.Response {
	a.t.Helper()
	response, err := a.client.Get(testOrigin + path)
	if err != nil {
		a.t.Fatal(err)
	}
	return response
}

func (a *testApplication) postForm(path string, values url.Values) *http.Response {
	a.t.Helper()
	response, err := a.client.PostForm(testOrigin+path, values)
	if err != nil {
		a.t.Fatal(err)
	}
	return response
}

func (a *testApplication) postGraphQL(query string) *http.Response {
	a.t.Helper()
	request, err := http.NewRequest(http.MethodPost, testOrigin+"/mobile/graphql", bytes.NewBufferString(query))
	if err != nil {
		a.t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := a.client.Do(request)
	if err != nil {
		a.t.Fatal(err)
	}
	return response
}

func (a *testApplication) csrf(path string) string {
	a.t.Helper()
	response := a.get(path)
	body := readBody(a.t, response)
	if response.StatusCode != http.StatusOK {
		a.t.Fatalf("GET %s: status = %d, body = %s", path, response.StatusCode, body)
	}
	match := csrfPattern.FindStringSubmatch(body)
	if len(match) != 2 {
		a.t.Fatalf("GET %s: CSRF field not found", path)
	}
	return match[1]
}

func readBody(t *testing.T, response *http.Response) string {
	t.Helper()
	defer func() {
		if err := response.Body.Close(); err != nil {
			t.Errorf("close response body: %v", err)
		}
	}()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func closeBody(t *testing.T, response *http.Response) {
	t.Helper()
	if err := response.Body.Close(); err != nil {
		t.Errorf("close response body: %v", err)
	}
}

type handlerTransport struct {
	handler http.Handler
}

func (transport handlerTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	recorder := httptest.NewRecorder()
	transport.handler.ServeHTTP(recorder, request)
	return recorder.Result(), nil
}
