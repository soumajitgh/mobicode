package app

import (
	"net/http"
	"os"

	appgraphql "github.com/soumajitgh/mobicode/internal/graphql"
	"github.com/soumajitgh/mobicode/internal/health"
	apphttp "github.com/soumajitgh/mobicode/internal/http"
	"github.com/soumajitgh/mobicode/internal/web/handlers"
)

// New assembles dependencies for the HTTP application.
func New() http.Handler {
	healthService := &health.Service{}
	resolver := &appgraphql.Resolver{HealthService: healthService}
	webHandler := &handlers.Handler{HealthService: healthService}
	return apphttp.NewRouter(
		resolver,
		webHandler,
		os.Getenv("MOBICODE_SERVER_PLAYGROUND") == "true",
		os.Getenv("MOBICODE_SERVER_DEV_ASSETS") == "true",
	)
}
