package services

import (
	"learnlang-api/database"
	"learnlang-api/models"
)

type UserProfileSummaryService struct{}

func NewUserProfileSummaryService() *UserProfileSummaryService {
	return &UserProfileSummaryService{}
}

func (ups *UserProfileSummaryService) GetUserProfileSummary(userID int64) (*models.UserProfileSummary, error) {
	var summary models.UserProfileSummary
	err := database.DB.Where("user_id = ?", userID).First(&summary).Error

	if err != nil {
		summary = models.UserProfileSummary{UserID: userID}
		if err := database.DB.Create(&summary).Error; err != nil {
			return nil, err
		}
	}

	return &summary, nil
}

func (ups *UserProfileSummaryService) UpdateUserProfileSummary(userID int64, summary string) (*models.UserProfileSummary, error) {
	profileSummary, err := ups.GetUserProfileSummary(userID)
	if err != nil {
		return nil, err
	}

	if summary != "" {
		profileSummary.Summary = summary
	}

	if err := database.DB.Save(profileSummary).Error; err != nil {
		return nil, err
	}

	return profileSummary, nil
}
