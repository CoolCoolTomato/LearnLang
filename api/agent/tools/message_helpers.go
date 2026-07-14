package tools

import (
	"context"
	"fmt"
	"learnlang-api/database"
	"learnlang-api/models"
	"time"
)

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

func formatMessages(messages []models.Message, timezone string) []string {
	loc := time.UTC
	if timezone != "" {
		if loaded, err := time.LoadLocation(timezone); err == nil {
			loc = loaded
		}
	}

	result := make([]string, 0, len(messages))
	for _, msg := range messages {
		result = append(result, fmt.Sprintf("[%s] %s: TextContent: %s Translation: %s", msg.CreatedAt.In(loc).Format("2006-01-02 15:04:05"), msg.Role, msg.TextContent, msg.Translation))
	}
	return result
}
