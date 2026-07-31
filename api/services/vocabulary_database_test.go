package services

import (
	"context"
	"errors"
	"learnlang-api/database"
	"learnlang-api/models"
	"testing"

	"gorm.io/gorm"
)

func setupVocabularyTestDB(t *testing.T) *VocabularyService {
	t.Helper()
	setupServiceTestDB(t,
		&models.Vocabulary{}, &models.VocabularyEntry{}, &models.VocabularyPronunciation{},
		&models.VocabularyMeaning{}, &models.VocabularyExample{}, &models.VocabularyEntryRelation{},
		&models.VocabularyAgentSelection{},
	)
	previousLock := acquireVocabularyUserLock
	acquireVocabularyUserLock = func(*gorm.DB, int64) error { return nil }
	t.Cleanup(func() { acquireVocabularyUserLock = previousLock })
	return NewVocabularyService(NewUserSettingsService())
}

func TestVocabularyServiceCRUDAndImport(t *testing.T) {
	service := setupVocabularyTestDB(t)
	settings := models.UserSettings{UserID: 1, TargetLanguage: "en-US", NativeLanguage: "zh-CN"}
	if err := database.DB.Create(&settings).Error; err != nil {
		t.Fatal(err)
	}

	first, err := service.Create(context.Background(), 1, VocabularyCreateInput{Name: " Default "})
	if err != nil {
		t.Fatalf("Create(default) error = %v", err)
	}
	if !first.IsDefault || first.TargetLanguage != "en-US" || first.NativeLanguage != "zh-CN" || first.Name != "Default" {
		t.Fatalf("default vocabulary = %#v", first)
	}
	second, err := service.Create(context.Background(), 1, VocabularyCreateInput{Name: "Review", TargetLanguage: "en", NativeLanguage: "zh"})
	if err != nil || second.IsDefault {
		t.Fatalf("Create(second) = %#v, %v", second, err)
	}
	if _, err := service.Create(context.Background(), 1, VocabularyCreateInput{Name: "Default", TargetLanguage: "en", NativeLanguage: "zh"}); !errors.Is(err, ErrVocabularyNameConflict) {
		t.Fatalf("duplicate name error = %v", err)
	}
	list, err := service.List(context.Background(), 1)
	if err != nil || len(list) != 2 || !list[0].IsDefault {
		t.Fatalf("List() = %#v, %v", list, err)
	}

	falseValue := false
	if _, err := service.Update(context.Background(), 1, first.ID, VocabularyUpdateInput{IsDefault: &falseValue}); !errors.Is(err, ErrVocabularyDefaultRequired) {
		t.Fatalf("demote default error = %v", err)
	}
	trueValue := true
	newName := "Review Updated"
	updated, err := service.Update(context.Background(), 1, second.ID, VocabularyUpdateInput{Name: &newName, IsDefault: &trueValue})
	if err != nil || !updated.IsDefault || updated.Name != newName {
		t.Fatalf("Update(second) = %#v, %v", updated, err)
	}
	if err := database.DB.First(first, first.ID).Error; err != nil || first.IsDefault {
		t.Fatalf("old default = %#v, %v", first, err)
	}

	input := VocabularyImportInput{
		TargetLanguage: "en", NativeLanguage: "zh",
		Entries: []VocabularyImportEntry{{
			TargetText: "Hello", EntryType: models.VocabularyEntryTypeWord, Tags: []string{"basic"}, Notes: "note",
			Meanings:       []VocabularyImportMeaning{{NativeText: "你好", PartOfSpeech: "interjection"}},
			Pronunciations: []VocabularyImportPronunciation{{Pronunciation: "həˈloʊ", PronunciationType: "ipa", Region: "us"}},
			Examples:       []VocabularyImportExample{{TargetText: "Hello!", NativeText: "你好！"}},
			RelatedPhrases: []VocabularyImportRelatedPhrase{{TargetText: "hello world", Meanings: []VocabularyImportMeaning{{NativeText: "你好世界"}}}},
		}},
	}
	result, err := service.Import(context.Background(), 1, second.ID, input)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if result.EntriesCreated != 2 || result.MeaningsCreated != 2 || result.PronunciationsCreated != 1 || result.ExamplesCreated != 1 || result.RelationsCreated != 1 {
		t.Fatalf("import result = %#v", result)
	}
	result, err = service.Import(context.Background(), 1, second.ID, input)
	if err != nil || result.EntriesUpdated == 0 || result.MeaningsCreated != 0 || result.RelationsCreated != 0 {
		t.Fatalf("repeat import result = %#v, %v", result, err)
	}

	page, err := service.Get(context.Background(), 1, second.ID, 0, 0, "")
	if err != nil || page.Page != 1 || page.PageSize != 20 || page.Total != 2 || len(page.Data) != 2 {
		t.Fatalf("Get() = %#v, %v", page, err)
	}
	if _, err := service.Get(context.Background(), 1, second.ID, 1, 20, string(make([]rune, 201))); !errors.Is(err, ErrVocabularyInvalidInput) {
		t.Fatalf("long query error = %v", err)
	}
	deleted, err := service.Clear(context.Background(), 1, second.ID)
	if err != nil || deleted != 2 {
		t.Fatalf("Clear() = %d, %v", deleted, err)
	}
	if err := service.Delete(context.Background(), 1, second.ID); err != nil {
		t.Fatalf("Delete(default) error = %v", err)
	}
	if err := database.DB.First(first, first.ID).Error; err != nil || !first.IsDefault {
		t.Fatalf("replacement default = %#v, %v", first, err)
	}
	if err := service.Delete(context.Background(), 2, first.ID); !errors.Is(err, ErrVocabularyNotFound) {
		t.Fatalf("foreign vocabulary delete error = %v", err)
	}
}

func TestVocabularyServiceValidation(t *testing.T) {
	service := setupVocabularyTestDB(t)
	if _, err := service.Create(context.Background(), 1, VocabularyCreateInput{}); !errors.Is(err, ErrVocabularyInvalidInput) {
		t.Fatalf("empty create error = %v", err)
	}
	if _, err := service.Create(context.Background(), 1, VocabularyCreateInput{Name: "name"}); !errors.Is(err, ErrVocabularyLanguageRequired) {
		t.Fatalf("missing languages error = %v", err)
	}
	if _, err := service.Get(context.Background(), 1, 999, 1, 20, ""); !errors.Is(err, ErrVocabularyNotFound) {
		t.Fatalf("missing Get() error = %v", err)
	}
	if _, err := service.Import(context.Background(), 1, 999, VocabularyImportInput{}); !errors.Is(err, ErrVocabularyInvalidImport) {
		t.Fatalf("invalid Import() error = %v", err)
	}
}
