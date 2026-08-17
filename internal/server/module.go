package server

import (
	"net/http"

	"go.uber.org/fx"
)

// Module provides the router and lifecycle-managed HTTP server.
var Module = fx.Module(
	"server",
	fx.Provide(NewRouter, NewHTTPServer),
	fx.Invoke(func(*http.Server) {}),
)
