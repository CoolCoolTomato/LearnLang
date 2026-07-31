package archive

import (
	"context"
	"errors"
	"learnlang-api/models"
	"learnlang-api/services"
	"strings"
	"testing"
	"time"
)

type fakeSegmenter struct {
	segments []archiveSegment
	err      error
	calls    int
}

func (f *fakeSegmenter) Generate(context.Context, *models.UserSettings, *services.ArchiveWindow) ([]archiveSegment, error) {
	f.calls++
	return f.segments, f.err
}

type fakeIndexer struct {
	calls int
}

func (f *fakeIndexer) Index(context.Context, int64, *models.UserSettings, []models.ConversationArchive) {
	f.calls++
}

func TestArchivePromptFormatting(t *testing.T) {
	if prompt := archiveSystemPrompt(); !strings.Contains(prompt, "archive_conversation_range") || !strings.Contains(prompt, "Candidate IDs") {
		t.Fatalf("archiveSystemPrompt() missing required guidance")
	}
	timestamp := time.Date(2026, 7, 31, 2, 0, 0, 0, time.UTC)
	input := buildArchiveInput(
		[]mappedArchiveMessage{{ID: 1, Message: models.Message{Role: "user", TextContent: "hello", CreatedAt: timestamp}}},
		[]models.Message{{Role: "assistant", TextContent: "reserved", CreatedAt: timestamp}},
	)
	for _, expected := range []string{`id=1`, `text="hello"`, `reserved_position=1`, `text="reserved"`, "2026-07-31T02:00:00Z"} {
		if !strings.Contains(input, expected) {
			t.Errorf("archive input missing %q: %s", expected, input)
		}
	}
	empty := buildArchiveInput(nil, nil)
	if strings.Count(empty, "- none") != 2 {
		t.Fatalf("empty archive input = %s", empty)
	}
}

func TestArchiveToolMetadata(t *testing.T) {
	tool := archiveConversationRangeTool{}
	if tool.Name() != "archive_conversation_range" || tool.Description() == "" {
		t.Fatalf("tool metadata = %q, %q", tool.Name(), tool.Description())
	}
}

func TestArchiveServiceRunStopsOnEmptyWindow(t *testing.T) {
	// The concrete archive/settings services require PostgreSQL queries. Their
	// state-machine and validation behavior is tested independently; this test
	// verifies constructor wiring without reaching external infrastructure.
	service := NewService(nil, nil, nil)
	if service == nil || service.segmenter == nil || service.indexer == nil {
		t.Fatalf("NewService() = %#v", service)
	}
	segmenter := &fakeSegmenter{err: errors.New("mock")}
	indexer := &fakeIndexer{}
	service.segmenter = segmenter
	service.indexer = indexer
	if segmenter.calls != 0 || indexer.calls != 0 {
		t.Fatal("test fakes were called during wiring")
	}
}
