package agent

import (
	"context"
	"io"
	"learnlang-api/agent/archive"
	"learnlang-api/agent/memory"
	"learnlang-api/config"
	"learnlang-api/models"
	"learnlang-api/services"
	"strings"
	"time"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

type ChatService struct {
	runtime     *services.ChatRuntimeService
	memoryStore *memory.Store
	agent       *Service
	archiver    *archive.Service
}

func NewChatService(runtime *services.ChatRuntimeService, memoryStore *memory.Store, archiver *archive.Service) *ChatService {
	return &ChatService{
		runtime:     runtime,
		memoryStore: memoryStore,
		agent: NewService(Config{
			MemoryStore: memoryStore,
			Runtime:     runtime,
		}),
		archiver: archiver,
	}
}

func NewMemoryStore(milvusCfg config.MilvusConfig, client *milvusclient.Client) *memory.Store {
	return memory.NewStore(memory.Config{
		Collection: milvusCfg.Collection,
		Dimension:  milvusCfg.Dimension,
	}, client)
}

func (s *ChatService) TranscribeAudio(ctx context.Context, userID int64, audioFile io.Reader) (string, *int64, error) {
	return s.runtime.TranscribeAudio(ctx, userID, audioFile)
}

func (s *ChatService) Chat(ctx context.Context, userID int64, userMessage string) (*models.Message, error) {
	message, err := s.runtime.CreateUserMessage(ctx, userID, userMessage, nil, "text")
	if err != nil {
		return nil, err
	}

	if err := s.runtime.CancelPendingWaitMessages(userID); err != nil {
		return nil, err
	}

	go s.processAIResponse(userID, message)
	return message, nil
}

func (s *ChatService) ChatWithVoice(ctx context.Context, userID int64, userMessage string, voiceFileID *int64) (*models.Message, error) {
	message, err := s.runtime.CreateUserMessage(ctx, userID, userMessage, voiceFileID, "voice")
	if err != nil {
		return nil, err
	}

	if err := s.runtime.CancelPendingWaitMessages(userID); err != nil {
		return nil, err
	}

	go s.processAIResponse(userID, message)
	return message, nil
}

func (s *ChatService) GetChatHistory(userID int64, beforeID *int64) ([]models.Message, error) {
	return s.runtime.GetChatHistory(userID, beforeID)
}

func (s *ChatService) ProcessInstantAIResponse(userID int64, messageID int64) {
	s.processInstantAIResponse(userID, messageID)
}

func (s *ChatService) processAIResponse(userID int64, userMessage *models.Message) {
	ctx := context.Background()
	result, err := s.runAgentChat(ctx, userID, userMessage.ID, userMessage.TextContent, false)
	if err != nil {
		return
	}

	if result.WaitForNextMsg {
		_ = s.runtime.ScheduleWaitMessage(ctx, userID, userMessage.ID, time.Now().Add(30*time.Second))
	}

	s.archiveConversation(userID)
}

func (s *ChatService) processInstantAIResponse(userID int64, messageID int64) {
	ctx := context.Background()
	recentMessages, err := s.runtime.GetRecentConversation(userID)
	if err != nil {
		return
	}

	var target *models.Message
	for i := range recentMessages {
		msg := recentMessages[i]
		if msg.ID == messageID && msg.Role == "user" {
			target = &msg
			break
		}
	}
	if target == nil {
		for i := len(recentMessages) - 1; i >= 0; i-- {
			msg := recentMessages[i]
			if msg.Role == "user" && msg.ID <= messageID {
				target = &msg
				break
			}
		}
	}
	if target == nil {
		return
	}

	_, err = s.runAgentChat(ctx, userID, target.ID, target.TextContent, true)
	if err != nil {
		return
	}

	s.archiveConversation(userID)
}

func (s *ChatService) archiveConversation(userID int64) {
	if s.archiver == nil {
		return
	}

	go func() {
		_ = s.archiver.Run(context.Background(), userID)
	}()
}

func (s *ChatService) runAgentChat(ctx context.Context, userID, currentMessageID int64, userInput string, instant bool) (*ChatResult, error) {
	settings, err := s.runtime.UserSettings(userID)
	if err != nil {
		return nil, err
	}

	timezone := strings.TrimSpace(settings.Timezone)
	if timezone == "" {
		timezone = "UTC"
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
		timezone = "UTC"
	}

	agentSettings := toAgentSettings(settings)
	result, err := s.agent.RunChat(ctx, ChatRequest{
		UserID:           userID,
		CurrentMessageID: currentMessageID,
		UserInput:        userInput,
		Settings:         agentSettings,
		Instant:          instant,
		CurrentTime:      time.Now().In(loc).Format(time.RFC3339),
		Timezone:         timezone,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func toAgentSettings(settings *models.UserSettings) UserSettings {
	apiKey := strings.TrimSpace(settings.APIKey)
	if apiKey == "" {
		apiKey = strings.TrimSpace(settings.STTAPIKey)
	}

	apiBaseURL := strings.TrimSpace(settings.APIBaseURL)
	if apiBaseURL == "" {
		apiBaseURL = strings.TrimSpace(settings.STTAPIBaseURL)
	}

	return UserSettings{
		APIKey:              apiKey,
		APIBaseURL:          apiBaseURL,
		Model:               settings.Model,
		LLMType:             settings.LLMType,
		EmbeddingAPIKey:     settings.EmbeddingAPIKey,
		EmbeddingAPIBaseURL: settings.EmbeddingAPIBaseURL,
		EmbeddingModel:      settings.EmbeddingModel,
		NativeLanguage:      settings.NativeLanguage,
		TargetLanguage:      settings.TargetLanguage,
	}
}
