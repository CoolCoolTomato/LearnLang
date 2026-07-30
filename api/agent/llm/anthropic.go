package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino/components/model"
)

func newClaudeCode(ctx context.Context, apiKey, apiBaseURL, modelName string, options Options) (model.ToolCallingChatModel, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("claudecode api key is required")
	}

	modelName = strings.TrimSpace(modelName)
	if modelName == "" {
		modelName = "claude-sonnet-4-5"
	}

	maxTokens := options.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}
	config := &claude.Config{
		APIKey:      apiKey,
		Model:       modelName,
		MaxTokens:   maxTokens,
		Temperature: options.Temperature,
	}
	if baseURL := strings.TrimSpace(apiBaseURL); baseURL != "" {
		config.BaseURL = &baseURL
	}
	return claude.NewChatModel(ctx, config)
}
