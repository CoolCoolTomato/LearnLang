package models

import "testing"

func TestNormalizeLLMType(t *testing.T) {
	for _, test := range []struct {
		input string
		want  string
		valid bool
	}{
		{"", LLMTypeOpenAI, true},
		{" OpenAI ", LLMTypeOpenAI, true},
		{"ANTHROPIC", LLMTypeAnthropic, true},
		{"claude", "", false},
	} {
		got, valid := NormalizeLLMType(test.input)
		if got != test.want || valid != test.valid {
			t.Errorf("NormalizeLLMType(%q) = %q, %t; want %q, %t", test.input, got, valid, test.want, test.valid)
		}
	}
}

func TestNormalizeTTSType(t *testing.T) {
	tests := []struct {
		input string
		want  string
		valid bool
	}{
		{"", TTSTypeOpenAI, true},
		{"OPENAI", TTSTypeOpenAI, true},
		{" fish-audio ", TTSTypeFishAudio, true},
		{"unknown", "", false},
	}

	for _, test := range tests {
		got, valid := NormalizeTTSType(test.input)
		if got != test.want || valid != test.valid {
			t.Errorf("NormalizeTTSType(%q) = %q, %t; want %q, %t", test.input, got, valid, test.want, test.valid)
		}
	}
}

func TestModelTableNames(t *testing.T) {
	tests := map[string]string{
		(User{}).TableName():                     "users",
		(UserSettings{}).TableName():             "user_settings",
		(Message{}).TableName():                  "messages",
		(ConversationArchive{}).TableName():      "conversation_archives",
		(UserProfileSummary{}).TableName():       "user_profile_summaries",
		(ScheduledTask{}).TableName():            "scheduled_tasks",
		(VoiceFile{}).TableName():                "voice_files",
		(Vocabulary{}).TableName():               "vocabularies",
		(VocabularyEntry{}).TableName():          "vocabulary_entries",
		(VocabularyPronunciation{}).TableName():  "vocabulary_pronunciations",
		(VocabularyMeaning{}).TableName():        "vocabulary_meanings",
		(VocabularyExample{}).TableName():        "vocabulary_examples",
		(VocabularyEntryRelation{}).TableName():  "vocabulary_entry_relations",
		(VocabularyAgentSelection{}).TableName(): "vocabulary_agent_selections",
	}
	if len(tests) != 14 {
		t.Fatalf("unexpected duplicate or missing table names: %#v", tests)
	}
	for got, want := range tests {
		if got != want {
			t.Errorf("TableName() = %q, want %q", got, want)
		}
	}
}
