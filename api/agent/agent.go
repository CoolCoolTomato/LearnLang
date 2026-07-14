package agent

import (
	"context"
	agentllm "learnlang-api/agent/llm"
	"learnlang-api/agent/memory"
	"learnlang-api/agent/prompts"
	agenttools "learnlang-api/agent/tools"
	"learnlang-api/services"
	"time"

	lcagents "github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	lctools "github.com/tmc/langchaingo/tools"
)

type Config struct {
	MemoryStore *memory.Store
	Runtime     *services.ChatRuntimeService
}

type Service struct {
	memoryStore *memory.Store
	runtime     *services.ChatRuntimeService
}

func NewService(cfg Config) *Service {
	return &Service{
		memoryStore: cfg.MemoryStore,
		runtime:     cfg.Runtime,
	}
}

func (s *Service) RunChat(ctx context.Context, req ChatRequest) (*ChatResult, error) {
	shortTermMessages, err := s.runtime.GetShortTermMemory(ctx, req.UserID, req.CurrentMessageID, time.Now().Add(-24*time.Hour))
	if err != nil {
		return nil, err
	}
	profile, err := s.runtime.UserProfileSummary(req.UserID)
	if err != nil {
		return nil, err
	}

	llm, err := agentllm.New(req.Settings.APIKey, req.Settings.APIBaseURL, req.Settings.Model, req.Settings.LLMType)
	if err != nil {
		return nil, err
	}

	turnState := agenttools.NewTurnState()
	tools := []lctools.Tool{
		agenttools.UserProfileSummaryTool{UserID: req.UserID},
		agenttools.LongTermMemorySearchTool{
			UserID:      req.UserID,
			Limit:       5,
			Timezone:    req.Timezone,
			Store:       s.memoryStore,
			APIKey:      req.Settings.EmbeddingAPIKey,
			APIBaseURL:  req.Settings.EmbeddingAPIBaseURL,
			Model:       req.Settings.EmbeddingModel,
			FallbackKey: req.Settings.APIKey,
			FallbackURL: req.Settings.APIBaseURL,
		},
		agenttools.ArchivedConversationKeywordSearchTool{
			UserID:   req.UserID,
			Timezone: req.Timezone,
		},
		agenttools.SendChatReplyTool{
			UserID:  req.UserID,
			Runtime: s.runtime,
			State:   turnState,
		},
		agenttools.CompleteChatTurnTool{
			State: turnState,
		},
		agenttools.ScheduleMessageTool{
			UserID:  req.UserID,
			Runtime: s.runtime,
		},
	}

	systemPrompt := prompts.ChatSystemPrompt(
		req.Settings.NativeLanguage,
		req.Settings.TargetLanguage,
		req.CurrentTime,
		req.Timezone,
		req.Instant,
		shortTermMessages,
		profile.Summary,
	)

	agent := lcagents.NewOpenAIFunctionsAgent(
		llm,
		tools,
		lcagents.NewOpenAIOption().WithSystemMessage(systemPrompt),
		lcagents.WithMaxIterations(12),
	)
	executor := lcagents.NewExecutor(agent, lcagents.WithMaxIterations(12))

	if _, err := chains.Run(ctx, executor, req.UserInput); err != nil {
		return nil, err
	}

	return turnState.Result(), nil
}
