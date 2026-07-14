package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"learnlang-api/database"
	"learnlang-api/models"
	"strings"
	"time"
)

type ArchivedConversationKeywordSearchTool struct {
	UserID   int64
	Timezone string
}

func (ArchivedConversationKeywordSearchTool) Name() string {
	return "search_archived_conversation_by_keyword"
}

func (ArchivedConversationKeywordSearchTool) Description() string {
	return `Search chat records by an exact keyword and return the complete archived conversation fragments containing matches. Matches without a conversation archive are treated as recent memory and their messages are intentionally omitted. Input must be JSON: {"keyword":"exact phrase","start_time":"optional RFC3339","end_time":"optional RFC3339","limit":5}. start_time and end_time filter the matching message, while returned fragments include the complete archive. limit defaults to 5 and is capped at 10.`
}

func (t ArchivedConversationKeywordSearchTool) Call(ctx context.Context, input string) (string, error) {
	var args struct {
		Keyword   string `json:"keyword"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Limit     int    `json:"limit"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return marshalToolResult(map[string]any{
			"status": "invalid_request",
			"error":  "tool input must be valid JSON",
		})
	}

	keyword := strings.TrimSpace(args.Keyword)
	if keyword == "" {
		return marshalToolResult(map[string]any{
			"status": "invalid_request",
			"error":  "keyword is required",
		})
	}
	startTime, err := parseKeywordSearchTime(args.StartTime, "start_time")
	if err != nil {
		return marshalToolResult(map[string]any{"status": "invalid_request", "error": err.Error()})
	}
	endTime, err := parseKeywordSearchTime(args.EndTime, "end_time")
	if err != nil {
		return marshalToolResult(map[string]any{"status": "invalid_request", "error": err.Error()})
	}
	if startTime != nil && endTime != nil && startTime.After(*endTime) {
		return marshalToolResult(map[string]any{
			"status": "invalid_request",
			"error":  "start_time must not be after end_time",
		})
	}
	limit := args.Limit
	if limit <= 0 {
		limit = 5
	}
	if limit > 10 {
		limit = 10
	}

	pattern := "%" + escapeLikeKeyword(keyword) + "%"
	archiveQuery := database.DB.WithContext(ctx).
		Model(&models.ConversationArchive{}).
		Distinct("conversation_archives.*").
		Joins("JOIN messages ON messages.user_id = conversation_archives.user_id AND messages.id BETWEEN conversation_archives.start_message_id AND conversation_archives.end_message_id").
		Where("conversation_archives.user_id = ?", t.UserID).
		Where(`(messages.text_content ILIKE ? ESCAPE '\' OR messages.translation ILIKE ? ESCAPE '\')`, pattern, pattern)
	if startTime != nil {
		archiveQuery = archiveQuery.Where("messages.created_at >= ?", *startTime)
	}
	if endTime != nil {
		archiveQuery = archiveQuery.Where("messages.created_at <= ?", *endTime)
	}

	var archives []models.ConversationArchive
	if err := archiveQuery.Order("conversation_archives.end_message_id DESC").Limit(limit).Find(&archives).Error; err != nil {
		return "", err
	}
	if len(archives) == 0 {
		messageQuery := database.DB.WithContext(ctx).
			Model(&models.Message{}).
			Where("user_id = ?", t.UserID).
			Where(`(text_content ILIKE ? ESCAPE '\' OR translation ILIKE ? ESCAPE '\')`, pattern, pattern)
		if startTime != nil {
			messageQuery = messageQuery.Where("created_at >= ?", *startTime)
		}
		if endTime != nil {
			messageQuery = messageQuery.Where("created_at <= ?", *endTime)
		}
		var count int64
		if err := messageQuery.Count(&count).Error; err != nil {
			return "", err
		}
		status := "not_found"
		if count > 0 {
			status = "recent_memory"
		}
		return marshalToolResult(map[string]any{
			"keyword":  keyword,
			"status":   status,
			"archives": []any{},
		})
	}

	type archiveResult struct {
		ArchiveID      int64    `json:"archive_id"`
		StartMessageID int64    `json:"start_message_id"`
		EndMessageID   int64    `json:"end_message_id"`
		Summary        string   `json:"summary"`
		Messages       []string `json:"messages"`
	}
	results := make([]archiveResult, 0, len(archives))
	for _, archive := range archives {
		messages, err := linkedMessages(ctx, t.UserID, archive.MessageIDs)
		if err != nil {
			return "", err
		}
		results = append(results, archiveResult{
			ArchiveID:      archive.ID,
			StartMessageID: archive.StartMessageID,
			EndMessageID:   archive.EndMessageID,
			Summary:        archive.Summary,
			Messages:       formatMessages(messages, t.Timezone),
		})
	}
	return marshalToolResult(map[string]any{
		"keyword":  keyword,
		"status":   "archived",
		"archives": results,
	})
}

func parseKeywordSearchTime(value, field string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fmt.Errorf("%s must be RFC3339", field)
	}
	return &parsed, nil
}

func escapeLikeKeyword(value string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(value)
}
