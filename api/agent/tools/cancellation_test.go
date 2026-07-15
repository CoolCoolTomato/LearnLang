package tools

import (
	"context"
	"errors"
	"testing"
)

func TestSendChatReplyToolRejectsCanceledRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := (SendChatReplyTool{}).Call(ctx, `{"messages":[{"original":"Hello","translation":"你好"}]}`)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestScheduleMessageToolRejectsCanceledRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := (ScheduleMessageTool{}).Call(ctx, `{"message":"Hello","translation":"你好","scheduled_at":"2026-07-16T12:00:00Z"}`)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
