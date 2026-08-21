package testutil

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/config"
	graphqlapi "github.com/soumajitgh/mobicode/internal/graphql"
	"github.com/soumajitgh/mobicode/internal/health"
	"github.com/soumajitgh/mobicode/internal/server"
	"github.com/soumajitgh/mobicode/internal/setup"
)

// App is an in-process Mobicode application backed by an isolated SQLite database.
type App struct {
	Handler http.Handler
	DB      *gorm.DB
	Config  *config.Config
	Owner   *auth.OwnerService
	Setup   *setup.Service
}

type appConfig struct {
	owner *KeyPair
}

// AppOption changes an explicitly supported aspect of a test application.
type AppOption func(*appConfig)

// WithOwner configures the supplied identity as the application's owner.
func WithOwner(owner KeyPair) AppOption {
	return func(cfg *appConfig) { cfg.owner = &owner }
}

// NewApp constructs the production HTTP routes with test-only infrastructure.
func NewApp(t *testing.T, options ...AppOption) *App {
	t.Helper()

	optionsConfig := appConfig{}
	for _, option := range options {
		option(&optionsConfig)
	}

	db := NewDB(t)
	cfg := &config.Config{Env: "test", PublicBaseURL: "http://mobicode.test"}
	owner := auth.NewOwnerService(auth.NewOwnerRepository(db))
	if optionsConfig.owner != nil {
		require.NoError(t, owner.Configure(t.Context(), optionsConfig.owner.PublicKey))
	}
	verifier := auth.NewNIP98Verifier(cfg)
	resolver := graphqlapi.NewResolver(owner)
	graphqlServer := graphqlapi.NewServer(resolver, zap.NewNop())
	setupService := setup.NewService(setup.NewRepository(db), owner)
	setupHandler, err := setup.NewHandler(setupService, verifier, cfg)
	require.NoError(t, err)

	router := server.NewRouter()
	mountRoutes(router, health.NewHandler(), graphqlServer, auth.RequireOwner(verifier, owner), setupHandler)
	return &App{Handler: router, DB: db, Config: cfg, Owner: owner, Setup: setupService}
}

func mountRoutes(router *chi.Mux, healthHandler *health.Handler, graphqlHandler http.Handler, authMiddleware func(http.Handler) http.Handler, setupHandler *setup.Handler) {
	router.Get("/health", healthHandler.Check)
	router.Get("/healthz", healthHandler.Check)
	router.With(authMiddleware).Handle("/query", graphqlHandler)
	router.With(authMiddleware).Handle("/graphql", graphqlHandler)
	router.Mount("/setup", setupHandler.Router())
}
