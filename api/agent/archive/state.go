package archive

import (
	"fmt"
	"learnlang-api/models"
	"strings"
)

type archiveState struct {
	mapping       *archiveMessageMapping
	reserved      []models.Message
	segments      []archiveSegment
	nextArchiveID int
}

type archiveMessageMapping struct {
	byArchiveID map[int]models.Message
	count       int
}

type archiveRangeInput struct {
	Summary string `json:"summary"`
	StartID int    `json:"start_id"`
	EndID   int    `json:"end_id"`
}

type archivePromptWindow struct {
	Candidates []mappedArchiveMessage
	Reserved   []models.Message
}

type mappedArchiveMessage struct {
	ID      int
	Message models.Message
}

func newArchiveMessageMapping(messages []models.Message) *archiveMessageMapping {
	byArchiveID := make(map[int]models.Message, len(messages))
	for index, message := range messages {
		byArchiveID[index+1] = message
	}
	return &archiveMessageMapping{byArchiveID: byArchiveID, count: len(messages)}
}

func (m *archiveMessageMapping) messageCount() int {
	return m.count
}

func (m *archiveMessageMapping) messageIDs(startID, endID int) []int64 {
	messageIDs := make([]int64, 0, endID-startID+1)
	for id := startID; id <= endID; id++ {
		messageIDs = append(messageIDs, m.byArchiveID[id].ID)
	}
	return messageIDs
}

func newArchiveState(candidates, reserved []models.Message) *archiveState {
	return &archiveState{
		mapping:       newArchiveMessageMapping(candidates),
		reserved:      append([]models.Message(nil), reserved...),
		segments:      []archiveSegment{},
		nextArchiveID: 1,
	}
}

func (s *archiveState) CurrentWindow() archivePromptWindow {
	candidates := make([]mappedArchiveMessage, 0, s.mapping.messageCount()-s.nextArchiveID+1)
	for id := s.nextArchiveID; id <= s.mapping.messageCount(); id++ {
		candidates = append(candidates, mappedArchiveMessage{ID: id, Message: s.mapping.byArchiveID[id]})
	}
	return archivePromptWindow{
		Candidates: candidates,
		Reserved:   append([]models.Message(nil), s.reserved...),
	}
}

func (s *archiveState) AddRanges(ranges []archiveRangeInput) error {
	if len(ranges) == 0 {
		return fmt.Errorf("ranges must contain at least one segment")
	}

	nextID := s.nextArchiveID
	newSegments := make([]archiveSegment, 0, len(ranges))
	for _, archiveRange := range ranges {
		summary := normalizeSummary(archiveRange.Summary)
		if summary == "" {
			return fmt.Errorf("archive segment summary is required")
		}
		if archiveRange.StartID < 1 || archiveRange.EndID < 1 || archiveRange.StartID > s.mapping.messageCount() || archiveRange.EndID > s.mapping.messageCount() {
			return fmt.Errorf("range must use candidate IDs from 1 to %d", s.mapping.messageCount())
		}
		if archiveRange.StartID != nextID {
			return fmt.Errorf("range must start at the next ID %d; archive ranges must cover the candidate prefix without skipping messages", nextID)
		}
		if archiveRange.EndID < archiveRange.StartID {
			return fmt.Errorf("end ID %d precedes start ID %d", archiveRange.EndID, archiveRange.StartID)
		}

		newSegments = append(newSegments, archiveSegment{
			Summary:    summary,
			MessageIDs: s.mapping.messageIDs(archiveRange.StartID, archiveRange.EndID),
		})
		nextID = archiveRange.EndID + 1
	}

	s.segments = append(s.segments, newSegments...)
	s.nextArchiveID = nextID
	return nil
}

func (s *archiveState) ExpectedStartID() *int {
	if s.nextArchiveID > s.mapping.messageCount() {
		return nil
	}
	id := s.nextArchiveID
	return &id
}

func (s *archiveState) Result() []archiveSegment {
	segments := make([]archiveSegment, len(s.segments))
	for index, segment := range s.segments {
		segments[index] = archiveSegment{
			Summary:    segment.Summary,
			MessageIDs: append([]int64(nil), segment.MessageIDs...),
		}
	}
	return segments
}

func normalizeSummary(summary string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(summary)), " ")
}
