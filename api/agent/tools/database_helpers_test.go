package tools

import (
	"context"
	"fmt"
	"learnlang-api/database"
	"learnlang-api/models"
	"learnlang-api/services"
	"reflect"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupToolsTestDB(t *testing.T, modelTypes ...any) *gorm.DB {
	t.Helper()
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:tools_%s?mode=memory&cache=shared", name)), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(modelTypes...); err != nil {
		t.Fatal(err)
	}
	previous := database.DB
	database.DB = db
	t.Cleanup(func() {
		database.DB = previous
		_ = sqlDB.Close()
	})
	return db
}

func TestLinkedMessagesPreservesRequestedOrderAndUser(t *testing.T) {
	setupToolsTestDB(t, &models.Message{})
	messages := []models.Message{
		{UserID: 1, Role: "user", TextContent: "first"},
		{UserID: 1, Role: "assistant", TextContent: "second"},
		{UserID: 2, Role: "user", TextContent: "other"},
	}
	for index := range messages {
		if err := database.DB.Create(&messages[index]).Error; err != nil {
			t.Fatal(err)
		}
	}
	got, err := linkedMessages(context.Background(), 1, []int64{messages[1].ID, messages[2].ID, messages[0].ID, 999})
	if err != nil || len(got) != 2 || got[0].ID != messages[1].ID || got[1].ID != messages[0].ID {
		t.Fatalf("linkedMessages() = %#v, %v", got, err)
	}
	empty, err := linkedMessages(context.Background(), 1, nil)
	if err != nil || len(empty) != 0 || empty == nil {
		t.Fatalf("linkedMessages(empty) = %#v, %v", empty, err)
	}
}

func TestFormatMessagesUsesTimezoneFallback(t *testing.T) {
	messages := []models.Message{{
		Role: "user", TextContent: "hello", Translation: "你好",
		CreatedAt: time.Date(2026, 7, 31, 2, 0, 0, 0, time.UTC),
	}}
	got := formatMessages(messages, "Asia/Shanghai")
	if len(got) != 1 || !strings.Contains(got[0], "2026-07-31 10:00:00") || !strings.Contains(got[0], "TextContent: hello Translation: 你好") {
		t.Fatalf("formatMessages() = %#v", got)
	}
	fallback := formatMessages(messages, "Not/A_Timezone")
	if !strings.Contains(fallback[0], "2026-07-31 02:00:00") {
		t.Fatalf("formatMessages() fallback = %#v", fallback)
	}
}

func TestUserProfileSummaryToolPersistsAndDetectsUnchanged(t *testing.T) {
	setupToolsTestDB(t, &models.UserProfileSummary{})
	tool := UserProfileSummaryTool{UserID: 3}
	output, err := tool.Call(context.Background(), `{"summary":"  基本事实：生日为7月31日。  "}`)
	if err != nil || !strings.Contains(output, `"status":"updated"`) {
		t.Fatalf("first Call() = %q, %v", output, err)
	}
	output, err = tool.Call(context.Background(), `{"summary":"基本事实：生日为7月31日。"}`)
	if err != nil || !strings.Contains(output, `"status":"unchanged"`) {
		t.Fatalf("second Call() = %q, %v", output, err)
	}
	var profile models.UserProfileSummary
	if err := database.DB.Where("user_id = ?", 3).First(&profile).Error; err != nil || profile.Summary == "" {
		t.Fatalf("stored profile = %#v, %v", profile, err)
	}
	for _, input := range []string{`{`, `{"summary":" "}`} {
		if _, err := tool.Call(context.Background(), input); err == nil {
			t.Errorf("Call(%q) unexpectedly succeeded", input)
		}
	}
	output, err = tool.Call(context.Background(), `{"summary":"用户询问过 Go"}`)
	if err != nil || !strings.Contains(output, `"status":"rejected"`) {
		t.Fatalf("invalid profile Call() = %q, %v", output, err)
	}
}

func TestArchivedKeywordSearchInputValidation(t *testing.T) {
	tool := ArchivedConversationKeywordSearchTool{UserID: 1}
	for _, tt := range []struct {
		input string
		want  string
	}{
		{`{`, "valid JSON"},
		{`{"keyword":" "}`, "keyword is required"},
		{`{"keyword":"go","start_time":"bad"}`, "start_time must be RFC3339"},
		{`{"keyword":"go","end_time":"bad"}`, "end_time must be RFC3339"},
		{`{"keyword":"go","start_time":"2026-08-01T00:00:00Z","end_time":"2026-07-31T00:00:00Z"}`, "must not be after"},
	} {
		output, err := tool.Call(context.Background(), tt.input)
		if err != nil || !strings.Contains(output, tt.want) {
			t.Errorf("Call(%q) = %q, %v", tt.input, output, err)
		}
	}
}

func TestKeywordSearchHelpers(t *testing.T) {
	if got, err := parseKeywordSearchTime(" ", "start"); err != nil || got != nil {
		t.Fatalf("empty parse = %v, %v", got, err)
	}
	got, err := parseKeywordSearchTime("2026-07-31T10:00:00+08:00", "start")
	if err != nil || got == nil || got.Hour() != 10 {
		t.Fatalf("valid parse = %v, %v", got, err)
	}
	if _, err := parseKeywordSearchTime("bad", "start"); err == nil {
		t.Fatal("invalid time accepted")
	}
	if escaped := escapeLikeKeyword(`50%_x\y`); escaped != `50\%\_x\\y` {
		t.Fatalf("escapeLikeKeyword() = %q", escaped)
	}
}

