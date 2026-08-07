package services

import (
	"errors"
	"fmt"
	"learnlang-api/database"
	"learnlang-api/models"

	"gorm.io/gorm"
)

type UserSettingsService struct{}

var ErrInvalidLLMType = errors.New("invalid LLM type")

func NewUserSettingsService() *UserSettingsService {
	return &UserSettingsService{}
}

func (uss *UserSettingsService) CreateUserSettings(userID int64) error {
	settings := models.UserSettings{
		UserID:  userID,
		LLMType: models.LLMTypeOpenAI,
	}
	return database.DB.Create(&settings).Error
}

func (uss *UserSettingsService) GetUserSettings(userID int64) (*models.UserSettings, error) {
	var settings models.UserSettings
	err := database.DB.Where("user_id = ?", userID).First(&settings).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := uss.CreateUserSettings(userID); err != nil {
			return nil, err
		}
		if err := database.DB.Where("user_id = ?", userID).First(&settings).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	normalizedType, valid := models.NormalizeLLMType(settings.LLMType)
	if !valid {
		normalizedType = models.LLMTypeOpenAI
	}
	if settings.LLMType != normalizedType {
		if err := database.DB.Model(&settings).Update("llm_type", normalizedType).Error; err != nil {
			return nil, err
		}
		settings.LLMType = normalizedType
	}
	return &settings, nil
}

func (uss *UserSettingsService) UpdateUserSettings(userID int64, updates map[string]interface{}) (*models.UserSettings, error) {
	var settings models.UserSettings
	if err := database.DB.Where("user_id = ?", userID).First(&settings).Error; err != nil {
		return nil, err
	}

	for key, value := range updates {
		switch key {
		case "api_base_url":
			if v, ok := value.(string); ok && v != "" {
				settings.APIBaseURL = v
			}
		case "api_key":
			if v, ok := value.(string); ok && v != "" {
				settings.APIKey = v
			}
		case "model":
			if v, ok := value.(string); ok && v != "" {
				settings.Model = v
			}
		case "llm_type":
			if v, ok := value.(string); ok {
				normalized, valid := models.NormalizeLLMType(v)
				if !valid {
					return nil, fmt.Errorf("%w: %q", ErrInvalidLLMType, v)
				}
				settings.LLMType = normalized
			}
		case "embedding_api_base_url":
			if v, ok := value.(string); ok && v != "" {
				settings.EmbeddingAPIBaseURL = v
			}
		case "embedding_api_key":
			if v, ok := value.(string); ok && v != "" {
				settings.EmbeddingAPIKey = v
			}
		case "embedding_model":
			if v, ok := value.(string); ok && v != "" {
				settings.EmbeddingModel = v
			}
		case "embedding_dimension":
			if v, ok := value.(int); ok && v > 0 {
				settings.EmbeddingDimension = v
			}
		case "stt_api_base_url":
			if v, ok := value.(string); ok && v != "" {
				settings.STTAPIBaseURL = v
			}
		case "stt_api_key":
			if v, ok := value.(string); ok && v != "" {
				settings.STTAPIKey = v
			}
		case "stt_model":
			if v, ok := value.(string); ok && v != "" {
				settings.STTModel = v
			}
		case "tts_api_base_url":
			if v, ok := value.(string); ok && v != "" {
				settings.TTSAPIBaseURL = v
			}
		case "tts_api_key":
			if v, ok := value.(string); ok && v != "" {
				settings.TTSAPIKey = v
			}
		case "tts_model":
			if v, ok := value.(string); ok && v != "" {
				settings.TTSModel = v
			}
		case "tts_voice":
			if v, ok := value.(string); ok && v != "" {
				settings.TTSVoice = v
			}
		case "native_language":
			if v, ok := value.(string); ok && v != "" {
				settings.NativeLanguage = v
			}
		case "target_language":
			if v, ok := value.(string); ok && v != "" {
				settings.TargetLanguage = v
			}
		case "timezone":
			if v, ok := value.(string); ok && v != "" {
				settings.Timezone = v
			}
		}
	}

	if err := database.DB.Save(&settings).Error; err != nil {
		return nil, err
	}

	return &settings, nil
}
