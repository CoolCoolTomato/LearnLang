package tools

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
	result ChatResult
}

func NewTurnState() *TurnState {
	return &TurnState{
		result: ChatResult{
			ReplySentences: []Sentence{},
			MessageIDs:     []int64{},
		},
	}
}

func (s *TurnState) Result() *ChatResult {
	result := s.result
	return &result
}

func (s *TurnState) AddReply(original, translation string, messageID int64) {
	s.result.ReplySentences = append(s.result.ReplySentences, Sentence{
		Original:    original,
		Translation: translation,
	})
	s.result.MessageIDs = append(s.result.MessageIDs, messageID)
}

func (s *TurnState) Complete(detectedLanguage string) {
	s.result.DetectedLanguage = detectedLanguage
}
