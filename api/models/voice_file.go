package models

import "time"

type VoiceFile struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"index;not null" json:"user_id"`
	VoiceRole string    `gorm:"size:64" json:"voice_role"`
	VoiceURL  string    `gorm:"size:512;not null" json:"voice_url"`
	Duration  int       `gorm:"default:0" json:"duration"`
	FileSize  int64     `gorm:"default:0" json:"file_size"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (VoiceFile) TableName() string {
	return "voice_files"
}
