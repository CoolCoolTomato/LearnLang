package services

import (
	"context"
	"learnlang-api/utils"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type ModelProviderService struct{}

func NewModelProviderService() *ModelProviderService {
	return &ModelProviderService{}
}

type ModelInfo struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

func (mps *ModelProviderService) GetCustomProviderModels(apiBaseURL, apiKey string) ([]ModelInfo, error) {
	if apiKey == "" {
		return nil, utils.ErrAPIKeyNotConfigured
	}

	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if apiBaseURL != "" {
		opts = append(opts, option.WithBaseURL(apiBaseURL))
	}

	client := openai.NewClient(opts...)
	modelPage, err := client.Models.List(context.Background())
	if err != nil {
		return nil, err
	}

	var models []ModelInfo
	for _, model := range modelPage.Data {
		models = append(models, ModelInfo{
			ID:      model.ID,
			Object:  string(model.Object),
			Created: model.Created,
			OwnedBy: model.OwnedBy,
		})
	}

	return models, nil
}
