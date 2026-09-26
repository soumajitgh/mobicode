package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestKeyedRateLimiter(t *testing.T) {
	limiter := NewKeyedRateLimiter(rate.Every(time.Hour), 10, 2*time.Hour)
	for range 10 {
		if !limiter.Allow("login:127.0.0.1") {
			t.Fatal("blocked before the burst was spent")
		}
	}
	if limiter.Allow("login:127.0.0.1") {
		t.Fatal("allowed a request after the burst was spent")
	}
	if !limiter.Allow("register:127.0.0.1") || !limiter.Allow("login:127.0.0.2") {
		t.Fatal("one key affected another key")
	}

	limiter.mu.Lock()
	limiter.entries["login:127.0.0.1"].lastSeen = time.Now().Add(-3 * time.Hour)
	limiter.nextCleanup = time.Time{}
	limiter.mu.Unlock()
	if !limiter.Allow("register:127.0.0.1") {
		t.Fatal("active key was lost during cleanup")
	}
	limiter.mu.Lock()
	_, exists := limiter.entries["login:127.0.0.1"]
	limiter.mu.Unlock()
	if exists {
		t.Fatal("idle key was not removed")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	limiter := NewKeyedRateLimiter(rate.Every(time.Hour), 1, 2*time.Hour)
	called := 0
	handler := RateLimit(limiter, ClientIP)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called++
		w.WriteHeader(http.StatusNoContent)
	}))
	for i, want := range []int{http.StatusNoContent, http.StatusTooManyRequests} {
		r := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		r.RemoteAddr = "192.0.2.1:1234"
		r.Header.Set("X-Forwarded-For", "192.0.2.99")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != want {
			t.Fatalf("request %d: status = %d, want %d", i+1, w.Code, want)
		}
	}
	if called != 1 {
		t.Fatalf("handler called %d times, want 1", called)
	}

	r := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	r.RemoteAddr = "192.0.2.2:5678"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent {
		t.Fatalf("different client status = %d, want %d", w.Code, http.StatusNoContent)
	}
}
