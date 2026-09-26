package app

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"

	"github.com/soumajitgh/mobicode/internal/store"
	"github.com/soumajitgh/mobicode/internal/store/model"
	"github.com/soumajitgh/mobicode/internal/store/repository"
)

func TestOnboardingFlow(t *testing.T) {
	t.Setenv("MOBICODE_SERVER_SECRET_TOKEN", "a-long-recovery-token-of-at-least-32-bytes")
	dbPath := filepath.Join(t.TempDir(), "onboarding.db")
	s, err := store.Open(context.Background(), store.Config{SQLitePath: dbPath, GORMLogLevel: logger.Silent})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	handler := New(s, zap.NewNop())
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:           jar,
		Transport:     localTransport{handler},
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}

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

	readBody := func(res *http.Response) string {
		t.Helper()
		defer func() { _ = res.Body.Close() }()
		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	token := func(path string) string {
		t.Helper()
		res := get(path)
		b := readBody(res)
		matches := regexp.MustCompile(`name="csrf_token" value="([^"]+)"`).FindSubmatch([]byte(b))
		if len(matches) != 2 {
			t.Fatalf("missing CSRF in %s: %s", path, b)
		}
		return string(matches[1])
	}

	// 1. Fresh installation behavior
	// Opening / must redirect to /onboarding
	res := get("/")
	if res.StatusCode != http.StatusSeeOther || res.Header.Get("Location") != "/onboarding" {
		t.Fatalf("expected redirect to /onboarding on fresh install, got %d %s", res.StatusCode, res.Header.Get("Location"))
	}
	_ = res.Body.Close()

	// Normal web routes must redirect to /onboarding
	for _, path := range []string{"/auth/login", "/auth/register", "/partials/status"} {
		res := get(path)
		if res.StatusCode != http.StatusSeeOther || res.Header.Get("Location") != "/onboarding" {
			t.Fatalf("expected %s to redirect to /onboarding, got %d %s", path, res.StatusCode, res.Header.Get("Location"))
		}
		_ = res.Body.Close()
	}

	// Static assets and health check must remain available
	for _, path := range []string{"/healthz", "/assets/css/app.css"} {
		res := get(path)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for %s, got %d", path, res.StatusCode)
		}
		_ = res.Body.Close()
	}

	// Mobile GraphQL should return 503 while onboarding is incomplete
	graphqlReq, _ := http.NewRequest(http.MethodPost, "http://mobicode.test/mobile/graphql", bytes.NewBufferString(`{"query":"{ health { status } }"}`))
	graphqlReq.Header.Set("Content-Type", "application/json")
	graphqlRes, err := client.Do(graphqlReq)
	if err != nil || graphqlRes.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 for GraphQL when unconfigured, got %d %v", graphqlRes.StatusCode, err)
	}
	_ = graphqlRes.Body.Close()

	// 2. Navigation restrictions before account creation
	// Visiting /onboarding?step=2 or step=3 directly must redirect to step=1
	for _, stepPath := range []string{"/onboarding?step=2", "/onboarding?step=3"} {
		res := get(stepPath)
		if res.StatusCode != http.StatusSeeOther || res.Header.Get("Location") != "/onboarding?step=1" {
			t.Fatalf("expected %s to redirect to /onboarding?step=1, got %d %s", stepPath, res.StatusCode, res.Header.Get("Location"))
		}
		_ = res.Body.Close()
	}

	// Attempting to finish onboarding without user must redirect to step=1
	csrf := token("/onboarding")
	res = post("/onboarding/finish", url.Values{"csrf_token": {csrf}})
	if res.StatusCode != http.StatusSeeOther || res.Header.Get("Location") != "/onboarding?step=1" {
		t.Fatalf("expected finish without user to redirect to /onboarding?step=1, got %d %s", res.StatusCode, res.Header.Get("Location"))
	}
	_ = res.Body.Close()

	// 3. Step 1 - Form presentation & Validation
	res = get("/onboarding")
	body := readBody(res)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on /onboarding, got %d", res.StatusCode)
	}
	for _, expectedText := range []string{"Create Initial User", "Create User", "Required", "Pair Phone", "Optional", "Connect GitHub"} {
		if !strings.Contains(body, expectedText) {
			t.Fatalf("expected onboarding page to contain %q", expectedText)
		}
	}

	// Invalid input: short password (< 12 characters)
	csrf = token("/onboarding")
	res = post("/onboarding", url.Values{"email": {"owner@example.com"}, "password": {"short"}, "confirm_password": {"short"}, "csrf_token": {csrf}})
	body = readBody(res)
	if res.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for short password, got %d", res.StatusCode)
	}
	if !strings.Contains(body, "Password must be at least 12 characters") {
		t.Fatalf("expected password length error in response, got %s", body)
	}

	// Invalid input: invalid email format
	csrf = token("/onboarding")
	res = post("/onboarding", url.Values{"email": {"not-an-email"}, "password": {"validpassword123"}, "confirm_password": {"validpassword123"}, "csrf_token": {csrf}})
	body = readBody(res)
	if res.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for invalid email, got %d", res.StatusCode)
	}
	if !strings.Contains(body, "Please enter a valid email address") {
		t.Fatalf("expected invalid email error in response, got %s", body)
	}

	// 4. Successful User Creation
	csrf = token("/onboarding")
	res = post("/onboarding", url.Values{"email": {" Owner@Example.com "}, "password": {"validpassword123"}, "confirm_password": {"validpassword123"}, "csrf_token": {csrf}})
	if res.StatusCode != http.StatusSeeOther || res.Header.Get("Location") != "/onboarding?step=2" {
		t.Fatalf("expected redirect to /onboarding?step=2, got %d %s", res.StatusCode, res.Header.Get("Location"))
	}
	_ = res.Body.Close()

	// Verify user exists in database
	count, err := s.Users.Count(context.Background())
	if err != nil || count != 1 {
		t.Fatalf("expected 1 user in database, got %d, err: %v", count, err)
	}

	// 5. Step 2 - Pair Your Phone Placeholder
	res = get("/onboarding?step=2")
	body = readBody(res)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on /onboarding?step=2, got %d", res.StatusCode)
	}
	for _, expectedText := range []string{"Pair Your Phone", "Mobile Device Companion", "Continue"} {
		if !strings.Contains(body, expectedText) {
			t.Fatalf("expected step 2 page to contain %q", expectedText)
		}
	}

	// Navigate Back to Step 1: Must show completed state and not re-prompt or duplicate
	res = get("/onboarding?step=1")
	body = readBody(res)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on /onboarding?step=1 after creation, got %d", res.StatusCode)
	}
	if !strings.Contains(body, "User Account Created") || !strings.Contains(body, "owner@example.com") {
		t.Fatalf("expected step 1 to show completed user account, got %s", body)
	}
	if !strings.Contains(body, "Continue to Step 2") {
		t.Fatalf("expected continue button on completed step 1, got %s", body)
	}

	// Submitting user creation again must not duplicate user
	csrf = token("/onboarding?step=3")
	res = post("/onboarding", url.Values{"email": {"duplicate@example.com"}, "password": {"validpassword123"}, "confirm_password": {"validpassword123"}, "csrf_token": {csrf}})
	if res.StatusCode != http.StatusSeeOther || res.Header.Get("Location") != "/onboarding?step=2" {
		t.Fatalf("expected redirect to step 2 on repeated post, got %d %s", res.StatusCode, res.Header.Get("Location"))
	}
	_ = res.Body.Close()
	count, _ = s.Users.Count(context.Background())
	if count != 1 {
		t.Fatalf("expected user count to remain 1, got %d", count)
	}

	// 6. Step 3 - Connect GitHub Placeholder
	res = get("/onboarding?step=3")
	body = readBody(res)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on /onboarding?step=3, got %d", res.StatusCode)
	}
	for _, expectedText := range []string{"Connect GitHub", "GitHub Integration", "Finish Setup"} {
		if !strings.Contains(body, expectedText) {
			t.Fatalf("expected step 3 page to contain %q", expectedText)
		}
	}

	// Normal web routes must still remain unavailable while onboarding session is in progress
	res = get("/partials/status")
	if res.StatusCode != http.StatusSeeOther || res.Header.Get("Location") != "/onboarding" {
		t.Fatalf("expected /partials/status to redirect to /onboarding while wizard in progress, got %d %s", res.StatusCode, res.Header.Get("Location"))
	}
	_ = res.Body.Close()

	// 7. Finish Onboarding
	csrf = token("/onboarding?step=3")
	res = post("/onboarding/finish", url.Values{"csrf_token": {csrf}})
	if res.StatusCode != http.StatusSeeOther || res.Header.Get("Location") != "/" {
		t.Fatalf("expected redirect to / on finish, got %d %s", res.StatusCode, res.Header.Get("Location"))
	}
	_ = res.Body.Close()

	// 8. Post-Onboarding Application Access
	// Normal route / is now accessible and authenticated!
	res = get("/")
	body = readBody(res)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on / after onboarding completion, got %d", res.StatusCode)
	}
	if !strings.Contains(body, "MobiCode") || !strings.Contains(body, "Log out") {
		t.Fatalf("expected home page with logout after completion, got %s", body)
	}

	// /onboarding route must no longer be available as an active setup flow
	res = get("/onboarding")
	if res.StatusCode != http.StatusSeeOther || res.Header.Get("Location") != "/" {
		t.Fatalf("expected /onboarding to redirect to / after completion, got %d %s", res.StatusCode, res.Header.Get("Location"))
	}
	_ = res.Body.Close()

	// Mobile GraphQL is now accessible
	graphqlRes, err = client.Do(graphqlReq)
	if err != nil || graphqlRes.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for GraphQL after completion, got %d %v", graphqlRes.StatusCode, err)
	}
	_ = graphqlRes.Body.Close()

	// 9. Reset behavior: Deleting all users resets installation back to onboarding
	s.Users = &emptyUserRepository{s.Users}
	// We reinitialize a client without session
	freshClient := &http.Client{
		Transport:     localTransport{New(s, zap.NewNop())},
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}
	resetRes, err := freshClient.Get("http://mobicode.test/")
	if err != nil || resetRes.StatusCode != http.StatusSeeOther || resetRes.Header.Get("Location") != "/onboarding" {
		t.Fatalf("expected redirect to /onboarding when users are 0, got %d", resetRes.StatusCode)
	}
	_ = resetRes.Body.Close()
}

type emptyUserRepository struct {
	repository.UserRepository
}

func (e *emptyUserRepository) Count(_ context.Context) (int64, error) {
	return 0, nil
}

func (e *emptyUserRepository) Create(ctx context.Context, u *model.User) error {
	return e.UserRepository.Create(ctx, u)
}

func (e *emptyUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return e.UserRepository.FindByEmail(ctx, email)
}

func (e *emptyUserRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	return e.UserRepository.FindByID(ctx, id)
}

func (e *emptyUserRepository) ReplacePassword(ctx context.Context, id uint, hash string) error {
	return e.UserRepository.ReplacePassword(ctx, id, hash)
}

func (e *emptyUserRepository) FindFirst(ctx context.Context) (*model.User, error) {
	return e.UserRepository.FindFirst(ctx)
}
