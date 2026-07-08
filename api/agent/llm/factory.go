package llm

import (
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

const (
	TypeOpenAI    = "openai"
	TypeAnthropic = "anthropic"
)

func New(apiKey, apiBaseURL, model, llmType string) (llms.Model, error) {
	switch normalizeType(llmType) {
	case "", TypeOpenAI:
		return newOpenAI(apiKey, apiBaseURL, model)
	case TypeAnthropic:
		return newClaudeCode(apiKey, apiBaseURL, model)
	default:
		return nil, fmt.Errorf("unsupported llm type: %s", llmType)
	}
}

func normalizeType(llmType string) string {
	return strings.ToLower(strings.TrimSpace(llmType))
}
