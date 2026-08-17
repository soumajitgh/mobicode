// Package app composes the application's Fx modules.
package app

import (
	"github.com/soumajitgh/mobicode/internal/config"
	"github.com/soumajitgh/mobicode/internal/database"
	graphqlapi "github.com/soumajitgh/mobicode/internal/graphql"
	"github.com/soumajitgh/mobicode/internal/middleware"
	"github.com/soumajitgh/mobicode/internal/migrations"
	"github.com/soumajitgh/mobicode/internal/observability"
	"github.com/soumajitgh/mobicode/internal/repository"
	"github.com/soumajitgh/mobicode/internal/router"
	"github.com/soumajitgh/mobicode/internal/server"
	"github.com/soumajitgh/mobicode/internal/service"

	"go.uber.org/fx"
)

var Module = fx.Module("app", config.Module, observability.Module, database.Module, migrations.Module, repository.Module, service.Module, graphqlapi.Module, middleware.Module, router.Module, server.Module)
