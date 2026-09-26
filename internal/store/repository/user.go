package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/soumajitgh/mobicode/internal/store/model"
)

var ErrDuplicateEmail = errors.New("duplicate email")

type UserRepository interface {
	Create(context.Context, *model.User) error
	FindByEmail(context.Context, string) (*model.User, error)
	FindByID(context.Context, uint) (*model.User, error)
	ReplacePassword(context.Context, uint, string) error
	Count(context.Context) (int64, error)
	FindFirst(context.Context) (*model.User, error)
}

type userRepository struct{ db *gorm.DB }

func NewUser(db *gorm.DB) UserRepository { return &userRepository{db} }
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
			return ErrDuplicateEmail
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) ReplacePassword(ctx context.Context, id uint, hash string) error {
	result := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(map[string]interface{}{"password_hash": hash, "session_version": gorm.Expr("session_version + 1"), "updated_at": gorm.Expr("CURRENT_TIMESTAMP")})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return count, nil
}

func (r *userRepository) FindFirst(ctx context.Context) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Order("id ASC").First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
