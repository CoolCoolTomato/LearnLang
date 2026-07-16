package services

import (
	"context"
	"fmt"
	"learnlang-api/database"
	"learnlang-api/models"
	"sort"
	"strings"
	"unicode"

	"gorm.io/gorm"
)

const (
	maxVocabularyLookupRunes      = 5000
	maxVocabularyLookupCandidates = 2000
	maxVocabularyLookupEntries    = 2000
	maxCJKCandidateRunes          = 8
)

type VocabularyLookupResult struct {
	MessageID int64                   `json:"message_id"`
	Text      string                  `json:"text"`
	Matches   []VocabularyLookupMatch `json:"matches"`
}

type VocabularyLookupMatch struct {
	Start   int                     `json:"start"`
	End     int                     `json:"end"`
	Text    string                  `json:"text"`
	Entries []VocabularyLookupEntry `json:"entries"`
}

type VocabularyLookupEntry struct {
	VocabularyID   int64                  `json:"vocabulary_id"`
	VocabularyName string                 `json:"vocabulary_name"`
	Entry          models.VocabularyEntry `json:"entry"`
}

type vocabularyLookupCandidate struct {
	start      int
	end        int
	normalized string
}

type vocabularyLookupToken struct {
	start int
	end   int
	cjk   bool
}

type vocabularyRuneMatch struct {
	start   int
	end     int
	entries []VocabularyLookupEntry
	seen    map[int64]struct{}
}

type vocabularyLookupSpan struct {
	start int
	end   int
}

func (s *VocabularyService) LookupMessage(ctx context.Context, userID, messageID int64) (*VocabularyLookupResult, error) {
	var message models.Message
	operation := database.DB.WithContext(ctx).
		Where("id = ? AND user_id = ?", messageID, userID).
		Limit(1).
		Find(&message)
	if operation.Error != nil {
		return nil, operation.Error
	}
	if operation.RowsAffected == 0 {
		return nil, ErrVocabularyMessageNotFound
	}

	text := message.TextContent
	result := &VocabularyLookupResult{
		MessageID: message.ID,
		Text:      text,
		Matches:   make([]VocabularyLookupMatch, 0),
	}
	if strings.TrimSpace(text) == "" {
		return result, nil
	}

	runes := []rune(text)
	if len(runes) > maxVocabularyLookupRunes {
		return nil, fmt.Errorf("%w: message is too long to query", ErrVocabularyInvalidInput)
	}
	candidates, err := buildVocabularyLookupCandidates(runes)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return result, nil
	}

	var vocabularies []models.Vocabulary
	if err := database.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&vocabularies).Error; err != nil {
		return nil, err
	}
	if len(vocabularies) == 0 {
		return result, nil
	}

	vocabularyIDs := make([]int64, 0, len(vocabularies))
	vocabularyNames := make(map[int64]string, len(vocabularies))
	for _, vocabulary := range vocabularies {
		vocabularyIDs = append(vocabularyIDs, vocabulary.ID)
		vocabularyNames[vocabulary.ID] = vocabulary.Name
	}

	candidateOccurrences := make(map[string][]vocabularyLookupCandidate)
	normalizedCandidates := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if _, exists := candidateOccurrences[candidate.normalized]; !exists {
			normalizedCandidates = append(normalizedCandidates, candidate.normalized)
		}
		candidateOccurrences[candidate.normalized] = append(candidateOccurrences[candidate.normalized], candidate)
	}

	var entries []models.VocabularyEntry
	query := database.DB.WithContext(ctx).
		Where("vocabulary_id IN ? AND normalized_target_text IN ?", vocabularyIDs, normalizedCandidates).
		Limit(maxVocabularyLookupEntries + 1)
	query = preloadVocabularyLookupEntry(query)
	if err := query.Find(&entries).Error; err != nil {
		return nil, err
	}
	if len(entries) > maxVocabularyLookupEntries {
		return nil, fmt.Errorf("%w: too many matching vocabulary entries", ErrVocabularyInvalidInput)
	}

	matches := make(map[vocabularyLookupSpan]*vocabularyRuneMatch)
	addMatch := func(start, end int, entry models.VocabularyEntry) {
		key := vocabularyLookupSpan{start: start, end: end}
		match, exists := matches[key]
		if !exists {
			match = &vocabularyRuneMatch{start: start, end: end, seen: make(map[int64]struct{})}
			matches[key] = match
		}
		if _, exists := match.seen[entry.ID]; exists {
			return
		}
		match.seen[entry.ID] = struct{}{}
		match.entries = append(match.entries, VocabularyLookupEntry{
			VocabularyID:   entry.VocabularyID,
			VocabularyName: vocabularyNames[entry.VocabularyID],
			Entry:          entry,
		})
	}

	for _, entry := range entries {
		for _, occurrence := range candidateOccurrences[entry.NormalizedTargetText] {
			addMatch(occurrence.start, occurrence.end, entry)
		}
		for _, relation := range entry.Relations {
			if relation.RelationType != models.VocabularyRelationPhrase || relation.RelatedEntry == nil {
				continue
			}
			for _, occurrence := range findVocabularyTextOccurrences(runes, []rune(relation.RelatedEntry.TargetText)) {
				addMatch(occurrence.start, occurrence.end, *relation.RelatedEntry)
				addMatch(occurrence.start, occurrence.end, entry)
			}
		}
	}

	selected := selectNonOverlappingVocabularyMatches(matches)
	utf16Offsets := vocabularyUTF16Offsets(runes)
	for _, match := range selected {
		sort.Slice(match.entries, func(i, j int) bool {
			if match.entries[i].VocabularyID == match.entries[j].VocabularyID {
				return match.entries[i].Entry.ID < match.entries[j].Entry.ID
			}
			return match.entries[i].VocabularyID < match.entries[j].VocabularyID
		})
		result.Matches = append(result.Matches, VocabularyLookupMatch{
			Start:   utf16Offsets[match.start],
			End:     utf16Offsets[match.end],
			Text:    string(runes[match.start:match.end]),
			Entries: match.entries,
		})
	}
	return result, nil
}

