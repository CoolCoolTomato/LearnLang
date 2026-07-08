package tools

import (
	"context"
	"learnlang-api/database"
	"learnlang-api/models"
)

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
