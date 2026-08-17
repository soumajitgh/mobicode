package controller

import (
	"mobicode/apps/server/internal/apperror"
	"mobicode/apps/server/internal/handler"
	"mobicode/apps/server/internal/schema"
	taskservice "mobicode/apps/server/internal/service/task"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type TaskController struct{ tasks taskservice.Service }

// NewTaskController creates the HTTP controller for Tasks.
func NewTaskController(tasks taskservice.Service) *TaskController {
	return &TaskController{tasks: tasks}
}

// taskID parses a positive Task identifier from the route.
func taskID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		handler.Respond(c, apperror.New("validation", "task ID must be a positive integer", err), nil)
		return 0, false
	}
	return uint(id), true
}

// Create handles creation of a Task.
func (c *TaskController) Create(ctx *gin.Context) {
	var req schema.CreateTaskRequest
	if !handler.BindAndValidate(ctx, &req) {
		return
	}
	out, err := c.tasks.Create(ctx.Request.Context(), req)
	if err != nil {
		handler.Respond(ctx, err, nil)
		return
	}
	handler.Success(ctx, http.StatusCreated, out)
}

// Get handles retrieval of one Task.
func (c *TaskController) Get(ctx *gin.Context) {
	id, ok := taskID(ctx)
	if !ok {
		return
	}
	out, err := c.tasks.Get(ctx.Request.Context(), id)
	handler.Respond(ctx, err, out)
}

// List handles retrieval of all Tasks.
func (c *TaskController) List(ctx *gin.Context) {
	out, err := c.tasks.List(ctx.Request.Context())
	handler.Respond(ctx, err, out)
}

// Update handles modification of a Task.
func (c *TaskController) Update(ctx *gin.Context) {
	id, ok := taskID(ctx)
	if !ok {
		return
	}
	var req schema.UpdateTaskRequest
	if !handler.BindAndValidate(ctx, &req) {
		return
	}
	out, err := c.tasks.Update(ctx.Request.Context(), id, req)
	handler.Respond(ctx, err, out)
}

// Delete handles removal of a Task.
func (c *TaskController) Delete(ctx *gin.Context) {
	id, ok := taskID(ctx)
	if !ok {
		return
	}
	err := c.tasks.Delete(ctx.Request.Context(), id)
	if err != nil {
		handler.Respond(ctx, err, nil)
		return
	}
	ctx.Status(http.StatusNoContent)
}

var Module = fx.Module("controller", fx.Provide(NewTaskController))
