package models

import (
	"time"
)

type ConversationSummary struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"uniqueIndex;not null" json:"user_id"`
	Summary   string    `gorm:"type:text" json:"summary"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ConversationSummary) TableName() string {
	return "conversation_summaries"
}
