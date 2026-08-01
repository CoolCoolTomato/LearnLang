package models

import "time"

type ConversationArchive struct {
	ID                 int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID             int64     `gorm:"index;not null" json:"user_id"`
	MessageIDs         []int64   `gorm:"serializer:json;type:jsonb;not null" json:"message_ids"`
	Summary            string    `gorm:"type:text;not null" json:"summary"`
	MessageCount       int       `gorm:"not null" json:"message_count"`
	EmbeddingID        string    `gorm:"size:64;index" json:"embedding_id"`
	EmbeddingDimension int       `gorm:"not null;default:0" json:"embedding_dimension"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (ConversationArchive) TableName() string {
	return "conversation_archives"
}
