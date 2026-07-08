package agent

import (
	"context"
	"fmt"
	"io"
	"learnlang-api/agent/memory"
	"learnlang-api/config"
	"learnlang-api/models"
	"learnlang-api/services"
	"net"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type ChatService struct {
	runtime     *services.ChatRuntimeService
	memoryStore *memory.Store
	agent       *Service
}

func NewChatService(runtime *services.ChatRuntimeService, milvusCfg config.MilvusConfig) *ChatService {
	memoryStore := memory.NewStore(memory.Config{
		Address:    milvusAddress(milvusCfg),
		Collection: milvusCfg.Collection,
		Dimension:  milvusCfg.Dimension,
	})

	return &ChatService{
		runtime:     runtime,
		memoryStore: memoryStore,
		agent: NewService(Config{
			MemoryStore: memoryStore,
			Runtime:     runtime,
		}),
	}
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
	result, settings, err := s.runAgentChat(ctx, userID, userMessage.TextContent, false)
	if err != nil {
		return
	}

	if result.WaitForNextMsg {
		_ = s.runtime.ScheduleWaitMessage(ctx, userID, userMessage.ID, time.Now().Add(30*time.Second))
	}

	s.updateMemory(ctx, userID, result, settings, append([]int64{userMessage.ID}, result.MessageIDs...))
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

	result, settings, err := s.runAgentChat(ctx, userID, target.TextContent, true)
	if err != nil {
		return
	}

	s.updateMemory(ctx, userID, result, settings, append([]int64{target.ID}, result.MessageIDs...))
}

func (s *ChatService) runAgentChat(ctx context.Context, userID int64, userInput string, instant bool) (*ChatResult, *models.UserSettings, error) {
	settings, err := s.runtime.UserSettings(userID)
	if err != nil {
		return nil, nil, err
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
		UserID:      userID,
		UserInput:   userInput,
		Settings:    agentSettings,
		Instant:     instant,
		CurrentTime: time.Now().In(loc).Format(time.RFC3339),
		Timezone:    timezone,
	})
	if err != nil {
		return nil, nil, err
	}

	return result, settings, nil
}

func (s *ChatService) updateMemory(ctx context.Context, userID int64, result *ChatResult, settings *models.UserSettings, messageIDs []int64) {
	if result == nil || result.Memory == nil || !result.Memory.ShouldStore {
		return
	}

	content := strings.TrimSpace(result.Memory.SemanticContent)
	if content == "" || s.memoryStore == nil {
		return
	}

	embedding, err := createMemoryEmbedding(ctx, settings, content)
	if err != nil {
		return
	}

	memoryType := strings.TrimSpace(result.Memory.MemoryType)
	if memoryType == "" {
		memoryType = "conversation"
	}

	_, _ = s.memoryStore.InsertSummary(ctx, userID, content, memoryType, result.Memory.Importance, messageIDs, embedding)
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

func createMemoryEmbedding(ctx context.Context, settings *models.UserSettings, text string) ([]float32, error) {
	apiKey := strings.TrimSpace(settings.EmbeddingAPIKey)
	if apiKey == "" {
		apiKey = strings.TrimSpace(settings.APIKey)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("embedding api key is required")
	}

	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	apiBaseURL := strings.TrimSpace(settings.EmbeddingAPIBaseURL)
	if apiBaseURL == "" {
		apiBaseURL = strings.TrimSpace(settings.APIBaseURL)
	}
	if apiBaseURL != "" {
		opts = append(opts, option.WithBaseURL(apiBaseURL))
	}

	model := strings.TrimSpace(settings.EmbeddingModel)
	if model == "" {
		model = "text-embedding-3-small"
	}

	client := openai.NewClient(opts...)
	resp, err := client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model: openai.EmbeddingModel(model),
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: []string{text},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("embedding response is empty")
	}

	vector := make([]float32, 0, len(resp.Data[0].Embedding))
	for _, value := range resp.Data[0].Embedding {
		vector = append(vector, float32(value))
	}
	return vector, nil
}

func milvusAddress(cfg config.MilvusConfig) string {
	host := strings.TrimSpace(cfg.Host)
	port := strings.TrimSpace(cfg.Port)
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "19530"
	}
	if strings.Contains(host, ":") {
		return host
	}
	return net.JoinHostPort(host, port)
}
