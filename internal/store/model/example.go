package model

import "time"

// Example demonstrates the shape of a persistence model. No table is created for it yet.
type Example struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
