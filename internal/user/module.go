package user

import "go.uber.org/fx"

// Module provides the user persistence and business-logic dependencies.
var Module = fx.Module(
	"user",
	fx.Provide(NewRepository, NewService),
)
