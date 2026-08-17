package user

import "time"

// User is the persisted user record.
type User struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
