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
	return `Archive one completed, contiguous conversation range. Call in chronological order and never include reserved messages. Input must be a JSON string: {"summary":"compact semantic summary","start_message_id":1,"end_message_id":3}. The range is inclusive.`
}

func (t archiveConversationRangeTool) Call(_ context.Context, input string) (string, error) {
	var args struct {
		Summary        string `json:"summary"`
		StartMessageID int64  `json:"start_message_id"`
		EndMessageID   int64  `json:"end_message_id"`
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
	if err := t.state.AddRange(args.Summary, args.StartMessageID, args.EndMessageID); err != nil {
		return archiveToolObservation(map[string]any{
			"status":                    "rejected",
			"error":                     err.Error(),
			"expected_start_message_id": t.state.ExpectedStartMessageID(),
		})
	}
	return archiveToolObservation(map[string]any{
		"status":                    "accepted",
		"start_message_id":          args.StartMessageID,
		"end_message_id":            args.EndMessageID,
		"expected_start_message_id": t.state.ExpectedStartMessageID(),
	})
}

type completeConversationArchiveTool struct {
	state *archiveState
}

func (completeConversationArchiveTool) Name() string {
	return "complete_conversation_archive"
}

func (completeConversationArchiveTool) Description() string {
	return `Finish the archive task after recording every completed conversation range. Call exactly once, including when no range can be archived. Input must be the JSON string {}.`
}

func (t completeConversationArchiveTool) Call(_ context.Context, _ string) (string, error) {
	if t.state == nil {
		return archiveToolObservation(map[string]any{
			"status": "rejected",
			"error":  "archive state is not configured",
		})
	}
	if err := t.state.Complete(); err != nil {
		return archiveToolObservation(map[string]any{
			"status": "rejected",
			"error":  err.Error(),
		})
	}
	return archiveToolObservation(map[string]any{
		"status": "completed",
	})
}

func archiveToolObservation(value map[string]any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
