package task

import (
	"context"
	"fmt"

	"mobicode/apps/server/internal/apperror"
	"mobicode/apps/server/internal/entity"
	"mobicode/apps/server/internal/schema"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Repository interface {
	Create(context.Context, *entity.Task) error
	GetByID(context.Context, uint) (*entity.Task, error)
	List(context.Context) ([]entity.Task, error)
	Update(context.Context, *entity.Task) error
	Delete(context.Context, uint) error
}
type Service interface {
	Create(context.Context, schema.CreateTaskRequest) (*schema.TaskResponse, error)
	Get(context.Context, uint) (*schema.TaskResponse, error)
	List(context.Context) ([]schema.TaskResponse, error)
	Update(context.Context, uint, schema.UpdateTaskRequest) (*schema.TaskResponse, error)
	Delete(context.Context, uint) error
}
type service struct {
	repo   Repository
	logger *zap.Logger
}

// New creates the Task service.
func New(repo Repository, logger *zap.Logger) Service { return &service{repo: repo, logger: logger} }

// response converts a persisted Task to its API representation.
func response(t *entity.Task) *schema.TaskResponse {
	return &schema.TaskResponse{ID: t.ID, Title: t.Title, Status: t.Status, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt}
}

// Create creates a new pending Task.
func (s *service) Create(ctx context.Context, req schema.CreateTaskRequest) (*schema.TaskResponse, error) {
	task := &entity.Task{Title: req.Title, Status: "pending"}
	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}
	s.logger.Info("task created", zap.Uint("task_id", task.ID))
	return response(task), nil
}

// Get returns one Task by identifier.
func (s *service) Get(ctx context.Context, id uint) (*schema.TaskResponse, error) {
	t, e := s.repo.GetByID(ctx, id)
	if e != nil {
		return nil, e
	}
	return response(t), nil
}

// List returns all Tasks in identifier order.
func (s *service) List(ctx context.Context) ([]schema.TaskResponse, error) {
	tasks, e := s.repo.List(ctx)
	if e != nil {
		return nil, e
	}
	out := make([]schema.TaskResponse, 0, len(tasks))
	for i := range tasks {
		out = append(out, *response(&tasks[i]))
	}
	return out, nil
}

// Update applies the supplied changes to a Task.
func (s *service) Update(ctx context.Context, id uint, req schema.UpdateTaskRequest) (*schema.TaskResponse, error) {
	t, e := s.repo.GetByID(ctx, id)
	if e != nil {
		return nil, e
	}
	if req.Title != nil {
		t.Title = *req.Title
	}
	if req.Status != nil {
		t.Status = *req.Status
	}
	if t.Title == "" {
		return nil, apperror.New("validation", "title is required", fmt.Errorf("empty title"))
	}
	if e = s.repo.Update(ctx, t); e != nil {
		return nil, e
	}
	return response(t), nil
}

// Delete removes a Task by identifier.
func (s *service) Delete(ctx context.Context, id uint) error { return s.repo.Delete(ctx, id) }

var Module = fx.Module("task-service", fx.Provide(New))
