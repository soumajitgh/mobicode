package middleware_test

import (
	"testing"

	webmiddleware "github.com/soumajitgh/mobicode/internal/web/middleware"
)

func TestSafeRedirects(t *testing.T) {
	for _, path := range []string{"/", "/partials/status?x=1"} {
		if got := webmiddleware.RedirectTarget(path); got != path {
			t.Fatalf("redirect target = %q, want %q", got, path)
		}
	}
	for _, path := range []string{"https://evil.example", "//evil.example", "/\\evil.example", "/\r\nLocation: evil"} {
		if webmiddleware.RedirectTarget(path) != "/" {
			t.Fatalf("accepted %q", path)
		}
	}
}
