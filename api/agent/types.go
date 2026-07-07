package agent

type ChatRequest struct {
	UserID      int64
	UserInput   string
	Settings    UserSettings
	Instant     bool
	CurrentTime string
	Timezone    string
}

type UserSettings struct {
	APIKey              string
	APIBaseURL          string
	Model               string
	EmbeddingAPIKey     string
	EmbeddingAPIBaseURL string
	EmbeddingModel      string
	NativeLanguage      string
	TargetLanguage      string
}

type ChatResult struct {
	ReplySentences   []Sentence    `json:"reply_sentences"`
	DetectedLanguage string        `json:"detected_language"`
	Memory           *MemoryInfo   `json:"memory"`
	Summary          *SummaryInfo  `json:"summary"`
	Function         *FunctionInfo `json:"function"`
	WaitForNextMsg   bool          `json:"wait_for_next_message"`
	TokensUsed       int           `json:"-"`
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

type FunctionInfo struct {
	CallFunction bool                   `json:"call_function"`
	FunctionName string                 `json:"function_name"`
	FunctionArgs map[string]interface{} `json:"function_args"`
}

type SummaryInfo struct {
	ShouldUpdate bool   `json:"should_update"`
	Content      string `json:"content"`
}
