package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Repository persists opaque refresh tokens.
type Repository interface {
	Create(ctx context.Context, token *RefreshToken) error
	FindActiveByHash(ctx context.Context, hash string) (*RefreshToken, error)
	Rotate(ctx context.Context, currentHash string, replacement *RefreshToken) error
	RevokeByHash(ctx context.Context, hash string) error
}

type gormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &gormRepository{db: db} }

func (r *gormRepository) Create(ctx context.Context, token *RefreshToken) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}
	return nil
}

func (r *gormRepository) FindActiveByHash(ctx context.Context, hash string) (*RefreshToken, error) {
	var token RefreshToken
	err := r.db.WithContext(ctx).Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", hash, time.Now()).First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidRefreshToken
	}
	if err != nil {
		return nil, fmt.Errorf("find refresh token: %w", err)
	}
	return &token, nil
}

func (r *gormRepository) Rotate(ctx context.Context, currentHash string, replacement *RefreshToken) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		result := tx.Model(&RefreshToken{}).Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", currentHash, now).Update("revoked_at", now)
		if result.Error != nil {
			return fmt.Errorf("revoke refresh token: %w", result.Error)
		}
		if result.RowsAffected != 1 {
			return ErrInvalidRefreshToken
		}
		if err := tx.Create(replacement).Error; err != nil {
			return fmt.Errorf("create rotated refresh token: %w", err)
		}
		return nil
	})
}

func (r *gormRepository) RevokeByHash(ctx context.Context, hash string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&RefreshToken{}).Where("token_hash = ? AND revoked_at IS NULL", hash).Update("revoked_at", now)
	if result.Error != nil {
		return fmt.Errorf("revoke refresh token: %w", result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrInvalidRefreshToken
	}
	return nil
}
