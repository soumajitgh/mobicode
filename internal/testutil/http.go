package testutil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Request sends a request to an in-process HTTP handler.
func Request(t *testing.T, handler http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, body)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}
