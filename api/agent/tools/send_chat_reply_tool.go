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

const maxChatReplyBatchSize = 20

func (t SendChatReplyTool) Name() string {
	return "send_chat_reply"
}

func (t SendChatReplyTool) Description() string {
	return `Send an ordered batch of assistant reply messages to the user and persist each message. Call once per chat turn with every user-visible reply sentence. Each array item must contain one short target-language sentence and its native-language translation. Input must be JSON: {"messages":[{"original":"target-language sentence","translation":"native-language translation"}]}. The array must contain 1-20 messages.`
}

func (t SendChatReplyTool) Call(ctx context.Context, input string) (string, error) {
	var args struct {
		Messages []Sentence `json:"messages"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return sendChatReplyRejection("tool input must be valid JSON")
	}
	if len(args.Messages) == 0 {
		return sendChatReplyRejection("messages must contain at least one reply")
	}
	if len(args.Messages) > maxChatReplyBatchSize {
		return sendChatReplyRejection(fmt.Sprintf("messages must contain at most %d replies", maxChatReplyBatchSize))
	}
	for index := range args.Messages {
		args.Messages[index].Original = strings.TrimSpace(args.Messages[index].Original)
		args.Messages[index].Translation = strings.TrimSpace(args.Messages[index].Translation)
		if args.Messages[index].Original == "" {
			return sendChatReplyRejection(fmt.Sprintf("messages[%d].original is required", index))
		}
		if args.Messages[index].Translation == "" {
			return sendChatReplyRejection(fmt.Sprintf("messages[%d].translation is required", index))
		}
	}
	if t.Runtime == nil {
		return "", fmt.Errorf("chat runtime service is not configured")
	}

	messageIDs := make([]int64, 0, len(args.Messages))
	for _, message := range args.Messages {
		messageID, err := t.Runtime.SaveAssistantReply(ctx, t.UserID, message.Original, message.Translation)
		if err != nil {
			return "", err
		}
		messageIDs = append(messageIDs, messageID)
		if t.State != nil {
			t.State.AddReply(message.Original, message.Translation, messageID)
		}
	}

	return marshalToolResult(map[string]any{
		"status":      "sent",
		"count":       len(messageIDs),
		"message_ids": messageIDs,
	})
}

func sendChatReplyRejection(reason string) (string, error) {
	return marshalToolResult(map[string]any{
		"status": "rejected",
		"error":  reason,
	})
}
