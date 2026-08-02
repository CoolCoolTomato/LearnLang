package models

import "time"

type Vocabulary struct {
	ID             int64             `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         int64             `gorm:"not null;uniqueIndex:idx_vocabularies_user_name,priority:1;index" json:"user_id"`
	Name           string            `gorm:"size:128;not null;uniqueIndex:idx_vocabularies_user_name,priority:2" json:"name"`
	TargetLanguage string            `gorm:"size:32;not null" json:"target_language"`
	NativeLanguage string            `gorm:"size:32;not null" json:"native_language"`
	User           *User             `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	Entries        []VocabularyEntry `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"entries,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

func (Vocabulary) TableName() string {
	return "vocabularies"
}
