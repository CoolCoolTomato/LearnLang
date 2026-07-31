package services

import (
	"context"
	"errors"
	"learnlang-api/models"
	"strings"
	"testing"
)

type fakeTranslationSettings struct {
	settings *models.UserSettings
	err      error
	userID   int64
}

func (f *fakeTranslationSettings) GetUserSettings(userID int64) (*models.UserSettings, error) {
	f.userID = userID
	return f.settings, f.err
}

func TestTranslationServiceValidationAndSameLanguage(t *testing.T) {
	provider := &fakeTranslationSettings{settings: &models.UserSettings{NativeLanguage: "en-US", TargetLanguage: "en-US"}}
	called := false
	service := &TranslationService{
		settings: provider,
		generate: func(context.Context, *models.UserSettings, string) (string, error) {
			called = true
			return "", nil
		},
	}

	if _, err := service.Translate(context.Background(), 5, "  "); !errors.Is(err, ErrTranslationTextRequired) {
		t.Fatalf("empty Translate() error = %v", err)
	}
	tooLong := strings.Repeat("界", MaxTranslationTextLength+1)
	if _, err := service.Translate(context.Background(), 5, tooLong); !errors.Is(err, ErrTranslationTextTooLong) {
		t.Fatalf("long Translate() error = %v", err)
	}
	got, err := service.Translate(context.Background(), 5, "  keep me  ")
	if err != nil || got != "keep me" {
		t.Fatalf("same-language Translate() = %q, %v", got, err)
	}
	if called {
		t.Fatal("same-language translation called the generator")
	}
	if provider.userID != 5 {
		t.Fatalf("settings requested for user %d", provider.userID)
	}
}

func TestTranslationServiceGeneratesPromptAndTrimsResult(t *testing.T) {
	settings := &models.UserSettings{NativeLanguage: "zh-CN", TargetLanguage: "en-US"}
	provider := &fakeTranslationSettings{settings: settings}
	var receivedPrompt string
	service := &TranslationService{
		settings: provider,
		generate: func(_ context.Context, gotSettings *models.UserSettings, prompt string) (string, error) {
			if gotSettings != settings {
				t.Fatal("generator received another settings value")
			}
			receivedPrompt = prompt
			return "  translated  ", nil
		},
	}
	got, err := service.Translate(context.Background(), 8, "say \"hello\"\nnow")
	if err != nil || got != "translated" {
		t.Fatalf("Translate() = %q, %v", got, err)
	}
	for _, expected := range []string{"from zh-CN to en-US", `"say \"hello\"\nnow"`} {
		if !strings.Contains(receivedPrompt, expected) {
			t.Errorf("prompt does not contain %q: %s", expected, receivedPrompt)
		}
	}
}

func TestTranslationServicePropagatesErrorsAndRejectsEmptyResponse(t *testing.T) {
	wantErr := errors.New("settings failed")
	service := &TranslationService{settings: &fakeTranslationSettings{err: wantErr}}
	if _, err := service.Translate(context.Background(), 1, "text"); !errors.Is(err, wantErr) {
		t.Fatalf("settings error = %v", err)
	}

	wantErr = errors.New("model failed")
	service = &TranslationService{
		settings: &fakeTranslationSettings{settings: &models.UserSettings{}},
		generate: func(context.Context, *models.UserSettings, string) (string, error) { return "", wantErr },
	}
	if _, err := service.Translate(context.Background(), 1, "text"); !errors.Is(err, wantErr) {
		t.Fatalf("generator error = %v", err)
	}
	service.generate = func(context.Context, *models.UserSettings, string) (string, error) { return " \n", nil }
	if _, err := service.Translate(context.Background(), 1, "text"); err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("empty response error = %v", err)
	}
}

func TestTranslationHelpers(t *testing.T) {
	if got := normalizeTranslationLanguage(" 0 ", "fallback"); got != "fallback" {
		t.Fatalf("normalizeTranslationLanguage(0) = %q", got)
	}
	if got := normalizeTranslationLanguage(" fr ", "fallback"); got != "fr" {
		t.Fatalf("normalizeTranslationLanguage(fr) = %q", got)
	}
	settings := &models.UserSettings{}
	if _, err := newTranslationLLM(context.Background(), settings); err == nil || !strings.Contains(err.Error(), "API key") {
		t.Fatalf("newTranslationLLM() error = %v", err)
	}
}
