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
	UserID   int64
	Timezone string
	Runtime  *services.ChatRuntimeService
}

const localDateTimeLayout = "2006-01-02T15:04:05"

func (t ScheduleMessageTool) Name() string {
	return "schedule_message"
}

func (t ScheduleMessageTool) Description() string {
	return fmt.Sprintf(`Schedule a future assistant message. Resolve relative dates and times from the current conversation before calling this tool. scheduled_at must be the local wall-clock time in the user's %s timezone, formatted as YYYY-MM-DDTHH:MM:SS with no Z or UTC offset. Input must be JSON: {"message":"target language message","translation":"native language translation","scheduled_at":"local datetime"}.`, normalizedTimezone(t.Timezone))
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

	scheduledAtUTC, err := scheduledTimeToUTC(args.ScheduledAt, t.Timezone)
	if err != nil {
		return "", fmt.Errorf("parse scheduled_at: %w", err)
	}

	taskID, err := t.Runtime.ScheduleMessage(ctx, t.UserID, args.Message, args.Translation, scheduledAtUTC)
	if err != nil {
		return "", err
	}

	return marshalToolResult(map[string]any{
		"status":           "scheduled",
		"task_id":          taskID,
		"scheduled_at_utc": scheduledAtUTC.Format(time.RFC3339),
	})
}

func scheduledTimeToUTC(value, timezone string) (time.Time, error) {
	timezone = normalizedTimezone(timezone)
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("load user timezone %q: %w", timezone, err)
	}

	localTime, err := time.ParseInLocation(localDateTimeLayout, value, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("expected local datetime in YYYY-MM-DDTHH:MM:SS format: %w", err)
	}
	if localTime.Format(localDateTimeLayout) != value {
		return time.Time{}, fmt.Errorf("local datetime %q does not exist in timezone %s", value, timezone)
	}
	for _, offset := range []time.Duration{-2 * time.Hour, -time.Hour, time.Hour, 2 * time.Hour} {
		alternative := localTime.Add(offset).In(loc)
		if alternative.Format(localDateTimeLayout) == value {
			return time.Time{}, fmt.Errorf("local datetime %q is ambiguous in timezone %s", value, timezone)
		}
	}

	return localTime.UTC(), nil
}

func normalizedTimezone(timezone string) string {
	timezone = strings.TrimSpace(timezone)
	if timezone == "" {
		return "UTC"
	}
	return timezone
}
