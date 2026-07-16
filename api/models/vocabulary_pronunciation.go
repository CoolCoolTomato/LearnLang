package models

import "time"

type VocabularyPronunciation struct {
	ID                int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	EntryID           int64     `gorm:"not null;uniqueIndex:idx_vocabulary_pronunciation_variant,priority:1" json:"entry_id"`
	Pronunciation     string    `gorm:"type:text;not null" json:"pronunciation"`
	PronunciationType string    `gorm:"size:32;not null;uniqueIndex:idx_vocabulary_pronunciation_variant,priority:2" json:"pronunciation_type"`
	Region            string    `gorm:"size:32;not null;default:'';uniqueIndex:idx_vocabulary_pronunciation_variant,priority:3" json:"region"`
	AudioURL          string    `gorm:"type:text" json:"audio_url"`
	SortOrder         int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (VocabularyPronunciation) TableName() string {
	return "vocabulary_pronunciations"
}
