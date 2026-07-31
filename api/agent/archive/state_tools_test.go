package archive

import (
	"context"
	"encoding/json"
	"learnlang-api/models"
	"reflect"
	"strings"
	"testing"
)

func testArchiveMessages(ids ...int64) []models.Message {
	messages := make([]models.Message, len(ids))
	for index, id := range ids {
		messages[index] = models.Message{ID: id, TextContent: "message"}
	}
	return messages
}

func TestArchiveStateAddsContiguousRanges(t *testing.T) {
	reserved := testArchiveMessages(99)
	state := newArchiveState(testArchiveMessages(10, 20, 30), reserved)
	window := state.CurrentWindow()
	if len(window.Candidates) != 3 || window.Candidates[0].ID != 1 || window.Reserved[0].ID != 99 {
		t.Fatalf("initial window = %#v", window)
	}
	window.Reserved[0].ID = 0
	if state.reserved[0].ID != 99 {
		t.Fatal("CurrentWindow() exposed the reserved slice")
	}

	if err := state.AddRanges([]archiveRangeInput{{Summary: " first   topic ", StartID: 1, EndID: 2}}); err != nil {
		t.Fatalf("AddRanges() error = %v", err)
	}
	if got := state.ExpectedStartID(); got == nil || *got != 3 {
		t.Fatalf("ExpectedStartID() = %v", got)
	}
	want := []archiveSegment{{Summary: "first topic", MessageIDs: []int64{10, 20}}}
	if got := state.Result(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Result() = %#v, want %#v", got, want)
	}

	if err := state.AddRanges([]archiveRangeInput{{Summary: "last", StartID: 3, EndID: 3}}); err != nil {
		t.Fatal(err)
	}
	if state.ExpectedStartID() != nil || len(state.CurrentWindow().Candidates) != 0 {
		t.Fatal("completed state still expects candidates")
	}

	result := state.Result()
	result[0].MessageIDs[0] = -1
	if state.segments[0].MessageIDs[0] != 10 {
		t.Fatal("Result() exposed internal message IDs")
	}
}

func TestArchiveStateRejectsInvalidRangesWithoutMutation(t *testing.T) {
	tests := []struct {
		name   string
		ranges []archiveRangeInput
		want   string
	}{
		{"empty", nil, "at least one"},
		{"blank summary", []archiveRangeInput{{StartID: 1, EndID: 1}}, "summary is required"},
		{"out of bounds", []archiveRangeInput{{Summary: "x", StartID: 1, EndID: 3}}, "from 1 to 2"},
		{"skipped prefix", []archiveRangeInput{{Summary: "x", StartID: 2, EndID: 2}}, "next ID 1"},
		{"reversed", []archiveRangeInput{{Summary: "x", StartID: 1, EndID: 0}}, "from 1 to 2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := newArchiveState(testArchiveMessages(1, 2), nil)
			err := state.AddRanges(tt.ranges)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("AddRanges() error = %v, want containing %q", err, tt.want)
			}
			if len(state.Result()) != 0 || state.nextArchiveID != 1 {
				t.Fatal("invalid ranges mutated state")
			}
		})
	}
}

func TestArchiveConversationRangeTool(t *testing.T) {
	tool := archiveConversationRangeTool{state: newArchiveState(testArchiveMessages(5), nil)}
	output, err := tool.Call(context.Background(), `{"ranges":[{"summary":"memory","start_id":1,"end_id":1}]}`)
	if err != nil {
		t.Fatal(err)
	}
	var observation map[string]any
	if err := json.Unmarshal([]byte(output), &observation); err != nil {
		t.Fatal(err)
	}
	if observation["status"] != "accepted" || observation["count"] != float64(1) {
		t.Fatalf("observation = %#v", observation)
	}

	for name, candidate := range map[string]archiveConversationRangeTool{
		"invalid JSON": {state: newArchiveState(nil, nil)},
		"nil state":    {},
	} {
		t.Run(name, func(t *testing.T) {
			input := `{"ranges":[]}`
			if name == "invalid JSON" {
				input = `{`
			}
			output, err := candidate.Call(context.Background(), input)
			if err != nil || !strings.Contains(output, `"status":"rejected"`) {
				t.Fatalf("Call() output = %q, error = %v", output, err)
			}
		})
	}
}
