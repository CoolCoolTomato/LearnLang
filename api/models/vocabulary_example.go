package models

import "time"

type VocabularyExample struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	EntryID    int64     `gorm:"not null;index" json:"entry_id"`
	MeaningID  *int64    `gorm:"index" json:"meaning_id,omitempty"`
	TargetText string    `gorm:"type:text;not null" json:"target_text"`
	NativeText string    `gorm:"type:text" json:"native_text"`
	Source     string    `gorm:"size:32;not null;check:source IN ('import','chat','manual','system')" json:"source"`
	SortOrder  int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (VocabularyExample) TableName() string {
	return "vocabulary_examples"
}
