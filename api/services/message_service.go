package services

import (
	"learnlang-api/database"
	"learnlang-api/models"
)

type MessageService struct{}

func NewMessageService() *MessageService {
	return &MessageService{}
}

type MessageListResult struct {
	Total    int64
	Messages []models.Message
}

func (ms *MessageService) CreateMessage(userID int64, role, textContent, translation string, voiceFileID *int64, inputType string, tokenCount int) (*models.Message, error) {
	message := models.Message{
		UserID:      userID,
		Role:        role,
		TextContent: textContent,
		Translation: translation,
		VoiceFileID: voiceFileID,
		InputType:   inputType,
		TokenCount:  tokenCount,
	}

	if err := database.DB.Create(&message).Error; err != nil {
		return nil, err
	}

	return &message, nil
}

func (ms *MessageService) GetMessage(messageID int64) (*models.Message, error) {
	var message models.Message
	if err := database.DB.First(&message, messageID).Error; err != nil {
		return nil, err
	}
	return &message, nil
}

func (ms *MessageService) ListMessages(page, pageSize int, userID string) (*MessageListResult, error) {
	query := database.DB.Model(&models.Message{})

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	var total int64
	query.Count(&total)

	var messages []models.Message
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&messages).Error; err != nil {
		return nil, err
	}

	return &MessageListResult{Total: total, Messages: messages}, nil
}

func (ms *MessageService) UpdateMessage(messageID int64, role, textContent, translation string, voiceFileID *int64, inputType string, tokenCount int) (*models.Message, error) {
	var message models.Message
	if err := database.DB.First(&message, messageID).Error; err != nil {
		return nil, err
	}

	if role != "" {
		message.Role = role
	}
	if textContent != "" {
		message.TextContent = textContent
	}
	if translation != "" {
		message.Translation = translation
	}
	if voiceFileID != nil {
		message.VoiceFileID = voiceFileID
	}
	if inputType != "" {
		message.InputType = inputType
	}
	if tokenCount > 0 {
		message.TokenCount = tokenCount
	}

	if err := database.DB.Save(&message).Error; err != nil {
		return nil, err
	}

	return &message, nil
}

func (ms *MessageService) DeleteMessage(messageID int64) error {
	return database.DB.Delete(&models.Message{}, messageID).Error
}
