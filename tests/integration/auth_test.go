package integration_test

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/soumajitgh/mobicode/internal/auth"
)

func TestPasswordResetInvalidatesBrowserSession(t *testing.T) {
	testApp := newTestApplication(t)
	service := auth.Service{Users: testApp.store.Users, RecoveryToken: testRecoveryToken}
	if _, err := service.CreateInitialUser(context.Background(), "owner@example.com", "initial-password"); err != nil {
		t.Fatal(err)
	}

	response := testApp.postForm("/auth/login", url.Values{
		"email":    {"owner@example.com"},
		"password": {"initial-password"},
	})
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("login without CSRF: status = %d", response.StatusCode)
	}
	closeBody(t, response)

	csrf := testApp.csrf("/auth/login")
	response = testApp.postForm("/auth/login", url.Values{
		"email":      {"owner@example.com"},
		"password":   {"initial-password"},
		"csrf_token": {csrf},
	})
	if response.StatusCode != http.StatusSeeOther || response.Header.Get("Location") != "/" {
		t.Fatalf("login: status = %d, location = %q", response.StatusCode, response.Header.Get("Location"))
	}
	closeBody(t, response)

	response = testApp.get("/")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("authenticated home: status = %d", response.StatusCode)
	}
	closeBody(t, response)

	csrf = testApp.csrf("/auth/reset-password")
	response = testApp.postForm("/auth/reset-password", url.Values{
		"email":          {"owner@example.com"},
		"password":       {"replacement-password"},
		"recovery_token": {testRecoveryToken},
		"csrf_token":     {csrf},
	})
	if response.StatusCode != http.StatusSeeOther || response.Header.Get("Location") != "/auth/login" {
		t.Fatalf("password reset: status = %d, location = %q", response.StatusCode, response.Header.Get("Location"))
	}
	closeBody(t, response)

	response = testApp.get("/")
	if response.StatusCode != http.StatusSeeOther || !strings.HasPrefix(response.Header.Get("Location"), "/auth/login?next=") {
		t.Fatalf("session after reset: status = %d, location = %q", response.StatusCode, response.Header.Get("Location"))
	}
	closeBody(t, response)
}
