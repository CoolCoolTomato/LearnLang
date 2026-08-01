package llm

import (
	"context"
	"fmt"
	"learnlang-api/models"
	"strings"

	"github.com/cloudwego/eino/components/model"
)

const (
	TypeOpenAI    = models.LLMTypeOpenAI
	TypeAnthropic = models.LLMTypeAnthropic
)

type Options struct {
	MaxTokens   int
	Temperature *float32
}

func New(ctx context.Context, apiKey, apiBaseURL, modelName, llmType string) (model.ToolCallingChatModel, error) {
	return NewWithOptions(ctx, apiKey, apiBaseURL, modelName, llmType, Options{})
}

func NewWithOptions(ctx context.Context, apiKey, apiBaseURL, modelName, llmType string, options Options) (model.ToolCallingChatModel, error) {
	switch normalizeType(llmType) {
	case "", TypeOpenAI:
		return newOpenAI(ctx, apiKey, apiBaseURL, modelName, options)
	case TypeAnthropic:
		return newClaudeCode(ctx, apiKey, apiBaseURL, modelName, options)
	default:
		return nil, fmt.Errorf("unsupported llm type: %s", llmType)
	}
}

func normalizeType(llmType string) string {
	return strings.ToLower(strings.TrimSpace(llmType))
}
