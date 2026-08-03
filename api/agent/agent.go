package agent

import (
	"context"
	"fmt"
	agentllm "learnlang-api/agent/llm"
	"learnlang-api/agent/memory"
	"learnlang-api/agent/prompts"
	agenttools "learnlang-api/agent/tools"
	"learnlang-api/aiusage"
	"learnlang-api/services"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

type Config struct {
	MemoryStore       *memory.Store
	Runtime           *services.ChatRuntimeService
	VocabularyService *services.VocabularyService
	UsageRecorder     aiusage.Recorder
}

type Service struct {
	memoryStore       *memory.Store
	runtime           *services.ChatRuntimeService
	vocabularyService *services.VocabularyService
	usageRecorder     aiusage.Recorder
}

const maxChatToolIterations = 12

func NewService(cfg Config) *Service {
	return &Service{
		memoryStore:       cfg.MemoryStore,
		runtime:           cfg.Runtime,
		vocabularyService: cfg.VocabularyService,
		usageRecorder:     cfg.UsageRecorder,
	}
}

func (s *Service) RunChat(ctx context.Context, req ChatRequest) (*ChatResult, error) {
	shortTermMessages, err := s.runtime.GetShortTermMemory(ctx, req.UserID, req.ContextBeforeMessageID, time.Now().Add(-24*time.Hour))
	if err != nil {
		return nil, err
	}
	profile, err := s.runtime.UserProfileSummary(req.UserID)
	if err != nil {
		return nil, err
	}
	llm, err := agentllm.New(ctx, req.Settings.APIKey, req.Settings.APIBaseURL, req.Settings.Model, req.Settings.LLMType)
	if err != nil {
		return nil, err
	}
	llm = aiusage.NewChatModel(llm, s.usageRecorder, req.UserID, "chat", req.Settings.Model)

	turnState := agenttools.NewTurnState()
	tools := []tool.BaseTool{
		agenttools.NewEinoTool(agenttools.UserProfileSummaryTool{UserID: req.UserID}, objectParams(map[string]*schema.ParameterInfo{
			"summary": {Type: schema.String, Desc: "complete updated user profile", Required: true},
		}), nil),
		agenttools.NewEinoTool(agenttools.LongTermMemorySearchTool{
			UserID:      req.UserID,
			Limit:       5,
			Timezone:    req.Timezone,
			Store:       s.memoryStore,
			APIKey:      req.Settings.EmbeddingAPIKey,
			APIBaseURL:  req.Settings.EmbeddingAPIBaseURL,
			Model:       req.Settings.EmbeddingModel,
			Dimension:   req.Settings.EmbeddingDimension,
			FallbackKey: req.Settings.APIKey,
			FallbackURL: req.Settings.APIBaseURL,
		}, objectParams(map[string]*schema.ParameterInfo{
			"query": {Type: schema.String, Desc: "standalone semantic retrieval query", Required: true},
		}), agenttools.JSONStringField("query")),
		agenttools.NewEinoTool(agenttools.ArchivedConversationKeywordSearchTool{
			UserID:   req.UserID,
			Timezone: req.Timezone,
		}, objectParams(map[string]*schema.ParameterInfo{
			"keyword":    {Type: schema.String, Desc: "exact phrase to search", Required: true},
			"start_time": {Type: schema.String, Desc: "optional RFC3339 start time"},
			"end_time":   {Type: schema.String, Desc: "optional RFC3339 end time"},
			"limit":      {Type: schema.Integer, Desc: "maximum archive count"},
		}), nil),
		agenttools.NewEinoTool(agenttools.SendChatReplyTool{
			UserID:  req.UserID,
			Runtime: s.runtime,
			State:   turnState,
		}, objectParams(map[string]*schema.ParameterInfo{
			"messages": {Type: schema.Array, Desc: "ordered reply sentences", Required: true, ElemInfo: &schema.ParameterInfo{Type: schema.Object, SubParams: map[string]*schema.ParameterInfo{
				"original":    {Type: schema.String, Desc: "target-language sentence", Required: true},
				"translation": {Type: schema.String, Desc: "native-language translation", Required: true},
			}}},
		}), nil),
		agenttools.NewEinoTool(agenttools.CompleteChatTurnTool{
			State: turnState,
		}, objectParams(map[string]*schema.ParameterInfo{
			"detected_language": {Type: schema.String, Desc: "detected language code", Required: true},
		}), nil),
		agenttools.NewEinoTool(agenttools.ScheduleMessageTool{
			UserID:   req.UserID,
			Timezone: req.Timezone,
			Runtime:  s.runtime,
		}, objectParams(map[string]*schema.ParameterInfo{
			"message":      {Type: schema.String, Desc: "target-language message", Required: true},
			"translation":  {Type: schema.String, Desc: "native-language translation", Required: true},
			"scheduled_at": {Type: schema.String, Desc: "local datetime", Required: true},
		}), nil),
		agenttools.NewEinoTool(agenttools.ScheduledTaskQueryTool{
			UserID:   req.UserID,
			Timezone: req.Timezone,
			Lister:   s.runtime,
		}, objectParams(map[string]*schema.ParameterInfo{
			"status":    {Type: schema.String, Desc: "unfinished, completed, or all"},
			"page":      {Type: schema.Integer, Desc: "page number starting at 1"},
			"page_size": {Type: schema.Integer, Desc: "items per page, maximum 50"},
		}), nil),
		agenttools.NewEinoTool(agenttools.RandomNewVocabularyWordTool{
			UserID:           req.UserID,
			RequestMessageID: req.ContextBeforeMessageID,
			TargetLanguage:   req.Settings.TargetLanguage,
			NativeLanguage:   req.Settings.NativeLanguage,
			Vocabulary:       s.vocabularyService,
			State:            turnState,
		}, objectParams(map[string]*schema.ParameterInfo{
			"count": {Type: schema.Integer, Desc: "requested word count"},
		}), nil),
		agenttools.NewEinoTool(agenttools.RandomOldVocabularyWordTool{
			UserID:           req.UserID,
			RequestMessageID: req.ContextBeforeMessageID,
			TargetLanguage:   req.Settings.TargetLanguage,
			NativeLanguage:   req.Settings.NativeLanguage,
			Vocabulary:       s.vocabularyService,
			State:            turnState,
		}, objectParams(map[string]*schema.ParameterInfo{
			"count": {Type: schema.Integer, Desc: "requested word count"},
		}), nil),
	}

	systemPrompt := prompts.ChatSystemPrompt(
		req.Settings.NativeLanguage,
		req.Settings.TargetLanguage,
		req.CurrentTime,
		req.Timezone,
		shortTermMessages,
		profile.Summary,
	)

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: llm,
		ToolsConfig:      compose.ToolsNodeConfig{Tools: tools},
		// Eino counts model and tool nodes separately. This permits the same
		// twelve tool-call iterations as the previous executor, plus its final
		// model response after the last tool result.
		MaxStep: maxChatToolIterations*2 + 1,
	})
	if err != nil {
		return nil, err
	}
	if _, err := agent.Generate(ctx, []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(req.UserInput),
	}); err != nil {
		return nil, err
	}
	if !turnState.IsCompleted() {
		return nil, fmt.Errorf("agent did not complete the chat turn")
	}

	return turnState.Result(), nil
}

func objectParams(params map[string]*schema.ParameterInfo) *schema.ParamsOneOf {
	return schema.NewParamsOneOfByParams(params)
}
