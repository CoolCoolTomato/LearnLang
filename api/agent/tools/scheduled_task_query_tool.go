package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"learnlang-api/services"
	"strings"
	"time"
)

const defaultScheduledTaskPageSize = 10

type ScheduledTaskLister interface {
	ListScheduledTasks(ctx context.Context, userID int64, filter string, page, pageSize int) (*services.ScheduledTaskPage, error)
}

type ScheduledTaskQueryTool struct {
	UserID   int64
	Timezone string
	Lister   ScheduledTaskLister
}

func (ScheduledTaskQueryTool) Name() string {
	return "list_scheduled_tasks"
}

func (ScheduledTaskQueryTool) Description() string {
	return `List the current user's scheduled tasks with pagination. status accepts "unfinished", "completed", or "all" and defaults to "all". unfinished includes every task not completed successfully, including pending and failed tasks; inspect each task's status to distinguish them. page defaults to 1, page_size defaults to 10 and is capped at 50. The result includes total_pages and has_next.`
}

func (t ScheduledTaskQueryTool) Call(ctx context.Context, input string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	var args struct {
		Status   string `json:"status"`
		Page     int    `json:"page"`
		PageSize int    `json:"page_size"`
	}
	if strings.TrimSpace(input) != "" {
		if err := json.Unmarshal([]byte(input), &args); err != nil {
			return marshalToolResult(map[string]any{"status": "invalid_request", "error": "tool input must be a valid JSON object"})
		}
	}
	filter := strings.ToLower(strings.TrimSpace(args.Status))
	if filter == "" {
		filter = services.ScheduledTaskFilterAll
	}
	if filter != services.ScheduledTaskFilterAll && filter != services.ScheduledTaskFilterUnfinished && filter != services.ScheduledTaskFilterCompleted {
		return marshalToolResult(map[string]any{"status": "invalid_request", "error": "status must be unfinished, completed, or all"})
	}
	if args.Page < 0 || args.PageSize < 0 {
		return marshalToolResult(map[string]any{"status": "invalid_request", "error": "page and page_size must be positive integers"})
	}
	if args.Page == 0 {
		args.Page = 1
	}
	if args.PageSize == 0 {
		args.PageSize = defaultScheduledTaskPageSize
	}
	if args.PageSize > 50 {
		args.PageSize = 50
	}
	if t.Lister == nil {
		return "", fmt.Errorf("scheduled task lister is not configured")
	}

	result, err := t.Lister.ListScheduledTasks(ctx, t.UserID, filter, args.Page, args.PageSize)
	if err != nil {
		return "", err
	}
	location, timezone := scheduledTaskLocation(t.Timezone)
	tasks := make([]map[string]any, 0, len(result.Tasks))
	for _, task := range result.Tasks {
		item := map[string]any{
			"id":                task.ID,
			"type":              task.FunctionName,
			"status":            task.Status,
			"scheduled_at":      task.ScheduledAt.In(location).Format(time.RFC3339),
			"created_at":        task.CreatedAt.In(location).Format(time.RFC3339),
			"details_available": false,
		}
		if task.FunctionName == "send_message" {
			var message services.SendMessageArgs
			if err := json.Unmarshal([]byte(task.Args), &message); err == nil {
				item["message"] = message.Message
				item["translation"] = message.Translation
				item["details_available"] = true
			}
		}
		tasks = append(tasks, item)
	}
	return marshalToolResult(map[string]any{
		"status":      "ok",
		"filter":      filter,
		"timezone":    timezone,
		"page":        result.Page,
		"page_size":   result.PageSize,
		"total":       result.Total,
		"total_pages": result.TotalPages,
		"has_next":    result.HasNext,
		"tasks":       tasks,
	})
}

func scheduledTaskLocation(timezone string) (*time.Location, string) {
	timezone = normalizedTimezone(timezone)
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.UTC, "UTC"
	}
	return location, timezone
}
