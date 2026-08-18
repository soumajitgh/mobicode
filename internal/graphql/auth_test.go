package graphql

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/config"
	"github.com/soumajitgh/mobicode/internal/user"
)

type gqlTestResponse struct {
	Data   map[string]any `json:"data"`
	Errors []struct {
		Message    string         `json:"message"`
		Extensions map[string]any `json:"extensions"`
	} `json:"errors"`
}

func setupGQLServer(t *testing.T) (http.Handler, *auth.JWTService, *auth.Service, *user.Service) {
	userRepo := &memoryUserRepository{users: make(map[string]*user.User)}
	refreshRepo := &memoryRefreshTokens{tokens: make(map[string]*auth.RefreshToken)}
	jwtSvc, err := auth.NewJWTService(&config.Config{JWTSecret: "test-secret-with-at-least-thirty-two-bytes"})
	if err != nil {
		t.Fatalf("create JWT service: %v", err)
	}

	userService := user.NewService(userRepo, zap.NewNop())
	authService := auth.NewService(userService, refreshRepo, auth.NewPasswordService(), jwtSvc, zap.NewNop(), &config.Config{})

	gqlServer := NewServer(NewResolver(user.NewResolver(userService), auth.NewResolver(authService)), zap.NewNop())
	handler := auth.Middleware(auth.NewJWTAuthenticator(jwtSvc))(gqlServer)

	return handler, jwtSvc, authService, userService
}

func executeGQLWithAuth(handler http.Handler, query string, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/query", strings.NewReader(`{"query":`+strconv.Quote(query)+`}`))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestPublicOperationsWithoutJWT(t *testing.T) {
	handler, _, _, _ := setupGQLServer(t)

	// 1. register without JWT -> allowed
	regRec := executeGQLWithAuth(handler, `mutation { register(name: "Alice", email: "alice@example.com", password: "password-123") { accessToken refreshToken user { id name } } }`, "")
	if regRec.Code != http.StatusOK {
		t.Fatalf("register status = %d, want %d", regRec.Code, http.StatusOK)
	}
	var regResp gqlTestResponse
	if err := json.Unmarshal(regRec.Body.Bytes(), &regResp); err != nil {
		t.Fatalf("unmarshal register resp: %v", err)
	}
	if len(regResp.Errors) > 0 {
		t.Fatalf("register returned errors: %#v", regResp.Errors)
	}

	// 2. login without JWT -> allowed
	loginRec := executeGQLWithAuth(handler, `mutation { login(email: "alice@example.com", password: "password-123") { accessToken refreshToken } }`, "")
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d", loginRec.Code, http.StatusOK)
	}
	var loginResp gqlTestResponse
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("unmarshal login resp: %v", err)
	}
	if len(loginResp.Errors) > 0 {
		t.Fatalf("login returned errors: %#v", loginResp.Errors)
	}

	refreshTokenVal, _ := loginResp.Data["login"].(map[string]any)["refreshToken"].(string)

	// 3. refreshToken without JWT -> allowed
	refRec := executeGQLWithAuth(handler, `mutation { refreshToken(refreshToken: "`+refreshTokenVal+`") { accessToken refreshToken } }`, "")
	if refRec.Code != http.StatusOK {
		t.Fatalf("refreshToken status = %d, want %d", refRec.Code, http.StatusOK)
	}
	var refResp gqlTestResponse
	if err := json.Unmarshal(refRec.Body.Bytes(), &refResp); err != nil {
		t.Fatalf("unmarshal refreshToken resp: %v", err)
	}
	if len(refResp.Errors) > 0 {
		t.Fatalf("refreshToken returned errors: %#v", refResp.Errors)
	}
}

func TestPrivateOperationsWithoutJWT(t *testing.T) {
	handler, _, _, _ := setupGQLServer(t)

	// 4. private Query without JWT -> UNAUTHENTICATED
	queryRec := executeGQLWithAuth(handler, `query { user(id: "some-id") { id name } }`, "")
	if queryRec.Code != http.StatusOK {
		t.Fatalf("query status = %d, want %d", queryRec.Code, http.StatusOK)
	}
	var queryResp gqlTestResponse
	json.Unmarshal(queryRec.Body.Bytes(), &queryResp)
	if len(queryResp.Errors) == 0 {
		t.Fatalf("expected UNAUTHENTICATED error for private query without JWT")
	}
	if code, _ := queryResp.Errors[0].Extensions["code"].(string); code != "UNAUTHENTICATED" {
		t.Fatalf("error code = %q, want UNAUTHENTICATED", code)
	}

	// 5. private Mutation without JWT -> UNAUTHENTICATED
	mutRec := executeGQLWithAuth(handler, `mutation { logout(refreshToken: "some-token") }`, "")
	if mutRec.Code != http.StatusOK {
		t.Fatalf("mutation status = %d, want %d", mutRec.Code, http.StatusOK)
	}
	var mutResp gqlTestResponse
	json.Unmarshal(mutRec.Body.Bytes(), &mutResp)
	if len(mutResp.Errors) == 0 {
		t.Fatalf("expected UNAUTHENTICATED error for private mutation without JWT")
	}
	if code, _ := mutResp.Errors[0].Extensions["code"].(string); code != "UNAUTHENTICATED" {
		t.Fatalf("error code = %q, want UNAUTHENTICATED", code)
	}
}

