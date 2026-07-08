package llm

import (
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func newOpenAI(apiKey, apiBaseURL, model string) (llms.Model, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("openai api key is required")
	}

	model = strings.TrimSpace(model)
	if model == "" {
		model = "gpt-4o-mini"
	}

	opts := []openai.Option{
		openai.WithToken(apiKey),
		openai.WithModel(model),
	}
	if strings.TrimSpace(apiBaseURL) != "" {
		opts = append(opts, openai.WithBaseURL(apiBaseURL))
	}

	return openai.New(opts...)
}
