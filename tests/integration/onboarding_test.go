package integration_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"
)

func TestOnboardingRejectsInvalidStateTransitions(t *testing.T) {
	testApp := newTestApplication(t)

	response := testApp.get("/")
	if response.StatusCode != http.StatusSeeOther || response.Header.Get("Location") != "/onboarding" {
		t.Fatalf("fresh installation: status = %d, location = %q", response.StatusCode, response.Header.Get("Location"))
	}
	closeBody(t, response)

	response = testApp.get("/onboarding?step=2")
	if response.StatusCode != http.StatusSeeOther || response.Header.Get("Location") != "/onboarding?step=1" {
		t.Fatalf("step before account creation: status = %d, location = %q", response.StatusCode, response.Header.Get("Location"))
	}
	closeBody(t, response)

	response = testApp.postForm("/onboarding", url.Values{
		"email":            {"owner@example.com"},
		"password":         {"validpassword123"},
		"confirm_password": {"validpassword123"},
	})
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("account creation without CSRF: status = %d", response.StatusCode)
	}
	closeBody(t, response)

	csrf := testApp.csrf("/onboarding")
	response = testApp.postForm("/onboarding", url.Values{
		"email":            {"owner@example.com"},
		"password":         {"short"},
		"confirm_password": {"short"},
		"csrf_token":       {csrf},
	})
	if response.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("invalid account: status = %d", response.StatusCode)
	}
	closeBody(t, response)
	count, err := testApp.store.Users.Count(context.Background())
	if err != nil || count != 0 {
		t.Fatalf("users after invalid account = %d, err = %v", count, err)
	}
}
