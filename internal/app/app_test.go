package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"go.uber.org/zap"

	"github.com/soumajitgh/mobicode/internal/app"
	"github.com/soumajitgh/mobicode/internal/config"
	"github.com/soumajitgh/mobicode/internal/store"
	"github.com/soumajitgh/mobicode/internal/store/model"
)

func TestDevAutoPair(t *testing.T) {
	persistence, err := store.Open(context.Background(), store.Config{SQLitePath: filepath.Join(t.TempDir(), "app.db")})
	if err != nil {
		t.Fatal(err)
	}
	defer persistence.Close()

	base := &config.Config{Environment: "development", Development: config.DevelopmentConfig{MobileAutoPair: true}}
	request := func(environment string, enabled bool) *httptest.ResponseRecorder {
		t.Helper()
		cfg := *base
		cfg.Environment = environment
		cfg.Development.MobileAutoPair = enabled
		r := httptest.NewRequest(http.MethodPost, "/mobile/dev/auto-pair", bytes.NewBufferString(`{"platform":"IOS"}`))
		w := httptest.NewRecorder()
		app.New(&cfg, persistence, zap.NewNop()).ServeHTTP(w, r)
		return w
	}

	// Onboarding blocks mobile routes before any account exists.
	if got := request("development", true).Code; got != http.StatusServiceUnavailable {
		t.Fatalf("no account: got %d", got)
	}
	for _, email := range []string{"first@example.com", "second@example.com"} {
		if err := persistence.Users.Create(context.Background(), &model.User{Email: email, PasswordHash: "test", SessionVersion: 1}); err != nil {
			t.Fatal(err)
		}
	}
	if got := request("development", false).Code; got != http.StatusNotFound {
		t.Fatalf("flag off: got %d", got)
	}
	if got := request("production", true).Code; got != http.StatusNotFound {
		t.Fatalf("production: got %d", got)
	}

	w := request("development", true)
	if w.Code != http.StatusOK {
		t.Fatalf("pair: got %d: %s", w.Code, w.Body.String())
	}
	var session struct {
		AccessToken string `json:"accessToken"`
		User        struct {
			Email string `json:"email"`
		} `json:"user"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if session.AccessToken == "" || session.User.Email != "first@example.com" {
		t.Fatalf("unexpected session: %+v", session)
	}

	r := httptest.NewRequest(http.MethodPost, "/mobile/graphql", bytes.NewBufferString(`{"query":"{ viewer { id email } }"}`))
	r.Header.Set("Authorization", "Bearer "+session.AccessToken)
	r.Header.Set("Content-Type", "application/json")
	viewer := httptest.NewRecorder()
	app.New(base, persistence, zap.NewNop()).ServeHTTP(viewer, r)
	if viewer.Code != http.StatusOK || !bytes.Contains(viewer.Body.Bytes(), []byte(`first@example.com`)) {
		t.Fatalf("viewer: got %d: %s", viewer.Code, viewer.Body.String())
	}
}
