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

const maxAgentRelatedPhrases = 20

type VocabularyRandomEntry struct {
	Vocabulary         models.Vocabulary      `json:"vocabulary"`
	Entry              models.VocabularyEntry `json:"entry"`
	RelatedPhraseCount int64                  `json:"related_phrase_count"`
}

func (s *VocabularyService) TakeRandomNewWord(ctx context.Context, userID int64, targetLanguage, nativeLanguage string) (*VocabularyRandomEntry, error) {
	var result *VocabularyRandomEntry
	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		vocabularies, err := findUserLanguageVocabularies(tx, userID, targetLanguage, nativeLanguage)
		if err != nil || len(vocabularies) == 0 {
			return err
		}

		var entry models.VocabularyEntry
		query := tx.
			Where("vocabulary_id IN ? AND entry_type = ? AND encountered = ?", vocabularyIDs(vocabularies), models.VocabularyEntryTypeWord, false).
			Order("RANDOM()").
			Limit(1).
			Clauses(clause.Locking{
				Strength: "UPDATE",
				Table:    clause.Table{Name: clause.CurrentTable},
				Options:  "SKIP LOCKED",
			})
		operation := query.Find(&entry)
		if operation.Error != nil {
			return operation.Error
		}
		if operation.RowsAffected == 0 {
			return nil
		}
		entryID := entry.ID

		now := time.Now().UTC()
		update := tx.Model(&models.VocabularyEntry{}).
			Where("id = ? AND encountered = ?", entryID, false).
			Updates(map[string]any{
				"encountered":          true,
				"encounter_count":      gorm.Expr("encounter_count + 1"),
				"first_encountered_at": gorm.Expr("COALESCE(first_encountered_at, ?)", now),
				"last_encountered_at":  now,
			})
		if update.Error != nil {
			return update.Error
		}
		if update.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		entry = models.VocabularyEntry{}
		detailQuery := preloadRandomVocabularyEntry(tx.Where("id = ?", entryID))
		if err := detailQuery.First(&entry).Error; err != nil {
			return err
		}
		relatedPhraseCount, err := countVocabularyEntryPhrases(tx, entry.ID)
		if err != nil {
			return err
		}

		vocabulary, ok := vocabularyByID(vocabularies, entry.VocabularyID)
		if !ok {
			return gorm.ErrRecordNotFound
		}
		result = &VocabularyRandomEntry{Vocabulary: vocabulary, Entry: entry, RelatedPhraseCount: relatedPhraseCount}
		return nil
	})
	return result, err
}

func (s *VocabularyService) GetRandomOldWord(ctx context.Context, userID int64, targetLanguage, nativeLanguage string) (*VocabularyRandomEntry, error) {
	vocabularies, err := findUserLanguageVocabularies(database.DB.WithContext(ctx), userID, targetLanguage, nativeLanguage)
	if err != nil || len(vocabularies) == 0 {
		return nil, err
	}

	var entry models.VocabularyEntry
	query := database.DB.WithContext(ctx).
		Where("vocabulary_id IN ? AND entry_type = ? AND encountered = ?", vocabularyIDs(vocabularies), models.VocabularyEntryTypeWord, true).
		Order("RANDOM()").
		Limit(1)
	query = preloadRandomVocabularyEntry(query)
	operation := query.Find(&entry)
	if operation.Error != nil {
		return nil, operation.Error
	}
	if operation.RowsAffected == 0 {
		return nil, nil
	}
	relatedPhraseCount, err := countVocabularyEntryPhrases(database.DB.WithContext(ctx), entry.ID)
	if err != nil {
		return nil, err
	}

	vocabulary, ok := vocabularyByID(vocabularies, entry.VocabularyID)
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &VocabularyRandomEntry{Vocabulary: vocabulary, Entry: entry, RelatedPhraseCount: relatedPhraseCount}, nil
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

func vocabularyByID(vocabularies []models.Vocabulary, vocabularyID int64) (models.Vocabulary, bool) {
	for _, vocabulary := range vocabularies {
		if vocabulary.ID == vocabularyID {
			return vocabulary, true
		}
	}
	return models.Vocabulary{}, false
}
