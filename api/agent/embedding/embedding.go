package embedding

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Config struct {
	APIKey      string
	APIBaseURL  string
	Model       string
	FallbackKey string
	FallbackURL string
}

func Create(ctx context.Context, cfg Config, text string) ([]float32, error) {
	apiKey := strings.TrimSpace(cfg.APIKey)
	if apiKey == "" {
		apiKey = strings.TrimSpace(cfg.FallbackKey)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("embedding api key is required")
	}

	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	apiBaseURL := strings.TrimSpace(cfg.APIBaseURL)
	if apiBaseURL == "" {
		apiBaseURL = strings.TrimSpace(cfg.FallbackURL)
	}
	if apiBaseURL != "" {
		opts = append(opts, option.WithBaseURL(apiBaseURL))
	}

	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = "text-embedding-3-small"
	}

	client := openai.NewClient(opts...)
	resp, err := client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model: openai.EmbeddingModel(model),
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: []string{text},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("embedding response is empty")
	}

	vector := make([]float32, 0, len(resp.Data[0].Embedding))
	for _, value := range resp.Data[0].Embedding {
		vector = append(vector, float32(value))
	}
	return vector, nil
}
