package service

import (
	"github.com/soumajitgh/mobicode/internal/service/task"

	"go.uber.org/fx"
)

var Module = fx.Module("service", task.Module)
