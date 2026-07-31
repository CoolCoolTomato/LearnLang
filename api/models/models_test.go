package models

import "testing"

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
