package models

import (
	"time"
)

type Message struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      int64      `gorm:"index:idx_messages_user_created,priority:1;not null" json:"user_id"`
	Role        string     `gorm:"size:16;not null" json:"role"`
	TextContent string     `gorm:"type:text" json:"text_content"`
	Translation string     `gorm:"type:text" json:"translation"`
	VoiceFileID *int64     `gorm:"index" json:"voice_file_id"`
	VoiceFile   *VoiceFile `gorm:"foreignKey:VoiceFileID" json:"voice_file,omitempty"`
	InputType   string     `gorm:"size:16" json:"input_type"`
	TokenCount  int        `json:"token_count"`
	CreatedAt   time.Time  `gorm:"index:idx_messages_user_created,priority:2" json:"created_at"`
}

func (Message) TableName() string {
	return "messages"
}
