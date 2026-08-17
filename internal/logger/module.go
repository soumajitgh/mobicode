package logger

import "go.uber.org/fx"

// Module provides logging dependencies.
var Module = fx.Module("logger", fx.Provide(New))
