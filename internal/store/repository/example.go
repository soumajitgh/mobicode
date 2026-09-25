package repository

import (
	"context"
	"fmt"

	"github.com/soumajitgh/mobicode/internal/store/model"
	"gorm.io/gorm"
)

// ExampleRepository is the persistence contract for the example model.
type ExampleRepository interface {
	Create(context.Context, *model.Example) error
	FindByID(context.Context, uint) (*model.Example, error)
	List(context.Context) ([]model.Example, error)
	Update(context.Context, *model.Example) error
	Delete(context.Context, uint) error
}

type exampleRepository struct{ db *gorm.DB }

func NewExample(db *gorm.DB) ExampleRepository { return &exampleRepository{db: db} }

func (r *exampleRepository) Create(ctx context.Context, example *model.Example) error {
	if err := r.db.WithContext(ctx).Create(example).Error; err != nil {
		return fmt.Errorf("create example: %w", err)
	}
	return nil
}

func (r *exampleRepository) FindByID(ctx context.Context, id uint) (*model.Example, error) {
	var example model.Example
	if err := r.db.WithContext(ctx).First(&example, id).Error; err != nil {
		return nil, fmt.Errorf("find example %d: %w", id, err)
	}
	return &example, nil
}

func (r *exampleRepository) List(ctx context.Context) ([]model.Example, error) {
	var examples []model.Example
	if err := r.db.WithContext(ctx).Order("id").Find(&examples).Error; err != nil {
		return nil, fmt.Errorf("list examples: %w", err)
	}
	return examples, nil
}

func (r *exampleRepository) Update(ctx context.Context, example *model.Example) error {
	result := r.db.WithContext(ctx).Model(example).Select("Name").Updates(example)
	if result.Error != nil {
		return fmt.Errorf("update example %d: %w", example.ID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update example %d: %w", example.ID, gorm.ErrRecordNotFound)
	}
	return nil
}

func (r *exampleRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.Example{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete example %d: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("delete example %d: %w", id, gorm.ErrRecordNotFound)
	}
	return nil
}

var _ ExampleRepository = (*exampleRepository)(nil)
