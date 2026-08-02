package services

import (
	"context"
	"errors"
	"learnlang-api/database"
	"learnlang-api/models"
	"testing"
)

func TestAIUsageRecordListAndSummary(t *testing.T) {
	setupServiceTestDB(t, &models.AIUsageEvent{})
	service := NewAIUsageService()
	ctx := context.Background()
	for _, record := range []AIUsageRecord{
		{UserID: 1, Operation: models.AIOperationChat, Model: "chat-model", Usage: 12, Unit: models.AIUsageUnitTokens, Status: models.AIUsageStatusSucceeded},
		{UserID: 1, Operation: models.AIOperationChat, Model: "chat-model", Usage: 3, Unit: models.AIUsageUnitTokens, Status: models.AIUsageStatusFailed},
		{UserID: 2, Operation: models.AIOperationTTS, Model: "tts-model", Usage: 30, Unit: models.AIUsageUnitCharacters, Status: models.AIUsageStatusSucceeded},
	} {
		if err := service.Record(ctx, record); err != nil {
			t.Fatalf("Record(%#v): %v", record, err)
		}
	}
	page, err := service.List(ctx, AIUsageQuery{UserID: 1})
	if err != nil || page.Total != 2 || len(page.Items) != 2 {
		t.Fatalf("List() = %#v, %v", page, err)
	}
	for _, event := range page.Items {
		if event.UserID != 1 {
			t.Fatalf("List() leaked user %d", event.UserID)
		}
	}
	summary, err := service.Summary(ctx, AIUsageQuery{UserID: 1})
	if err != nil || len(summary) != 1 || summary[0].Usage != 15 || summary[0].RequestCount != 2 {
		t.Fatalf("Summary() = %#v, %v", summary, err)
	}
}

func TestAIUsageValidation(t *testing.T) {
	setupServiceTestDB(t, &models.AIUsageEvent{})
	service := NewAIUsageService()
	valid := AIUsageRecord{UserID: 1, Operation: models.AIOperationSTT, Model: "whisper", Usage: 1.5, Unit: models.AIUsageUnitSeconds, Status: models.AIUsageStatusSucceeded}
	for name, mutate := range map[string]func(*AIUsageRecord){
		"missing user":      func(record *AIUsageRecord) { record.UserID = 0 },
		"unknown operation": func(record *AIUsageRecord) { record.Operation = "archive" },
		"wrong unit":        func(record *AIUsageRecord) { record.Unit = models.AIUsageUnitTokens },
		"negative usage":    func(record *AIUsageRecord) { record.Usage = -1 },
		"unknown status":    func(record *AIUsageRecord) { record.Status = "pending" },
	} {
		t.Run(name, func(t *testing.T) {
			record := valid
			mutate(&record)
			if err := service.Record(context.Background(), record); !errors.Is(err, ErrAIUsageInvalidInput) {
				t.Fatalf("Record() error = %v", err)
			}
		})
	}
	var count int64
	if err := database.DB.Model(&models.AIUsageEvent{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("invalid records count = %d, %v", count, err)
	}
}
