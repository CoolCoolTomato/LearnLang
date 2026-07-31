package llm

import (
	"context"
	"strings"
	"testing"
)

func TestNormalizeType(t *testing.T) {
	for input, want := range map[string]string{
		"":          "",
		" OpenAI ":  TypeOpenAI,
		"ANTHROPIC": TypeAnthropic,
	} {
		if got := normalizeType(input); got != want {
			t.Errorf("normalizeType(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestNewWithOptionsRejectsUnsupportedType(t *testing.T) {
	model, err := NewWithOptions(context.Background(), "key", "", "model", "local", Options{})
	if err == nil || !strings.Contains(err.Error(), "unsupported llm type") {
		t.Fatalf("NewWithOptions() model = %v, error = %v", model, err)
	}
	if model != nil {
		t.Fatalf("NewWithOptions() model = %v, want nil", model)
	}
}

func TestNewModelConstructors(t *testing.T) {
	ctx := context.Background()
	temperature := float32(0.25)
	for _, tt := range []struct {
		name    string
		llmType string
	}{
		{"default OpenAI", ""},
		{"explicit OpenAI", " OPENAI "},
		{"Anthropic", " ANTHROPIC "},
	} {
		t.Run(tt.name, func(t *testing.T) {
			model, err := NewWithOptions(ctx, " key ", " https://example.test ", " ", tt.llmType, Options{MaxTokens: 128, Temperature: &temperature})
			if err != nil || model == nil {
				t.Fatalf("NewWithOptions() = %v, %v", model, err)
			}
		})
	}
	model, err := New(ctx, "key", "", "model", TypeOpenAI)
	if err != nil || model == nil {
		t.Fatalf("New() = %v, %v", model, err)
	}
}

func TestModelConstructorsRequireAPIKey(t *testing.T) {
	ctx := context.Background()
	if model, err := newOpenAI(ctx, " ", "", "", Options{}); err == nil || model != nil || !strings.Contains(err.Error(), "api key") {
		t.Fatalf("newOpenAI() = %v, %v", model, err)
	}
	if model, err := newClaudeCode(ctx, " ", "", "", Options{}); err == nil || model != nil || !strings.Contains(err.Error(), "api key") {
		t.Fatalf("newClaudeCode() = %v, %v", model, err)
	}
	model, err := newClaudeCode(ctx, "key", "", "", Options{})
	if err != nil || model == nil {
		t.Fatalf("newClaudeCode(defaults) = %v, %v", model, err)
	}
}
