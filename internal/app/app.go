package app

import (
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"

	"github.com/soumajitgh/mobicode/internal/auth"

	"go.uber.org/zap"

	appgraphql "github.com/soumajitgh/mobicode/internal/graphql"
	"github.com/soumajitgh/mobicode/internal/health"
	apphttp "github.com/soumajitgh/mobicode/internal/http"
	"github.com/soumajitgh/mobicode/internal/store"
	"github.com/soumajitgh/mobicode/internal/web/handlers"
)

// New assembles dependencies for the HTTP application.
func New(persistence *store.Store, log *zap.Logger) http.Handler {
	healthService := &health.Service{}
	resolver := &appgraphql.Resolver{HealthService: healthService, Store: persistence}
	sessions := scs.New()
	sessions.Store = persistence.Sessions
	sessions.Lifetime = 24 * time.Hour
	sessions.Cookie.Name = "mobicode_session"
	sessions.Cookie.HttpOnly = true
	sessions.Cookie.SameSite = http.SameSiteLaxMode
	sessions.Cookie.Path = "/"
	sessions.Cookie.Secure = os.Getenv("MOBICODE_SERVER_ENV") == "production"
	browserAuth := handlers.NewAuth(&auth.Service{Users: persistence.Users, RecoveryToken: os.Getenv("MOBICODE_SERVER_SECRET_TOKEN")}, sessions, persistence.Users)
	onboarding := handlers.NewOnboarding(browserAuth, persistence.Users, sessions)
	webHandler := &handlers.Handler{HealthService: healthService, Sessions: sessions}
	return apphttp.NewRouter(
		resolver,
		webHandler,
		os.Getenv("MOBICODE_SERVER_PLAYGROUND") == "true",
		os.Getenv("MOBICODE_SERVER_DEV_ASSETS") == "true",
		log,
		browserAuth,
		onboarding,
	)
}
