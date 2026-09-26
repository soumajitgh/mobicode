package http

import (
	"net"
	"net/http"
	"strings"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	graphqlmiddleware "github.com/soumajitgh/mobicode/internal/graphql/middleware"
	"github.com/soumajitgh/mobicode/internal/graphql/resolver"
	"github.com/soumajitgh/mobicode/internal/http/handler"
	"github.com/soumajitgh/mobicode/internal/http/middleware"
	"github.com/soumajitgh/mobicode/internal/store/repository"
	"github.com/soumajitgh/mobicode/internal/web"
	webhandlers "github.com/soumajitgh/mobicode/internal/web/handlers"
	webmiddleware "github.com/soumajitgh/mobicode/internal/web/middleware"
)

// NewRouter builds the server's HTTP handler.
func NewRouter(resolver *resolver.Resolver, mobile repository.MobileRepository, webHandler *webhandlers.Handler, enablePlayground, devAssets bool, log *zap.Logger, browserAuth *webhandlers.Auth, onboarding *webhandlers.Onboarding) *chi.Mux {
	r := chi.NewRouter()
	middleware.Apply(r, log)
	r.Use(browserAuth.Sessions.LoadAndSave)
	r.Use(webmiddleware.GateOnboarding(onboarding))
	r.Get("/healthz", handler.Health)
	web.RegisterRoutes(r, webHandler, devAssets, browserAuth, onboarding)

	r.Route("/mobile", func(r chi.Router) {
		r.Use(graphqlmiddleware.Authenticate(browserAuth.Sessions, browserAuth.Users, mobile))
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
