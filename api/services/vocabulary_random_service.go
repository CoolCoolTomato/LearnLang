package services

import (
	"context"
	"learnlang-api/database"
	"learnlang-api/models"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	maxAgentRelatedPhrases  = 20
	MaxAgentVocabularyWords = 5
)

type VocabularyRandomEntry struct {
	Vocabulary         models.Vocabulary      `json:"vocabulary"`
	Entry              models.VocabularyEntry `json:"entry"`
	RelatedPhraseCount int64                  `json:"related_phrase_count"`
}

type VocabularyAgentSelectionResult struct {
	RequestedCount int
	Entries        []VocabularyRandomEntry
}

func (s *VocabularyService) SelectAgentVocabularyWords(ctx context.Context, userID, requestMessageID int64, selectionType, targetLanguage, nativeLanguage string, count int) (*VocabularyAgentSelectionResult, error) {
	if requestMessageID <= 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if selectionType != models.VocabularyAgentSelectionNew && selectionType != models.VocabularyAgentSelectionOld {
		return nil, gorm.ErrInvalidData
	}
	if count < 1 {
		count = 1
	}
	if count > MaxAgentVocabularyWords {
		count = MaxAgentVocabularyWords
	}

	result := &VocabularyAgentSelectionResult{}
	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// The request message is the transaction lock shared by interrupted or concurrent retries.
		var message models.Message
		messageQuery := tx.Select("id").
			Where("id = ? AND user_id = ?", requestMessageID, userID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Find(&message)
		if messageQuery.Error != nil {
			return messageQuery.Error
		}
		if messageQuery.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		var selection models.VocabularyAgentSelection
		selectionQuery := tx.Where("user_id = ? AND request_message_id = ? AND selection_type = ?", userID, requestMessageID, selectionType).
			Limit(1).
			Find(&selection)
		if selectionQuery.Error != nil {
			return selectionQuery.Error
		}
		if selectionQuery.RowsAffected > 0 {
			result.RequestedCount = selection.RequestedCount
			var loadErr error
			result.Entries, loadErr = loadVocabularyRandomEntries(tx, userID, selection.EntryIDs)
			return loadErr
		}

		vocabularies, err := findUserLanguageVocabularies(tx, userID, targetLanguage, nativeLanguage)
		if err != nil {
			return err
		}
		entryIDs := make([]int64, 0, count)
		if len(vocabularies) > 0 {
			encountered := selectionType == models.VocabularyAgentSelectionOld
			var entries []models.VocabularyEntry
			query := tx.Select("id").
				Where("vocabulary_id IN ? AND entry_type = ? AND encountered = ?", vocabularyIDs(vocabularies), models.VocabularyEntryTypeWord, encountered).
				Order("RANDOM()").
				Limit(count)
			if selectionType == models.VocabularyAgentSelectionNew {
				query = query.Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: clause.CurrentTable}, Options: "SKIP LOCKED"})
			}
			if err := query.Find(&entries).Error; err != nil {
				return err
			}
			for _, entry := range entries {
				entryIDs = append(entryIDs, entry.ID)
			}
		}

		if selectionType == models.VocabularyAgentSelectionNew && len(entryIDs) > 0 {
			now := time.Now().UTC()
			update := tx.Model(&models.VocabularyEntry{}).
				Where("id IN ? AND encountered = ?", entryIDs, false).
				Updates(map[string]any{
					"encountered":          true,
					"encounter_count":      gorm.Expr("encounter_count + 1"),
					"first_encountered_at": gorm.Expr("COALESCE(first_encountered_at, ?)", now),
					"last_encountered_at":  now,
				})
			if update.Error != nil {
				return update.Error
			}
			if update.RowsAffected != int64(len(entryIDs)) {
				return gorm.ErrRecordNotFound
			}
		}

		selection = models.VocabularyAgentSelection{
			UserID:           userID,
			RequestMessageID: requestMessageID,
			SelectionType:    selectionType,
			RequestedCount:   count,
			EntryIDs:         entryIDs,
		}
		if err := tx.Create(&selection).Error; err != nil {
			return err
		}
		result.RequestedCount = selection.RequestedCount
		result.Entries, err = loadVocabularyRandomEntries(tx, userID, entryIDs)
		return err
	})
	return result, err
}

func loadVocabularyRandomEntries(tx *gorm.DB, userID int64, entryIDs []int64) ([]VocabularyRandomEntry, error) {
	results := make([]VocabularyRandomEntry, 0, len(entryIDs))
	for _, entryID := range entryIDs {
		var entry models.VocabularyEntry
		entryQuery := preloadRandomVocabularyEntry(tx.Where("vocabulary_entries.id = ?", entryID)).
			Joins("JOIN vocabularies ON vocabularies.id = vocabulary_entries.vocabulary_id AND vocabularies.user_id = ?", userID).
			Limit(1).
			Find(&entry)
		if entryQuery.Error != nil {
			return nil, entryQuery.Error
		}
		if entryQuery.RowsAffected == 0 {
			continue
		}
		var vocabulary models.Vocabulary
		if err := tx.First(&vocabulary, "id = ? AND user_id = ?", entry.VocabularyID, userID).Error; err != nil {
			return nil, err
		}
		relatedPhraseCount, err := countVocabularyEntryPhrases(tx, entry.ID)
		if err != nil {
			return nil, err
		}
		results = append(results, VocabularyRandomEntry{Vocabulary: vocabulary, Entry: entry, RelatedPhraseCount: relatedPhraseCount})
	}
	return results, nil
}

func preloadRandomVocabularyEntry(query *gorm.DB) *gorm.DB {
	return query.
		Preload("Pronunciations", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Preload("Meanings", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Preload("Examples", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Preload("Relations", func(db *gorm.DB) *gorm.DB {
			return db.Where("relation_type = ?", models.VocabularyRelationPhrase).
				Order("sort_order ASC, id ASC").
				Limit(maxAgentRelatedPhrases)
		}).
		Preload("Relations.RelatedEntry.Meanings", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") })
}

func countVocabularyEntryPhrases(db *gorm.DB, entryID int64) (int64, error) {
	var count int64
	err := db.Model(&models.VocabularyEntryRelation{}).
		Where("entry_id = ? AND relation_type = ?", entryID, models.VocabularyRelationPhrase).
		Count(&count).Error
	return count, err
}

func findUserLanguageVocabularies(db *gorm.DB, userID int64, targetLanguage, nativeLanguage string) ([]models.Vocabulary, error) {
	targetLanguage = strings.TrimSpace(targetLanguage)
	nativeLanguage = strings.TrimSpace(nativeLanguage)
	query := db.Where("user_id = ?", userID)
	if targetLanguage != "" {
		query = query.Where("LOWER(target_language) = LOWER(?)", targetLanguage)
	}
	if nativeLanguage != "" {
		query = query.Where("LOWER(native_language) = LOWER(?)", nativeLanguage)
	}
	var vocabularies []models.Vocabulary
	if err := query.Order("is_default DESC, id ASC").Find(&vocabularies).Error; err != nil {
		return nil, err
	}
	return vocabularies, nil
}

func vocabularyIDs(vocabularies []models.Vocabulary) []int64 {
	ids := make([]int64, 0, len(vocabularies))
	for _, vocabulary := range vocabularies {
		ids = append(ids, vocabulary.ID)
	}
	return ids
}
