package web

import "testing"

func TestSafeRedirects(t *testing.T) {
	for _, path := range []string{"/", "/partials/status?x=1"} {
		if !safeNext(path) {
			t.Fatalf("rejected %q", path)
		}
	}
	for _, path := range []string{"https://evil.example", "//evil.example", "/\\evil.example", "/\r\nLocation: evil"} {
		if safeNext(path) || redirectTarget(path) != "/" {
			t.Fatalf("accepted %q", path)
		}
	}
}

func TestEndpointRateLimit(t *testing.T) {
	limiter := &rateLimiter{entries: make(map[string]*rateEntry)}
	for range 10 {
		if !limiter.Allow("login:127.0.0.1") {
			t.Fatal("blocked early")
		}
	}
	if limiter.Allow("login:127.0.0.1") {
		t.Fatal("allowed eleventh request")
	}
	if !limiter.Allow("register:127.0.0.1") {
		t.Fatal("another endpoint was limited")
	}
}
