package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"learnlang-api/models"
)

type LegacyVocabularyImportDocument struct {
	Entries []LegacyVocabularyImportEntry `json:"entries"`
}

type LegacyVocabularyImportEntry struct {
	Word           string                                `json:"word"`
	US             string                                `json:"us"`
	UK             string                                `json:"uk"`
	Pronunciations []LegacyVocabularyImportPronunciation `json:"pronunciations"`
	Translations   []LegacyVocabularyImportTranslation   `json:"translations"`
	Phrases        []LegacyVocabularyImportPhrase        `json:"phrases"`
	Sentences      []LegacyVocabularyImportSentence      `json:"sentences"`
	Tags           []string                              `json:"tags"`
	Notes          string                                `json:"notes"`
}

type LegacyVocabularyImportPronunciation struct {
	Pronunciation     string `json:"pronunciation"`
	PronunciationType string `json:"type"`
	Region            string `json:"region"`
	AudioURL          string `json:"audio_url"`
}

type LegacyVocabularyImportTranslation struct {
	Translation string `json:"translation"`
	Type        string `json:"type"`
}

type LegacyVocabularyImportPhrase struct {
	Phrase      string `json:"phrase"`
	Translation string `json:"translation"`
	Type        string `json:"type"`
}

type LegacyVocabularyImportSentence struct {
	Sentence    string `json:"sentence"`
	Translation string `json:"translation"`
}

func decodeLegacyVocabularyImport(data []byte) (*VocabularyImportInput, error) {
	var entries []LegacyVocabularyImportEntry
	if bytes.HasPrefix(data, []byte("[")) {
		if err := json.Unmarshal(data, &entries); err != nil {
			return nil, fmt.Errorf("%w: invalid legacy entry array: %v", ErrVocabularyInvalidImport, err)
		}
	} else {
		var document LegacyVocabularyImportDocument
		if err := json.Unmarshal(data, &document); err != nil {
			return nil, fmt.Errorf("%w: invalid legacy document: %v", ErrVocabularyInvalidImport, err)
		}
		if len(document.Entries) > 0 {
			entries = document.Entries
		} else {
			var entry LegacyVocabularyImportEntry
			if err := json.Unmarshal(data, &entry); err != nil || strings.TrimSpace(entry.Word) == "" {
				return nil, fmt.Errorf("%w: legacy format requires word or entries", ErrVocabularyInvalidImport)
			}
			entries = []LegacyVocabularyImportEntry{entry}
		}
	}

	input := &VocabularyImportInput{Entries: make([]VocabularyImportEntry, 0, len(entries))}
	for _, entry := range entries {
		input.Entries = append(input.Entries, convertLegacyVocabularyEntry(entry))
	}
	return input, nil
}

func convertLegacyVocabularyEntry(entry LegacyVocabularyImportEntry) VocabularyImportEntry {
	converted := VocabularyImportEntry{
		TargetText: entry.Word,
		EntryType:  models.VocabularyEntryTypeWord,
		Tags:       entry.Tags,
		Notes:      entry.Notes,
	}
	for _, meaning := range entry.Translations {
		converted.Meanings = append(converted.Meanings, VocabularyImportMeaning{
			NativeText:   meaning.Translation,
			PartOfSpeech: meaning.Type,
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
	if strings.TrimSpace(entry.US) != "" {
		converted.Pronunciations = append(converted.Pronunciations, VocabularyImportPronunciation{
			Pronunciation:     entry.US,
			PronunciationType: "ipa",
			Region:            "us",
		})
	}
	if strings.TrimSpace(entry.UK) != "" {
		converted.Pronunciations = append(converted.Pronunciations, VocabularyImportPronunciation{
			Pronunciation:     entry.UK,
			PronunciationType: "ipa",
			Region:            "uk",
		})
	}
	for _, phrase := range entry.Phrases {
		converted.RelatedPhrases = append(converted.RelatedPhrases, VocabularyImportRelatedPhrase{
			TargetText: phrase.Phrase,
			Meanings: []VocabularyImportMeaning{{
				NativeText:   phrase.Translation,
				PartOfSpeech: phrase.Type,
			}},
		})
	}
	for _, sentence := range entry.Sentences {
		converted.Examples = append(converted.Examples, VocabularyImportExample{
			TargetText: sentence.Sentence,
			NativeText: sentence.Translation,
		})
	}
	return converted
}
