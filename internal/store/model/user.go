package model

import "time"

type User struct {
	ID             uint   `gorm:"primaryKey"`
	Email          string `gorm:"not null;uniqueIndex"`
	PasswordHash   string `gorm:"not null"`
	SessionVersion int64  `gorm:"not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
