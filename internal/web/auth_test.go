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
