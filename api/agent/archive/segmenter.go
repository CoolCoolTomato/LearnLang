package archive

import (
	"context"
	"fmt"
	"learnlang-api/agent/llm"
	"learnlang-api/models"
	"learnlang-api/services"
	"strings"

	"github.com/cloudwego/eino/schema"
	agenttools "learnlang-api/agent/tools"
)

const archiveBatchPlanAttempts = 2

type archiveSegment struct {
	Summary    string
	MessageIDs []int64
}

type llmSegmenter struct{}

func (llmSegmenter) Generate(ctx context.Context, settings *models.UserSettings, window *services.ArchiveWindow) ([]archiveSegment, error) {
	chatModel, err := llm.New(ctx, settings.APIKey, settings.APIBaseURL, settings.Model, settings.LLMType)
	if err != nil {
		return nil, err
	}

	state := newArchiveState(window.Candidates, window.Reserved)
	rangeTool := archiveConversationRangeTool{state: state}
	einoTool := agenttools.NewEinoTool(rangeTool, schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"ranges": {Type: schema.Array, Desc: "chronological archive ranges", Required: true, ElemInfo: &schema.ParameterInfo{Type: schema.Object, SubParams: map[string]*schema.ParameterInfo{
			"summary":  {Type: schema.String, Desc: "retrieval-oriented archive summary", Required: true},
			"start_id": {Type: schema.Integer, Desc: "first candidate ID", Required: true},
			"end_id":   {Type: schema.Integer, Desc: "last candidate ID", Required: true},
		}}},
	}), nil)
	toolInfo, err := einoTool.Info(ctx)
	if err != nil {
		return nil, err
	}
	modelWithTools, err := chatModel.WithTools([]*schema.ToolInfo{toolInfo})
	if err != nil {
		return nil, err
	}
	promptWindow := state.CurrentWindow()
	input := buildArchiveInput(promptWindow.Candidates, promptWindow.Reserved)
	var lastObservation string

	for attempt := 0; attempt < archiveBatchPlanAttempts; attempt++ {
		planInput := input
		if lastObservation != "" {
			planInput += "\nThe previous attempt did not produce an accepted archive tool call. Correct it using this observation:\n" + lastObservation
		}
		response, err := modelWithTools.Generate(ctx, []*schema.Message{
			schema.SystemMessage(archiveSystemPrompt()),
			schema.UserMessage(planInput),
		})
		if err != nil {
			return nil, err
		}
		if response == nil || len(response.ToolCalls) == 0 {
			lastObservation = `{"status":"rejected","error":"call archive_conversation_range instead of returning plain text"}`
			continue
		}

		matchedTool := false
		for _, action := range response.ToolCalls {
			if !strings.EqualFold(action.Function.Name, rangeTool.Name()) {
				continue
			}
			matchedTool = true
			observation, err := einoTool.InvokableRun(ctx, action.Function.Arguments)
			if err != nil {
				return nil, err
			}
			lastObservation = observation
		}
		if !matchedTool {
			lastObservation = `{"status":"rejected","error":"only archive_conversation_range is available"}`
			continue
		}

		segments := state.Result()
		if len(segments) > 0 {
			return segments, nil
		}
	}

	return nil, fmt.Errorf("archive agent made no progress after %d attempts: %s", archiveBatchPlanAttempts, lastObservation)
}
