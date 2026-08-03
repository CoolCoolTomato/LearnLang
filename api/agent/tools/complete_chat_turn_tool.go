package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

type CompleteChatTurnTool struct {
	State *TurnState
}

func (t CompleteChatTurnTool) Name() string {
	return "complete_chat_turn"
}

func (t CompleteChatTurnTool) Description() string {
	return `Finish the current chat turn after sending all required reply sentences, including an immediate confirmation after scheduling a future message. Call exactly once. Input must be JSON: {"detected_language":"language code"}.`
}

func (t CompleteChatTurnTool) Call(ctx context.Context, input string) (string, error) {
	var args struct {
		DetectedLanguage string `json:"detected_language"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("parse complete_chat_turn input: %w", err)
	}

	if t.State != nil {
		if !t.State.Complete(args.DetectedLanguage) {
			return "", fmt.Errorf("complete_chat_turn may only be called once")
		}
	}

	return marshalToolResult(map[string]any{
		"status": "completed",
	})
}
