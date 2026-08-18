package auth

import "go.uber.org/fx"

// Module provides password, token, authenticator, repository, service, GraphQL resolver, and HTTP middleware dependencies.
var Module = fx.Module(
	"auth",
	fx.Provide(
		NewPasswordService,
		NewJWTService,
		NewAuthenticator,
		NewRepository,
		NewService,
		NewResolver,
		Middleware,
	),
)
