package services

import (
	"learnlang-api/database"
	"learnlang-api/models"
)

type VoiceFileService struct{}

func NewVoiceFileService() *VoiceFileService {
	return &VoiceFileService{}
}

func (vfs *VoiceFileService) CreateVoiceFile(voiceFile *models.VoiceFile) error {
	return database.DB.Create(voiceFile).Error
}

func (vfs *VoiceFileService) GetVoiceFile(id int64) (*models.VoiceFile, error) {
	var voiceFile models.VoiceFile
	if err := database.DB.First(&voiceFile, id).Error; err != nil {
		return nil, err
	}
	return &voiceFile, nil
}

func (vfs *VoiceFileService) ListVoiceFiles(userID *int64, page int, size int) ([]models.VoiceFile, error) {
	var voiceFiles []models.VoiceFile
	query := database.DB.Model(&models.VoiceFile{})

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	offset := (page - 1) * size
	if err := query.Order("created_at DESC").Limit(size).Offset(offset).Find(&voiceFiles).Error; err != nil {
		return nil, err
	}
	return voiceFiles, nil
}

func (vfs *VoiceFileService) UpdateVoiceFile(voiceFile *models.VoiceFile, updates map[string]interface{}) error {
	return database.DB.Model(voiceFile).Updates(updates).Error
}

func (vfs *VoiceFileService) DeleteVoiceFile(id int64) error {
	return database.DB.Delete(&models.VoiceFile{}, id).Error
}
