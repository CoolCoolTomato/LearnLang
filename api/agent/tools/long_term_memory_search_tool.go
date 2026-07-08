package tools

import (
	"context"
	"fmt"
	"learnlang-api/agent/memory"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type LongTermMemorySearchTool struct {
	UserID      int64
	Limit       int
	Timezone    string
	Store       *memory.Store
	APIKey      string
	APIBaseURL  string
	Model       string
	FallbackKey string
	FallbackURL string
}

func (t LongTermMemorySearchTool) Name() string {
	return "search_long_term_memory"
}

func (t LongTermMemorySearchTool) Description() string {
	return "Search long-term memory summaries and return linked chat records. Input should be the current user message or search query."
}

func (t LongTermMemorySearchTool) Call(ctx context.Context, input string) (string, error) {
	limit := t.Limit
	if limit <= 0 {
		limit = 3
	}

	if t.Store == nil {
		return marshalToolResult(map[string]any{
			"query":    input,
			"memories": []any{},
			"error":    "memory store is not configured",
		})
	}

	type memoryResult struct {
		ID              string   `json:"id"`
		Summary         string   `json:"summary"`
		Type            string   `json:"type"`
		ImportanceScore float64  `json:"importance_score"`
		Score           float32  `json:"score"`
		Messages        []string `json:"messages"`
	}

	embedding, err := t.createEmbedding(ctx, input)
	if err != nil {
		return marshalToolResult(map[string]any{
			"query":    input,
			"memories": []memoryResult{},
			"error":    err.Error(),
		})
	}

	memories, err := t.Store.Search(ctx, t.UserID, embedding, limit)
	if err != nil {
		return marshalToolResult(map[string]any{
			"query":    input,
			"memories": []memoryResult{},
			"error":    err.Error(),
		})
	}

	results := make([]memoryResult, 0, len(memories))
	for _, item := range memories {
		messages, err := linkedMessages(ctx, t.UserID, item.MessageIDs)
		if err != nil {
			return "", err
		}

		results = append(results, memoryResult{
			ID:              item.ID,
			Summary:         item.Summary,
			Type:            item.MemoryType,
			ImportanceScore: item.ImportanceScore,
			Score:           item.Score,
			Messages:        formatMessages(messages, t.Timezone),
		})
	}

	return marshalToolResult(map[string]any{
		"query":    input,
		"memories": results,
	})
}

func (t LongTermMemorySearchTool) createEmbedding(ctx context.Context, text string) ([]float32, error) {
	apiKey := strings.TrimSpace(t.APIKey)
	if apiKey == "" {
		apiKey = strings.TrimSpace(t.FallbackKey)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("embedding api key is required")
	}

	apiBaseURL := strings.TrimSpace(t.APIBaseURL)
	if apiBaseURL == "" {
		apiBaseURL = strings.TrimSpace(t.FallbackURL)
	}

	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if apiBaseURL != "" {
		opts = append(opts, option.WithBaseURL(apiBaseURL))
	}

	model := strings.TrimSpace(t.Model)
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
	for _, v := range resp.Data[0].Embedding {
		vector = append(vector, float32(v))
	}
	return vector, nil
}
