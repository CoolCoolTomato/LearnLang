package models

import "time"

type VocabularyMeaning struct {
	ID             int64               `gorm:"primaryKey;autoIncrement" json:"id"`
	EntryID        int64               `gorm:"not null;index" json:"entry_id"`
	NativeText     string              `gorm:"type:text;not null" json:"native_text"`
	NativeLanguage string              `gorm:"size:32;not null" json:"native_language"`
	PartOfSpeech   string              `gorm:"size:64" json:"part_of_speech"`
	SortOrder      int                 `gorm:"not null;default:0" json:"sort_order"`
	Examples       []VocabularyExample `gorm:"foreignKey:MeaningID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"examples,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

func (VocabularyMeaning) TableName() string {
	return "vocabulary_meanings"
}
