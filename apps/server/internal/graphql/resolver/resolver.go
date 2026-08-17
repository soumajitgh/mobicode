package resolver

import (
	"strconv"
	"strings"

	"mobicode/apps/server/internal/apperror"
	"mobicode/apps/server/internal/graphql/model"
	"mobicode/apps/server/internal/service/task"

	"go.uber.org/fx"
)

// Resolver holds services used by generated GraphQL resolvers.
type Resolver struct{ TaskService task.Service }

// New creates the GraphQL resolver root.
func New(tasks task.Service) *Resolver { return &Resolver{TaskService: tasks} }

// parseTaskID validates and converts a GraphQL ID to an application identifier.
func parseTaskID(value string) (uint, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, apperror.New("validation", "task ID must be a positive integer", err)
	}
	return uint(id), nil
}

// mapTask converts an application Task to its GraphQL representation.
func mapTask(value *task.Task) *model.Task {
	return &model.Task{
		ID:        strconv.FormatUint(uint64(value.ID), 10),
		Title:     value.Title,
		Status:    model.TaskStatus(strings.ToUpper(value.Status)),
		CreatedAt: value.CreatedAt,
		UpdatedAt: value.UpdatedAt,
	}
}

var Module = fx.Module("graphql-resolver", fx.Provide(New))
