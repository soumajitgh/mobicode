package http

import (
	"net"
	"net/http"
	"strings"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	appgraphql "github.com/soumajitgh/mobicode/internal/graphql"
	"github.com/soumajitgh/mobicode/internal/http/handler"
	"github.com/soumajitgh/mobicode/internal/http/middleware"
)

// NewRouter builds the server's HTTP handler.
func NewRouter(resolver *appgraphql.Resolver, enablePlayground bool) *chi.Mux {
	r := chi.NewRouter()
	middleware.Apply(r)
	r.Get("/healthz", handler.Health)

	r.Route("/mobile", func(r chi.Router) {
		graphqlHandler := newGraphQLHandler(resolver)
		r.Method(http.MethodPost, "/graphql", graphqlHandler)
		r.Method(http.MethodOptions, "/graphql", graphqlHandler)
		if enablePlayground {
			r.With(localOnly).Get("/graphql/playground", playground.Handler("Mobile GraphQL", "/mobile/graphql").ServeHTTP)
		}
	})
	return r
}

func localOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		ip := net.ParseIP(host)
		requestHost := r.Host
		if name, _, splitErr := net.SplitHostPort(requestHost); splitErr == nil {
			requestHost = name
		}
		requestIP := net.ParseIP(requestHost)
		localHost := strings.EqualFold(requestHost, "localhost") || requestIP != nil && requestIP.IsLoopback()
		if err != nil || ip == nil || !ip.IsLoopback() || !localHost {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
