package app

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
	"strings"
	"testing"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"

	"github.com/soumajitgh/mobicode/internal/store"
)

func TestBrowserAuth(t *testing.T) {
	t.Setenv("MOBICODE_SERVER_SECRET_TOKEN", "a-long-recovery-token-of-at-least-32-bytes")
	dbPath := filepath.Join(t.TempDir(), "auth.db")
	s, err := store.Open(context.Background(), store.Config{SQLitePath: dbPath, GORMLogLevel: logger.Silent})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Errorf("close store: %v", err)
		}
	})
	closeBody := func(res *http.Response) {
		t.Helper()
		if err := res.Body.Close(); err != nil {
			t.Errorf("close response body: %v", err)
		}
	}
	handler := New(s, zap.NewNop())
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Transport: localTransport{handler}, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	get := func(path string) *http.Response {
		t.Helper()
		res, err := client.Get("http://mobicode.test" + path)
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	post := func(path string, values url.Values) *http.Response {
		t.Helper()
		res, err := client.PostForm("http://mobicode.test"+path, values)
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	token := func(path string) string {
		t.Helper()
		res := get(path)
		defer closeBody(res)
		b, _ := io.ReadAll(res.Body)
		matches := regexp.MustCompile(`name="csrf_token" value="([^"]+)"`).FindSubmatch(b)
		if len(matches) != 2 {
			t.Fatalf("missing CSRF in %s: %s", path, b)
		}
		return string(matches[1])
	}
	if res := get("/"); res.StatusCode != 303 || !strings.Contains(res.Header.Get("Location"), "/auth/login?next=") {
		t.Fatalf("home: %d %s", res.StatusCode, res.Header.Get("Location"))
	}
	for _, path := range []string{"/healthz", "/assets/css/app.css"} {
		res := get(path)
		if res.StatusCode != 200 {
			t.Fatalf("%s: %d", path, res.StatusCode)
		}
		closeBody(res)
	}
	res := post("/auth/register", url.Values{"email": {"a@example.com"}, "password": {"123456789012"}, "recovery_token": {"a-long-recovery-token-of-at-least-32-bytes"}})
	if res.StatusCode != 403 {
		t.Fatalf("CSRF: %d", res.StatusCode)
	}
	closeBody(res)
	graphqlReq, _ := http.NewRequest(http.MethodPost, "http://mobicode.test/mobile/graphql", bytes.NewBufferString(`{"query":"{ health { status } }"}`))
	graphqlReq.Header.Set("Content-Type", "application/json")
	graphqlRes, err := client.Do(graphqlReq)
	if err != nil || graphqlRes.StatusCode != 200 {
		t.Fatalf("public GraphQL: %v %+v", err, graphqlRes)
	}
	closeBody(graphqlRes)
	csrf := token("/auth/register?next=%2Fpartials%2Fstatus")
	res = post("/auth/register", url.Values{"email": {" A@Example.com "}, "password": {"123456789012"}, "recovery_token": {"a-long-recovery-token-of-at-least-32-bytes"}, "csrf_token": {csrf}, "next": {"//evil.example"}})
	if res.StatusCode != 303 || res.Header.Get("Location") != "/" {
		t.Fatalf("register redirect: %d %s", res.StatusCode, res.Header.Get("Location"))
	}
	closeBody(res)
	registerCookie := jar.Cookies(&url.URL{Scheme: "http", Host: "mobicode.test"})[0].Value
	if res = get("/"); res.StatusCode != 200 {
		t.Fatalf("authenticated home: %d", res.StatusCode)
	}
	closeBody(res)
	csrf = token("/auth/reset-password")
	res = post("/auth/reset-password", url.Values{"email": {"a@example.com"}, "password": {"new-password-123"}, "recovery_token": {"a-long-recovery-token-of-at-least-32-bytes"}, "csrf_token": {csrf}})
	if res.StatusCode != 303 {
		t.Fatalf("reset: %d", res.StatusCode)
	}
	closeBody(res)
	if res = get("/"); res.StatusCode != 303 {
		t.Fatalf("old session after reset: %d", res.StatusCode)
	}
	closeBody(res)
	csrf = token("/auth/login")
	res = post("/auth/login", url.Values{"email": {"a@example.com"}, "password": {"123456789012"}, "csrf_token": {csrf}})
	if res.StatusCode != 422 {
		t.Fatalf("old password: %d", res.StatusCode)
	}
	closeBody(res)
	res = post("/auth/login", url.Values{"email": {"a@example.com"}, "password": {"new-password-123"}, "csrf_token": {csrf}})
	if res.StatusCode != 303 {
		t.Fatalf("new password: %d", res.StatusCode)
	}
	closeBody(res)
	loginCookie := jar.Cookies(&url.URL{Scheme: "http", Host: "mobicode.test"})[0].Value
	if loginCookie == registerCookie {
		t.Fatal("session cookie did not rotate")
	}
	res = post("/auth/logout", url.Values{"csrf_token": {csrf}})
	if res.StatusCode != 303 {
		t.Fatalf("logout: %d", res.StatusCode)
	}
	closeBody(res)
	if res = get("/"); res.StatusCode != 303 {
		t.Fatalf("after logout: %d", res.StatusCode)
	}
	closeBody(res)
}

type localTransport struct{ handler http.Handler }

func (t localTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	recorder := httptest.NewRecorder()
	t.handler.ServeHTTP(recorder, req)
	return recorder.Result(), nil
}
