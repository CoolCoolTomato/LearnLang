package services

import (
	"context"
	"fmt"
	"learnlang-api/agent/embedding"
	"learnlang-api/agent/memory"
	"learnlang-api/database"
	"learnlang-api/models"
	"strings"
	"time"
)

type DeveloperArchiveSearchService struct {
	memoryStore *memory.Store
	settings    *UserSettingsService
}

type DeveloperArchiveSearchResult struct {
	EmbeddingID string           `json:"embedding_id"`
	ArchiveID   int64            `json:"archive_id"`
	Score       float32          `json:"score"`
	Summary     string           `json:"summary"`
	MessageIDs  []int64          `json:"message_ids"`
	Messages    []models.Message `json:"messages"`
	CreatedAt   time.Time        `json:"created_at"`
}

func NewDeveloperArchiveSearchService(memoryStore *memory.Store, settings *UserSettingsService) *DeveloperArchiveSearchService {
	return &DeveloperArchiveSearchService{memoryStore: memoryStore, settings: settings}
}

func (s *DeveloperArchiveSearchService) Search(ctx context.Context, userID int64, query string, limit int) ([]DeveloperArchiveSearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if s.memoryStore == nil {
		return nil, fmt.Errorf("memory store is not configured")
	}
	if limit < 1 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	settings, err := s.settings.GetUserSettings(userID)
	if err != nil {
		return nil, err
	}
	vector, err := embedding.Create(ctx, embedding.Config{
		APIKey:      settings.EmbeddingAPIKey,
		APIBaseURL:  settings.EmbeddingAPIBaseURL,
		Model:       settings.EmbeddingModel,
		FallbackKey: settings.APIKey,
		FallbackURL: settings.APIBaseURL,
	}, query)
	if err != nil {
		return nil, err
	}

	memories, err := s.memoryStore.Search(ctx, userID, vector, limit*4)
	if err != nil {
		return nil, err
	}
	if len(memories) == 0 {
		return []DeveloperArchiveSearchResult{}, nil
	}

	embeddingIDs := make([]string, 0, len(memories))
	for _, memory := range memories {
		embeddingIDs = append(embeddingIDs, memory.ID)
	}
	var archives []models.ConversationArchive
	if err := database.DB.WithContext(ctx).
		Where("user_id = ? AND embedding_id IN ?", userID, embeddingIDs).
		Find(&archives).Error; err != nil {
		return nil, err
	}
	archiveByEmbeddingID := make(map[string]models.ConversationArchive, len(archives))
	for _, archive := range archives {
		archiveByEmbeddingID[archive.EmbeddingID] = archive
	}
	liveMemories := make([]memory.Summary, 0, limit)
	orphanIDs := make([]string, 0)
	for _, item := range memories {
		if _, ok := archiveByEmbeddingID[item.ID]; !ok {
			orphanIDs = append(orphanIDs, item.ID)
			continue
		}
		liveMemories = append(liveMemories, item)
		if len(liveMemories) == limit {
			break
		}
	}
	_ = s.memoryStore.DeleteArchives(ctx, orphanIDs)

	allMessageIDs := make([]int64, 0)
	seenMessageIDs := make(map[int64]struct{})
	for _, item := range liveMemories {
		for _, messageID := range item.MessageIDs {
			if _, seen := seenMessageIDs[messageID]; seen {
				continue
			}
			seenMessageIDs[messageID] = struct{}{}
			allMessageIDs = append(allMessageIDs, messageID)
		}
	}
	var messages []models.Message
	if len(allMessageIDs) > 0 {
		if err := database.DB.WithContext(ctx).
			Where("user_id = ? AND id IN ?", userID, allMessageIDs).
			Find(&messages).Error; err != nil {
			return nil, err
		}
	}
	messageByID := make(map[int64]models.Message, len(messages))
	for _, message := range messages {
		messageByID[message.ID] = message
	}

	results := make([]DeveloperArchiveSearchResult, 0, len(liveMemories))
	for _, item := range liveMemories {
		linkedMessages := make([]models.Message, 0, len(item.MessageIDs))
		for _, messageID := range item.MessageIDs {
			if message, ok := messageByID[messageID]; ok {
				linkedMessages = append(linkedMessages, message)
			}
		}
		archive := archiveByEmbeddingID[item.ID]
		results = append(results, DeveloperArchiveSearchResult{
			EmbeddingID: item.ID,
			ArchiveID:   archive.ID,
			Score:       item.Score,
			Summary:     item.Summary,
			MessageIDs:  append([]int64(nil), item.MessageIDs...),
			Messages:    linkedMessages,
			CreatedAt:   item.CreatedAt,
		})
	}
	return results, nil
}
