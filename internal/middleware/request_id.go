package middleware

import (
	"net/http"

	"github.com/soumajitgh/mobicode/internal/requestctx"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

const requestIDHeader = "X-Request-ID"

// RequestContext exposes Chi's request ID through the app-owned context contract.
func RequestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chimiddleware.GetReqID(r.Context())
		if id != "" {
			w.Header().Set(requestIDHeader, id)
		}
		next.ServeHTTP(w, r.WithContext(requestctx.WithRequestID(r.Context(), id)))
	})
}
