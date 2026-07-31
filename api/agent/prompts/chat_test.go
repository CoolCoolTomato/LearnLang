package prompts

import (
	"encoding/json"
	"learnlang-api/models"
	"strings"
	"testing"
	"time"
)

func TestNormalizeLanguage(t *testing.T) {
	if got := normalizeLanguage("  ", "fallback"); got != "fallback" {
		t.Fatalf("normalizeLanguage(empty) = %q", got)
	}
	if got := normalizeLanguage(" en-US ", "fallback"); got != "en-US" {
		t.Fatalf("normalizeLanguage(value) = %q", got)
	}
}

func TestFormatUserProfileEscapesData(t *testing.T) {
	got := formatUserProfile("  line \"quoted\"\nnext  ")
	var decoded map[string]string
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("formatUserProfile() returned invalid JSON: %v", err)
	}
	if decoded["summary"] != "line \"quoted\"\nnext" {
		t.Fatalf("summary = %q", decoded["summary"])
	}
}

func TestFormatShortTermMemoryUsesTimezoneAndOmitsEmptyTranslation(t *testing.T) {
	messages := []models.Message{{
		ID: 7, Role: "user", TextContent: "hello",
		CreatedAt: time.Date(2026, 7, 31, 2, 3, 4, 0, time.UTC),
	}}
	got := formatShortTermMemory(messages, "Asia/Shanghai")
	if !strings.Contains(got, `"time":"2026-07-31T10:03:04+08:00"`) {
		t.Fatalf("formatted memory = %s", got)
	}
	if strings.Contains(got, "translation") {
		t.Fatalf("empty translation should be omitted: %s", got)
	}

	invalidTimezone := formatShortTermMemory(messages, "Not/A_Timezone")
	if !strings.Contains(invalidTimezone, `"time":"2026-07-31T02:03:04Z"`) {
		t.Fatalf("invalid timezone did not fall back to UTC: %s", invalidTimezone)
	}
}

func TestChatSystemPromptIncludesDefaultsAndData(t *testing.T) {
	prompt := ChatSystemPrompt("", "", "2026-07-31 10:00", "Asia/Shanghai", nil, "likes Go")
	for _, expected := range []string{
		"User's native language: zh-CN",
		"Target language: en-US",
		"likes Go",
		"2026-07-31 10:00",
		"Asia/Shanghai",
		"application-selected conversation context",
		"send_chat_reply exactly once",
		"complete_chat_turn",
		"scheduled_at",
	} {
		if !strings.Contains(prompt, expected) {
			t.Errorf("prompt does not contain %q", expected)
		}
	}
}
