package task

import (
	"context"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/soumajitgh/mobicode/internal/apperror"
	"github.com/soumajitgh/mobicode/internal/entity"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Repository interface {
	Create(context.Context, *entity.Task) error
	GetByID(context.Context, uint) (*entity.Task, error)
	List(context.Context, uint, uint) ([]entity.Task, error)
	Update(context.Context, *entity.Task) error
	Delete(context.Context, uint) error
}
type Service interface {
	Create(context.Context, CreateInput) (*Task, error)
	Get(context.Context, uint) (*Task, error)
	List(context.Context, ListInput) (*TaskConnection, error)
	Update(context.Context, uint, UpdateInput) (*Task, error)
	Delete(context.Context, uint) error
}

// CreateInput contains the business input needed to create a Task.
type CreateInput struct{ Title string }

// UpdateInput contains the optional business changes for a Task.
type UpdateInput struct{ Title, Status *string }

// ListInput describes a bounded Task page request.
type ListInput struct {
	First int
	After *string
}

// Task is the application representation of a persisted Task.
type Task struct {
	ID                   uint
	Title, Status        string
	CreatedAt, UpdatedAt time.Time
}

// TaskConnection is a cursor-based page of Tasks.
type TaskConnection struct {
	Nodes       []Task
	HasNextPage bool
	EndCursor   *string
}
type service struct {
	repo   Repository
	logger *zap.Logger
}

// New creates the Task service.
func New(repo Repository, logger *zap.Logger) Service { return &service{repo: repo, logger: logger} }

// taskResponse converts a persisted Task to its application representation.
func taskResponse(t *entity.Task) *Task {
	return &Task{ID: t.ID, Title: t.Title, Status: t.Status, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt}
}

// validateTitle enforces the Task title business constraints.
func validateTitle(title string) error {
	length := utf8.RuneCountInString(title)
	if strings.TrimSpace(title) == "" || length < 3 || length > 200 {
		return apperror.New("validation", "title must be between 3 and 200 characters", nil)
	}
	return nil
}

// Create creates a new pending Task.
func (s *service) Create(ctx context.Context, input CreateInput) (*Task, error) {
	if err := validateTitle(input.Title); err != nil {
		return nil, err
	}
	task := &entity.Task{Title: input.Title, Status: "pending"}
	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}
	s.logger.Info("task created", zap.Uint("task_id", task.ID))
	return taskResponse(task), nil
}

// Get returns one Task by identifier.
func (s *service) Get(ctx context.Context, id uint) (*Task, error) {
	t, e := s.repo.GetByID(ctx, id)
	if e != nil {
		return nil, e
	}
	return taskResponse(t), nil
}

// List returns a bounded cursor page of Tasks.
func (s *service) List(ctx context.Context, input ListInput) (*TaskConnection, error) {
	first := input.First
	if first == 0 {
		first = 20
	}
	if first < 1 || first > 100 {
		return nil, apperror.New("validation", "first must be between 1 and 100", nil)
	}
	var after uint
	if input.After != nil {
		parsed, err := strconv.ParseUint(*input.After, 10, 64)
		if err != nil || parsed == 0 {
			return nil, apperror.New("validation", "after must be a valid task cursor", err)
		}
		after = uint(parsed)
	}
	tasks, e := s.repo.List(ctx, after, uint(first+1))
	if e != nil {
		return nil, e
	}
	hasNextPage := len(tasks) > first
	if hasNextPage {
		tasks = tasks[:first]
	}
	nodes := make([]Task, 0, len(tasks))
	for i := range tasks {
		nodes = append(nodes, *taskResponse(&tasks[i]))
	}
	var endCursor *string
	if len(nodes) > 0 {
		cursor := strconv.FormatUint(uint64(nodes[len(nodes)-1].ID), 10)
		endCursor = &cursor
	}
	return &TaskConnection{Nodes: nodes, HasNextPage: hasNextPage, EndCursor: endCursor}, nil
}

// Update applies the supplied changes to a Task.
func (s *service) Update(ctx context.Context, id uint, input UpdateInput) (*Task, error) {
	t, e := s.repo.GetByID(ctx, id)
	if e != nil {
		return nil, e
	}
	if input.Title != nil {
		if err := validateTitle(*input.Title); err != nil {
			return nil, err
		}
		t.Title = *input.Title
	}
	if input.Status != nil {
		t.Status = *input.Status
	}
	if e = s.repo.Update(ctx, t); e != nil {
		return nil, e
	}
	return taskResponse(t), nil
}

// Delete removes a Task by identifier.
func (s *service) Delete(ctx context.Context, id uint) error { return s.repo.Delete(ctx, id) }

var Module = fx.Module("task-service", fx.Provide(New))
