package services

import "testing"

func TestNormalizeVocabularyLanguage(t *testing.T) {
	tests := map[string]string{
		"en":      "en",
		"en-US":   "en",
		"en_US":   "en",
		" JA-jp ": "ja",
		"":        "",
	}

	for input, want := range tests {
		if got := normalizeVocabularyLanguage(input); got != want {
			t.Errorf("normalizeVocabularyLanguage(%q) = %q, want %q", input, got, want)
		}
	}
}
