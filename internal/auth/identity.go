package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	ErrSetupRequired = errors.New("server setup required")
	ErrForbidden     = errors.New("authenticated key is not the owner")
	ErrReplay        = errors.New("authentication proof already used")
)

// OwnerIdentity is the sole principal allowed to use a MobiCode deployment.
type OwnerIdentity struct {
	ID        int       `gorm:"primaryKey"`
	PublicKey string    `gorm:"uniqueIndex"`
	CreatedAt time.Time `json:"createdAt"`
}

// TableName matches the singular production table created by the SQL migration.
func (OwnerIdentity) TableName() string { return "owner_identity" }

type AuthReplay struct {
	EventID   string    `gorm:"primaryKey"`
	ExpiresAt time.Time `gorm:"index"`
	CreatedAt time.Time
}

type OwnerRepository struct{ db *gorm.DB }

func NewOwnerRepository(db *gorm.DB) *OwnerRepository { return &OwnerRepository{db: db} }

func (r *OwnerRepository) Get(ctx context.Context) (*OwnerIdentity, error) {
	var owner OwnerIdentity
	err := r.db.WithContext(ctx).First(&owner, "id = 1").Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSetupRequired
	}
	if err != nil {
		return nil, fmt.Errorf("read owner identity: %w", err)
	}
	return &owner, nil
}

func (r *OwnerRepository) Set(ctx context.Context, publicKey string) error {
	publicKey = strings.ToLower(strings.TrimSpace(publicKey))
	if !validPublicKey(publicKey) {
		return fmt.Errorf("invalid owner public key")
	}
	return r.db.WithContext(ctx).Create(&OwnerIdentity{ID: 1, PublicKey: publicKey}).Error
}

func (r *OwnerRepository) Reset(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("id = 1").Delete(&OwnerIdentity{}).Error
}

func (r *OwnerRepository) ClaimReplay(ctx context.Context, eventID string, expiresAt time.Time) error {
	if err := r.db.WithContext(ctx).Where("expires_at <= ?", time.Now()).Delete(&AuthReplay{}).Error; err != nil {
		return fmt.Errorf("prune auth replays: %w", err)
	}
	if err := r.db.WithContext(ctx).Create(&AuthReplay{EventID: eventID, ExpiresAt: expiresAt}).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(strings.ToLower(err.Error()), "unique constraint") {
			return ErrReplay
		}
		return fmt.Errorf("store auth replay: %w", err)
	}
	return nil
}

// OwnerService owns the configured single identity and proof replay records.
type OwnerService struct{ repo *OwnerRepository }

func NewOwnerService(repo *OwnerRepository) *OwnerService { return &OwnerService{repo: repo} }

func (s *OwnerService) Owner(ctx context.Context) (*OwnerIdentity, error) { return s.repo.Get(ctx) }
func (s *OwnerService) Configure(ctx context.Context, publicKey string) error {
	return s.repo.Set(ctx, publicKey)
}
func (s *OwnerService) Reset(ctx context.Context) error { return s.repo.Reset(ctx) }
func (s *OwnerService) ClaimReplay(ctx context.Context, eventID string, expiresAt time.Time) error {
	return s.repo.ClaimReplay(ctx, eventID, expiresAt)
}
