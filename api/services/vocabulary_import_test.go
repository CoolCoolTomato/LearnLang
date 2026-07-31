package services

import (
	"errors"
	"learnlang-api/models"
	"reflect"
	"strings"
	"testing"
)

func TestDecodeLearnLangVocabularyImport(t *testing.T) {
	data := []byte(`{
		"format":"learnlang-vocabulary","version":1,
		"target_language":" en-US ","native_language":" zh-CN ",
		"entries":[{
			"target_text":"hello","entry_type":"word","tags":["greeting"],"notes":"note",
			"pronunciations":[{"pronunciation":"həˈloʊ","type":"ipa","region":"us","audio_url":"https://example.test/a.mp3"}],
			"meanings":[{"native_text":"你好","part_of_speech":"interjection"}],
			"related_phrases":[{"target_text":"hello world","meanings":[{"native_text":"你好世界","part_of_speech":"phrase"}]}],
			"examples":[{"target_text":"Hello!","native_text":"你好！"}]
		}]
	}`)
	got, err := DecodeVocabularyImport(data)
	if err != nil {
		t.Fatalf("DecodeVocabularyImport() error = %v", err)
	}
	if got.TargetLanguage != "en-US" || got.NativeLanguage != "zh-CN" || len(got.Entries) != 1 {
		t.Fatalf("decoded document = %#v", got)
	}
	entry := got.Entries[0]
	if entry.TargetText != "hello" || len(entry.Pronunciations) != 1 || len(entry.Meanings) != 1 || len(entry.RelatedPhrases) != 1 || len(entry.Examples) != 1 {
		t.Fatalf("decoded entry = %#v", entry)
	}
}

func TestDecodeLegacyVocabularyImport(t *testing.T) {
	data := []byte(`[{
		"word":"hello","us":"həˈloʊ","uk":"həˈləʊ",
		"translations":[{"translation":"你好","type":"int."}],
		"pronunciations":[{"pronunciation":"custom","type":"phonetic","region":"x","audio_url":"url"}],
		"phrases":[{"phrase":"hello world","translation":"你好世界","type":"phrase"}],
		"sentences":[{"sentence":"Hello there","translation":"你好"}],
		"tags":["basic"],"notes":"note"
	}]`)
	got, err := DecodeVocabularyImport(data)
	if err != nil {
		t.Fatal(err)
	}
	entry := got.Entries[0]
	if entry.EntryType != models.VocabularyEntryTypeWord || len(entry.Pronunciations) != 3 || len(entry.Meanings) != 1 || len(entry.RelatedPhrases) != 1 || len(entry.Examples) != 1 {
		t.Fatalf("legacy entry = %#v", entry)
	}

	object, err := DecodeVocabularyImport([]byte(`{"entries":[{"word":"one"}]}`))
	if err != nil || len(object.Entries) != 1 || object.Entries[0].TargetText != "one" {
		t.Fatalf("legacy document = %#v, %v", object, err)
	}
	single, err := DecodeVocabularyImport([]byte(`{"word":"one"}`))
	if err != nil || len(single.Entries) != 1 {
		t.Fatalf("single legacy entry = %#v, %v", single, err)
	}
}

func TestDecodeVocabularyImportErrors(t *testing.T) {
	tests := []string{
		``, `null`, `x`, `{`, `{"format":"unknown"}`,
		`{"format":"learnlang-vocabulary","version":2,"target_language":"en","native_language":"zh"}`,
		`{"format":"learnlang-vocabulary","version":1,"target_language":"","native_language":"zh"}`,
		`{"entries":[]}`,
	}
	for _, input := range tests {
		if _, err := DecodeVocabularyImport([]byte(input)); !errors.Is(err, ErrVocabularyInvalidImport) {
			t.Errorf("DecodeVocabularyImport(%q) error = %v", input, err)
		}
	}
}

func TestVocabularyNormalizationHelpers(t *testing.T) {
	if got := normalizeVocabularyText("  Hello  WORLD "); got != "hello  world" {
		t.Fatalf("normalizeVocabularyText() = %q", got)
	}
	if got := normalizeVocabularyTags([]string{" Go ", "", "go", "Lang"}); !reflect.DeepEqual(got, []string{"Go", "Lang"}) {
		t.Fatalf("normalizeVocabularyTags() = %#v", got)
	}
	if got := mergeVocabularyTags([]string{"go", "old"}, []string{"Go", "new"}); !reflect.DeepEqual(got, []string{"go", "old", "new"}) {
		t.Fatalf("mergeVocabularyTags() = %#v", got)
	}
	if got := escapeVocabularyLikePattern(`50%_x\y`); got != `50\%\_x\\y` {
		t.Fatalf("escapeVocabularyLikePattern() = %q", got)
	}
	if meaningKey(" Text ", " Noun ", " en-US ") != "text\x00noun\x00en-us" {
		t.Fatal("meaningKey() did not normalize fields")
	}
	if pronunciationKey(" IPA ", " US ") != "ipa\x00us" || exampleKey(" One ", " Two ") != "one\x00two" {
		t.Fatal("composite key helpers did not normalize fields")
	}
}

func TestValidateVocabularyInfoAndImport(t *testing.T) {
	if err := validateVocabularyInfo(" Name ", "en", "zh"); err != nil {
		t.Fatalf("valid vocabulary info rejected: %v", err)
	}
	for _, input := range [][3]string{{"", "en", "zh"}, {strings.Repeat("x", 129), "en", "zh"}, {"name", "", "zh"}, {"name", "en", ""}} {
		if err := validateVocabularyInfo(input[0], input[1], input[2]); err == nil {
			t.Errorf("validateVocabularyInfo(%q, %q, %q) succeeded", input[0], input[1], input[2])
		}
	}
	valid := VocabularyImportInput{TargetLanguage: "en", NativeLanguage: "zh", Entries: []VocabularyImportEntry{{TargetText: "hello", Meanings: []VocabularyImportMeaning{{NativeText: "你好"}}}}}
	if err := validateVocabularyImport(valid); err != nil {
		t.Fatalf("valid import rejected: %v", err)
	}
	invalid := valid
	invalid.Entries = nil
	if err := validateVocabularyImport(invalid); err == nil {
		t.Fatal("empty import accepted")
	}
}
