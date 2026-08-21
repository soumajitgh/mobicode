package auth

import "go.uber.org/fx"

// Module provides single-owner NIP-98 authentication dependencies.
var Module = fx.Module(
	"auth",
	fx.Provide(
		NewOwnerRepository,
		NewOwnerService,
		NewNIP98Verifier,
		RequireOwner,
	),
	fx.Invoke(BootstrapDevelopmentIdentity),
)
