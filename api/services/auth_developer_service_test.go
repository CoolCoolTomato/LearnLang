package services

import (
	"context"
	"errors"
	"learnlang-api/config"
	"learnlang-api/database"
	"learnlang-api/models"
	"learnlang-api/utils"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func TestAuthServiceWithMockRedisAndDatabase(t *testing.T) {
	setupServiceTestDB(t)
	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })
	manager := utils.NewTokenManager(redisClient)
	service := NewAuthService(&config.Config{JWT: config.JWTConfig{Secret: "secret"}}, manager)

	email := " User@Example.com "
	user, token, err := service.Register(&email, nil, "password")
	if err != nil || user.Email == nil || *user.Email != "user@example.com" || user.Username != "user" || token == "" {
		t.Fatalf("Register() = %#v, %q, %v", user, token, err)
	}
	if valid, err := manager.ValidateToken(user.ID, token); err != nil || !valid {
		t.Fatalf("registered token = %v, %v", valid, err)
	}
	loggedIn, loginToken, err := service.Login(" USER@example.COM ", "password")
	if err != nil || loggedIn.ID != user.ID || loginToken == "" || loggedIn.LastActiveAt == nil {
		t.Fatalf("Login() = %#v, %q, %v", loggedIn, loginToken, err)
	}
	if _, _, err := service.Login("user@example.com", "wrong"); !errors.Is(err, utils.ErrInvalidCredentials) {
		t.Fatalf("wrong password error = %v", err)
	}
	if err := service.ChangePassword(user.ID, "new-password"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := service.Login("user@example.com", "new-password"); err != nil {
		t.Fatalf("new password login error = %v", err)
	}
	if err := service.Logout(user.ID); err != nil {
		t.Fatal(err)
	}
	if valid, _ := manager.ValidateToken(user.ID, loginToken); valid {
		t.Fatal("logout left token valid")
	}
	if err := service.ChangePassword(999, "password"); !errors.Is(err, utils.ErrUserNotFound) {
		t.Fatalf("missing user password error = %v", err)
	}
}

func TestAuthRegistrationValidation(t *testing.T) {
	service := &AuthService{}
	if _, _, err := service.Register(nil, nil, "password"); !errors.Is(err, utils.ErrRegistrationContact) {
		t.Fatalf("missing contact error = %v", err)
	}
	badEmail := "bad"
	if _, _, err := service.Register(&badEmail, nil, "password"); !errors.Is(err, utils.ErrInvalidEmail) {
		t.Fatalf("bad email error = %v", err)
	}
	longPhone := "123456789012345678901234567890123"
	if _, _, err := service.Register(nil, &longPhone, "password"); !errors.Is(err, utils.ErrInvalidPhone) {
		t.Fatalf("bad phone error = %v", err)
	}
}

func TestDeveloperDataServiceCRUD(t *testing.T) {
	setupServiceTestDB(t, &models.ConversationArchive{})
	user := models.User{Username: "dev", PasswordHash: "hash", Role: "developer"}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	message := models.Message{UserID: user.ID, Role: "user", TextContent: "hello"}
	if err := database.DB.Create(&message).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.DB.Create(&models.ScheduledTask{UserID: user.ID, FunctionName: "x", Status: "completed", ScheduledAt: time.Now()}).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.DB.Create(&models.VoiceFile{UserID: user.ID, VoiceURL: "x", FileSize: 123}).Error; err != nil {
		t.Fatal(err)
	}
	vectors := &fakeArchiveVectors{}
	service := NewDeveloperDataService(vectors)
	dashboard, err := service.Dashboard(user.ID)
	if err != nil || dashboard.Messages != 1 || dashboard.CompletedTasks != 1 || dashboard.VoiceFiles != 1 || dashboard.VoiceFileBytes != 123 {
		t.Fatalf("Dashboard() = %#v, %v", dashboard, err)
	}
	page, err := service.List(DeveloperResourceMessages, 0, 200)
	if err != nil || page.Total != 1 || page.Page != 1 || page.Size != 100 {
		t.Fatalf("List() = %#v, %v", page, err)
	}
	got, err := service.Get(DeveloperResourceMessages, message.ID)
	if err != nil || got.(*models.Message).ID != message.ID {
		t.Fatalf("Get() = %#v, %v", got, err)
	}
	created, err := service.Create(DeveloperResourceMessages, map[string]any{"user_id": user.ID, "role": "assistant", "text_content": "answer", "id": 999})
	if err != nil || created.(*models.Message).ID == 999 {
		t.Fatalf("Create() = %#v, %v", created, err)
	}
	createdMessage := created.(*models.Message)
	updated, err := service.Update(DeveloperResourceMessages, createdMessage.ID, map[string]any{"text_content": "updated"})
	if err != nil || updated.(*models.Message).TextContent != "updated" {
		t.Fatalf("Update() = %#v, %v", updated, err)
	}
	if _, err := service.Update(DeveloperResourceMessages, 999, map[string]any{"text_content": "x"}); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("missing Update() error = %v", err)
	}
	if err := service.Delete(context.Background(), DeveloperResourceMessages, createdMessage.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.DeleteMany(context.Background(), DeveloperResourceMessages, nil); err == nil {
		t.Fatal("DeleteMany(empty) succeeded")
	}
	deleted, err := service.DeleteMany(context.Background(), DeveloperResourceMessages, []int64{message.ID})
	if err != nil || deleted != 1 {
		t.Fatalf("DeleteMany() = %d, %v", deleted, err)
	}

	archive := models.ConversationArchive{UserID: user.ID, MessageIDs: []int64{1}, Summary: "summary", MessageCount: 1, EmbeddingID: "vector-id"}
	if err := database.DB.Create(&archive).Error; err != nil {
		t.Fatal(err)
	}
	if err := service.Delete(context.Background(), DeveloperResourceConversationArchives, archive.ID); err != nil {
		t.Fatal(err)
	}
	if len(vectors.ids) != 1 || vectors.ids[0] != "vector-id" {
		t.Fatalf("deleted vectors = %#v", vectors.ids)
	}
}

func TestDeveloperArchiveSearchValidation(t *testing.T) {
	service := NewDeveloperArchiveSearchService(nil, nil)
	if _, err := service.Search(context.Background(), 1, " ", 5); err == nil || err.Error() != "query is required" {
		t.Fatalf("empty query error = %v", err)
	}
	if _, err := service.Search(context.Background(), 1, "query", 5); err == nil || err.Error() != "memory store is not configured" {
		t.Fatalf("nil store error = %v", err)
	}
}
