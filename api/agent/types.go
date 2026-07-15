package agent

import agenttools "learnlang-api/agent/tools"

type ChatRequest struct {
	UserID                 int64
	ContextBeforeMessageID int64
	UserInput              string
	Settings               UserSettings
	CurrentTime            string
	Timezone               string
}

type UserSettings struct {
	APIKey              string
	APIBaseURL          string
	Model               string
	LLMType             string
	EmbeddingAPIKey     string
	EmbeddingAPIBaseURL string
	EmbeddingModel      string
	NativeLanguage      string
	TargetLanguage      string
}

type ChatResult = agenttools.ChatResult
type Sentence = agenttools.Sentence
