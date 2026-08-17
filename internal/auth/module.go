package auth

import "go.uber.org/fx"

// Module provides password, token, repository, service, and GraphQL resolver dependencies.
var Module = fx.Module("auth", fx.Provide(NewPasswordService, NewJWTService, NewRepository, NewService, NewResolver))
