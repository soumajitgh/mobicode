package entity

import "time"

type Task struct {
	ID        uint   `gorm:"primaryKey"`
	Title     string `gorm:"not null"`
	Status    string `gorm:"not null;default:pending"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
