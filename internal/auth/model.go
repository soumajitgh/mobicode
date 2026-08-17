// Package auth provides password and token authentication.
package auth

import "time"

// RefreshToken is a per-device, opaque refresh-token record.
type RefreshToken struct {
	ID        string `gorm:"primaryKey"`
	UserID    string `gorm:"index"`
	TokenHash string `gorm:"uniqueIndex"`
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func (RefreshToken) TableName() string { return "refresh_tokens" }
