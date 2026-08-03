package agent

import (
	"context"
	"io"
	"learnlang-api/agent/archive"
	"learnlang-api/agent/memory"
	"learnlang-api/aiusage"
	"learnlang-api/config"
	"learnlang-api/models"
	"learnlang-api/services"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

type ChatService struct {
	runtime      *services.ChatRuntimeService
	memoryStore  *memory.Store
	agent        *Service
	archiver     *archive.Service
	agentRuns    agentRunCoordinator
	archiveLocks sync.Map
}

func NewChatService(runtime *services.ChatRuntimeService, memoryStore *memory.Store, archiver *archive.Service, vocabularyService *services.VocabularyService, usage ...aiusage.Recorder) *ChatService {
	var usageRecorder aiusage.Recorder
	if len(usage) > 0 {
		usageRecorder = usage[0]
	}
	return &ChatService{
		runtime:     runtime,
		memoryStore: memoryStore,
		agent: NewService(Config{
			MemoryStore:       memoryStore,
			Runtime:           runtime,
			VocabularyService: vocabularyService,
			UsageRecorder:     usageRecorder,
		}),
		archiver: archiver,
	}
}

func NewMemoryStore(milvusCfg config.MilvusConfig, client *milvusclient.Client) *memory.Store {
	return memory.NewStore(memory.Config{
		Collection: milvusCfg.Collection,
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

	s.archiveConversation(userID)
	s.restartAIResponse(userID, message)
	return message, nil
}

func (s *ChatService) ChatWithVoice(ctx context.Context, userID int64, userMessage string, voiceFileID *int64) (*models.Message, error) {
	message, err := s.runtime.CreateUserMessage(ctx, userID, userMessage, voiceFileID, "voice")
	if err != nil {
		return nil, err
	}

	s.archiveConversation(userID)
	s.restartAIResponse(userID, message)
	return message, nil
}

func (s *ChatService) GetChatHistory(ctx context.Context, userID int64, beforeID *int64) ([]models.Message, error) {
	return s.runtime.GetChatHistory(ctx, userID, beforeID)
}

func (s *ChatService) restartAIResponse(userID int64, userMessage *models.Message) {
	s.agentRuns.Submit(userID, queuedUserInput{
		MessageID: userMessage.ID,
		Text:      userMessage.TextContent,
	}, func(ctx context.Context, inputs []queuedUserInput) {
		s.processAIResponse(ctx, userID, inputs)
	})
}

func (s *ChatService) processAIResponse(ctx context.Context, userID int64, inputs []queuedUserInput) {
	contextBeforeMessageID, userInput := mergeQueuedUserInputs(inputs)
	if contextBeforeMessageID == 0 {
		return
	}

	if _, err := s.runAgentChat(ctx, userID, contextBeforeMessageID, userInput); err != nil {
		if ctx.Err() != nil {
			return
		}

		log.Printf("Agent chat failed for user %d: %v", userID, err)
		s.runtime.SendAgentError(userID)
		return
	}
	s.runtime.SendAgentTurnCompleted(userID)
}

func (s *ChatService) archiveConversation(userID int64) {
	if s.archiver == nil {
		return
	}

	go func() {
		lockValue, _ := s.archiveLocks.LoadOrStore(userID, &sync.Mutex{})
		userLock := lockValue.(*sync.Mutex)
		userLock.Lock()
		defer userLock.Unlock()

		if err := s.archiver.Run(context.Background(), userID); err != nil {
			log.Printf("conversation archive failed for user %d: %v", userID, err)
		}
	}()
}

func (s *ChatService) runAgentChat(ctx context.Context, userID, contextBeforeMessageID int64, userInput string) (*ChatResult, error) {
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
		UserID:                 userID,
		ContextBeforeMessageID: contextBeforeMessageID,
		UserInput:              userInput,
		Settings:               agentSettings,
		CurrentTime:            time.Now().In(loc).Format(time.RFC3339),
		Timezone:               timezone,
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
		EmbeddingDimension:  settings.EmbeddingDimension,
		NativeLanguage:      settings.NativeLanguage,
		TargetLanguage:      settings.TargetLanguage,
	}
}