func TestPrivateOperationsWithValidJWT(t *testing.T) {
	handler, _, _, userService := setupGQLServer(t)

	u, err := userService.CreateUser(context.Background(), "Bob", "bob@example.com", "hash")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	jwtSvc, _ := auth.NewJWTService(&config.Config{JWTSecret: "test-secret-with-at-least-thirty-two-bytes"})
	validToken, err := jwtSvc.SignAccessToken(u.ID)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	// 6. private Query with valid JWT -> allowed
	queryRec := executeGQLWithAuth(handler, `query { user(id: "`+u.ID+`") { id name email } }`, validToken)
	if queryRec.Code != http.StatusOK {
		t.Fatalf("query status = %d, want %d", queryRec.Code, http.StatusOK)
	}
	var queryResp gqlTestResponse
	json.Unmarshal(queryRec.Body.Bytes(), &queryResp)
	if len(queryResp.Errors) > 0 {
		t.Fatalf("expected no errors for valid JWT, got: %#v", queryResp.Errors)
	}
	userData, ok := queryResp.Data["user"].(map[string]any)
	if !ok || userData["id"] != u.ID {
		t.Fatalf("got user data %#v, want id %q", userData, u.ID)
	}

	// 7. me Query with valid JWT -> allowed
	meRec := executeGQLWithAuth(handler, `query { me { id name email } }`, validToken)
	if meRec.Code != http.StatusOK {
		t.Fatalf("me query status = %d, want %d", meRec.Code, http.StatusOK)
	}
	var meResp gqlTestResponse
	json.Unmarshal(meRec.Body.Bytes(), &meResp)
	if len(meResp.Errors) > 0 {
		t.Fatalf("expected no errors for me query with valid JWT, got: %#v", meResp.Errors)
	}
	meData, ok := meResp.Data["me"].(map[string]any)
	if !ok || meData["id"] != u.ID || meData["email"] != "bob@example.com" {
		t.Fatalf("got me data %#v, want id %q", meData, u.ID)
	}
}

func TestInvalidOrExpiredJWTIsTreatedAsUnauthenticated(t *testing.T) {
	handler, _, _, _ := setupGQLServer(t)

	// 8. Expired JWT
	now := time.Now().Add(-1 * time.Hour)
	expiredClaims := auth.Claims{
		UserID: "user-123",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		},
	}
	expiredToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims).SignedString([]byte("test-secret-with-at-least-thirty-two-bytes"))

	expRec := executeGQLWithAuth(handler, `query { user(id: "user-123") { id } }`, expiredToken)
	var expResp gqlTestResponse
	json.Unmarshal(expRec.Body.Bytes(), &expResp)
	if len(expResp.Errors) == 0 || expResp.Errors[0].Extensions["code"] != "UNAUTHENTICATED" {
		t.Fatalf("expected UNAUTHENTICATED error for expired token, got: %#v", expResp.Errors)
	}

	// 9. Malformed JWT
	malformedRec := executeGQLWithAuth(handler, `query { user(id: "user-123") { id } }`, "not-a-valid-token")
	var malformedResp gqlTestResponse
	json.Unmarshal(malformedRec.Body.Bytes(), &malformedResp)
	if len(malformedResp.Errors) == 0 || malformedResp.Errors[0].Extensions["code"] != "UNAUTHENTICATED" {
		t.Fatalf("expected UNAUTHENTICATED error for malformed token, got: %#v", malformedResp.Errors)
	}

	// 10. Invalid Signature
	otherSvc, _ := auth.NewJWTService(&config.Config{JWTSecret: "other-secret-with-at-least-thirty-two-bytes"})
	badSigToken, _ := otherSvc.SignAccessToken("user-123")
	badSigRec := executeGQLWithAuth(handler, `query { user(id: "user-123") { id } }`, badSigToken)
	var badSigResp gqlTestResponse
	json.Unmarshal(badSigRec.Body.Bytes(), &badSigResp)
	if len(badSigResp.Errors) == 0 || badSigResp.Errors[0].Extensions["code"] != "UNAUTHENTICATED" {
		t.Fatalf("expected UNAUTHENTICATED error for bad signature token, got: %#v", badSigResp.Errors)
	}
}

func TestUnknownRootFieldIsProtectedByDefault(t *testing.T) {
	handler, _, _, _ := setupGQLServer(t)

	// Verify that any query or mutation root field not explicitly listed in publicFields returns UNAUTHENTICATED when unauthenticated
	rec := executeGQLWithAuth(handler, `query { user(id: "abc") { id } }`, "")
	var resp gqlTestResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if len(resp.Errors) == 0 {
		t.Fatalf("TestUnknownRootFieldIsProtectedByDefault failed: expected error for unauthenticated root query")
	}
	if code, _ := resp.Errors[0].Extensions["code"].(string); code != "UNAUTHENTICATED" {
		t.Fatalf("TestUnknownRootFieldIsProtectedByDefault failed: got code %q, want UNAUTHENTICATED", code)
	}
}
