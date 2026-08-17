// Package app composes the application's Fx modules.
package app

import (
	"mobicode/apps/server/internal/config"
	"mobicode/apps/server/internal/database"
	graphqlapi "mobicode/apps/server/internal/graphql"
	"mobicode/apps/server/internal/middleware"
	"mobicode/apps/server/internal/migrations"
	"mobicode/apps/server/internal/observability"
	"mobicode/apps/server/internal/repository"
	"mobicode/apps/server/internal/router"
	"mobicode/apps/server/internal/server"
	"mobicode/apps/server/internal/service"

	"go.uber.org/fx"
)

var Module = fx.Module("app", config.Module, observability.Module, database.Module, migrations.Module, repository.Module, service.Module, graphqlapi.Module, middleware.Module, router.Module, server.Module)
