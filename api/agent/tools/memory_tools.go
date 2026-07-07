package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"learnlang-api/database"
	"learnlang-api/models"
	"sort"
	"time"
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
	UserID   int64
	Limit    int
	Timezone string
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

	var memories []models.UserMemory
	if err := database.DB.WithContext(ctx).
		Where("user_id = ?", t.UserID).
		Order("importance_score DESC, updated_at DESC, created_at DESC").
		Limit(limit).
		Find(&memories).Error; err != nil {
		return "", err
	}

	type memoryResult struct {
		ID       int64    `json:"id"`
		Summary  string   `json:"summary"`
		Type     string   `json:"type"`
		Messages []string `json:"messages"`
	}

	results := make([]memoryResult, 0, len(memories))
	for _, memory := range memories {
		messages, err := linkedMessages(ctx, memory.ID)
		if err != nil {
			return "", err
		}

		results = append(results, memoryResult{
			ID:       memory.ID,
			Summary:  memory.Summary,
			Type:     memory.MemoryType,
			Messages: formatMessages(messages, t.Timezone),
		})
	}

	return marshalToolResult(map[string]any{
		"query":    input,
		"memories": results,
	})
}

func linkedMessages(ctx context.Context, memoryID int64) ([]models.Message, error) {
	var links []models.UserMemoryMessage
	if err := database.DB.WithContext(ctx).
		Preload("Message").
		Where("user_memory_id = ?", memoryID).
		Joins("JOIN messages ON messages.id = user_memory_messages.message_id").
		Order("messages.created_at ASC").
		Find(&links).Error; err != nil {
		return nil, err
	}

	messages := make([]models.Message, 0, len(links))
	for _, link := range links {
		messages = append(messages, link.Message)
	}

	return messages, nil
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
