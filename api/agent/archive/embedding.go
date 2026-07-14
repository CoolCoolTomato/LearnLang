package archive

import (
	"context"
	"fmt"
	"learnlang-api/agent/memory"
	"learnlang-api/models"
	"learnlang-api/services"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type archiveIndexerImpl struct {
	archiveService *services.ConversationArchiveService
	memoryStore    *memory.Store
}

func newArchiveIndexer(archiveService *services.ConversationArchiveService, memoryStore *memory.Store) archiveIndexer {
	return archiveIndexerImpl{
		archiveService: archiveService,
		memoryStore:    memoryStore,
	}
}

func (i archiveIndexerImpl) Index(ctx context.Context, userID int64, settings *models.UserSettings, archives []models.ConversationArchive) {
	if i.memoryStore == nil || len(archives) == 0 {
		return
	}

	for _, archive := range archives {
		embedding, err := createEmbedding(ctx, settings, archive.Summary)
		if err != nil {
			continue
		}

		embeddingID, err := i.memoryStore.InsertArchive(ctx, userID, archive.Summary, archive.MessageIDs, embedding)
		if err != nil {
			continue
		}

		_ = i.archiveService.UpdateEmbeddingID(ctx, archive.ID, embeddingID)
	}
}

func createEmbedding(ctx context.Context, settings *models.UserSettings, text string) ([]float32, error) {
	apiKey := strings.TrimSpace(settings.EmbeddingAPIKey)
	if apiKey == "" {
		apiKey = strings.TrimSpace(settings.APIKey)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("embedding api key is required")
	}

	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	apiBaseURL := strings.TrimSpace(settings.EmbeddingAPIBaseURL)
	if apiBaseURL == "" {
		apiBaseURL = strings.TrimSpace(settings.APIBaseURL)
	}
	if apiBaseURL != "" {
		opts = append(opts, option.WithBaseURL(apiBaseURL))
	}

	model := strings.TrimSpace(settings.EmbeddingModel)
	if model == "" {
		model = "text-embedding-3-small"
	}

	client := openai.NewClient(opts...)
	resp, err := client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model: openai.EmbeddingModel(model),
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: []string{text},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("embedding response is empty")
	}

	vector := make([]float32, 0, len(resp.Data[0].Embedding))
	for _, value := range resp.Data[0].Embedding {
		vector = append(vector, float32(value))
	}
	return vector, nil
}
