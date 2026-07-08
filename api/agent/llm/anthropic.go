package llm

import (
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

func newClaudeCode(apiKey, apiBaseURL, model string) (llms.Model, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("claudecode api key is required")
	}

	model = strings.TrimSpace(model)
	if model == "" {
		model = "claude-sonnet-4-5"
	}

	opts := []anthropic.Option{
		anthropic.WithToken(apiKey),
		anthropic.WithModel(model),
	}
	if strings.TrimSpace(apiBaseURL) != "" {
		opts = append(opts, anthropic.WithBaseURL(apiBaseURL))
	}

	return anthropic.New(opts...)
}
