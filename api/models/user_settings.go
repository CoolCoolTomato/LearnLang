package models

import (
	"time"
)

type UserSettings struct {
	ID                  int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID              int64     `gorm:"uniqueIndex;not null" json:"user_id"`
	APIBaseURL          string    `gorm:"size:255" json:"api_base_url"`
	APIKey              string    `gorm:"size:255" json:"api_key"`
	Model               string    `gorm:"size:128" json:"model"`
	EmbeddingAPIBaseURL string    `gorm:"size:255" json:"embedding_api_base_url"`
	EmbeddingAPIKey     string    `gorm:"size:255" json:"embedding_api_key"`
	EmbeddingModel      string    `gorm:"size:128" json:"embedding_model"`
	STTAPIBaseURL       string    `gorm:"size:255" json:"stt_api_base_url"`
	STTAPIKey           string    `gorm:"size:255" json:"stt_api_key"`
	STTModel            string    `gorm:"size:128" json:"stt_model"`
	TTSAPIBaseURL       string    `gorm:"size:255" json:"tts_api_base_url"`
	TTSAPIKey           string    `gorm:"size:255" json:"tts_api_key"`
	TTSModel            string    `gorm:"size:128" json:"tts_model"`
	TTSVoice            string    `gorm:"size:128" json:"tts_voice"`
	Theme               string    `gorm:"size:32;default:'system'" json:"theme"`
	Language            string    `gorm:"size:32" json:"language"`
	NativeLanguage      string    `gorm:"size:32" json:"native_language"`
	TargetLanguage      string    `gorm:"size:32" json:"target_language"`
	Timezone            string    `gorm:"size:64;default:'UTC'" json:"timezone"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func (UserSettings) TableName() string {
	return "user_settings"
}
