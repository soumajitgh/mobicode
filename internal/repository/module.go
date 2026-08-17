package repository

import (
	"github.com/soumajitgh/mobicode/internal/repository/task"

	"go.uber.org/fx"
)

var Module = fx.Module("repository", task.Module)
