package agent

import (
	"learnlang-api/config"
	"learnlang-api/models"
	"testing"
)

func TestNewServiceAndObjectParams(t *testing.T) {
	service := NewService(Config{})
	if service == nil {
		t.Fatal("NewService() returned nil")
	}
	params := objectParams(nil)
	if params == nil {
		t.Fatal("objectParams() returned nil")
	}
}

func TestNewChatServiceAndMemoryStore(t *testing.T) {
	service := NewChatService(nil, nil, nil, nil)
	if service == nil || service.agent == nil {
		t.Fatalf("NewChatService() = %#v", service)
	}
	store := NewMemoryStore(config.MilvusConfig{Collection: "custom", Dimension: 8}, nil)
	if store == nil {
		t.Fatal("NewMemoryStore() returned nil")
	}
}

func TestToAgentSettingsUsesFallbacks(t *testing.T) {
	settings := &models.UserSettings{
		STTAPIKey: " stt-key ", STTAPIBaseURL: " https://stt.example ",
		Model: "model", LLMType: "openai", EmbeddingAPIKey: "embed",
		EmbeddingAPIBaseURL: "https://embed.example", EmbeddingModel: "embedding",
		NativeLanguage: "zh", TargetLanguage: "en",
	}
	got := toAgentSettings(settings)
	if got.APIKey != "stt-key" || got.APIBaseURL != "https://stt.example" || got.Model != "model" || got.EmbeddingModel != "embedding" || got.NativeLanguage != "zh" || got.TargetLanguage != "en" {
		t.Fatalf("toAgentSettings() = %#v", got)
	}
	settings.APIKey = " primary "
	settings.APIBaseURL = " https://primary.example "
	got = toAgentSettings(settings)
	if got.APIKey != "primary" || got.APIBaseURL != "https://primary.example" {
		t.Fatalf("primary settings = %#v", got)
	}
}

func TestArchiveConversationWithoutArchiver(t *testing.T) {
	service := &ChatService{}
	service.archiveConversation(1)
	if _, exists := service.archiveLocks.Load(1); exists {
		t.Fatal("nil archiver created an archive lock")
	}
}
