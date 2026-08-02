package archive

import (
	"context"
	"learnlang-api/agent/embedding"
	"learnlang-api/agent/memory"
	"learnlang-api/aiusage"
	"learnlang-api/models"
	"learnlang-api/services"
	"log"
)

type archiveIndexerImpl struct {
	archiveService *services.ConversationArchiveService
	memoryStore    *memory.Store
	usageRecorder  aiusage.Recorder
}

func newArchiveIndexer(archiveService *services.ConversationArchiveService, memoryStore *memory.Store, usage ...aiusage.Recorder) archiveIndexer {
	var recorder aiusage.Recorder
	if len(usage) > 0 {
		recorder = usage[0]
	}
	return archiveIndexerImpl{
		archiveService: archiveService,
		memoryStore:    memoryStore,
		usageRecorder:  recorder,
	}
}

func (i archiveIndexerImpl) Index(ctx context.Context, userID int64, settings *models.UserSettings, archives []models.ConversationArchive) {
	if i.memoryStore == nil || len(archives) == 0 {
		return
	}
	if settings.EmbeddingDimension <= 0 {
		log.Printf("conversation archive embedding skipped for user %d: embedding dimension is required", userID)
		return
	}

	for _, archive := range archives {
		embedding, err := createEmbedding(ctx, settings, archive.Summary)
		if i.usageRecorder != nil {
			status := models.AIUsageStatusSucceeded
			if err != nil {
				status = models.AIUsageStatusFailed
			}
			modelName := settings.EmbeddingModel
			if modelName == "" {
				modelName = "text-embedding-3-small"
			}
			if recordErr := i.usageRecorder.RecordAIUsage(ctx, aiusage.Record{UserID: userID, Operation: models.AIOperationEmbedding, Model: modelName, Usage: float64(len([]rune(archive.Summary))), Unit: models.AIUsageUnitTokens, Status: status}); recordErr != nil {
				log.Printf("record embedding AI usage failed for user %d: %v", userID, recordErr)
			}
		}
		if err != nil {
			log.Printf("conversation archive embedding failed for archive %d: %v", archive.ID, err)
			continue
		}

		embeddingID, err := i.memoryStore.InsertArchive(ctx, userID, archive.Summary, archive.MessageIDs, embedding, settings.EmbeddingDimension)
		if err != nil {
			log.Printf("conversation archive vector insert failed for archive %d: %v", archive.ID, err)
			continue
		}

		if err := i.archiveService.UpdateEmbeddingID(ctx, archive.ID, embeddingID, settings.EmbeddingDimension); err != nil {
			log.Printf("conversation archive embedding ID update failed for archive %d: %v", archive.ID, err)
		}
	}
}

func createEmbedding(ctx context.Context, settings *models.UserSettings, text string) ([]float32, error) {
	return embedding.Create(ctx, embedding.Config{
		APIKey:      settings.EmbeddingAPIKey,
		APIBaseURL:  settings.EmbeddingAPIBaseURL,
		Model:       settings.EmbeddingModel,
		FallbackKey: settings.APIKey,
		FallbackURL: settings.APIBaseURL,
	}, text)
}
