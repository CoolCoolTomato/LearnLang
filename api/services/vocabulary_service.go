package services

import (
	"context"
	"errors"
	"fmt"
	"learnlang-api/database"
	"learnlang-api/models"
	"strings"

	"golang.org/x/text/unicode/norm"
	"gorm.io/gorm"
)

const (
	maxVocabularyImportEntries = 100000
	vocabularyAdvisoryLockBase = int64(740000000000000000)
)

var (
	ErrVocabularyInvalidImport    = errors.New("invalid vocabulary import")
	ErrVocabularyInvalidInput     = errors.New("invalid vocabulary input")
	ErrVocabularyNotFound         = errors.New("vocabulary not found")
	ErrVocabularyNameConflict     = errors.New("a vocabulary with this name already exists")
	ErrVocabularyLanguageRequired = errors.New("target language and native language are required")
	ErrVocabularyDefaultRequired  = errors.New("a default vocabulary is required")
)

type VocabularyService struct {
	userSettingsService *UserSettingsService
}

type VocabularyCreateInput struct {
	Name           string `json:"name"`
	TargetLanguage string `json:"target_language"`
	NativeLanguage string `json:"native_language"`
	IsDefault      bool   `json:"is_default"`
}

type VocabularyUpdateInput struct {
	Name           *string `json:"name"`
	TargetLanguage *string `json:"target_language"`
	NativeLanguage *string `json:"native_language"`
	IsDefault      *bool   `json:"is_default"`
}

type VocabularyImportResult struct {
	Vocabulary            models.Vocabulary `json:"vocabulary"`
	EntriesCreated        int               `json:"entries_created"`
	EntriesUpdated        int               `json:"entries_updated"`
	MeaningsCreated       int               `json:"meanings_created"`
	PronunciationsCreated int               `json:"pronunciations_created"`
	ExamplesCreated       int               `json:"examples_created"`
	RelationsCreated      int               `json:"relations_created"`
}

