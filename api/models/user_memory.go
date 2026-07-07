package models

import (
	"time"

	"github.com/pgvector/pgvector-go"
)

type UserMemory struct {
	ID              int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          int64           `gorm:"index;not null" json:"user_id"`
	Summary         string          `gorm:"type:text;column:summary" json:"summary"`
	Embedding       pgvector.Vector `gorm:"type:vector(1024)" json:"embedding"`
	MemoryType      string          `gorm:"size:32" json:"memory_type"`
	ImportanceScore float64         `json:"importance_score"`
	MessageCount    int             `json:"message_count"`
	StartedAt       *time.Time      `gorm:"index" json:"started_at"`
	EndedAt         *time.Time      `gorm:"index" json:"ended_at"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

func (UserMemory) TableName() string {
	return "user_memories"
}

type UserMemoryMessage struct {
	ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserMemoryID int64      `gorm:"uniqueIndex:idx_user_memory_message;not null" json:"user_memory_id"`
	UserMemory   UserMemory `gorm:"foreignKey:UserMemoryID" json:"-"`
	MessageID    int64      `gorm:"uniqueIndex:idx_user_memory_message;index;not null" json:"message_id"`
	Message      Message    `gorm:"foreignKey:MessageID" json:"message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (UserMemoryMessage) TableName() string {
	return "user_memory_messages"
}
