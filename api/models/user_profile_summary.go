package models

import (
	"time"
)

type UserProfileSummary struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"uniqueIndex;not null" json:"user_id"`
	Summary   string    `gorm:"type:text" json:"summary"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (UserProfileSummary) TableName() string {
	return "user_profile_summaries"
}
