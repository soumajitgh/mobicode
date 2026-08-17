package repository

import (
	"mobicode/apps/server/internal/repository/task"

	"go.uber.org/fx"
)

var Module = fx.Module("repository", task.Module)
