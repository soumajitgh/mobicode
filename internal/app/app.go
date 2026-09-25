package app

import (
	"net/http"
	"os"

	appgraphql "github.com/soumajitgh/mobicode/internal/graphql"
	"github.com/soumajitgh/mobicode/internal/health"
	apphttp "github.com/soumajitgh/mobicode/internal/http"
)

// New assembles dependencies for the HTTP application.
func New() http.Handler {
	resolver := &appgraphql.Resolver{HealthService: &health.Service{}}
	return apphttp.NewRouter(resolver, os.Getenv("MOBICODE_SERVER_PLAYGROUND") == "true")
}
