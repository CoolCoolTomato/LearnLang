package archive

import (
	"context"
	"encoding/json"
)

type archiveConversationRangeTool struct {
	state *archiveState
}

func (archiveConversationRangeTool) Name() string {
	return "archive_conversation_range"
}

func (archiveConversationRangeTool) Description() string {
	return `Archive one or more retrieval topics from a contiguous prefix of the current candidate messages. summary is embedded for long-term-memory retrieval: write a standalone semantic passage containing the user's need, exact entities and terms, and the useful answer, decision, outcome, constraint, preference, or unresolved state. Submit every archiveable topic in one call as chronological ranges, but choose the stopping boundary so messages remain for the next batch. Never include reserved messages. Input must be JSON: {"ranges":[{"summary":"embedding-oriented semantic memory","start_id":1,"end_id":3},{"summary":"another retrieval memory","start_id":4,"end_id":8}]}. Ranges are inclusive and must form a contiguous prefix starting at candidate ID 1; do not skip an initial message.`
}

func (t archiveConversationRangeTool) Call(_ context.Context, input string) (string, error) {
	var args struct {
		Ranges []archiveRangeInput `json:"ranges"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return archiveToolObservation(map[string]any{
			"status": "rejected",
			"error":  "tool input must be valid JSON",
		})
	}
	if t.state == nil {
		return archiveToolObservation(map[string]any{
			"status": "rejected",
			"error":  "archive state is not configured",
		})
	}
	if err := t.state.AddRanges(args.Ranges); err != nil {
		return archiveToolObservation(map[string]any{
			"status":            "rejected",
			"error":             err.Error(),
			"expected_start_id": t.state.ExpectedStartID(),
		})
	}
	return archiveToolObservation(map[string]any{
		"status":            "accepted",
		"count":             len(args.Ranges),
		"expected_start_id": t.state.ExpectedStartID(),
	})
}

func archiveToolObservation(value map[string]any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
