package archive

import (
	"context"
	"encoding/json"
	"fmt"
	"learnlang-api/agent/llm"
	"learnlang-api/agent/memory"
	"learnlang-api/models"
	"learnlang-api/services"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/tmc/langchaingo/llms"
)

const maxArchiveSummaryRunes = 120

type Service struct {
	archiveService *services.ConversationArchiveService
	settings       *services.UserSettingsService
	memoryStore    *memory.Store
}

func NewService(archiveService *services.ConversationArchiveService, settings *services.UserSettingsService, memoryStore *memory.Store) *Service {
	return &Service{
		archiveService: archiveService,
		settings:       settings,
		memoryStore:    memoryStore,
	}
}

type archiveResponse struct {
	Segments []archiveSegment `json:"segments"`
}

type archiveSegment struct {
	Summary    string  `json:"summary"`
	MessageIDs []int64 `json:"message_ids"`
}

func (s *Service) Run(ctx context.Context, userID int64) error {
	window, err := s.archiveService.GetArchiveWindow(ctx, userID)
	if err != nil {
		return err
	}
	if len(window.Candidates) == 0 {
		return nil
	}

	settings, err := s.settings.GetUserSettings(userID)
	if err != nil {
		return err
	}

	model, err := llm.New(settings.APIKey, settings.APIBaseURL, settings.Model, settings.LLMType)
	if err != nil {
		return err
	}

	output, err := llms.GenerateFromSinglePrompt(ctx, model, buildPrompt(window.Candidates, window.Reserved))
	if err != nil {
		return err
	}

	segments, err := parseSegments(output)

	if err != nil {
		return err
	}
	if len(segments) == 0 {
		return nil
	}

	inputs := make([]services.ArchiveSegmentInput, 0, len(segments))
	for _, segment := range segments {
		inputs = append(inputs, services.ArchiveSegmentInput{
			Summary:    strings.TrimSpace(segment.Summary),
			MessageIDs: segment.MessageIDs,
		})
	}

	archives, err := s.archiveService.SaveArchiveSegments(ctx, userID, window.Candidates, inputs)
	if err != nil {
		return err
	}

	s.storeArchiveEmbeddings(ctx, userID, settings, archives)
	return nil
}

func (s *Service) storeArchiveEmbeddings(ctx context.Context, userID int64, settings *models.UserSettings, archives []models.ConversationArchive) {
	if s.memoryStore == nil || len(archives) == 0 {
		return
	}

	for _, archive := range archives {
		embedding, err := createEmbedding(ctx, settings, archive.Summary)
		if err != nil {
			continue
		}

		embeddingID, err := s.memoryStore.InsertArchive(ctx, userID, archive.Summary, archive.MessageIDs, embedding)
		if err != nil {
			continue
		}

		_ = s.archiveService.UpdateEmbeddingID(ctx, archive.ID, embeddingID)
	}
}

func createEmbedding(ctx context.Context, settings *models.UserSettings, text string) ([]float32, error) {
	apiKey := strings.TrimSpace(settings.EmbeddingAPIKey)
	if apiKey == "" {
		apiKey = strings.TrimSpace(settings.APIKey)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("embedding api key is required")
	}

	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	apiBaseURL := strings.TrimSpace(settings.EmbeddingAPIBaseURL)
	if apiBaseURL == "" {
		apiBaseURL = strings.TrimSpace(settings.APIBaseURL)
	}
	if apiBaseURL != "" {
		opts = append(opts, option.WithBaseURL(apiBaseURL))
	}

	model := strings.TrimSpace(settings.EmbeddingModel)
	if model == "" {
		model = "text-embedding-3-small"
	}

	client := openai.NewClient(opts...)
	resp, err := client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model: openai.EmbeddingModel(model),
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: []string{text},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("embedding response is empty")
	}

	vector := make([]float32, 0, len(resp.Data[0].Embedding))
	for _, value := range resp.Data[0].Embedding {
		vector = append(vector, float32(value))
	}
	return vector, nil
}

func parseSegments(output string) ([]archiveSegment, error) {
	clean := strings.TrimSpace(output)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)

	var resp archiveResponse
	if err := json.Unmarshal([]byte(clean), &resp); err != nil {
		return nil, fmt.Errorf("parse archive response: %w: %s", err, output)
	}

	segments := make([]archiveSegment, 0, len(resp.Segments))
	for _, segment := range resp.Segments {
		segment.Summary = normalizeSummary(segment.Summary)
		if segment.Summary == "" || len(segment.MessageIDs) == 0 {
			continue
		}
		if utf8.RuneCountInString(segment.Summary) > maxArchiveSummaryRunes {
			return nil, fmt.Errorf("archive segment summary exceeds %d characters", maxArchiveSummaryRunes)
		}
		ids := append([]int64(nil), segment.MessageIDs...)
		sort.Slice(ids, func(i, j int) bool {
			return ids[i] < ids[j]
		})
		segment.MessageIDs = ids
		segments = append(segments, segment)
	}

	return segments, nil
}

func normalizeSummary(summary string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(summary)), " ")
}

func buildPrompt(candidates, reserved []models.Message) string {
	var b strings.Builder
	b.WriteString(`# Conversation Archive Agent

You archive completed parts of a chat into semantic summaries.

Rules:
- Read candidate messages in chronological order from oldest to newest.
- Split only completed conversation portions into summary segments.
- A message can belong to at most one segment.
- Each segment must contain contiguous message IDs from the candidate list.
- Keep segment order chronological.
- Do not include reserved latest messages in any segment.
- If the candidate messages appear to be one ongoing unfinished topic, return an empty segments array.
- Summaries are for embedding retrieval. Write one compact, standalone paragraph targeted at 30-90 characters. Never pad a short summary with invented detail, quote the dialogue, or repeat conversational filler.
- Use only facts stated in the messages. Make the topic and named entities explicit, then retain only details useful for future retrieval: user facts or preferences, goals, decisions, outcomes, unresolved work, and important constraints.
- Prefer this compact format, omitting empty parts: "Topic: ...; Fact/decision: ...; Keywords: ...". Keywords must contain concrete names, technologies, concepts, or task terms mentioned in the dialogue.
- Return JSON only, with no markdown.

Output schema:
{"segments":[{"summary":"semantic archive summary","message_ids":[1,2,3]}]}

Candidate messages that may be archived:
`)

	writeMessages(&b, candidates)

	b.WriteString("\nReserved latest messages for context only. Never archive these:\n")
	writeMessages(&b, reserved)

	return b.String()
}

func writeMessages(b *strings.Builder, messages []models.Message) {
	if len(messages) == 0 {
		b.WriteString("- none\n")
		return
	}

	for _, message := range messages {
		b.WriteString(fmt.Sprintf("- id=%d time=%s role=%s text=%q translation=%q\n",
			message.ID,
			message.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
			message.Role,
			message.TextContent,
			message.Translation,
		))
	}
}
