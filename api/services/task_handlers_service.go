package services

import (
	"context"
	"encoding/json"
)

type SendMessageArgs struct {
	UserID      int64  `json:"user_id"`
	Message     string `json:"message"`
	Translation string `json:"translation"`
}

func NewSendMessageHandler(chatRuntimeService *ChatRuntimeService) TaskHandler {
	return func(args string) error {
		var msgArgs SendMessageArgs
		if err := json.Unmarshal([]byte(args), &msgArgs); err != nil {
			return err
		}

		_, err := chatRuntimeService.SaveAssistantReply(context.Background(), msgArgs.UserID, msgArgs.Message, msgArgs.Translation)
		return err
	}
}
