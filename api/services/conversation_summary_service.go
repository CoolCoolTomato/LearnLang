package services

import (
	"learnlang-api/database"
	"learnlang-api/models"
)

type ConversationSummaryService struct{}

func NewConversationSummaryService() *ConversationSummaryService {
	return &ConversationSummaryService{}
}

func (css *ConversationSummaryService) GetConversationSummary(userID int64) (*models.ConversationSummary, error) {
	var summary models.ConversationSummary
	err := database.DB.Where("user_id = ?", userID).First(&summary).Error

	if err != nil {
		summary = models.ConversationSummary{UserID: userID}
		if err := database.DB.Create(&summary).Error; err != nil {
			return nil, err
		}
	}

	return &summary, nil
}

func (css *ConversationSummaryService) UpdateConversationSummary(userID int64, summary string) (*models.ConversationSummary, error) {
	conversationSummary, err := css.GetConversationSummary(userID)
	if err != nil {
		return nil, err
	}

	if summary != "" {
		conversationSummary.Summary = summary
	}

	if err := database.DB.Save(conversationSummary).Error; err != nil {
		return nil, err
	}

	return conversationSummary, nil
}
