package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

var (
	ErrInvalidPairing = errors.New("invalid pairing token")
	ErrExpiredPairing = errors.New("pairing expired")
	ErrClaimedPairing = errors.New("pairing already claimed")
)

type MobileDevice struct {
	ID         string
	UserID     uint
	Name       string
	Platform   string
	CreatedAt  time.Time
	LastSeenAt sql.NullTime
}

type MobileIdentity struct {
	UserID uint
	Email  string
}

type MobileRepository interface {
	CreatePairing(context.Context, string, uint, string, time.Time, time.Time) error
	ClaimPairing(context.Context, string, string, string, string, string, string, time.Time) (MobileIdentity, MobileDevice, error)
	FindMobileIdentity(context.Context, string) (*MobileIdentity, error)
}

type mobileRepository struct{ db *gorm.DB }

func NewMobile(db *gorm.DB) MobileRepository { return &mobileRepository{db: db} }

func (r *mobileRepository) CreatePairing(ctx context.Context, id string, userID uint, tokenHash string, now, expires time.Time) error {
	result := r.db.WithContext(ctx).Exec(`INSERT INTO device_pairings (id,user_id,token_hash,created_at,expires_at) VALUES (?,?,?,?,?)`, id, userID, tokenHash, now, expires)
	if result.Error != nil {
		return fmt.Errorf("create device pairing: %w", result.Error)
	}
	return nil
}

// ClaimPairing creates the mobile device and session while atomically consuming the pairing.
func (r *mobileRepository) ClaimPairing(ctx context.Context, tokenHash, deviceID, name, platform, sessionID, sessionHash string, now time.Time) (MobileIdentity, MobileDevice, error) {
	var identity MobileIdentity
	var device MobileDevice
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var pairingID string
		var expires time.Time
		var consumed sql.NullTime
		err := tx.Raw(`SELECT p.id,p.user_id,u.email,p.expires_at,p.consumed_at FROM device_pairings p JOIN users u ON u.id=p.user_id WHERE p.token_hash=?`, tokenHash).Row().Scan(&pairingID, &identity.UserID, &identity.Email, &expires, &consumed)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidPairing
		}
		if err != nil {
			return fmt.Errorf("find device pairing: %w", err)
		}
		if consumed.Valid {
			return ErrClaimedPairing
		}
		if !now.Before(expires) {
			return ErrExpiredPairing
		}

		device = MobileDevice{ID: deviceID, UserID: identity.UserID, Name: name, Platform: platform, CreatedAt: now}
		if err := tx.Exec(`INSERT INTO mobile_devices (id,user_id,name,platform,created_at) VALUES (?,?,?,?,?)`, deviceID, identity.UserID, name, platform, now).Error; err != nil {
			return fmt.Errorf("create mobile device: %w", err)
		}
		if err := tx.Exec(`INSERT INTO mobile_sessions (id,device_id,token_hash,created_at) VALUES (?,?,?,?)`, sessionID, deviceID, sessionHash, now).Error; err != nil {
			return fmt.Errorf("create mobile session: %w", err)
		}
		result := tx.Exec(`UPDATE device_pairings SET consumed_at=? WHERE id=? AND consumed_at IS NULL AND expires_at>?`, now, pairingID, now)
		if result.Error != nil {
			return fmt.Errorf("consume device pairing: %w", result.Error)
		}
		if result.RowsAffected != 1 {
			return ErrClaimedPairing
		}
		return nil
	})
	if err != nil {
		return MobileIdentity{}, MobileDevice{}, err
	}
	return identity, device, nil
}

func (r *mobileRepository) FindMobileIdentity(ctx context.Context, tokenHash string) (*MobileIdentity, error) {
	var identity MobileIdentity
	err := r.db.WithContext(ctx).Raw(`SELECT u.id,u.email FROM mobile_sessions s JOIN mobile_devices d ON d.id=s.device_id JOIN users u ON u.id=d.user_id WHERE s.token_hash=? AND s.revoked_at IS NULL`, tokenHash).Row().Scan(&identity.UserID, &identity.Email)
	if err != nil {
		return nil, err
	}
	return &identity, nil
}
