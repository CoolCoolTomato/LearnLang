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
	return `Finish the current chat turn after sending all reply sentences, or finish without replies when waiting for more user input. Call exactly once. Input must be JSON: {"detected_language":"language code","wait_for_next_message":false}.`
}

func (t CompleteChatTurnTool) Call(ctx context.Context, input string) (string, error) {
	var args struct {
		DetectedLanguage string `json:"detected_language"`
		WaitForNextMsg   bool   `json:"wait_for_next_message"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("parse complete_chat_turn input: %w", err)
	}

	if t.State != nil {
		t.State.Complete(args.DetectedLanguage, args.WaitForNextMsg)
	}

	return marshalToolResult(map[string]any{
		"status": "completed",
	})
}
