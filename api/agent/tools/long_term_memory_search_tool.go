package tools

import (
	"context"
	"learnlang-api/agent/embedding"
	"learnlang-api/agent/memory"
	"learnlang-api/database"
	"learnlang-api/models"
	"log"
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
	return "Search long-term memory summaries and return linked chat records. Input must be a standalone semantic search query describing the memory needed. Resolve pronouns and omitted context using the current conversation, and include exact project names, people, technologies, errors, paths, commands, goals, decisions, or constraints that should match the memory. Do not pass a vague latest message verbatim."
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

	memories, err := t.Store.Search(ctx, t.UserID, embedding, limit*4)
	if err != nil {
		return marshalToolResult(map[string]any{
			"query":    input,
			"memories": []memoryResult{},
			"error":    err.Error(),
		})
	}

	embeddingIDs := make([]string, 0, len(memories))
	for _, item := range memories {
		embeddingIDs = append(embeddingIDs, item.ID)
	}
	var existingIDs []string
	if len(embeddingIDs) > 0 {
		if err := database.DB.WithContext(ctx).
			Model(&models.ConversationArchive{}).
			Where("user_id = ? AND embedding_id IN ?", t.UserID, embeddingIDs).
			Pluck("embedding_id", &existingIDs).Error; err != nil {
			return "", err
		}
	}
	existing := make(map[string]struct{}, len(existingIDs))
	for _, id := range existingIDs {
		existing[id] = struct{}{}
	}
	orphanIDs := make([]string, 0)
	results := make([]memoryResult, 0, limit)
	for _, item := range memories {
		if _, ok := existing[item.ID]; !ok {
			orphanIDs = append(orphanIDs, item.ID)
			continue
		}
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
		if len(results) == limit {
			break
		}
	}
	if err := t.Store.DeleteArchives(ctx, orphanIDs); err != nil {
		log.Printf("failed to delete %d orphaned archive vectors for user %d: %v", len(orphanIDs), t.UserID, err)
	}

	return marshalToolResult(map[string]any{
		"query":    input,
		"memories": results,
	})
}

func (t LongTermMemorySearchTool) createEmbedding(ctx context.Context, text string) ([]float32, error) {
	return embedding.Create(ctx, embedding.Config{
		APIKey:      t.APIKey,
		APIBaseURL:  t.APIBaseURL,
		Model:       t.Model,
		FallbackKey: t.FallbackKey,
		FallbackURL: t.FallbackURL,
	}, text)
}
