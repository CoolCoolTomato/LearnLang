package archive

import (
	"context"
	"fmt"
	"learnlang-api/agent/llm"
	"learnlang-api/models"
	"learnlang-api/services"

	lcagents "github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	lctools "github.com/tmc/langchaingo/tools"
)

const archiveAgentMaxIterations = 24

type archiveSegment struct {
	Summary    string
	MessageIDs []int64
}

type llmSegmenter struct{}

func (llmSegmenter) Generate(ctx context.Context, settings *models.UserSettings, window *services.ArchiveWindow) ([]archiveSegment, error) {
	model, err := llm.New(settings.APIKey, settings.APIBaseURL, settings.Model, settings.LLMType)
	if err != nil {
		return nil, err
	}

	state := newArchiveState(window.Candidates)
	tools := []lctools.Tool{
		archiveConversationRangeTool{state: state},
		completeConversationArchiveTool{state: state},
	}
	agent := lcagents.NewOpenAIFunctionsAgent(
		model,
		tools,
		lcagents.NewOpenAIOption().WithSystemMessage(archiveSystemPrompt()),
		lcagents.WithMaxIterations(archiveAgentMaxIterations),
	)
	executor := lcagents.NewExecutor(agent, lcagents.WithMaxIterations(archiveAgentMaxIterations))

	if _, err := chains.Run(ctx, executor, buildArchiveInput(window.Candidates, window.Reserved)); err != nil {
		return nil, err
	}

	segments, completed := state.Result()
	if !completed {
		return nil, fmt.Errorf("archive agent finished without calling complete_conversation_archive")
	}
	return segments, nil
}
