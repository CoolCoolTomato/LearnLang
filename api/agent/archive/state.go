package archive

import (
	"fmt"
	"learnlang-api/models"
	"strings"
)

type archiveState struct {
	candidates         []models.Message
	candidateIndexByID map[int64]int
	segments           []archiveSegment
	nextCandidateIndex int
	completed          bool
}

func newArchiveState(candidates []models.Message) *archiveState {
	indexByID := make(map[int64]int, len(candidates))
	for index, message := range candidates {
		indexByID[message.ID] = index
	}
	return &archiveState{
		candidates:         candidates,
		candidateIndexByID: indexByID,
		segments:           []archiveSegment{},
	}
}

func (s *archiveState) AddRange(summary string, startMessageID, endMessageID int64) error {
	if s.completed {
		return fmt.Errorf("conversation archive is already completed")
	}
	summary = normalizeSummary(summary)
	if summary == "" {
		return fmt.Errorf("summary is required")
	}
	if s.nextCandidateIndex >= len(s.candidates) {
		return fmt.Errorf("all candidate messages are already archived")
	}

	startIndex, ok := s.candidateIndexByID[startMessageID]
	if !ok {
		return fmt.Errorf("start message %d is not in the candidate window", startMessageID)
	}
	endIndex, ok := s.candidateIndexByID[endMessageID]
	if !ok {
		return fmt.Errorf("end message %d is not in the candidate window", endMessageID)
	}
	if startIndex != s.nextCandidateIndex {
		return fmt.Errorf("range must start at the next candidate message %d", s.candidates[s.nextCandidateIndex].ID)
	}
	if endIndex < startIndex {
		return fmt.Errorf("end message %d precedes start message %d", endMessageID, startMessageID)
	}

	messageIDs := make([]int64, 0, endIndex-startIndex+1)
	for _, message := range s.candidates[startIndex : endIndex+1] {
		messageIDs = append(messageIDs, message.ID)
	}
	s.segments = append(s.segments, archiveSegment{
		Summary:    summary,
		MessageIDs: messageIDs,
	})
	s.nextCandidateIndex = endIndex + 1
	return nil
}

func (s *archiveState) Complete() error {
	if s.completed {
		return fmt.Errorf("conversation archive is already completed")
	}
	s.completed = true
	return nil
}

func (s *archiveState) ExpectedStartMessageID() *int64 {
	if s.nextCandidateIndex >= len(s.candidates) {
		return nil
	}
	messageID := s.candidates[s.nextCandidateIndex].ID
	return &messageID
}

func (s *archiveState) Result() ([]archiveSegment, bool) {
	segments := make([]archiveSegment, len(s.segments))
	for index, segment := range s.segments {
		segments[index] = archiveSegment{
			Summary:    segment.Summary,
			MessageIDs: append([]int64(nil), segment.MessageIDs...),
		}
	}
	return segments, s.completed
}

func normalizeSummary(summary string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(summary)), " ")
}
