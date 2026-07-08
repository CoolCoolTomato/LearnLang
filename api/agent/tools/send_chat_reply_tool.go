package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"learnlang-api/services"
	"strings"
)

type SendChatReplyTool struct {
	UserID  int64
	Runtime *services.ChatRuntimeService
	State   *TurnState
}

func (t SendChatReplyTool) Name() string {
	return "send_chat_reply"
}

func (t SendChatReplyTool) Description() string {
	return `Send one assistant reply sentence to the user and persist it. Call once for each sentence you want the user to see. Input must be JSON: {"original":"target-language sentence","translation":"native-language translation"}. Do not put multiple sentences in one call.`
}

func (t SendChatReplyTool) Call(ctx context.Context, input string) (string, error) {
	var args struct {
		Original    string `json:"original"`
		Translation string `json:"translation"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("parse send_chat_reply input: %w", err)
	}

	args.Original = strings.TrimSpace(args.Original)
	args.Translation = strings.TrimSpace(args.Translation)
	if args.Original == "" {
		return "", fmt.Errorf("original is required")
	}
	if t.Runtime == nil {
		return "", fmt.Errorf("chat runtime service is not configured")
	}

	messageID, err := t.Runtime.SaveAssistantReply(ctx, t.UserID, args.Original, args.Translation)
	if err != nil {
		return "", err
	}

	if t.State != nil {
		t.State.AddReply(args.Original, args.Translation, messageID)
	}

	return marshalToolResult(map[string]any{
		"status":     "sent",
		"message_id": messageID,
	})
}
