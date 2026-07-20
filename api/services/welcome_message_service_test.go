package services

import "testing"

func TestWelcomeMessageUsesLanguagePreferences(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		native      string
		wantText    string
		wantMeaning string
	}{
		{name: "English for Chinese speaker", target: "en", native: "zh-CN", wantText: welcomeGreetings["en"].Text, wantMeaning: welcomeGreetings["zh"].Text},
		{name: "Japanese for English speaker", target: "ja-JP", native: "en-US", wantText: welcomeGreetings["ja"].Text, wantMeaning: welcomeGreetings["en"].Text},
		{name: "Unsupported target falls back to English", target: "de", native: "fr", wantText: welcomeGreetings["en"].Text, wantMeaning: welcomeGreetings["fr"].Text},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			text, meaning := welcomeMessage(test.target, test.native)
			if text != test.wantText || meaning != test.wantMeaning {
				t.Fatalf("welcomeMessage(%q, %q) = (%q, %q)", test.target, test.native, text, meaning)
			}
		})
	}
}

func TestWelcomeMessageOmitsDuplicateTranslation(t *testing.T) {
	text, translation := welcomeMessage("en-US", "en")
	if text == "" || translation != "" {
		t.Fatalf("expected an English greeting without duplicate translation, got (%q, %q)", text, translation)
	}
}