func preloadVocabularyLookupEntry(query *gorm.DB) *gorm.DB {
	return query.
		Preload("Pronunciations", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Preload("Meanings", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Preload("Examples", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Preload("Relations", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Preload("Relations.RelatedEntry.Pronunciations", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Preload("Relations.RelatedEntry.Meanings", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Preload("Relations.RelatedEntry.Examples", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") })
}

func buildVocabularyLookupCandidates(runes []rune) ([]vocabularyLookupCandidate, error) {
	tokens := splitVocabularyLookupTokens(runes)
	candidates := make([]vocabularyLookupCandidate, 0, len(tokens))
	seen := make(map[vocabularyLookupSpan]struct{})
	add := func(start, end int) error {
		normalized := normalizeVocabularyText(string(runes[start:end]))
		if normalized == "" {
			return nil
		}
		key := vocabularyLookupSpan{start: start, end: end}
		if _, exists := seen[key]; exists {
			return nil
		}
		if len(candidates) >= maxVocabularyLookupCandidates {
			return fmt.Errorf("%w: message contains too many lookup candidates", ErrVocabularyInvalidInput)
		}
		seen[key] = struct{}{}
		candidates = append(candidates, vocabularyLookupCandidate{start: start, end: end, normalized: normalized})
		return nil
	}

	for index, token := range tokens {
		if err := add(token.start, token.end); err != nil {
			return nil, err
		}
		if !token.cjk {
			continue
		}
		end := token.end
		for next := index + 1; next < len(tokens) && next < index+maxCJKCandidateRunes; next++ {
			if !tokens[next].cjk || tokens[next].start != end {
				break
			}
			end = tokens[next].end
			if err := add(token.start, end); err != nil {
				return nil, err
			}
		}
	}
	return candidates, nil
}

func splitVocabularyLookupTokens(runes []rune) []vocabularyLookupToken {
	tokens := make([]vocabularyLookupToken, 0)
	for index := 0; index < len(runes); {
		if isCJKLookupRune(runes[index]) {
			tokens = append(tokens, vocabularyLookupToken{start: index, end: index + 1, cjk: true})
			index++
			continue
		}
		if !isVocabularyWordRune(runes[index]) {
			index++
			continue
		}
		start := index
		index++
		for index < len(runes) {
			if isCJKLookupRune(runes[index]) {
				break
			}
			if isVocabularyWordRune(runes[index]) {
				index++
				continue
			}
			if isVocabularyApostrophe(runes[index]) && index+1 < len(runes) && isVocabularyWordRune(runes[index+1]) {
				index++
				continue
			}
			break
		}
		tokens = append(tokens, vocabularyLookupToken{start: start, end: index})
	}
	return tokens
}

func findVocabularyTextOccurrences(text, target []rune) []vocabularyLookupCandidate {
	if len(target) == 0 || len(target) > len(text) {
		return nil
	}
	boundarySensitive := containsBoundarySensitiveLookupRune(target)
	result := make([]vocabularyLookupCandidate, 0)
	for start := 0; start+len(target) <= len(text); start++ {
		end := start + len(target)
		if !strings.EqualFold(string(text[start:end]), string(target)) {
			continue
		}
		if boundarySensitive {
			if start > 0 && isVocabularyLookupContinuation(text[start-1]) {
				continue
			}
			if end < len(text) && isVocabularyLookupContinuation(text[end]) {
				continue
			}
		}
		result = append(result, vocabularyLookupCandidate{start: start, end: end})
	}
	return result
}

func selectNonOverlappingVocabularyMatches(matches map[vocabularyLookupSpan]*vocabularyRuneMatch) []*vocabularyRuneMatch {
	candidates := make([]*vocabularyRuneMatch, 0, len(matches))
	for _, match := range matches {
		candidates = append(candidates, match)
	}
	sort.Slice(candidates, func(i, j int) bool {
		leftLength := candidates[i].end - candidates[i].start
		rightLength := candidates[j].end - candidates[j].start
		if leftLength == rightLength {
			return candidates[i].start < candidates[j].start
		}
		return leftLength > rightLength
	})

	maxEnd := 0
	for _, candidate := range candidates {
		if candidate.end > maxEnd {
			maxEnd = candidate.end
		}
	}
	occupied := make([]bool, maxEnd)
	selected := make([]*vocabularyRuneMatch, 0, len(candidates))
	for _, candidate := range candidates {
		overlaps := false
		for index := candidate.start; index < candidate.end; index++ {
			if occupied[index] {
				overlaps = true
				break
			}
		}
		if !overlaps {
			selected = append(selected, candidate)
			for index := candidate.start; index < candidate.end; index++ {
				occupied[index] = true
			}
		}
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].start < selected[j].start })
	return selected
}

func vocabularyUTF16Offsets(runes []rune) []int {
	offsets := make([]int, len(runes)+1)
	for index, value := range runes {
		width := 1
		if value > 0xffff {
			width = 2
		}
		offsets[index+1] = offsets[index] + width
	}
	return offsets
}

func isCJKLookupRune(value rune) bool {
	return unicode.In(value, unicode.Han, unicode.Hiragana, unicode.Katakana, unicode.Hangul)
}

func isVocabularyWordRune(value rune) bool {
	return unicode.IsLetter(value) || unicode.IsNumber(value) || unicode.IsMark(value)
}

func isVocabularyApostrophe(value rune) bool {
	return value == '\'' || value == '’'
}

func containsBoundarySensitiveLookupRune(values []rune) bool {
	for _, value := range values {
		if (unicode.IsLetter(value) || unicode.IsNumber(value)) && !isCJKLookupRune(value) {
			return true
		}
	}
	return false
}

func isVocabularyLookupContinuation(value rune) bool {
	return (!isCJKLookupRune(value) && isVocabularyWordRune(value)) || isVocabularyApostrophe(value)
}
