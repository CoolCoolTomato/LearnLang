package services

import (
	"context"
	"fmt"
	"learnlang-api/database"
	"learnlang-api/models"
	"sort"
	"strings"

	"gorm.io/gorm"
)

const (
	defaultArchiveLookbackLimit = 120
	defaultArchiveReserveCount  = 12
)

type ConversationArchiveService struct{}

func NewConversationArchiveService() *ConversationArchiveService {
	return &ConversationArchiveService{}
}

type ArchiveWindow struct {
	Candidates []models.Message
	Reserved   []models.Message
}

type ArchiveSegmentInput struct {
	Summary    string
	MessageIDs []int64
}

func (s *ConversationArchiveService) GetArchiveWindow(ctx context.Context, userID int64) (*ArchiveWindow, error) {
	latestEndID, err := s.latestArchiveEndMessageID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var messages []models.Message
	if err := database.DB.WithContext(ctx).
		Where("user_id = ? AND id > ?", userID, latestEndID).
		Order("id ASC").
		Limit(defaultArchiveLookbackLimit + defaultArchiveReserveCount).
		Find(&messages).Error; err != nil {
		return nil, err
	}

	if len(messages) <= defaultArchiveReserveCount {
		return &ArchiveWindow{
			Candidates: []models.Message{},
			Reserved:   messages,
		}, nil
	}

	cutoff := len(messages) - defaultArchiveReserveCount
	return &ArchiveWindow{
		Candidates: messages[:cutoff],
		Reserved:   messages[cutoff:],
	}, nil
}

func (s *ConversationArchiveService) SaveArchiveSegments(ctx context.Context, userID int64, candidates []models.Message, segments []ArchiveSegmentInput) ([]models.ConversationArchive, error) {
	if len(candidates) == 0 || len(segments) == 0 {
		return []models.ConversationArchive{}, nil
	}
	if err := validateArchiveSegments(candidates, segments); err != nil {
		return nil, err
	}

	latestEndID, err := s.latestArchiveEndMessageID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var archives []models.ConversationArchive
	err = database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		currentEndID, err := s.latestArchiveEndMessageIDWithTx(tx, userID)
		if err != nil {
			return err
		}
		if currentEndID != latestEndID {
			return fmt.Errorf("conversation archive changed while archiving")
		}

		for _, segment := range segments {
			messageIDs := append([]int64(nil), segment.MessageIDs...)
			sort.Slice(messageIDs, func(i, j int) bool { return messageIDs[i] < messageIDs[j] })

			archive := models.ConversationArchive{
				UserID:         userID,
				StartMessageID: messageIDs[0],
				EndMessageID:   messageIDs[len(messageIDs)-1],
				MessageIDs:     messageIDs,
				Summary:        strings.TrimSpace(segment.Summary),
				MessageCount:   len(messageIDs),
			}
			if err := tx.Create(&archive).Error; err != nil {
				return err
			}

			archives = append(archives, archive)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return archives, nil
}

func (s *ConversationArchiveService) UpdateEmbeddingID(ctx context.Context, archiveID int64, embeddingID string) error {
	embeddingID = strings.TrimSpace(embeddingID)
	if embeddingID == "" {
		return nil
	}

	return database.DB.WithContext(ctx).
		Model(&models.ConversationArchive{}).
		Where("id = ?", archiveID).
		Update("embedding_id", embeddingID).Error
}

func (s *ConversationArchiveService) latestArchiveEndMessageID(ctx context.Context, userID int64) (int64, error) {
	return s.latestArchiveEndMessageIDWithTx(database.DB.WithContext(ctx), userID)
}

func (s *ConversationArchiveService) latestArchiveEndMessageIDWithTx(tx *gorm.DB, userID int64) (int64, error) {
	var archive models.ConversationArchive
	err := tx.Where("user_id = ?", userID).
		Order("end_message_id DESC").
		First(&archive).Error
	if err == nil {
		return archive.EndMessageID, nil
	}
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	return 0, err
}

func validateArchiveSegments(candidates []models.Message, segments []ArchiveSegmentInput) error {
	candidateIndexByID := make(map[int64]int, len(candidates))
	for idx, message := range candidates {
		candidateIndexByID[message.ID] = idx
	}

	seen := make(map[int64]struct{})
	expectedNextIndex := 0
	for _, segment := range segments {
		if strings.TrimSpace(segment.Summary) == "" {
			return fmt.Errorf("archive segment summary is required")
		}
		if len(segment.MessageIDs) == 0 {
			return fmt.Errorf("archive segment has no messages")
		}

		previousIndex := -1
		for _, messageID := range segment.MessageIDs {
			currentIndex, ok := candidateIndexByID[messageID]
			if !ok {
				return fmt.Errorf("message %d is not archivable in current window", messageID)
			}
			if _, ok := seen[messageID]; ok {
				return fmt.Errorf("message %d appears in multiple archive segments", messageID)
			}
			if currentIndex != expectedNextIndex {
				return fmt.Errorf("archive segments must form a contiguous prefix of candidate messages")
			}
			if previousIndex >= 0 && currentIndex != previousIndex+1 {
				return fmt.Errorf("archive segment messages must be contiguous")
			}

			seen[messageID] = struct{}{}
			previousIndex = currentIndex
			expectedNextIndex++
		}
	}

	return nil
}
