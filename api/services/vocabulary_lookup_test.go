package services

import (
	"reflect"
	"testing"
)

func TestSplitAndBuildVocabularyLookupCandidates(t *testing.T) {
	runes := []rune("Go can't 学中文!")
	tokens := splitVocabularyLookupTokens(runes)
	if len(tokens) != 5 {
		t.Fatalf("splitVocabularyLookupTokens() = %#v", tokens)
	}
	candidates, err := buildVocabularyLookupCandidates(runes)
	if err != nil {
		t.Fatal(err)
	}
	normalized := make([]string, len(candidates))
	for index := range candidates {
		normalized[index] = candidates[index].normalized
	}
	want := []string{"go", "can't", "学", "学中", "学中文", "中", "中文", "文"}
	if !reflect.DeepEqual(normalized, want) {
		t.Fatalf("candidates = %#v, want %#v", normalized, want)
	}
}

func TestFindVocabularyTextOccurrencesHonorsBoundaries(t *testing.T) {
	got := findVocabularyTextOccurrences([]rune("go gopher GO"), []rune("go"))
	if len(got) != 2 || got[0].start != 0 || got[1].start != 10 {
		t.Fatalf("word occurrences = %#v", got)
	}
	cjk := findVocabularyTextOccurrences([]rune("中文中文"), []rune("中文"))
	if len(cjk) != 2 || cjk[1].start != 2 {
		t.Fatalf("CJK occurrences = %#v", cjk)
	}
	if got := findVocabularyTextOccurrences([]rune("abc"), nil); got != nil {
		t.Fatalf("empty target occurrences = %#v", got)
	}
}

func TestSelectNonOverlappingVocabularyMatchesPrefersLongest(t *testing.T) {
	long := &vocabularyRuneMatch{start: 0, end: 3}
	short := &vocabularyRuneMatch{start: 0, end: 1}
	last := &vocabularyRuneMatch{start: 3, end: 4}
	matches := map[vocabularyLookupSpan]*vocabularyRuneMatch{
		{start: 0, end: 3}: long,
		{start: 0, end: 1}: short,
		{start: 3, end: 4}: last,
	}
	got := selectNonOverlappingVocabularyMatches(matches)
	if !reflect.DeepEqual(got, []*vocabularyRuneMatch{long, last}) {
		t.Fatalf("selected = %#v", got)
	}
	if got := selectNonOverlappingVocabularyMatches(nil); len(got) != 0 {
		t.Fatalf("empty selected = %#v", got)
	}
}

func TestVocabularyUTF16OffsetsAndRuneClassification(t *testing.T) {
	if got := vocabularyUTF16Offsets([]rune("A😀中")); !reflect.DeepEqual(got, []int{0, 1, 3, 4}) {
		t.Fatalf("vocabularyUTF16Offsets() = %#v", got)
	}
	if !isCJKLookupRune('中') || isCJKLookupRune('A') {
		t.Fatal("isCJKLookupRune() classification is wrong")
	}
	if !isVocabularyWordRune('é') || !isVocabularyApostrophe('’') || !containsBoundarySensitiveLookupRune([]rune("Go")) || containsBoundarySensitiveLookupRune([]rune("中文")) {
		t.Fatal("lookup rune helpers returned unexpected classifications")
	}
	if !isVocabularyLookupContinuation('a') || isVocabularyLookupContinuation('中') {
		t.Fatal("continuation classification is wrong")
	}
}