func TestLongTermMemorySearchWithoutStore(t *testing.T) {
	tool := LongTermMemorySearchTool{UserID: 1}
	output, err := tool.Call(context.Background(), "project decision")
	if err != nil || !strings.Contains(output, "memory store is not configured") || !strings.Contains(output, "project decision") {
		t.Fatalf("Call() = %q, %v", output, err)
	}
}

func TestParseVocabularyToolCount(t *testing.T) {
	for _, tt := range []struct {
		input       string
		want        int
		wantInvalid bool
	}{
		{"", 1, false},
		{`{}`, 1, false},
		{`{"count":3}`, 3, false},
		{`{"count":99}`, services.MaxAgentVocabularyWords, false},
		{`{"count":0}`, 0, true},
		{`{`, 0, true},
	} {
		got, invalid, err := parseVocabularyToolCount(tt.input)
		if err != nil || got != tt.want || (invalid != "") != tt.wantInvalid {
			t.Errorf("parseVocabularyToolCount(%q) = %d, %q, %v", tt.input, got, invalid, err)
		}
	}
}

func TestMarshalVocabularyWordsResultAndEntry(t *testing.T) {
	phrase := &models.VocabularyEntry{TargetText: "hello world", Meanings: []models.VocabularyMeaning{{NativeText: "你好世界", PartOfSpeech: "phrase"}}}
	entry := models.VocabularyEntry{
		ID: 9, TargetText: "hello", Tags: []string{"basic"}, Notes: "note", EncounterCount: 2,
		Meanings:       []models.VocabularyMeaning{{NativeText: "你好", PartOfSpeech: "interjection"}},
		Pronunciations: []models.VocabularyPronunciation{{Pronunciation: "həˈloʊ", PronunciationType: "ipa", Region: "us", AudioURL: "audio"}},
		Examples:       []models.VocabularyExample{{TargetText: "Hello!", NativeText: "你好！"}},
		Relations: []models.VocabularyEntryRelation{
			{RelationType: models.VocabularyRelationPhrase, RelatedEntry: phrase},
			{RelationType: models.VocabularyRelationDerived, RelatedEntry: &models.VocabularyEntry{TargetText: "ignored"}},
			{RelationType: models.VocabularyRelationPhrase},
		},
	}
	result := services.VocabularyRandomEntry{
		Vocabulary: models.Vocabulary{ID: 2, Name: "default", TargetLanguage: "en", NativeLanguage: "zh"},
		Entry:      entry, RelatedPhraseCount: 4,
	}
	converted := makeVocabularyToolEntry(&result)
	if converted.Entry.ID != 9 || len(converted.Entry.Meanings) != 1 || len(converted.Entry.Pronunciations) != 1 || len(converted.Entry.Examples) != 1 || len(converted.Entry.RelatedPhrases) != 1 || !converted.Entry.RelatedPhrasesTruncated {
		t.Fatalf("makeVocabularyToolEntry() = %#v", converted)
	}
	output, err := marshalVocabularyWordsResult(2, []services.VocabularyRandomEntry{result}, "none")
	if err != nil || !strings.Contains(output, `"status":"found"`) || !strings.Contains(output, `"actual_count":1`) {
		t.Fatalf("found result = %q, %v", output, err)
	}
	empty, err := marshalVocabularyWordsResult(1, nil, "none available")
	if err != nil || !strings.Contains(empty, `"status":"empty"`) || !strings.Contains(empty, "none available") {
		t.Fatalf("empty result = %q, %v", empty, err)
	}
}

func TestVocabularyToolRequiresServiceAndUsesCache(t *testing.T) {
	if _, err := callRandomVocabularyWords(context.Background(), `{}`, vocabularyToolCallConfig{}); err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("nil service error = %v", err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := callRandomVocabularyWords(canceled, `{}`, vocabularyToolCallConfig{}); err != context.Canceled {
		t.Fatalf("canceled error = %v", err)
	}
	state := NewTurnState()
	state.SetVocabularyToolResult(models.VocabularyAgentSelectionNew, "cached")
	got, err := callRandomVocabularyWords(context.Background(), `{}`, vocabularyToolCallConfig{
		selectionType: models.VocabularyAgentSelectionNew,
		vocabulary:    &services.VocabularyService{},
		state:         state,
	})
	if err != nil || got != "cached" {
		t.Fatalf("cached result = %q, %v", got, err)
	}
}

func TestToolMetadata(t *testing.T) {
	tools := []interface {
		Name() string
		Description() string
	}{
		ArchivedConversationKeywordSearchTool{}, LongTermMemorySearchTool{},
		RandomNewVocabularyWordTool{}, RandomOldVocabularyWordTool{}, UserProfileSummaryTool{},
		SendChatReplyTool{}, ScheduleMessageTool{}, CompleteChatTurnTool{},
	}
	for _, tool := range tools {
		if strings.TrimSpace(tool.Name()) == "" || strings.TrimSpace(tool.Description()) == "" {
			t.Errorf("empty metadata for %T", tool)
		}
	}
	if got := []string{tools[0].Name(), tools[1].Name()}; reflect.DeepEqual(got, []string{"", ""}) {
		t.Fatal("tool names were unexpectedly empty")
	}
}
