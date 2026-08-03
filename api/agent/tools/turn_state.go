package tools

import "sync"

type ChatResult struct {
	ReplySentences   []Sentence `json:"reply_sentences"`
	DetectedLanguage string     `json:"detected_language"`
	TokensUsed       int        `json:"-"`
	MessageIDs       []int64    `json:"-"`
}

type Sentence struct {
	Original    string `json:"original"`
	Translation string `json:"translation"`
}

type TurnState struct {
	mu                    sync.Mutex
	result                ChatResult
	completed             bool
	vocabularyToolResults map[string]string
}

func NewTurnState() *TurnState {
	return &TurnState{
		result: ChatResult{
			ReplySentences: []Sentence{},
			MessageIDs:     []int64{},
		},
		vocabularyToolResults: make(map[string]string),
	}
}

func (s *TurnState) Result() *ChatResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := s.result
	result.ReplySentences = append([]Sentence(nil), s.result.ReplySentences...)
	result.MessageIDs = append([]int64(nil), s.result.MessageIDs...)
	return &result
}

func (s *TurnState) AddReply(original, translation string, messageID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.result.ReplySentences = append(s.result.ReplySentences, Sentence{
		Original:    original,
		Translation: translation,
	})
	s.result.MessageIDs = append(s.result.MessageIDs, messageID)
}

func (s *TurnState) Complete(detectedLanguage string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.completed {
		return false
	}
	s.result.DetectedLanguage = detectedLanguage
	s.completed = true
	return true
}

func (s *TurnState) IsCompleted() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.completed
}

func (s *TurnState) VocabularyToolResult(selectionType string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result, ok := s.vocabularyToolResults[selectionType]
	return result, ok
}

func (s *TurnState) SetVocabularyToolResult(selectionType, result string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.vocabularyToolResults[selectionType]; !exists {
		s.vocabularyToolResults[selectionType] = result
	}
}
