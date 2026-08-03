package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestTurnStateCopiesResultsAndKeepsFirstVocabularyResult(t *testing.T) {
	state := NewTurnState()
	state.AddReply("hello", "你好", 10)
	state.Complete("en")
	state.SetVocabularyToolResult("new", "first")
	state.SetVocabularyToolResult("new", "second")

	result := state.Result()
	if len(result.ReplySentences) != 1 || result.ReplySentences[0].Original != "hello" || result.MessageIDs[0] != 10 || result.DetectedLanguage != "en" {
		t.Fatalf("Result() = %#v", result)
	}
	result.ReplySentences[0].Original = "changed"
	result.MessageIDs[0] = -1
	if current := state.Result(); current.ReplySentences[0].Original != "hello" || current.MessageIDs[0] != 10 {
		t.Fatal("Result() exposed internal slices")
	}
	if got, ok := state.VocabularyToolResult("new"); !ok || got != "first" {
		t.Fatalf("VocabularyToolResult() = %q, %v", got, ok)
	}
	if _, ok := state.VocabularyToolResult("missing"); ok {
		t.Fatal("missing vocabulary result reported as present")
	}
}

func TestCompleteChatTurnTool(t *testing.T) {
	state := NewTurnState()
	tool := CompleteChatTurnTool{State: state}
	output, err := tool.Call(context.Background(), `{"detected_language":"fr"}`)
	if err != nil || !strings.Contains(output, `"status":"completed"`) {
		t.Fatalf("Call() output = %q, error = %v", output, err)
	}
	if state.Result().DetectedLanguage != "fr" {
		t.Fatalf("detected language = %q", state.Result().DetectedLanguage)
	}
	if !state.IsCompleted() {
		t.Fatal("turn state is not marked completed")
	}
	if _, err := tool.Call(context.Background(), `{"detected_language":"fr"}`); err == nil {
		t.Fatal("CompleteChatTurnTool accepted a second completion")
	}
	if _, err := tool.Call(context.Background(), `{`); err == nil {
		t.Fatal("Call() accepted invalid JSON")
	}
}

func TestSendChatReplyToolValidation(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"invalid JSON", `{`, "valid JSON"},
		{"empty", `{"messages":[]}`, "at least one"},
		{"missing original", `{"messages":[{"translation":"x"}]}`, "original is required"},
		{"missing translation", `{"messages":[{"original":"x"}]}`, "translation is required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := (SendChatReplyTool{}).Call(context.Background(), tt.input)
			if err != nil || !strings.Contains(output, tt.want) {
				t.Fatalf("Call() output = %q, error = %v", output, err)
			}
		})
	}

	messages := make([]Sentence, maxChatReplyBatchSize+1)
	for index := range messages {
		messages[index] = Sentence{Original: "a", Translation: "b"}
	}
	data, _ := json.Marshal(map[string]any{"messages": messages})
	output, err := (SendChatReplyTool{}).Call(context.Background(), string(data))
	if err != nil || !strings.Contains(output, "at most") {
		t.Fatalf("oversized Call() output = %q, error = %v", output, err)
	}

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := (SendChatReplyTool{}).Call(canceled, `{}`); err != context.Canceled {
		t.Fatalf("canceled Call() error = %v", err)
	}

	if _, err := (SendChatReplyTool{}).Call(context.Background(), `{"messages":[{"original":"a","translation":"b"}]}`); err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("nil runtime error = %v", err)
	}
}

func TestScheduledTimeToUTC(t *testing.T) {
	got, err := scheduledTimeToUTC("2026-07-31T10:30:00", "Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 7, 31, 2, 30, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("scheduledTimeToUTC() = %v, want %v", got, want)
	}
	if normalizedTimezone("  ") != "UTC" || normalizedTimezone(" Asia/Shanghai ") != "Asia/Shanghai" {
		t.Fatal("normalizedTimezone() returned unexpected value")
	}
	for _, tt := range []struct{ value, timezone string }{
		{"bad", "UTC"},
		{"2026-07-31T10:30:00Z", "UTC"},
		{"2026-07-31T10:30:00", "Not/A_Timezone"},
		{"2026-03-08T02:30:00", "America/New_York"},
		{"2026-11-01T01:30:00", "America/New_York"},
	} {
		if _, err := scheduledTimeToUTC(tt.value, tt.timezone); err == nil {
			t.Errorf("scheduledTimeToUTC(%q, %q) unexpectedly succeeded", tt.value, tt.timezone)
		}
	}
}

func TestScheduleMessageToolValidation(t *testing.T) {
	tool := ScheduleMessageTool{Timezone: "Asia/Shanghai"}
	if !strings.Contains(tool.Description(), "Asia/Shanghai") {
		t.Fatalf("Description() = %q", tool.Description())
	}
	for _, tt := range []struct {
		input string
		want  string
	}{
		{`{`, "parse schedule_message input"},
		{`{"scheduled_at":"2026-01-01T10:00:00"}`, "message is required"},
		{`{"message":"x"}`, "scheduled_at is required"},
		{`{"message":"x","scheduled_at":"2026-01-01T10:00:00"}`, "not configured"},
	} {
		if _, err := tool.Call(context.Background(), tt.input); err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("Call(%q) error = %v, want containing %q", tt.input, err, tt.want)
		}
	}
}

func TestValidateUserProfileSummary(t *testing.T) {
	for _, summary := range []string{"今天开始学习", "Tomorrow is my birthday", "用户询问过 Go", "Topic not specified"} {
		if reason := validateUserProfileSummary(summary); reason == "" {
			t.Errorf("validateUserProfileSummary(%q) accepted invalid summary", summary)
		}
	}
	if reason := validateUserProfileSummary("基本事实：生日为7月31日。兴趣与偏好：喜欢 Go。"); reason != "" {
		t.Fatalf("valid summary rejected: %s", reason)
	}
}
