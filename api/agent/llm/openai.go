package llm

import (
	"context"
	"fmt"
	"strings"

	einoopenai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

func newOpenAI(ctx context.Context, apiKey, apiBaseURL, modelName string, options Options) (model.ToolCallingChatModel, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("openai api key is required")
	}

	modelName = strings.TrimSpace(modelName)
	if modelName == "" {
		modelName = "gpt-4o-mini"
	}

	config := &einoopenai.ChatModelConfig{
		APIKey:      apiKey,
		BaseURL:     strings.TrimSpace(apiBaseURL),
		Model:       modelName,
		Temperature: options.Temperature,
	}
	if options.MaxTokens > 0 {
		maxTokens := options.MaxTokens
		config.MaxTokens = &maxTokens
	}
	return einoopenai.NewChatModel(ctx, config)
}
