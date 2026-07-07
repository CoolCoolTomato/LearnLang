package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"learnlang-api/agent/memory"
	"learnlang-api/agent/prompts"
	agenttools "learnlang-api/agent/tools"
	"strings"

	lcagents "github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	lctools "github.com/tmc/langchaingo/tools"
)

type Config struct {
	MemoryStore *memory.Store
}

type Service struct {
	memoryStore *memory.Store
}

func NewService(cfg Config) *Service {
	return &Service{
		memoryStore: cfg.MemoryStore,
	}
}

func (s *Service) RunChat(ctx context.Context, req ChatRequest) (*ChatResult, error) {
	llm, err := s.newLLM(req.Settings)
	if err != nil {
		return nil, err
	}

	tools := []lctools.Tool{
		agenttools.UserProfileSummaryTool{UserID: req.UserID},
		agenttools.RecentConversationTool{UserID: req.UserID, Limit: 100, Timezone: req.Timezone},
		agenttools.LongTermMemorySearchTool{
			UserID:      req.UserID,
			Limit:       3,
			Timezone:    req.Timezone,
			Store:       s.memoryStore,
			APIKey:      req.Settings.EmbeddingAPIKey,
			APIBaseURL:  req.Settings.EmbeddingAPIBaseURL,
			Model:       req.Settings.EmbeddingModel,
			FallbackKey: req.Settings.APIKey,
			FallbackURL: req.Settings.APIBaseURL,
		},
	}

	systemPrompt := prompts.ChatSystemPrompt(
		req.Settings.NativeLanguage,
		req.Settings.TargetLanguage,
		req.CurrentTime,
		req.Timezone,
		req.Instant,
	)

	agent := lcagents.NewOpenAIFunctionsAgent(
		llm,
		tools,
		lcagents.NewOpenAIOption().WithSystemMessage(systemPrompt),
		lcagents.WithMaxIterations(5),
	)
	executor := lcagents.NewExecutor(agent, lcagents.WithMaxIterations(5))

	output, err := chains.Run(ctx, executor, req.UserInput)
	if err != nil {
		return nil, err
	}

	result, err := parseChatResult(output)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) newLLM(settings UserSettings) (*openai.LLM, error) {
	apiKey := settings.APIKey
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("agent api key is required")
	}

	model := settings.Model
	if strings.TrimSpace(model) == "" {
		model = "gpt-4o-mini"
	}

	opts := []openai.Option{
		openai.WithToken(apiKey),
		openai.WithModel(model),
	}
	if strings.TrimSpace(settings.APIBaseURL) != "" {
		opts = append(opts, openai.WithBaseURL(settings.APIBaseURL))
	}

	return openai.New(opts...)
}

func parseChatResult(output string) (*ChatResult, error) {
	clean := strings.TrimSpace(output)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)

	var result ChatResult
	if err := json.Unmarshal([]byte(clean), &result); err != nil {
		return nil, fmt.Errorf("parse agent chat result: %w: %s", err, output)
	}

	return &result, nil
}
