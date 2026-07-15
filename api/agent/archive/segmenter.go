package archive

import (
	"context"
	"fmt"
	"learnlang-api/agent/llm"
	"learnlang-api/models"
	"learnlang-api/services"
	"strings"

	lcagents "github.com/tmc/langchaingo/agents"
	lctools "github.com/tmc/langchaingo/tools"
)

const archiveBatchPlanAttempts = 2

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

	state := newArchiveState(window.Candidates, window.Reserved)
	rangeTool := archiveConversationRangeTool{state: state}
	tools := []lctools.Tool{rangeTool}
	promptWindow := state.CurrentWindow()
	input := buildArchiveInput(promptWindow.Candidates, promptWindow.Reserved)
	var lastObservation string

	for attempt := 0; attempt < archiveBatchPlanAttempts; attempt++ {
		agent := lcagents.NewOpenAIFunctionsAgent(
			model,
			tools,
			lcagents.NewOpenAIOption().WithSystemMessage(archiveSystemPrompt()),
		)

		planInput := input
		if lastObservation != "" {
			planInput += "\nThe previous tool call was rejected. Correct the input using this observation:\n" + lastObservation
		}
		actions, _, err := agent.Plan(ctx, nil, map[string]string{"input": planInput})
		if err != nil {
			return nil, err
		}
		if len(actions) == 0 {
			return nil, fmt.Errorf("archive agent did not call archive_conversation_range")
		}

		for _, action := range actions {
			if !strings.EqualFold(action.Tool, rangeTool.Name()) {
				continue
			}
			observation, err := rangeTool.Call(ctx, action.ToolInput)
			if err != nil {
				return nil, err
			}
			lastObservation = observation
		}

		segments := state.Result()
		if len(segments) > 0 {
			return segments, nil
		}
	}

	return nil, fmt.Errorf("archive agent made no progress after %d attempts: %s", archiveBatchPlanAttempts, lastObservation)
}