type VocabularyEntryPage struct {
	Vocabulary *models.Vocabulary       `json:"vocabulary"`
	Data       []models.VocabularyEntry `json:"data"`
	Total      int64                    `json:"total"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
}

type VocabularySummary struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	TargetLanguage string `json:"target_language"`
	NativeLanguage string `json:"native_language"`
	IsDefault      bool   `json:"is_default"`
	EntryCount     int64  `json:"entry_count"`
}

func NewVocabularyService(userSettingsService *UserSettingsService) *VocabularyService {
	return &VocabularyService{userSettingsService: userSettingsService}
}

func (s *VocabularyService) List(ctx context.Context, userID int64) ([]VocabularySummary, error) {
	vocabularies := make([]VocabularySummary, 0)
	err := database.DB.WithContext(ctx).
		Table("vocabularies").
		Select("vocabularies.id, vocabularies.name, vocabularies.target_language, vocabularies.native_language, vocabularies.is_default, COUNT(vocabulary_entries.id) AS entry_count").
		Joins("LEFT JOIN vocabulary_entries ON vocabulary_entries.vocabulary_id = vocabularies.id").
		Where("vocabularies.user_id = ?", userID).
		Group("vocabularies.id").
		Order("vocabularies.is_default DESC, vocabularies.updated_at DESC, vocabularies.id DESC").
		Scan(&vocabularies).Error
	return vocabularies, err
}

func (s *VocabularyService) Create(ctx context.Context, userID int64, input VocabularyCreateInput) (*models.Vocabulary, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" || len([]rune(name)) > 128 {
		return nil, fmt.Errorf("%w: name is required and must not exceed 128 characters", ErrVocabularyInvalidInput)
	}
	targetLanguage, nativeLanguage, err := s.resolveLanguages(userID, input.TargetLanguage, input.NativeLanguage)
	if err != nil {
		return nil, err
	}
	if err := validateVocabularyInfo(name, targetLanguage, nativeLanguage); err != nil {
		return nil, err
	}

	var vocabulary models.Vocabulary
	err = database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockUserVocabulary(tx, userID); err != nil {
			return err
		}
		if err := ensureVocabularyNameAvailable(tx, userID, name, 0); err != nil {
			return err
		}

		var defaultCount int64
		if err := tx.Model(&models.Vocabulary{}).Where("user_id = ? AND is_default = ?", userID, true).Count(&defaultCount).Error; err != nil {
			return err
		}
		isDefault := input.IsDefault || defaultCount == 0
		if isDefault {
			if err := tx.Model(&models.Vocabulary{}).Where("user_id = ? AND is_default = ?", userID, true).Update("is_default", false).Error; err != nil {
				return err
			}
		}

		vocabulary = models.Vocabulary{
			UserID:         userID,
			Name:           name,
			TargetLanguage: targetLanguage,
			NativeLanguage: nativeLanguage,
			IsDefault:      isDefault,
		}
		return tx.Create(&vocabulary).Error
	})
	if err != nil {
		return nil, err
	}
	return &vocabulary, nil
}

func (s *VocabularyService) Update(ctx context.Context, userID, vocabularyID int64, input VocabularyUpdateInput) (*models.Vocabulary, error) {
	var vocabulary *models.Vocabulary
	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockUserVocabulary(tx, userID); err != nil {
			return err
		}
		current, err := findOwnedVocabulary(tx, userID, vocabularyID)
		if err != nil {
			return err
		}

		name := current.Name
		if input.Name != nil {
			name = strings.TrimSpace(*input.Name)
		}
		targetLanguage := current.TargetLanguage
		if input.TargetLanguage != nil {
			targetLanguage = strings.TrimSpace(*input.TargetLanguage)
		}
		nativeLanguage := current.NativeLanguage
		if input.NativeLanguage != nil {
			nativeLanguage = strings.TrimSpace(*input.NativeLanguage)
		}
		if err := validateVocabularyInfo(name, targetLanguage, nativeLanguage); err != nil {
			return err
		}
		if err := ensureVocabularyNameAvailable(tx, userID, name, vocabularyID); err != nil {
			return err
		}

		if targetLanguage != current.TargetLanguage {
			if err := tx.Model(&models.VocabularyEntry{}).
				Where("vocabulary_id = ?", vocabularyID).
				Update("target_language", targetLanguage).Error; err != nil {
				return err
			}
		}
		if nativeLanguage != current.NativeLanguage {
			entryIDs := tx.Model(&models.VocabularyEntry{}).
				Select("id").
				Where("vocabulary_id = ?", vocabularyID)
			if err := tx.Model(&models.VocabularyMeaning{}).
				Where("entry_id IN (?)", entryIDs).
				Update("native_language", nativeLanguage).Error; err != nil {
				return err
			}
		}

		isDefault := current.IsDefault
		if input.IsDefault != nil {
			if !*input.IsDefault && current.IsDefault {
				return ErrVocabularyDefaultRequired
			}
			if *input.IsDefault && !current.IsDefault {
				if err := tx.Model(&models.Vocabulary{}).Where("user_id = ? AND is_default = ?", userID, true).Update("is_default", false).Error; err != nil {
					return err
				}
				isDefault = true
			}
		}

		if err := tx.Model(current).Updates(map[string]any{
			"name":            name,
			"target_language": targetLanguage,
			"native_language": nativeLanguage,
			"is_default":      isDefault,
		}).Error; err != nil {
			return err
		}
		current.Name = name
		current.TargetLanguage = targetLanguage
		current.NativeLanguage = nativeLanguage
		current.IsDefault = isDefault
		vocabulary = current
		return nil
	})
	if err != nil {
		return nil, err
	}
	return vocabulary, nil
}

func (s *VocabularyService) Delete(ctx context.Context, userID, vocabularyID int64) error {
	return database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockUserVocabulary(tx, userID); err != nil {
			return err
		}
		vocabulary, err := findOwnedVocabulary(tx, userID, vocabularyID)
		if err != nil {
			return err
		}
		if err := tx.Delete(vocabulary).Error; err != nil {
			return err
		}
		if vocabulary.IsDefault {
			var replacement models.Vocabulary
			operation := tx.Where("user_id = ?", userID).Order("created_at ASC, id ASC").Limit(1).Find(&replacement)
			if operation.Error != nil {
				return operation.Error
			}
			if operation.RowsAffected > 0 {
				return tx.Model(&replacement).Update("is_default", true).Error
			}
		}
		return nil
	})
}

func (s *VocabularyService) Import(ctx context.Context, userID, vocabularyID int64, input VocabularyImportInput) (*VocabularyImportResult, error) {
	if err := validateVocabularyImport(input); err != nil {
		return nil, err
	}

	result := &VocabularyImportResult{}
	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockUserVocabulary(tx, userID); err != nil {
			return err
		}
		vocabulary, err := findOwnedVocabulary(tx, userID, vocabularyID)
		if err != nil {
			return err
		}
		if input.TargetLanguage != "" && !strings.EqualFold(input.TargetLanguage, vocabulary.TargetLanguage) {
			return fmt.Errorf("%w: file target_language %q does not match vocabulary target language %q", ErrVocabularyInvalidImport, input.TargetLanguage, vocabulary.TargetLanguage)
		}
		if input.NativeLanguage != "" && !strings.EqualFold(input.NativeLanguage, vocabulary.NativeLanguage) {
			return fmt.Errorf("%w: file native_language %q does not match vocabulary native language %q", ErrVocabularyInvalidImport, input.NativeLanguage, vocabulary.NativeLanguage)
		}

		for index := range input.Entries {
			if err := importVocabularyEntry(tx, vocabulary.ID, vocabulary.TargetLanguage, vocabulary.NativeLanguage, input.Entries[index], result); err != nil {
				return err
			}
		}

		result.Vocabulary = *vocabulary
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *VocabularyService) Clear(ctx context.Context, userID, vocabularyID int64) (int64, error) {
	var deleted int64
	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockUserVocabulary(tx, userID); err != nil {
			return err
		}

		vocabulary, err := findOwnedVocabulary(tx, userID, vocabularyID)
		if err != nil {
			return err
		}

		operation := tx.Where("vocabulary_id = ?", vocabulary.ID).Delete(&models.VocabularyEntry{})
		deleted = operation.RowsAffected
		return operation.Error
	})
	return deleted, err
}

func (s *VocabularyService) Get(ctx context.Context, userID, vocabularyID int64, page, pageSize int) (*VocabularyEntryPage, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	vocabulary, err := findOwnedVocabulary(database.DB.WithContext(ctx), userID, vocabularyID)
	if err != nil {
		return nil, err
	}
	result := &VocabularyEntryPage{
		Vocabulary: vocabulary,
		Data:       make([]models.VocabularyEntry, 0),
		Page:       page,
		PageSize:   pageSize,
	}
	query := database.DB.WithContext(ctx).
		Model(&models.VocabularyEntry{}).
		Where("vocabulary_id = ?", vocabulary.ID)
	if err := query.Count(&result.Total).Error; err != nil {
		return nil, err
	}
	if err := query.
		Preload("Pronunciations", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Preload("Meanings", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Preload("Examples", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Preload("Relations", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Preload("Relations.RelatedEntry.Meanings", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Order("id ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&result.Data).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (s *VocabularyService) resolveLanguages(userID int64, targetLanguage, nativeLanguage string) (string, string, error) {
	targetLanguage = strings.TrimSpace(targetLanguage)
	nativeLanguage = strings.TrimSpace(nativeLanguage)
	if targetLanguage != "" && nativeLanguage != "" {
		return targetLanguage, nativeLanguage, nil
	}

	settings, err := s.userSettingsService.GetUserSettings(userID)
	if err != nil {
		return "", "", err
	}
	if targetLanguage == "" {
		targetLanguage = strings.TrimSpace(settings.TargetLanguage)
	}
	if nativeLanguage == "" {
		nativeLanguage = strings.TrimSpace(settings.NativeLanguage)
	}
	if targetLanguage == "" || nativeLanguage == "" {
		return "", "", ErrVocabularyLanguageRequired
	}
	return targetLanguage, nativeLanguage, nil
}

func validateVocabularyImport(input VocabularyImportInput) error {
	if len(input.Entries) == 0 {
		return fmt.Errorf("%w: entries are required", ErrVocabularyInvalidImport)
	}
	totalEntries := len(input.Entries)
	for index, entry := range input.Entries {
		if strings.TrimSpace(entry.TargetText) == "" {
			return fmt.Errorf("%w: entries[%d].target text is required", ErrVocabularyInvalidImport, index)
		}
		if len([]rune(strings.TrimSpace(entry.TargetText))) > 500 {
			return fmt.Errorf("%w: entries[%d].target text is too long", ErrVocabularyInvalidImport, index)
		}
		entryType := strings.TrimSpace(entry.EntryType)
		if entryType != "" && entryType != models.VocabularyEntryTypeWord && entryType != models.VocabularyEntryTypePhrase {
			return fmt.Errorf("%w: entries[%d].entry_type must be word or phrase", ErrVocabularyInvalidImport, index)
		}
		if len(entry.Meanings) == 0 {
			return fmt.Errorf("%w: entries[%d].meanings are required", ErrVocabularyInvalidImport, index)
		}
		for meaningIndex, meaning := range entry.Meanings {
			if strings.TrimSpace(meaning.NativeText) == "" {
				return fmt.Errorf("%w: entries[%d].meanings[%d].native text is required", ErrVocabularyInvalidImport, index, meaningIndex)
			}
			if len([]rune(strings.TrimSpace(meaning.PartOfSpeech))) > 64 {
				return fmt.Errorf("%w: entries[%d].meanings[%d].part_of_speech is too long", ErrVocabularyInvalidImport, index, meaningIndex)
			}
		}
		for pronunciationIndex, pronunciation := range entry.Pronunciations {
			if strings.TrimSpace(pronunciation.Pronunciation) == "" || strings.TrimSpace(pronunciation.PronunciationType) == "" {
				return fmt.Errorf("%w: entries[%d].pronunciations[%d] requires pronunciation and type", ErrVocabularyInvalidImport, index, pronunciationIndex)
			}
			if len([]rune(strings.TrimSpace(pronunciation.PronunciationType))) > 32 || len([]rune(strings.TrimSpace(pronunciation.Region))) > 32 {
				return fmt.Errorf("%w: entries[%d].pronunciations[%d] type or region is too long", ErrVocabularyInvalidImport, index, pronunciationIndex)
			}
		}
		for phraseIndex, phrase := range entry.RelatedPhrases {
			if strings.TrimSpace(phrase.TargetText) == "" || len(phrase.Meanings) == 0 {
				return fmt.Errorf("%w: entries[%d].related_phrases[%d] requires target_text and meanings", ErrVocabularyInvalidImport, index, phraseIndex)
			}
			if len([]rune(strings.TrimSpace(phrase.TargetText))) > 500 {
				return fmt.Errorf("%w: entries[%d].related_phrases[%d].target_text is too long", ErrVocabularyInvalidImport, index, phraseIndex)
			}
			for meaningIndex, meaning := range phrase.Meanings {
				if strings.TrimSpace(meaning.NativeText) == "" || len([]rune(strings.TrimSpace(meaning.PartOfSpeech))) > 64 {
					return fmt.Errorf("%w: entries[%d].related_phrases[%d].meanings[%d] is invalid", ErrVocabularyInvalidImport, index, phraseIndex, meaningIndex)
				}
			}
		}
		for exampleIndex, example := range entry.Examples {
			if strings.TrimSpace(example.TargetText) == "" {
				return fmt.Errorf("%w: entries[%d].examples[%d].target_text is required", ErrVocabularyInvalidImport, index, exampleIndex)
			}
		}
		totalEntries += len(entry.RelatedPhrases)
	}
	if totalEntries > maxVocabularyImportEntries {
		return fmt.Errorf("%w: import cannot exceed %d words and phrases", ErrVocabularyInvalidImport, maxVocabularyImportEntries)
	}
	return nil
}

func validateVocabularyInfo(name, targetLanguage, nativeLanguage string) error {
	if name == "" || len([]rune(name)) > 128 {
		return fmt.Errorf("%w: name is required and must not exceed 128 characters", ErrVocabularyInvalidInput)
	}
	if targetLanguage == "" || nativeLanguage == "" {
		return ErrVocabularyLanguageRequired
	}
	if len([]rune(targetLanguage)) > 32 || len([]rune(nativeLanguage)) > 32 {
		return fmt.Errorf("%w: language codes must not exceed 32 characters", ErrVocabularyInvalidInput)
	}
	return nil
}

func findOwnedVocabulary(db *gorm.DB, userID, vocabularyID int64) (*models.Vocabulary, error) {
	var vocabulary models.Vocabulary
	operation := db.Where("user_id = ? AND id = ?", userID, vocabularyID).Limit(1).Find(&vocabulary)
	if operation.Error != nil {
		return nil, operation.Error
	}
	if operation.RowsAffected == 0 {
		return nil, ErrVocabularyNotFound
	}
	return &vocabulary, nil
}

func ensureVocabularyNameAvailable(tx *gorm.DB, userID int64, name string, excludeID int64) error {
	query := tx.Model(&models.Vocabulary{}).Where("user_id = ? AND name = ?", userID, name)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrVocabularyNameConflict
	}
	return nil
}

func importVocabularyEntry(tx *gorm.DB, vocabularyID int64, targetLanguage, nativeLanguage string, input VocabularyImportEntry, result *VocabularyImportResult) error {
	entryType := strings.TrimSpace(input.EntryType)
	if entryType == "" {
		entryType = models.VocabularyEntryTypeWord
	}
	entry, created, err := upsertVocabularyEntry(tx, vocabularyID, targetLanguage, entryType, input.TargetText, input.Tags, input.Notes)
	if err != nil {
		return err
	}
	if created {
		result.EntriesCreated++
	} else {
		result.EntriesUpdated++
	}

	if err := mergeMeanings(tx, entry.ID, nativeLanguage, input.Meanings, result); err != nil {
		return err
	}
	if err := mergePronunciations(tx, entry.ID, input.Pronunciations, result); err != nil {
		return err
	}
	if err := mergeExamples(tx, entry.ID, input.Examples, result); err != nil {
		return err
	}

	for _, phrase := range input.RelatedPhrases {
		phraseEntry, phraseCreated, err := upsertVocabularyEntry(tx, vocabularyID, targetLanguage, models.VocabularyEntryTypePhrase, phrase.TargetText, nil, "")
		if err != nil {
			return err
		}
		if phraseCreated {
			result.EntriesCreated++
		} else {
			result.EntriesUpdated++
		}
		if err := mergeMeanings(tx, phraseEntry.ID, nativeLanguage, phrase.Meanings, result); err != nil {
			return err
		}
		if phraseEntry.ID != entry.ID {
			created, err := ensureEntryRelation(tx, entry.ID, phraseEntry.ID, models.VocabularyRelationPhrase)
			if err != nil {
				return err
			}
			if created {
				result.RelationsCreated++
			}
		}
	}
	return nil
}

func upsertVocabularyEntry(tx *gorm.DB, vocabularyID int64, targetLanguage, entryType, targetText string, tags []string, notes string) (*models.VocabularyEntry, bool, error) {
	targetText = strings.TrimSpace(targetText)
	normalized := normalizeVocabularyText(targetText)
	var entry models.VocabularyEntry
	operation := tx.Where(
		"vocabulary_id = ? AND normalized_target_text = ? AND target_language = ?",
		vocabularyID,
		normalized,
		targetLanguage,
	).Limit(1).Find(&entry)
	if operation.Error != nil {
		return nil, false, operation.Error
	}
	if operation.RowsAffected == 0 {
		entry = models.VocabularyEntry{
			VocabularyID:         vocabularyID,
			TargetText:           targetText,
			NormalizedTargetText: normalized,
			TargetLanguage:       targetLanguage,
			EntryType:            entryType,
			Tags:                 normalizeVocabularyTags(tags),
			Notes:                strings.TrimSpace(notes),
			Source:               models.VocabularySourceImport,
		}
		if err := tx.Create(&entry).Error; err != nil {
			return nil, false, err
		}
		return &entry, true, nil
	}

	updates := make(map[string]any)
	if entry.EntryType == models.VocabularyEntryTypePhrase && entryType == models.VocabularyEntryTypeWord {
		updates["entry_type"] = models.VocabularyEntryTypeWord
		entry.EntryType = models.VocabularyEntryTypeWord
	}
	mergedTags := mergeVocabularyTags(entry.Tags, tags)
	if len(mergedTags) != len(entry.Tags) {
		updates["tags"] = mergedTags
		entry.Tags = mergedTags
	}
	if note := strings.TrimSpace(notes); note != "" && note != entry.Notes {
		updates["notes"] = note
		entry.Notes = note
	}
	if len(updates) > 0 {
		if err := tx.Model(&entry).Updates(updates).Error; err != nil {
			return nil, false, err
		}
	}
	return &entry, false, nil
}

func mergeMeanings(tx *gorm.DB, entryID int64, nativeLanguage string, meanings []VocabularyImportMeaning, result *VocabularyImportResult) error {
	var existing []models.VocabularyMeaning
	if err := tx.Where("entry_id = ?", entryID).Find(&existing).Error; err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(existing)+len(meanings))
	for _, meaning := range existing {
		seen[meaningKey(meaning.NativeText, meaning.PartOfSpeech, meaning.NativeLanguage)] = struct{}{}
	}
	for _, input := range meanings {
		nativeText := strings.TrimSpace(input.NativeText)
		partOfSpeech := strings.TrimSpace(input.PartOfSpeech)
		key := meaningKey(nativeText, partOfSpeech, nativeLanguage)
		if _, ok := seen[key]; ok {
			continue
		}
		meaning := models.VocabularyMeaning{
			EntryID:        entryID,
			NativeText:     nativeText,
			NativeLanguage: nativeLanguage,
			PartOfSpeech:   partOfSpeech,
			SortOrder:      len(existing),
		}
		if err := tx.Create(&meaning).Error; err != nil {
			return err
		}
		existing = append(existing, meaning)
		seen[key] = struct{}{}
		result.MeaningsCreated++
	}
	return nil
}

func mergePronunciations(tx *gorm.DB, entryID int64, pronunciations []VocabularyImportPronunciation, result *VocabularyImportResult) error {
	var existing []models.VocabularyPronunciation
	if err := tx.Where("entry_id = ?", entryID).Find(&existing).Error; err != nil {
		return err
	}
	byVariant := make(map[string]*models.VocabularyPronunciation, len(existing))
	for index := range existing {
		byVariant[pronunciationKey(existing[index].PronunciationType, existing[index].Region)] = &existing[index]
	}
	for _, input := range pronunciations {
		pronunciationType := strings.ToLower(strings.TrimSpace(input.PronunciationType))
		region := strings.ToLower(strings.TrimSpace(input.Region))
		value := strings.TrimSpace(input.Pronunciation)
		key := pronunciationKey(pronunciationType, region)
		if current, ok := byVariant[key]; ok {
			updates := make(map[string]any)
			if value != current.Pronunciation {
				updates["pronunciation"] = value
			}
			if audioURL := strings.TrimSpace(input.AudioURL); audioURL != "" && audioURL != current.AudioURL {
				updates["audio_url"] = audioURL
			}
			if len(updates) > 0 {
				if err := tx.Model(current).Updates(updates).Error; err != nil {
					return err
				}
			}
			continue
		}
		pronunciation := models.VocabularyPronunciation{
			EntryID:           entryID,
			Pronunciation:     value,
			PronunciationType: pronunciationType,
			Region:            region,
			AudioURL:          strings.TrimSpace(input.AudioURL),
			SortOrder:         len(existing),
		}
		if err := tx.Create(&pronunciation).Error; err != nil {
			return err
		}
		existing = append(existing, pronunciation)
		byVariant[key] = &existing[len(existing)-1]
		result.PronunciationsCreated++
	}
	return nil
}

func mergeExamples(tx *gorm.DB, entryID int64, examples []VocabularyImportExample, result *VocabularyImportResult) error {
	var existing []models.VocabularyExample
	if err := tx.Where("entry_id = ?", entryID).Find(&existing).Error; err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(existing)+len(examples))
	for _, example := range existing {
		seen[exampleKey(example.TargetText, example.NativeText)] = struct{}{}
	}
	for _, input := range examples {
		targetText := strings.TrimSpace(input.TargetText)
		nativeText := strings.TrimSpace(input.NativeText)
		key := exampleKey(targetText, nativeText)
		if _, ok := seen[key]; ok {
			continue
		}
		example := models.VocabularyExample{
			EntryID:    entryID,
			TargetText: targetText,
			NativeText: nativeText,
			Source:     models.VocabularySourceImport,
			SortOrder:  len(existing),
		}
		if err := tx.Create(&example).Error; err != nil {
			return err
		}
		existing = append(existing, example)
		seen[key] = struct{}{}
		result.ExamplesCreated++
	}
	return nil
}

func ensureEntryRelation(tx *gorm.DB, entryID, relatedEntryID int64, relationType string) (bool, error) {
	var count int64
	if err := tx.Model(&models.VocabularyEntryRelation{}).
		Where("entry_id = ? AND related_entry_id = ? AND relation_type = ?", entryID, relatedEntryID, relationType).
		Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return false, nil
	}
	var relationCount int64
	if err := tx.Model(&models.VocabularyEntryRelation{}).Where("entry_id = ?", entryID).Count(&relationCount).Error; err != nil {
		return false, err
	}
	relation := models.VocabularyEntryRelation{
		EntryID:        entryID,
		RelatedEntryID: relatedEntryID,
		RelationType:   relationType,
		SortOrder:      int(relationCount),
	}
	if err := tx.Create(&relation).Error; err != nil {
		return false, err
	}
	return true, nil
}

func normalizeVocabularyText(value string) string {
	return strings.ToLower(strings.TrimSpace(norm.NFKC.String(value)))
}

func normalizeVocabularyTags(tags []string) []string {
	return mergeVocabularyTags(nil, tags)
}

func mergeVocabularyTags(existing, additions []string) []string {
	result := append([]string(nil), existing...)
	seen := make(map[string]struct{}, len(existing)+len(additions))
	for _, tag := range existing {
		seen[strings.ToLower(strings.TrimSpace(tag))] = struct{}{}
	}
	for _, tag := range additions {
		tag = strings.TrimSpace(tag)
		key := strings.ToLower(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, tag)
	}
	return result
}

func meaningKey(nativeText, partOfSpeech, nativeLanguage string) string {
	return normalizeVocabularyText(nativeText) + "\x00" + strings.ToLower(strings.TrimSpace(partOfSpeech)) + "\x00" + strings.ToLower(strings.TrimSpace(nativeLanguage))
}

func pronunciationKey(pronunciationType, region string) string {
	return strings.ToLower(strings.TrimSpace(pronunciationType)) + "\x00" + strings.ToLower(strings.TrimSpace(region))
}

func exampleKey(targetText, nativeText string) string {
	return normalizeVocabularyText(targetText) + "\x00" + normalizeVocabularyText(nativeText)
}

func lockUserVocabulary(tx *gorm.DB, userID int64) error {
	return tx.Exec("SELECT pg_advisory_xact_lock(?)", vocabularyAdvisoryLockBase+userID).Error
}
