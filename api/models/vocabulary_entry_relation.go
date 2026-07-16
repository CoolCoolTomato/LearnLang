package models

import "time"

const (
	VocabularyRelationPhrase      = "phrase"
	VocabularyRelationCollocation = "collocation"
	VocabularyRelationDerived     = "derived"
)

type VocabularyEntryRelation struct {
	ID             int64            `gorm:"primaryKey;autoIncrement" json:"id"`
	EntryID        int64            `gorm:"not null;uniqueIndex:idx_vocabulary_entry_relation,priority:1" json:"entry_id"`
	RelatedEntryID int64            `gorm:"not null;uniqueIndex:idx_vocabulary_entry_relation,priority:2;index" json:"related_entry_id"`
	RelationType   string           `gorm:"size:32;not null;uniqueIndex:idx_vocabulary_entry_relation,priority:3" json:"relation_type"`
	SortOrder      int              `gorm:"not null;default:0" json:"sort_order"`
	RelatedEntry   *VocabularyEntry `gorm:"foreignKey:RelatedEntryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"related_entry,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

func (VocabularyEntryRelation) TableName() string {
	return "vocabulary_entry_relations"
}
