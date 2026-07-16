package services

import (
	"bytes"
	"encoding/json"
	"fmt"
)

const (
	LearnLangVocabularyFormat  = "learnlang-vocabulary"
	LearnLangVocabularyVersion = 1
)

type VocabularyImportInput struct {
	TargetLanguage string
	NativeLanguage string
	Entries        []VocabularyImportEntry
}

type VocabularyImportEntry struct {
	TargetText     string
	EntryType      string
	Pronunciations []VocabularyImportPronunciation
	Meanings       []VocabularyImportMeaning
	RelatedPhrases []VocabularyImportRelatedPhrase
	Examples       []VocabularyImportExample
	Tags           []string
	Notes          string
}

type VocabularyImportPronunciation struct {
	Pronunciation     string
	PronunciationType string
	Region            string
	AudioURL          string
}

type VocabularyImportMeaning struct {
	NativeText   string
	PartOfSpeech string
}

type VocabularyImportRelatedPhrase struct {
	TargetText string
	Meanings   []VocabularyImportMeaning
}

type VocabularyImportExample struct {
	TargetText string
	NativeText string
}

func DecodeVocabularyImport(data []byte) (*VocabularyImportInput, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, fmt.Errorf("%w: import body is empty", ErrVocabularyInvalidImport)
	}
	if data[0] == '[' {
		return decodeLegacyVocabularyImport(data)
	}
	if data[0] != '{' {
		return nil, fmt.Errorf("%w: import must be a JSON object or array", ErrVocabularyInvalidImport)
	}

	var probe struct {
		Format string `json:"format"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return nil, fmt.Errorf("%w: invalid JSON: %v", ErrVocabularyInvalidImport, err)
	}
	if probe.Format == "" {
		return decodeLegacyVocabularyImport(data)
	}
	if probe.Format == LearnLangVocabularyFormat {
		return decodeLearnLangVocabularyImport(data)
	}
	return nil, fmt.Errorf("%w: unsupported format %q", ErrVocabularyInvalidImport, probe.Format)
}
