package setup

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

const (
	stateWaiting  = "waiting"
	statePending  = "pending_confirmation"
	stateComplete = "complete"
)

type Session struct {
	ID                 string `gorm:"primaryKey"`
	BrowserTokenHash   string `gorm:"uniqueIndex"`
	PairingTokenHash   string `gorm:"uniqueIndex"`
	CSRFTokenHash      string
	CandidatePublicKey string
	State              string
	ExpiresAt          time.Time `gorm:"index"`
	CreatedAt          time.Time
}

func (Session) TableName() string { return "setup_sessions" }

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Active(ctx context.Context) (*Session, error) {
	var session Session
	err := r.db.WithContext(ctx).Where("expires_at > ?", time.Now()).Order("created_at ASC").First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find setup session: %w", err)
	}
	return &session, nil
}

func (r *Repository) Create(ctx context.Context, session *Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *Repository) RefreshTokens(ctx context.Context, id, pairingHash, csrfHash string) error {
	return r.db.WithContext(ctx).Model(&Session{}).Where("id = ? AND state = ? AND expires_at > ?", id, stateWaiting, time.Now()).
		Updates(map[string]any{"pairing_token_hash": pairingHash, "csrf_token_hash": csrfHash}).Error
}

func (r *Repository) ByBrowserToken(ctx context.Context, hash string) (*Session, error) {
	var session Session
	err := r.db.WithContext(ctx).Where("browser_token_hash = ? AND expires_at > ?", hash, time.Now()).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find browser setup session: %w", err)
	}
	return &session, nil
}

func (r *Repository) Pair(ctx context.Context, tokenHash, publicKey string) error {
	result := r.db.WithContext(ctx).Model(&Session{}).
		Where("pairing_token_hash = ? AND state = ? AND expires_at > ?", tokenHash, stateWaiting, time.Now()).
		Updates(map[string]any{"candidate_public_key": publicKey, "state": statePending})
	if result.Error != nil {
		return fmt.Errorf("store pairing candidate: %w", result.Error)
	}
	if result.RowsAffected != 1 {
		return errors.New("pairing session is unavailable")
	}
	return nil
}

func (r *Repository) PairStatus(ctx context.Context, tokenHash, publicKey string) (string, error) {
	var session Session
	err := r.db.WithContext(ctx).Where("pairing_token_hash = ? AND candidate_public_key = ? AND expires_at > ?", tokenHash, publicKey, time.Now()).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", errors.New("pairing session is unavailable")
	}
	if err != nil {
		return "", fmt.Errorf("find pairing status: %w", err)
	}
	return session.State, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&Session{}).Error
}

func (r *Repository) Complete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&Session{}).Where("id = ?", id).Update("state", stateComplete).Error
}

func (r *Repository) DeleteAll(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("1 = 1").Delete(&Session{}).Error
}
