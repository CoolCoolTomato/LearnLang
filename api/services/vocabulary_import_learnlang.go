package services

import (
	"encoding/json"
	"fmt"
	"strings"
)

type LearnLangVocabularyImportDocument struct {
	Format         string                           `json:"format"`
	Version        int                              `json:"version"`
	TargetLanguage string                           `json:"target_language"`
	NativeLanguage string                           `json:"native_language"`
	Entries        []LearnLangVocabularyImportEntry `json:"entries"`
}

type LearnLangVocabularyImportEntry struct {
	TargetText     string                                   `json:"target_text"`
	EntryType      string                                   `json:"entry_type"`
	Pronunciations []LearnLangVocabularyImportPronunciation `json:"pronunciations"`
	Meanings       []LearnLangVocabularyImportMeaning       `json:"meanings"`
	RelatedPhrases []LearnLangVocabularyImportRelatedPhrase `json:"related_phrases"`
	Examples       []LearnLangVocabularyImportExample       `json:"examples"`
	Tags           []string                                 `json:"tags"`
	Notes          string                                   `json:"notes"`
}

type LearnLangVocabularyImportPronunciation struct {
	Pronunciation     string `json:"pronunciation"`
	PronunciationType string `json:"type"`
	Region            string `json:"region"`
	AudioURL          string `json:"audio_url"`
}

type LearnLangVocabularyImportMeaning struct {
	NativeText   string `json:"native_text"`
	PartOfSpeech string `json:"part_of_speech"`
}

type LearnLangVocabularyImportRelatedPhrase struct {
	TargetText string                             `json:"target_text"`
	Meanings   []LearnLangVocabularyImportMeaning `json:"meanings"`
}

type LearnLangVocabularyImportExample struct {
	TargetText string `json:"target_text"`
	NativeText string `json:"native_text"`
}

func decodeLearnLangVocabularyImport(data []byte) (*VocabularyImportInput, error) {
	var document LearnLangVocabularyImportDocument
	if err := json.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("%w: invalid LearnLang document: %v", ErrVocabularyInvalidImport, err)
	}
	if document.Version != LearnLangVocabularyVersion {
		return nil, fmt.Errorf("%w: unsupported LearnLang vocabulary version %d", ErrVocabularyInvalidImport, document.Version)
	}
	if strings.TrimSpace(document.TargetLanguage) == "" || strings.TrimSpace(document.NativeLanguage) == "" {
		return nil, fmt.Errorf("%w: LearnLang format requires target_language and native_language", ErrVocabularyInvalidImport)
	}

	input := &VocabularyImportInput{
		TargetLanguage: strings.TrimSpace(document.TargetLanguage),
		NativeLanguage: strings.TrimSpace(document.NativeLanguage),
		Entries:        make([]VocabularyImportEntry, 0, len(document.Entries)),
	}
	for _, entry := range document.Entries {
		input.Entries = append(input.Entries, convertLearnLangVocabularyEntry(entry))
	}
	return input, nil
}

func convertLearnLangVocabularyEntry(entry LearnLangVocabularyImportEntry) VocabularyImportEntry {
	converted := VocabularyImportEntry{
		TargetText: entry.TargetText,
		EntryType:  entry.EntryType,
		Tags:       entry.Tags,
		Notes:      entry.Notes,
	}
	for _, meaning := range entry.Meanings {
		converted.Meanings = append(converted.Meanings, VocabularyImportMeaning{
			NativeText:   meaning.NativeText,
			PartOfSpeech: meaning.PartOfSpeech,
		})
	}
	for _, pronunciation := range entry.Pronunciations {
		converted.Pronunciations = append(converted.Pronunciations, VocabularyImportPronunciation{
			Pronunciation:     pronunciation.Pronunciation,
			PronunciationType: pronunciation.PronunciationType,
			Region:            pronunciation.Region,
			AudioURL:          pronunciation.AudioURL,
		})
	}
	for _, phrase := range entry.RelatedPhrases {
		convertedPhrase := VocabularyImportRelatedPhrase{TargetText: phrase.TargetText}
		for _, meaning := range phrase.Meanings {
			convertedPhrase.Meanings = append(convertedPhrase.Meanings, VocabularyImportMeaning{
				NativeText:   meaning.NativeText,
				PartOfSpeech: meaning.PartOfSpeech,
			})
		}
		converted.RelatedPhrases = append(converted.RelatedPhrases, convertedPhrase)
	}
	for _, example := range entry.Examples {
		converted.Examples = append(converted.Examples, VocabularyImportExample{
			TargetText: example.TargetText,
			NativeText: example.NativeText,
		})
	}
	return converted
}
