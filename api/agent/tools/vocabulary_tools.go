package tools

import (
	"context"
	"fmt"
	"learnlang-api/models"
	"learnlang-api/services"
)

type vocabularyToolMeaning struct {
	Text         string `json:"text"`
	PartOfSpeech string `json:"part_of_speech,omitempty"`
}

type vocabularyToolPronunciation struct {
	Text     string `json:"text"`
	Type     string `json:"type"`
	Region   string `json:"region,omitempty"`
	AudioURL string `json:"audio_url,omitempty"`
}

type vocabularyToolExample struct {
	TargetText string `json:"target_text"`
	NativeText string `json:"native_text,omitempty"`
}

type vocabularyToolRelatedPhrase struct {
	TargetText string                  `json:"target_text"`
	Meanings   []vocabularyToolMeaning `json:"meanings"`
}

type vocabularyToolEntry struct {
	ID                      int64                         `json:"id"`
	TargetText              string                        `json:"target_text"`
	Meanings                []vocabularyToolMeaning       `json:"meanings"`
	Pronunciations          []vocabularyToolPronunciation `json:"pronunciations"`
	Examples                []vocabularyToolExample       `json:"examples"`
	RelatedPhrases          []vocabularyToolRelatedPhrase `json:"related_phrases"`
	RelatedPhraseCount      int64                         `json:"related_phrase_count"`
	RelatedPhrasesTruncated bool                          `json:"related_phrases_truncated"`
	Tags                    []string                      `json:"tags,omitempty"`
	Notes                   string                        `json:"notes,omitempty"`
	EncounterCount          int                           `json:"encounter_count"`
}

type RandomNewVocabularyWordTool struct {
	UserID         int64
	TargetLanguage string
	NativeLanguage string
	Vocabulary     *services.VocabularyService
}

func (RandomNewVocabularyWordTool) Name() string {
	return "get_random_new_vocabulary_word"
}

func (RandomNewVocabularyWordTool) Description() string {
	return `Get one random unseen word from the user's vocabulary for the current target/native language pair, including meanings, pronunciations, examples, notes, tags, and related phrases. The selected word is atomically marked as encountered. Use this when introducing a new vocabulary item. Input must be an empty JSON object: {}.`
}

func (t RandomNewVocabularyWordTool) Call(ctx context.Context, _ string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if t.Vocabulary == nil {
		return "", fmt.Errorf("vocabulary service is not configured")
	}

	result, err := t.Vocabulary.TakeRandomNewWord(ctx, t.UserID, t.TargetLanguage, t.NativeLanguage)
	if err != nil {
		return "", err
	}
	if result == nil {
		return marshalToolResult(map[string]any{
			"status": "empty",
			"reason": "no unseen words are available for the current language pair",
		})
	}
	return marshalVocabularyWordResult(result)
}

type RandomOldVocabularyWordTool struct {
	UserID         int64
	TargetLanguage string
	NativeLanguage string
	Vocabulary     *services.VocabularyService
}

func (RandomOldVocabularyWordTool) Name() string {
	return "get_random_old_vocabulary_word"
}

func (RandomOldVocabularyWordTool) Description() string {
	return `Get one random previously encountered word from the user's vocabulary for the current target/native language pair, including meanings, pronunciations, examples, notes, tags, and related phrases. This tool is read-only and does not change encounter statistics. Use this for review or practice. Input must be an empty JSON object: {}.`
}

func (t RandomOldVocabularyWordTool) Call(ctx context.Context, _ string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if t.Vocabulary == nil {
		return "", fmt.Errorf("vocabulary service is not configured")
	}

	result, err := t.Vocabulary.GetRandomOldWord(ctx, t.UserID, t.TargetLanguage, t.NativeLanguage)
	if err != nil {
		return "", err
	}
	if result == nil {
		return marshalToolResult(map[string]any{
			"status": "empty",
			"reason": "no previously encountered words are available for the current language pair",
		})
	}
	return marshalVocabularyWordResult(result)
}

func marshalVocabularyWordResult(result *services.VocabularyRandomEntry) (string, error) {
	entry := vocabularyToolEntry{
		ID:                      result.Entry.ID,
		TargetText:              result.Entry.TargetText,
		Meanings:                make([]vocabularyToolMeaning, 0, len(result.Entry.Meanings)),
		Pronunciations:          make([]vocabularyToolPronunciation, 0, len(result.Entry.Pronunciations)),
		Examples:                make([]vocabularyToolExample, 0, len(result.Entry.Examples)),
		RelatedPhrases:          make([]vocabularyToolRelatedPhrase, 0, len(result.Entry.Relations)),
		RelatedPhraseCount:      result.RelatedPhraseCount,
		RelatedPhrasesTruncated: result.RelatedPhraseCount > int64(len(result.Entry.Relations)),
		Tags:                    result.Entry.Tags,
		Notes:                   result.Entry.Notes,
		EncounterCount:          result.Entry.EncounterCount,
	}
	for _, meaning := range result.Entry.Meanings {
		entry.Meanings = append(entry.Meanings, vocabularyToolMeaning{
			Text:         meaning.NativeText,
			PartOfSpeech: meaning.PartOfSpeech,
		})
	}
	for _, pronunciation := range result.Entry.Pronunciations {
		entry.Pronunciations = append(entry.Pronunciations, vocabularyToolPronunciation{
			Text:     pronunciation.Pronunciation,
			Type:     pronunciation.PronunciationType,
			Region:   pronunciation.Region,
			AudioURL: pronunciation.AudioURL,
		})
	}
	for _, example := range result.Entry.Examples {
		entry.Examples = append(entry.Examples, vocabularyToolExample{
			TargetText: example.TargetText,
			NativeText: example.NativeText,
		})
	}
	for _, relation := range result.Entry.Relations {
		if relation.RelatedEntry == nil || relation.RelationType != models.VocabularyRelationPhrase {
			continue
		}
		phrase := vocabularyToolRelatedPhrase{
			TargetText: relation.RelatedEntry.TargetText,
			Meanings:   make([]vocabularyToolMeaning, 0, len(relation.RelatedEntry.Meanings)),
		}
		for _, meaning := range relation.RelatedEntry.Meanings {
			phrase.Meanings = append(phrase.Meanings, vocabularyToolMeaning{
				Text:         meaning.NativeText,
				PartOfSpeech: meaning.PartOfSpeech,
			})
		}
		entry.RelatedPhrases = append(entry.RelatedPhrases, phrase)
	}

	return marshalToolResult(map[string]any{
		"status": "found",
		"vocabulary": map[string]any{
			"id":              result.Vocabulary.ID,
			"name":            result.Vocabulary.Name,
			"target_language": result.Vocabulary.TargetLanguage,
			"native_language": result.Vocabulary.NativeLanguage,
		},
		"entry": entry,
	})
}
