package service

import (
	"mobicode/apps/server/internal/service/task"

	"go.uber.org/fx"
)

var Module = fx.Module("service", task.Module)
