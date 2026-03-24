package services

import (
	"context"
	"encoding/json"
	"learnlang-api/database"
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

func NewSendMessageHandler(messageService *MessageService, chatService *ChatService, userSettingsService *UserSettingsService, wsHub WSHub) TaskHandler {
	return func(args string) error {
		var msgArgs SendMessageArgs
		if err := json.Unmarshal([]byte(args), &msgArgs); err != nil {
			return err
		}

		settings, err := userSettingsService.GetUserSettings(msgArgs.UserID)
		if err != nil {
			return err
		}

		voiceFileID, _ := chatService.TextToSpeech(context.Background(), msgArgs.UserID, msgArgs.Message, settings)

		message, err := messageService.CreateMessage(msgArgs.UserID, "assistant", msgArgs.Message, msgArgs.Translation, voiceFileID, "text", 0)
		if err != nil {
			return err
		}

		database.DB.Preload("VoiceFile").First(message, message.ID)

		if message.VoiceFile != nil {
			message.VoiceFile.VoiceURL = ""
		}

		messageJSON, _ := json.Marshal(message)
		wsHub.SendToUser(msgArgs.UserID, messageJSON)
		return nil
	}
}

func NewWaitMessageHandler(chatService *ChatService) TaskHandler {
	return func(args string) error {
		var msgArgs WaitMessageArgs
		if err := json.Unmarshal([]byte(args), &msgArgs); err != nil {
			return err
		}
		chatService.processInstantAIResponse(msgArgs.UserID, msgArgs.MessageID)
		return nil
	}
}
