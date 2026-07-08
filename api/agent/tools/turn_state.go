package tools

type ChatResult struct {
	ReplySentences   []Sentence   `json:"reply_sentences"`
	DetectedLanguage string       `json:"detected_language"`
	Memory           *MemoryInfo  `json:"memory"`
	Summary          *SummaryInfo `json:"summary"`
	WaitForNextMsg   bool         `json:"wait_for_next_message"`
	TokensUsed       int          `json:"-"`
	MessageIDs       []int64      `json:"-"`
}

type Sentence struct {
	Original    string `json:"original"`
	Translation string `json:"translation"`
}

type MemoryInfo struct {
	ShouldStore     bool    `json:"should_store"`
	SemanticContent string  `json:"semantic_content"`
	Importance      float64 `json:"importance"`
	MemoryType      string  `json:"memory_type"`
	Language        string  `json:"language"`
}

type SummaryInfo struct {
	ShouldUpdate bool   `json:"should_update"`
	Content      string `json:"content"`
}

type TurnState struct {
	result ChatResult
}

func NewTurnState() *TurnState {
	return &TurnState{
		result: ChatResult{
			ReplySentences: []Sentence{},
			Memory:         &MemoryInfo{},
			Summary:        &SummaryInfo{},
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

func (s *TurnState) Complete(detectedLanguage string, waitForNextMessage bool, memory *MemoryInfo, summary *SummaryInfo) {
	s.result.DetectedLanguage = detectedLanguage
	s.result.WaitForNextMsg = waitForNextMessage
	if memory != nil {
		s.result.Memory = memory
	}
	if summary != nil {
		s.result.Summary = summary
	}
}
