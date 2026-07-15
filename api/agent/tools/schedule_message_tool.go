package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"learnlang-api/services"
	"strings"
	"time"
)

type ScheduleMessageTool struct {
	UserID  int64
	Runtime *services.ChatRuntimeService
}

func (t ScheduleMessageTool) Name() string {
	return "schedule_message"
}

func (t ScheduleMessageTool) Description() string {
	return `Schedule a future assistant message. Input must be JSON: {"message":"target language message","translation":"native language translation","scheduled_at":"UTC RFC3339 timestamp ending with Z"}.`
}

func (t ScheduleMessageTool) Call(ctx context.Context, input string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	var args struct {
		Message     string `json:"message"`
		Translation string `json:"translation"`
		ScheduledAt string `json:"scheduled_at"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("parse schedule_message input: %w", err)
	}

	args.Message = strings.TrimSpace(args.Message)
	args.Translation = strings.TrimSpace(args.Translation)
	args.ScheduledAt = strings.TrimSpace(args.ScheduledAt)
	if args.Message == "" {
		return "", fmt.Errorf("message is required")
	}
	if args.ScheduledAt == "" {
		return "", fmt.Errorf("scheduled_at is required")
	}
	if t.Runtime == nil {
		return "", fmt.Errorf("chat runtime service is not configured")
	}

	scheduledAt, err := time.Parse(time.RFC3339, args.ScheduledAt)
	if err != nil {
		return "", fmt.Errorf("parse scheduled_at: %w", err)
	}

	taskID, err := t.Runtime.ScheduleMessage(ctx, t.UserID, args.Message, args.Translation, scheduledAt.UTC())
	if err != nil {
		return "", err
	}

	return marshalToolResult(map[string]any{
		"status":  "scheduled",
		"task_id": taskID,
	})
}
