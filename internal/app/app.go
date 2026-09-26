package app

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"go.uber.org/zap"

	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/config"
	"github.com/soumajitgh/mobicode/internal/graphql/resolver"
	"github.com/soumajitgh/mobicode/internal/health"
	apphttp "github.com/soumajitgh/mobicode/internal/http"
	"github.com/soumajitgh/mobicode/internal/pairing"
	"github.com/soumajitgh/mobicode/internal/store"
	"github.com/soumajitgh/mobicode/internal/web/handlers"
)

// New assembles dependencies for the HTTP application.
func New(cfg *config.Config, persistence *store.Store, log *zap.Logger) http.Handler {
	healthService := &health.Service{}
	graphqlResolver := &resolver.Resolver{Pairing: &pairing.Service{Repository: persistence.Mobile, BaseURL: cfg.Server.ResolvedBaseURL}}
	sessions := scs.New()
	sessions.Store = persistence.Sessions
	sessions.Lifetime = 24 * time.Hour
	sessions.Cookie.Name = "mobicode_session"
	sessions.Cookie.HttpOnly = true
	sessions.Cookie.SameSite = http.SameSiteLaxMode
	sessions.Cookie.Path = "/"
	sessions.Cookie.Secure = cfg.Environment == "production"
	browserAuth := handlers.NewAuth(&auth.Service{Users: persistence.Users, RecoveryToken: cfg.Settings.SecretToken}, sessions, persistence.Users)
	onboarding := handlers.NewOnboarding(browserAuth, persistence.Users, sessions)
	webHandler := &handlers.Handler{HealthService: healthService, Sessions: sessions}
	return apphttp.NewRouter(
		graphqlResolver,
		persistence.Mobile,
		webHandler,
		cfg.Server.Playground,
		cfg.Server.DevAssets,
		log,
		browserAuth,
		onboarding,
	)
}
