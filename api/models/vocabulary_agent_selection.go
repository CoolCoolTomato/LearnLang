package models

import "time"

const (
	VocabularyAgentSelectionNew = "new"
	VocabularyAgentSelectionOld = "old"
)

type VocabularyAgentSelection struct {
	ID               int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           int64     `gorm:"not null;uniqueIndex:idx_vocabulary_agent_selection_request,priority:1" json:"user_id"`
	RequestMessageID int64     `gorm:"not null;uniqueIndex:idx_vocabulary_agent_selection_request,priority:2" json:"request_message_id"`
	SelectionType    string    `gorm:"size:16;not null;uniqueIndex:idx_vocabulary_agent_selection_request,priority:3;check:selection_type IN ('new','old')" json:"selection_type"`
	RequestedCount   int       `gorm:"not null;check:requested_count >= 1 AND requested_count <= 5" json:"requested_count"`
	EntryIDs         []int64   `gorm:"type:jsonb;serializer:json;not null" json:"entry_ids"`
	User             *User     `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	RequestMessage   *Message  `gorm:"foreignKey:RequestMessageID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	CreatedAt        time.Time `json:"created_at"`
}

func (VocabularyAgentSelection) TableName() string {
	return "vocabulary_agent_selections"
}
