package services

import (
	"context"
	"errors"
	"fmt"
	"learnlang-api/aiusage"
	"learnlang-api/database"
	"learnlang-api/models"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"
)

var ErrAIUsageInvalidInput = errors.New("invalid AI usage input")

type AIUsageRecord struct {
	UserID    int64
	Operation string
	Model     string
	Usage     float64
	Unit      string
	Status    string
}

type AIUsageQuery struct {
	UserID    int64
	Operation string
	Page      int
	PageSize  int
}

type AIUsagePage struct {
	Items    []models.AIUsageEvent `json:"items"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}

type AIUsageSummary struct {
	Operation    string  `json:"operation"`
	Unit         string  `json:"unit"`
	Usage        float64 `json:"usage"`
	RequestCount int64   `json:"request_count"`
}

type AIUsageService struct{}

var _ aiusage.Recorder = (*AIUsageService)(nil)

func (s *AIUsageService) RecordAIUsage(ctx context.Context, input aiusage.Record) error {
	return s.Record(ctx, AIUsageRecord{UserID: input.UserID, Operation: input.Operation, Model: input.Model, Usage: input.Usage, Unit: input.Unit, Status: input.Status})
}

func NewAIUsageService() *AIUsageService {
	return &AIUsageService{}
}

func (s *AIUsageService) Record(ctx context.Context, input AIUsageRecord) error {
	input.Operation = strings.TrimSpace(input.Operation)
	input.Model = strings.TrimSpace(input.Model)
	input.Unit = strings.TrimSpace(input.Unit)
	input.Status = strings.TrimSpace(input.Status)
	if err := validateAIUsageRecord(input); err != nil {
		return err
	}
	event := &models.AIUsageEvent{
		UserID: input.UserID, Operation: input.Operation, Model: input.Model,
		Usage: input.Usage, Unit: input.Unit, Status: input.Status, CreatedAt: time.Now().UTC(),
	}
	return database.DB.WithContext(ctx).Create(event).Error
}

func (s *AIUsageService) List(ctx context.Context, query AIUsageQuery) (*AIUsagePage, error) {
	db, query, err := buildAIUsageQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}
	items := make([]models.AIUsageEvent, 0)
	if err := db.Order("created_at DESC, id DESC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&items).Error; err != nil {
		return nil, err
	}
	return &AIUsagePage{Items: items, Total: total, Page: query.Page, PageSize: query.PageSize}, nil
}

func (s *AIUsageService) Summary(ctx context.Context, query AIUsageQuery) ([]AIUsageSummary, error) {
	db, _, err := buildAIUsageQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	result := make([]AIUsageSummary, 0)
	if err := db.Select("operation, unit, COALESCE(SUM(usage), 0) AS usage, COUNT(*) AS request_count").
		Group("operation, unit").Order("operation ASC, unit ASC").Scan(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func validateAIUsageRecord(input AIUsageRecord) error {
	if input.UserID <= 0 || input.Model == "" || len(input.Model) > 128 {
		return ErrAIUsageInvalidInput
	}
	if math.IsNaN(input.Usage) || math.IsInf(input.Usage, 0) || input.Usage < 0 {
		return ErrAIUsageInvalidInput
	}
	expectedUnit, ok := aiUsageUnitForOperation(input.Operation)
	if !ok || input.Unit != expectedUnit {
		return ErrAIUsageInvalidInput
	}
	if input.Status != models.AIUsageStatusSucceeded && input.Status != models.AIUsageStatusFailed {
		return ErrAIUsageInvalidInput
	}
	return nil
}

func aiUsageUnitForOperation(operation string) (string, bool) {
	switch operation {
	case models.AIOperationChat, models.AIOperationTranslation, models.AIOperationEmbedding:
		return models.AIUsageUnitTokens, true
	case models.AIOperationSTT:
		return models.AIUsageUnitSeconds, true
	case models.AIOperationTTS:
		return models.AIUsageUnitCharacters, true
	default:
		return "", false
	}
}

func buildAIUsageQuery(ctx context.Context, query AIUsageQuery) (*gorm.DB, AIUsageQuery, error) {
	if query.UserID <= 0 {
		return nil, query, ErrAIUsageInvalidInput
	}
	query.Operation = strings.TrimSpace(query.Operation)
	if query.Operation != "" {
		if _, ok := aiUsageUnitForOperation(query.Operation); !ok {
			return nil, query, fmt.Errorf("%w: unsupported operation", ErrAIUsageInvalidInput)
		}
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}
	db := database.DB.WithContext(ctx).Model(&models.AIUsageEvent{}).Where("user_id = ?", query.UserID)
	if query.Operation != "" {
		db = db.Where("operation = ?", query.Operation)
	}
	return db, query, nil
}
