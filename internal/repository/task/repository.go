package task

import (
	"context"
	"errors"
	"github.com/soumajitgh/mobicode/internal/apperror"
	"github.com/soumajitgh/mobicode/internal/entity"
	taskservice "github.com/soumajitgh/mobicode/internal/service/task"
	"strings"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }

// New creates a Task repository backed by GORM.
func New(db *gorm.DB) *Repository { return &Repository{db: db} }

// persistence translates database errors to stable application errors.
func persistence(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperror.ErrNotFound
	}
	if strings.Contains(strings.ToLower(err.Error()), "unique constraint") {
		return apperror.ErrConflict
	}
	return apperror.New("internal", "persistence operation failed", err)
}

// Create persists a new Task.
func (r *Repository) Create(c context.Context, t *entity.Task) error {
	if e := r.db.WithContext(c).Create(t).Error; e != nil {
		return persistence(e)
	}
	return nil
}

// GetByID returns a Task by identifier.
func (r *Repository) GetByID(c context.Context, id uint) (*entity.Task, error) {
	var t entity.Task
	if e := r.db.WithContext(c).First(&t, id).Error; e != nil {
		return nil, persistence(e)
	}
	return &t, nil
}

// List returns a bounded page of stored Tasks in identifier order.
func (r *Repository) List(c context.Context, after uint, limit uint) ([]entity.Task, error) {
	var t []entity.Task
	query := r.db.WithContext(c).Order("id asc").Limit(int(limit))
	if after > 0 {
		query = query.Where("id > ?", after)
	}
	if e := query.Find(&t).Error; e != nil {
		return nil, persistence(e)
	}
	return t, nil
}

// Update saves changes to an existing Task.
func (r *Repository) Update(c context.Context, t *entity.Task) error {
	if e := r.db.WithContext(c).Save(t).Error; e != nil {
		return persistence(e)
	}
	return nil
}

// Delete removes a Task by identifier.
func (r *Repository) Delete(c context.Context, id uint) error {
	result := r.db.WithContext(c).Delete(&entity.Task{}, id)
	if result.Error != nil {
		return persistence(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

var Module = fx.Module("task-repository", fx.Provide(fx.Annotate(New, fx.As(new(taskservice.Repository)))))
