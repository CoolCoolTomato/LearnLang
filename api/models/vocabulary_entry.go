package models

import "time"

const (
	VocabularySourceImport = "import"
	VocabularySourceChat   = "chat"
	VocabularySourceManual = "manual"
	VocabularySourceSystem = "system"

	VocabularyEntryTypeWord   = "word"
	VocabularyEntryTypePhrase = "phrase"
)

type VocabularyEntry struct {
	ID                   int64                     `gorm:"primaryKey;autoIncrement" json:"id"`
	VocabularyID         int64                     `gorm:"not null;uniqueIndex:idx_vocabulary_entries_identity,priority:1" json:"vocabulary_id"`
	TargetText           string                    `gorm:"type:text;not null" json:"target_text"`
	NormalizedTargetText string                    `gorm:"size:512;not null;uniqueIndex:idx_vocabulary_entries_identity,priority:2" json:"-"`
	TargetLanguage       string                    `gorm:"size:32;not null;uniqueIndex:idx_vocabulary_entries_identity,priority:3" json:"target_language"`
	EntryType            string                    `gorm:"size:16;not null;default:'word';check:entry_type IN ('word','phrase')" json:"entry_type"`
	Tags                 []string                  `gorm:"type:jsonb;serializer:json" json:"tags"`
	Notes                string                    `gorm:"type:text" json:"notes"`
	Source               string                    `gorm:"size:32;not null;check:source IN ('import','chat','manual','system')" json:"source"`
	Encountered          bool                      `gorm:"not null;default:false;index" json:"encountered"`
	EncounterCount       int                       `gorm:"not null;default:0;check:encounter_count >= 0" json:"encounter_count"`
	FirstEncounteredAt   *time.Time                `json:"first_encountered_at"`
	LastEncounteredAt    *time.Time                `json:"last_encountered_at"`
	Pronunciations       []VocabularyPronunciation `gorm:"foreignKey:EntryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"pronunciations,omitempty"`
	Meanings             []VocabularyMeaning       `gorm:"foreignKey:EntryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"meanings,omitempty"`
	Examples             []VocabularyExample       `gorm:"foreignKey:EntryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"examples,omitempty"`
	Relations            []VocabularyEntryRelation `gorm:"foreignKey:EntryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"relations,omitempty"`
	CreatedAt            time.Time                 `json:"created_at"`
	UpdatedAt            time.Time                 `json:"updated_at"`
}

func (VocabularyEntry) TableName() string {
	return "vocabulary_entries"
}
