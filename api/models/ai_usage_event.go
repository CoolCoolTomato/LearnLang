package models

import "time"

const (
	AIOperationChat        = "chat"
	AIOperationTTS         = "tts"
	AIOperationSTT         = "stt"
	AIOperationEmbedding   = "embedding"
	AIOperationTranslation = "translation"

	AIUsageUnitTokens     = "tokens"
	AIUsageUnitSeconds    = "seconds"
	AIUsageUnitCharacters = "characters"

	AIUsageStatusSucceeded = "succeeded"
	AIUsageStatusFailed    = "failed"
)

// AIUsageEvent records the user-visible usage of one external AI request.
type AIUsageEvent struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"-"`
	UserID    int64     `gorm:"not null;index:idx_ai_usage_events_user_created,priority:1" json:"-"`
	Operation string    `gorm:"size:16;not null;index" json:"operation"`
	Model     string    `gorm:"size:128;not null" json:"model"`
	Usage     float64   `gorm:"not null;default:0;check:usage >= 0" json:"usage"`
	Unit      string    `gorm:"size:16;not null" json:"unit"`
	Status    string    `gorm:"size:16;not null;index" json:"status"`
	CreatedAt time.Time `gorm:"not null;index:idx_ai_usage_events_user_created,priority:2" json:"created_at"`
}

func (AIUsageEvent) TableName() string {
	return "ai_usage_events"
}
