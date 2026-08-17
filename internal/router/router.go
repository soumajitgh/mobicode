package router

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/soumajitgh/mobicode/internal/config"
	"github.com/soumajitgh/mobicode/internal/middleware"
	"github.com/soumajitgh/mobicode/internal/requestctx"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/fx"
)

// New builds the HTTP router and exposes only the standard library contract.
func New(cfg config.Config, db *sql.DB, graphqlHandler http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.RequestContext)
	r.Use(chimiddleware.ClientIPFromRemoteAddr)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(
		middleware.SecurityHeaders(),
		middleware.BodyLimit(cfg.Server.MaxBodyBytes),
		middleware.CORS(cfg.CORS.AllowedOrigins),
	)
	r.Method(http.MethodPost, "/graphql", graphqlHandler)
	r.Get("/health/live", healthHandler(nil))
	r.Get("/health/ready", healthHandler(db))
	if cfg.Environment == "development" {
		r.Handle("/playground", playground.Handler("GraphQL Playground", "/graphql"))
	}
	return r
}

type healthResponse struct {
	Data      *healthData  `json:"data"`
	Error     *healthError `json:"error"`
	RequestID string       `json:"request_id"`
}

type healthData struct {
	Status string `json:"status"`
}

type healthError struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Fields  []string `json:"fields"`
}

func healthHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := healthResponse{Data: &healthData{Status: "ok"}, RequestID: requestctx.RequestID(r.Context())}
		status := http.StatusOK
		if db != nil {
			if err := db.PingContext(r.Context()); err != nil {
				status = http.StatusServiceUnavailable
				response.Data = nil
				response.Error = &healthError{Code: "dependency_unavailable", Message: "database unavailable", Fields: []string{}}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(response)
	}
}

var Module = fx.Module("router", fx.Provide(fx.Annotate(New, fx.ParamTags("", "", `name:"graphql"`))))
