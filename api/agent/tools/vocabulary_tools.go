package tools

import (
	"context"
	"encoding/json"
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

type vocabularyToolArgs struct {
	Count int `json:"count"`
}

type vocabularyToolVocabulary struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	TargetLanguage string `json:"target_language"`
	NativeLanguage string `json:"native_language"`
}

type vocabularyToolSelectionEntry struct {
	Vocabulary vocabularyToolVocabulary `json:"vocabulary"`
	Entry      vocabularyToolEntry      `json:"entry"`
}

type RandomNewVocabularyWordTool struct {
	UserID           int64
	RequestMessageID int64
	TargetLanguage   string
	NativeLanguage   string
	Vocabulary       *services.VocabularyService
	State            *TurnState
}

func (RandomNewVocabularyWordTool) Name() string {
	return "get_random_new_vocabulary_word"
}

func (RandomNewVocabularyWordTool) Description() string {
	return `Get a batch of random unseen words from the user's vocabulary for the current target/native language pair. Input is {"count": integer}; count defaults to 1 and is capped at 5. Call once with the full quantity the user requested, never repeatedly to accumulate words. Selected words are atomically marked as encountered. Repeated calls in the same user turn replay the first batch.`
}

func (t RandomNewVocabularyWordTool) Call(ctx context.Context, input string) (string, error) {
	return callRandomVocabularyWords(ctx, input, vocabularyToolCallConfig{
		userID: t.UserID, requestMessageID: t.RequestMessageID,
		targetLanguage: t.TargetLanguage, nativeLanguage: t.NativeLanguage,
		selectionType: models.VocabularyAgentSelectionNew,
		emptyReason:   "no unseen words are available for the current language pair",
		vocabulary:    t.Vocabulary, state: t.State,
	})
}

type RandomOldVocabularyWordTool struct {
	UserID           int64
	RequestMessageID int64
	TargetLanguage   string
	NativeLanguage   string
	Vocabulary       *services.VocabularyService
	State            *TurnState
}

func (RandomOldVocabularyWordTool) Name() string {
	return "get_random_old_vocabulary_word"
}

func (RandomOldVocabularyWordTool) Description() string {
	return `Get a batch of random previously encountered words from the user's vocabulary for review. Input is {"count": integer}; count defaults to 1 and is capped at 5. Call once with the full quantity the user requested, never repeatedly to accumulate words. This tool does not change encounter statistics. Repeated calls in the same user turn replay the first batch.`
}

func (t RandomOldVocabularyWordTool) Call(ctx context.Context, input string) (string, error) {
	return callRandomVocabularyWords(ctx, input, vocabularyToolCallConfig{
		userID: t.UserID, requestMessageID: t.RequestMessageID,
		targetLanguage: t.TargetLanguage, nativeLanguage: t.NativeLanguage,
		selectionType: models.VocabularyAgentSelectionOld,
		emptyReason:   "no previously encountered words are available for the current language pair",
		vocabulary:    t.Vocabulary, state: t.State,
	})
}

type vocabularyToolCallConfig struct {
	userID           int64
	requestMessageID int64
	targetLanguage   string
	nativeLanguage   string
	selectionType    string
	emptyReason      string
	vocabulary       *services.VocabularyService
	state            *TurnState
}

func callRandomVocabularyWords(ctx context.Context, input string, cfg vocabularyToolCallConfig) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if cfg.vocabulary == nil {
		return "", fmt.Errorf("vocabulary service is not configured")
	}
	if cfg.state != nil {
		if cached, ok := cfg.state.VocabularyToolResult(cfg.selectionType); ok {
			return cached, nil
		}
	}

	count, invalidResult, err := parseVocabularyToolCount(input)
	if err != nil {
		return "", err
	}
	if invalidResult != "" {
		return invalidResult, nil
	}

	selection, err := cfg.vocabulary.SelectAgentVocabularyWords(ctx, cfg.userID, cfg.requestMessageID, cfg.selectionType, cfg.targetLanguage, cfg.nativeLanguage, count)
	if err != nil {
		return "", err
	}
	result, err := marshalVocabularyWordsResult(selection.RequestedCount, selection.Entries, cfg.emptyReason)
	if err != nil {
		return "", err
	}
	if cfg.state != nil {
		cfg.state.SetVocabularyToolResult(cfg.selectionType, result)
	}
	return result, nil
}

func parseVocabularyToolCount(input string) (int, string, error) {
	args := vocabularyToolArgs{Count: 1}
	if input != "" {
		if err := json.Unmarshal([]byte(input), &args); err != nil {
			result, marshalErr := marshalToolResult(map[string]any{"status": "invalid_request", "error": "input must be a JSON object with an optional integer count"})
			return 0, result, marshalErr
		}
	}
	if args.Count < 1 {
		result, err := marshalToolResult(map[string]any{"status": "invalid_request", "error": "count must be at least 1"})
		return 0, result, err
	}
	if args.Count > services.MaxAgentVocabularyWords {
		args.Count = services.MaxAgentVocabularyWords
	}
	return args.Count, "", nil
}

func marshalVocabularyWordsResult(requestedCount int, results []services.VocabularyRandomEntry, emptyReason string) (string, error) {
	entries := make([]vocabularyToolSelectionEntry, 0, len(results))
	for i := range results {
		entries = append(entries, makeVocabularyToolEntry(&results[i]))
	}
	status := "found"
	if len(entries) == 0 {
		status = "empty"
	}
	payload := map[string]any{
		"status":          status,
		"requested_count": requestedCount,
		"actual_count":    len(entries),
		"entries":         entries,
	}
	if len(entries) == 0 {
		payload["reason"] = emptyReason
	}
	return marshalToolResult(payload)
}

func makeVocabularyToolEntry(result *services.VocabularyRandomEntry) vocabularyToolSelectionEntry {
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

	return vocabularyToolSelectionEntry{
		Vocabulary: vocabularyToolVocabulary{
			ID: result.Vocabulary.ID, Name: result.Vocabulary.Name,
			TargetLanguage: result.Vocabulary.TargetLanguage, NativeLanguage: result.Vocabulary.NativeLanguage,
		},
		Entry: entry,
	}
}
