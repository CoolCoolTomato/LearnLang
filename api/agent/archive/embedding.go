package archive

import (
	"context"
	"learnlang-api/agent/embedding"
	"learnlang-api/agent/memory"
	"learnlang-api/models"
	"learnlang-api/services"
	"log"
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
			log.Printf("conversation archive embedding failed for archive %d: %v", archive.ID, err)
			continue
		}

		embeddingID, err := i.memoryStore.InsertArchive(ctx, userID, archive.Summary, archive.MessageIDs, embedding)
		if err != nil {
			log.Printf("conversation archive vector insert failed for archive %d: %v", archive.ID, err)
			continue
		}

		if err := i.archiveService.UpdateEmbeddingID(ctx, archive.ID, embeddingID); err != nil {
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
