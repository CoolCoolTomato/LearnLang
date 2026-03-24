package models

import (
	"time"
)

type User struct {
	ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Email        *string    `gorm:"uniqueIndex;size:255" json:"email"`
	Phone        *string    `gorm:"uniqueIndex;size:32" json:"phone"`
	Username     string     `gorm:"size:64;not null" json:"username"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	AvatarURL    string     `gorm:"type:text" json:"avatar_url"`
	LastActiveAt *time.Time `json:"last_active_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Role         string     `gorm:"size:16;default:'user'" json:"role"`
}

func (User) TableName() string {
	return "users"
}
