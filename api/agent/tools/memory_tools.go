package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"learnlang-api/agent/memory"
	"learnlang-api/database"
	"learnlang-api/models"
	"sort"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type UserProfileSummaryTool struct {
	UserID int64
}

func (t UserProfileSummaryTool) Name() string {
	return "get_user_profile_summary"
}

func (t UserProfileSummaryTool) Description() string {
	return "Read the user's stable profile summary. Input can be empty."
}

func (t UserProfileSummaryTool) Call(ctx context.Context, input string) (string, error) {
	var summary models.ConversationSummary
	if err := database.DB.WithContext(ctx).Where("user_id = ?", t.UserID).First(&summary).Error; err != nil {
		return `{"summary":""}`, nil
	}

	return marshalToolResult(map[string]any{
		"summary": summary.Summary,
	})
}

type RecentConversationTool struct {
	UserID   int64
	Limit    int
	Timezone string
}

func (t RecentConversationTool) Name() string {
	return "get_recent_conversation"
}

func (t RecentConversationTool) Description() string {
	return "Read recent short-term conversation messages for the current user. Input can be empty."
}

func (t RecentConversationTool) Call(ctx context.Context, input string) (string, error) {
	limit := t.Limit
	if limit <= 0 {
		limit = 100
	}

	var allMessages []models.Message
	if err := database.DB.WithContext(ctx).
		Where("user_id = ?", t.UserID).
		Order("created_at DESC").
		Limit(limit).
		Find(&allMessages).Error; err != nil {
		return "", err
	}

	recentMessages := continuousRecentMessages(allMessages)
	return marshalToolResult(map[string]any{
		"messages": formatMessages(recentMessages, t.Timezone),
	})
}

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

func linkedMessages(ctx context.Context, userID int64, messageIDs []int64) ([]models.Message, error) {
	if len(messageIDs) == 0 {
		return []models.Message{}, nil
	}

	var messages []models.Message
	if err := database.DB.WithContext(ctx).
		Where("user_id = ? AND id IN ?", userID, messageIDs).
		Find(&messages).Error; err != nil {
		return nil, err
	}

	byID := make(map[int64]models.Message, len(messages))
	for _, message := range messages {
		byID[message.ID] = message
	}

	ordered := make([]models.Message, 0, len(messages))
	for _, id := range messageIDs {
		if message, ok := byID[id]; ok {
			ordered = append(ordered, message)
		}
	}

	return ordered, nil
}

func continuousRecentMessages(allMessages []models.Message) []models.Message {
	if len(allMessages) == 0 {
		return []models.Message{}
	}

	const maxInterval = 60 * 60
	recentMessages := make([]models.Message, 0, len(allMessages))
	for i := 0; i < len(allMessages); i++ {
		if i == 0 {
			recentMessages = append(recentMessages, allMessages[i])
			continue
		}

		interval := allMessages[i-1].CreatedAt.Unix() - allMessages[i].CreatedAt.Unix()
		if interval > maxInterval {
			break
		}
		recentMessages = append(recentMessages, allMessages[i])
	}

	sort.Slice(recentMessages, func(i, j int) bool {
		return recentMessages[i].CreatedAt.Before(recentMessages[j].CreatedAt)
	})

	return recentMessages
}

func formatMessages(messages []models.Message, timezone string) []string {
	loc := time.UTC
	if timezone != "" {
		if loaded, err := time.LoadLocation(timezone); err == nil {
			loc = loaded
		}
	}

	result := make([]string, 0, len(messages))
	for _, msg := range messages {
		result = append(result, fmt.Sprintf("[%s] %s: %s", msg.CreatedAt.In(loc).Format("2006-01-02 15:04:05"), msg.Role, msg.TextContent))
	}
	return result
}

func marshalToolResult(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
