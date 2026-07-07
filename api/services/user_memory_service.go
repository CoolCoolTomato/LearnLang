package services

import (
	"learnlang-api/database"
	"learnlang-api/models"
	"time"

	"gorm.io/gorm"
)

type UserMemoryService struct{}

func NewUserMemoryService() *UserMemoryService {
	return &UserMemoryService{}
}

type UserMemoryListResult struct {
	Total    int64
	Memories []models.UserMemory
}

type UserMemorySearchResult struct {
	Memory   models.UserMemory
	Messages []models.Message
	Distance float64
}

func (ums *UserMemoryService) CreateUserMemory(userID int64, content, embedding, memoryType string, importanceScore float64) (*models.UserMemory, error) {
	return ums.CreateUserMemorySummary(userID, content, embedding, memoryType, importanceScore, nil)
}

func (ums *UserMemoryService) CreateUserMemorySummaryFromMessages(userID int64, summary, embedding, memoryType string, importanceScore float64, messages []models.Message) (*models.UserMemory, error) {
	messageIDs := make([]int64, 0, len(messages))
	for _, message := range messages {
		if message.UserID == userID {
			messageIDs = append(messageIDs, message.ID)
		}
	}

	return ums.CreateUserMemorySummary(userID, summary, embedding, memoryType, importanceScore, messageIDs)
}

func (ums *UserMemoryService) CreateUserMemorySummary(userID int64, summary, embedding, memoryType string, importanceScore float64, messageIDs []int64) (*models.UserMemory, error) {
	var startedAt *time.Time
	var endedAt *time.Time
	if len(messageIDs) > 0 {
		var bounds struct {
			StartedAt time.Time
			EndedAt   time.Time
		}
		if err := database.DB.Model(&models.Message{}).
			Select("MIN(created_at) AS started_at, MAX(created_at) AS ended_at").
			Where("user_id = ? AND id IN ?", userID, messageIDs).
			Scan(&bounds).Error; err != nil {
			return nil, err
		}
		if !bounds.StartedAt.IsZero() {
			startedAt = &bounds.StartedAt
		}
		if !bounds.EndedAt.IsZero() {
			endedAt = &bounds.EndedAt
		}
	}

	memory := models.UserMemory{
		UserID:          userID,
		Summary:         summary,
		VectorID:        "",
		MemoryType:      memoryType,
		ImportanceScore: importanceScore,
		MessageCount:    len(messageIDs),
		StartedAt:       startedAt,
		EndedAt:         endedAt,
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&memory).Error; err != nil {
			return err
		}

		return ums.linkMessages(tx, memory.ID, messageIDs)
	}); err != nil {
		return nil, err
	}

	return &memory, nil
}

func (ums *UserMemoryService) LinkMessages(memoryID int64, messageIDs []int64) error {
	return ums.linkMessages(database.DB, memoryID, messageIDs)
}

func (ums *UserMemoryService) ReplaceMemoryMessages(memoryID int64, messageIDs []int64) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_memory_id = ?", memoryID).Delete(&models.UserMemoryMessage{}).Error; err != nil {
			return err
		}

		if err := ums.linkMessages(tx, memoryID, messageIDs); err != nil {
			return err
		}

		return tx.Model(&models.UserMemory{}).Where("id = ?", memoryID).Update("message_count", len(messageIDs)).Error
	})
}

func (ums *UserMemoryService) linkMessages(db *gorm.DB, memoryID int64, messageIDs []int64) error {
	if len(messageIDs) == 0 {
		return nil
	}

	links := make([]models.UserMemoryMessage, 0, len(messageIDs))
	for _, messageID := range messageIDs {
		links = append(links, models.UserMemoryMessage{
			UserMemoryID: memoryID,
			MessageID:    messageID,
		})
	}

	return db.Create(&links).Error
}

func (ums *UserMemoryService) GetUserMemory(memoryID int64) (*models.UserMemory, error) {
	var memory models.UserMemory
	if err := database.DB.First(&memory, memoryID).Error; err != nil {
		return nil, err
	}
	return &memory, nil
}

func (ums *UserMemoryService) UpdateUserMemoryVectorID(memoryID int64, vectorID string) error {
	return database.DB.Model(&models.UserMemory{}).
		Where("id = ?", memoryID).
		Update("vector_id", vectorID).Error
}

func (ums *UserMemoryService) GetUserMemoriesByVectorIDs(userID int64, vectorIDs []string) ([]models.UserMemory, error) {
	if len(vectorIDs) == 0 {
		return []models.UserMemory{}, nil
	}

	var memories []models.UserMemory
	if err := database.DB.Where("user_id = ? AND vector_id IN ?", userID, vectorIDs).Find(&memories).Error; err != nil {
		return nil, err
	}

	return memories, nil
}

func (ums *UserMemoryService) GetUserMemoryMessages(memoryID int64) ([]models.Message, error) {
	var links []models.UserMemoryMessage
	if err := database.DB.Preload("Message").
		Where("user_memory_id = ?", memoryID).
		Joins("JOIN messages ON messages.id = user_memory_messages.message_id").
		Order("messages.created_at ASC").
		Find(&links).Error; err != nil {
		return nil, err
	}

	messages := make([]models.Message, 0, len(links))
	for _, link := range links {
		messages = append(messages, link.Message)
	}

	return messages, nil
}

func (ums *UserMemoryService) ListUnlinkedMessagesForSummary(userID int64, limit int) ([]models.Message, error) {
	if limit <= 0 {
		limit = 20
	}

	var messages []models.Message
	if err := database.DB.Model(&models.Message{}).
		Joins("LEFT JOIN user_memory_messages ON user_memory_messages.message_id = messages.id").
		Where("messages.user_id = ? AND user_memory_messages.id IS NULL", userID).
		Order("messages.created_at ASC").
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}

func (ums *UserMemoryService) SearchUserMemorySummaries(userID int64, embedding string, limit int, maxDistance *float64) ([]UserMemorySearchResult, error) {
	if limit <= 0 {
		limit = 3
	}

	return []UserMemorySearchResult{}, nil
}

func (ums *UserMemoryService) ListUserMemories(page, pageSize int, userID, memoryType string) (*UserMemoryListResult, error) {
	query := database.DB.Model(&models.UserMemory{})

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if memoryType != "" {
		query = query.Where("memory_type = ?", memoryType)
	}

	var total int64
	query.Count(&total)

	var memories []models.UserMemory
	offset := (page - 1) * pageSize
	if err := query.Order("importance_score DESC, created_at DESC").Offset(offset).Limit(pageSize).Find(&memories).Error; err != nil {
		return nil, err
	}

	return &UserMemoryListResult{Total: total, Memories: memories}, nil
}

func (ums *UserMemoryService) UpdateUserMemory(memoryID int64, content, embedding, memoryType string, importanceScore float64) (*models.UserMemory, error) {
	var memory models.UserMemory
	if err := database.DB.First(&memory, memoryID).Error; err != nil {
		return nil, err
	}

	if content != "" {
		memory.Summary = content
	}
	if memoryType != "" {
		memory.MemoryType = memoryType
	}
	if importanceScore > 0 {
		memory.ImportanceScore = importanceScore
	}

	if err := database.DB.Save(&memory).Error; err != nil {
		return nil, err
	}

	return &memory, nil
}

func (ums *UserMemoryService) DeleteUserMemory(memoryID int64) error {
	return database.DB.Delete(&models.UserMemory{}, memoryID).Error
}
