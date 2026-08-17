// Package app composes the application's Fx modules.
package app

import (
	"mobicode/apps/server/internal/config"
	"mobicode/apps/server/internal/controller"
	"mobicode/apps/server/internal/database"
	"mobicode/apps/server/internal/middleware"
	"mobicode/apps/server/internal/observability"
	"mobicode/apps/server/internal/repository"
	"mobicode/apps/server/internal/router"
	"mobicode/apps/server/internal/server"
	"mobicode/apps/server/internal/service"

	"go.uber.org/fx"
)

var Module = fx.Module("app", config.Module, observability.Module, database.Module, repository.Module, service.Module, controller.Module, middleware.Module, router.Module, server.Module)
