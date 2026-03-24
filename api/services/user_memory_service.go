package services

import (
	"encoding/json"
	"learnlang-api/database"
	"learnlang-api/models"

	"github.com/pgvector/pgvector-go"
)

type UserMemoryService struct{}

func NewUserMemoryService() *UserMemoryService {
	return &UserMemoryService{}
}

type UserMemoryListResult struct {
	Total    int64
	Memories []models.UserMemory
}

func (ums *UserMemoryService) CreateUserMemory(userID int64, content, embedding, memoryType string, importanceScore float64) (*models.UserMemory, error) {
	var embeddingSlice []float32
	if err := json.Unmarshal([]byte(embedding), &embeddingSlice); err != nil {
		return nil, err
	}

	memory := models.UserMemory{
		UserID:          userID,
		Content:         content,
		Embedding:       pgvector.NewVector(embeddingSlice),
		MemoryType:      memoryType,
		ImportanceScore: importanceScore,
	}

	if err := database.DB.Create(&memory).Error; err != nil {
		return nil, err
	}

	return &memory, nil
}

func (ums *UserMemoryService) GetUserMemory(memoryID int64) (*models.UserMemory, error) {
	var memory models.UserMemory
	if err := database.DB.First(&memory, memoryID).Error; err != nil {
		return nil, err
	}
	return &memory, nil
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
		memory.Content = content
	}
	if embedding != "" {
		var embeddingSlice []float32
		if err := json.Unmarshal([]byte(embedding), &embeddingSlice); err != nil {
			return nil, err
		}
		memory.Embedding = pgvector.NewVector(embeddingSlice)
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
