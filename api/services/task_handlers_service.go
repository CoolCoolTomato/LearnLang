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

type WaitMessageArgs struct {
	UserID    int64 `json:"user_id"`
	MessageID int64 `json:"message_id"`
}

type InstantMessageResponder interface {
	ProcessInstantAIResponse(userID int64, messageID int64)
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

func NewWaitMessageHandler(responder InstantMessageResponder) TaskHandler {
	return func(args string) error {
		var msgArgs WaitMessageArgs
		if err := json.Unmarshal([]byte(args), &msgArgs); err != nil {
			return err
		}
		responder.ProcessInstantAIResponse(msgArgs.UserID, msgArgs.MessageID)
		return nil
	}
}
