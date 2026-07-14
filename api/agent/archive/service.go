package archive

import (
	"context"
	"learnlang-api/agent/memory"
	"learnlang-api/models"
	"learnlang-api/services"
)

// Service coordinates the archive workflow and delegates each processing step
// to a focused component.
type Service struct {
	archiveService *services.ConversationArchiveService
	settings       *services.UserSettingsService
	segmenter      segmenter
	indexer        archiveIndexer
}

type segmenter interface {
	Generate(context.Context, *models.UserSettings, *services.ArchiveWindow) ([]archiveSegment, error)
}

type archiveIndexer interface {
	Index(context.Context, int64, *models.UserSettings, []models.ConversationArchive)
}

func NewService(archiveService *services.ConversationArchiveService, settings *services.UserSettingsService, memoryStore *memory.Store) *Service {
	return &Service{
		archiveService: archiveService,
		settings:       settings,
		segmenter:      llmSegmenter{},
		indexer:        newArchiveIndexer(archiveService, memoryStore),
	}
}

func (s *Service) Run(ctx context.Context, userID int64) error {
	window, err := s.archiveService.GetArchiveWindow(ctx, userID)
	if err != nil {
		return err
	}
	if len(window.Candidates) == 0 {
		return nil
	}

	settings, err := s.settings.GetUserSettings(userID)
	if err != nil {
		return err
	}

	segments, err := s.segmenter.Generate(ctx, settings, window)
	if err != nil {
		return err
	}
	if len(segments) == 0 {
		return nil
	}

	inputs := make([]services.ArchiveSegmentInput, 0, len(segments))
	for _, segment := range segments {
		inputs = append(inputs, services.ArchiveSegmentInput{
			Summary:    segment.Summary,
			MessageIDs: segment.MessageIDs,
		})
	}

	archives, err := s.archiveService.SaveArchiveSegments(ctx, userID, window.Candidates, inputs)
	if err != nil {
		return err
	}

	s.indexer.Index(ctx, userID, settings, archives)
	return nil
}
