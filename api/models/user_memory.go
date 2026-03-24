package models

import (
	"time"

	"github.com/pgvector/pgvector-go"
)

type UserMemory struct {
	ID              int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          int64           `gorm:"index;not null" json:"user_id"`
	Content         string          `gorm:"type:text" json:"content"`
	Embedding       pgvector.Vector `gorm:"type:vector(1024)" json:"embedding"`
	MemoryType      string          `gorm:"size:32" json:"memory_type"`
	ImportanceScore float64         `json:"importance_score"`
	CreatedAt       time.Time       `json:"created_at"`
}

func (UserMemory) TableName() string {
	return "user_memories"
}
