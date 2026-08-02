package services

import (
	"context"
	"errors"
	"fmt"
	"learnlang-api/database"
	"learnlang-api/models"
	"learnlang-api/utils"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupServiceTestDB(t *testing.T, additionalModels ...any) *gorm.DB {
	t.Helper()
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", name)), &gorm.Config{})
	if err != nil {
		t.Fatalf("open in-memory database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)

	modelTypes := []any{
		&models.User{}, &models.UserSettings{}, &models.UserProfileSummary{},
		&models.VoiceFile{}, &models.Message{}, &models.ScheduledTask{},
	}
	modelTypes = append(modelTypes, additionalModels...)
	if err := db.AutoMigrate(modelTypes...); err != nil {
		t.Fatalf("migrate in-memory database: %v", err)
	}

	previous := database.DB
	database.DB = db
	t.Cleanup(func() {
		database.DB = previous
		_ = sqlDB.Close()
	})
	return db
}

func stringPointer(value string) *string { return &value }

func TestUserServiceCRUD(t *testing.T) {
	setupServiceTestDB(t)
	service := NewUserService()
	email := "first@example.com"
	phone := "123"
	user, err := service.CreateUser(&email, &phone, "first", "password", "")
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if user.ID == 0 || user.Role != "user" || !utils.CheckPassword("password", user.PasswordHash) {
		t.Fatalf("created user = %#v", user)
	}
	var settingsCount, summaryCount int64
	database.DB.Model(&models.UserSettings{}).Where("user_id = ?", user.ID).Count(&settingsCount)
	database.DB.Model(&models.UserProfileSummary{}).Where("user_id = ?", user.ID).Count(&summaryCount)
	if settingsCount != 1 || summaryCount != 1 {
		t.Fatalf("related rows = settings %d, summaries %d", settingsCount, summaryCount)
	}

	if _, err := service.CreateUser(&email, nil, "duplicate", "password", "user"); !errors.Is(err, utils.ErrEmailExists) {
		t.Fatalf("duplicate email error = %v", err)
	}
	if _, err := service.CreateUser(nil, &phone, "duplicate", "password", "user"); !errors.Is(err, utils.ErrPhoneExists) {
		t.Fatalf("duplicate phone error = %v", err)
	}

	secondEmail := "second@example.com"
	second, err := service.CreateUser(&secondEmail, nil, "second", "password", "admin")
	if err != nil {
		t.Fatal(err)
	}
	list, err := service.ListUsers(1, 10, "first@", "", "first")
	if err != nil || list.Total != 1 || len(list.Users) != 1 || list.Users[0].ID != user.ID {
		t.Fatalf("ListUsers() = %#v, %v", list, err)
	}

	newEmail := "updated@example.com"
	newPhone := "456"
	updated, err := service.UpdateUser(user.ID, &newEmail, &newPhone, "updated", "new-password", "admin")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Username != "updated" || updated.Role != "admin" || !utils.CheckPassword("new-password", updated.PasswordHash) {
		t.Fatalf("updated user = %#v", updated)
	}
	if _, err := service.UpdateUser(user.ID, &secondEmail, nil, "", "", ""); !errors.Is(err, utils.ErrEmailExists) {
		t.Fatalf("conflicting update error = %v", err)
	}
	if _, err := service.UpdateProfile(user.ID, nil, nil, "profile-name"); err != nil {
		t.Fatal(err)
	}
	avatar, err := service.UpdateAvatar(user.ID, "avatar.png")
	if err != nil || avatar.AvatarURL != "avatar.png" {
		t.Fatalf("UpdateAvatar() = %#v, %v", avatar, err)
	}
	if err := service.DeleteUser(second.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.GetUser(second.ID); !errors.Is(err, utils.ErrUserNotFound) {
		t.Fatalf("GetUser(deleted) error = %v", err)
	}
	if err := service.DeleteUser(9999); !errors.Is(err, utils.ErrUserNotFound) {
		t.Fatalf("DeleteUser(missing) error = %v", err)
	}
}

func TestUserSettingsService(t *testing.T) {
	setupServiceTestDB(t)
	service := NewUserSettingsService()
	settings, err := service.GetUserSettings(22)
	if err != nil || settings.UserID != 22 {
		t.Fatalf("GetUserSettings() = %#v, %v", settings, err)
	}
	updated, err := service.UpdateUserSettings(22, map[string]interface{}{
		"api_base_url": "https://api.example", "api_key": "secret", "model": "model",
		"llm_type": "anthropic", "embedding_api_base_url": "https://embed.example",
		"embedding_api_key": "embed-key", "embedding_model": "embed-model", "embedding_dimension": 1024,
		"stt_api_base_url": "https://stt.example", "stt_api_key": "stt-key", "stt_model": "stt-model",
		"tts_api_base_url": "https://tts.example", "tts_api_key": "tts-key", "tts_model": "tts-model", "tts_voice": "voice",
		"native_language": "zh-CN", "target_language": "en-US", "timezone": "Asia/Shanghai",
		"ignored": "value", "theme": "dark", "api_key_wrong_type": 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.APIKey != "secret" || updated.LLMType != models.LLMTypeAnthropic || updated.EmbeddingModel != "embed-model" || updated.EmbeddingDimension != 1024 || updated.STTModel != "stt-model" || updated.TTSVoice != "voice" || updated.Timezone != "Asia/Shanghai" {
		t.Fatalf("updated settings = %#v", updated)
	}
	if _, err := service.UpdateUserSettings(22, map[string]interface{}{"llm_type": "claude"}); !errors.Is(err, ErrInvalidLLMType) {
		t.Fatalf("invalid LLM type error = %v", err)
	}
	if _, err := service.UpdateUserSettings(999, map[string]interface{}{"model": "x"}); err == nil {
		t.Fatal("UpdateUserSettings() updated a missing row")
	}
}

func TestMessageServiceCRUD(t *testing.T) {
	setupServiceTestDB(t)
	service := NewMessageService()
	voice := models.VoiceFile{UserID: 1, VoiceURL: "voice.mp3"}
	if err := database.DB.Create(&voice).Error; err != nil {
		t.Fatal(err)
	}
	message, err := service.CreateMessage(context.Background(), 1, "user", "hello", "你好", &voice.ID, "voice")
	if err != nil || message.ID == 0 {
		t.Fatalf("CreateMessage() = %#v, %v", message, err)
	}
	_, _ = service.CreateMessage(context.Background(), 2, "user", "other", "", nil, "text")
	got, err := service.GetMessage(message.ID)
	if err != nil || got.TextContent != "hello" {
		t.Fatalf("GetMessage() = %#v, %v", got, err)
	}
	list, err := service.ListMessages(1, 10, "1")
	if err != nil || list.Total != 1 || len(list.Messages) != 1 {
		t.Fatalf("ListMessages() = %#v, %v", list, err)
	}
	updated, err := service.UpdateMessage(message.ID, "assistant", "answer", "回答", nil, "text")
	if err != nil || updated.Role != "assistant" || updated.TextContent != "answer" {
		t.Fatalf("UpdateMessage() = %#v, %v", updated, err)
	}
	if err := service.DeleteMessage(message.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.GetMessage(message.ID); err == nil {
		t.Fatal("deleted message still exists")
	}
}

func TestVoiceFileServiceCRUD(t *testing.T) {
	setupServiceTestDB(t)
	service := NewVoiceFileService()
	first := &models.VoiceFile{UserID: 1, VoiceRole: "user", VoiceURL: "one.mp3", Duration: 2, FileSize: 10}
	second := &models.VoiceFile{UserID: 2, VoiceURL: "two.mp3"}
	if err := service.CreateVoiceFile(first); err != nil {
		t.Fatal(err)
	}
	if err := service.CreateVoiceFile(second); err != nil {
		t.Fatal(err)
	}
	got, err := service.GetVoiceFile(first.ID)
	if err != nil || got.VoiceURL != "one.mp3" {
		t.Fatalf("GetVoiceFile() = %#v, %v", got, err)
	}
	list, err := service.ListVoiceFiles(&first.UserID, 1, 10)
	if err != nil || len(list) != 1 || list[0].ID != first.ID {
		t.Fatalf("ListVoiceFiles() = %#v, %v", list, err)
	}
	if err := service.UpdateVoiceFile(first, map[string]interface{}{"duration": 5}); err != nil {
		t.Fatal(err)
	}
	got, _ = service.GetVoiceFile(first.ID)
	if got.Duration != 5 {
		t.Fatalf("updated duration = %d", got.Duration)
	}
	if err := service.DeleteVoiceFile(first.ID); err != nil {
		t.Fatal(err)
	}
}

func TestScheduledTaskService(t *testing.T) {
	setupServiceTestDB(t)
	service := NewScheduledTaskService()
	scheduledAt := time.Now().Add(-time.Minute)
	task, err := service.CreateTask(context.Background(), 3, "success", "args", scheduledAt)
	if err != nil || task.Status != "pending" || task.ScheduledAt.Location() != time.UTC {
		t.Fatalf("CreateTask() = %#v, %v", task, err)
	}
	called := ""
	service.RegisterHandler("success", func(args string) error { called = args; return nil })
	service.executeTask(task)
	if called != "args" || task.Status != "completed" {
		t.Fatalf("successful task = %#v, called %q", task, called)
	}

	failed, _ := service.CreateTask(context.Background(), 3, "failed", "bad", scheduledAt)
	service.RegisterHandler("failed", func(string) error { return errors.New("mock error") })
	service.executeTask(failed)
	if failed.Status != "failed" {
		t.Fatalf("failed task status = %q", failed.Status)
	}
	missing, _ := service.CreateTask(context.Background(), 4, "missing", "", scheduledAt)
	service.executeTask(missing)
	if missing.Status != "failed" {
		t.Fatalf("missing-handler task status = %q", missing.Status)
	}

	pending, _ := service.CreateTask(context.Background(), 3, "success", "pending", scheduledAt)
	service.processPendingTasks()
	got, err := service.GetTask(pending.ID)
	if err != nil || got.Status != "completed" {
		t.Fatalf("processed task = %#v, %v", got, err)
	}
	status := "completed"
	userID := int64(3)
	list, err := service.ListTasks(&userID, &status, 1, 10)
	if err != nil || len(list) < 2 {
		t.Fatalf("ListTasks() = %#v, %v", list, err)
	}
	if err := service.UpdateTask(got, map[string]interface{}{"status": "pending"}); err != nil {
		t.Fatal(err)
	}
	if err := service.DeleteTask(got.ID); err != nil {
		t.Fatal(err)
	}
}

func TestScheduledTaskStartSchedulerStopsWithContext(t *testing.T) {
	service := NewScheduledTaskService()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan struct{})
	go func() {
		service.StartScheduler(ctx)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("StartScheduler() did not stop after cancellation")
	}
}

func TestUserProfileSummaryService(t *testing.T) {
	setupServiceTestDB(t)
	service := NewUserProfileSummaryService()
	summary, err := service.GetUserProfileSummary(4)
	if err != nil || summary.UserID != 4 || summary.Summary != "" {
		t.Fatalf("GetUserProfileSummary() = %#v, %v", summary, err)
	}
	updated, err := service.UpdateUserProfileSummary(4, "likes Go")
	if err != nil || updated.Summary != "likes Go" {
		t.Fatalf("UpdateUserProfileSummary() = %#v, %v", updated, err)
	}
	unchanged, err := service.UpdateUserProfileSummary(4, "")
	if err != nil || unchanged.Summary != "likes Go" {
		t.Fatalf("empty update = %#v, %v", unchanged, err)
	}
}

func TestEnsureWelcomeMessage(t *testing.T) {
	setupServiceTestDB(t)
	user := models.User{Username: "u", PasswordHash: "hash", Role: "user"}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	settings := models.UserSettings{UserID: user.ID, Timezone: "Asia/Shanghai", NativeLanguage: "zh-CN", TargetLanguage: "en-US"}
	if err := database.DB.Create(&settings).Error; err != nil {
		t.Fatal(err)
	}
	runtime := &ChatRuntimeService{}
	if err := runtime.ensureWelcomeMessage(context.Background(), user.ID); err != nil {
		t.Fatal(err)
	}
	if err := runtime.ensureWelcomeMessage(context.Background(), user.ID); err != nil {
		t.Fatal(err)
	}
	var messages []models.Message
	if err := database.DB.Where("user_id = ?", user.ID).Find(&messages).Error; err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 || messages[0].Role != "assistant" || messages[0].TextContent == "" || messages[0].Translation == "" {
		t.Fatalf("welcome messages = %#v", messages)
	}
}
