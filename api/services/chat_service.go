package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"learnlang-api/database"
	"learnlang-api/models"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/shared"
	"github.com/pgvector/pgvector-go"
)

type ChatService struct {
	messageService             *MessageService
	conversationSummaryService *ConversationSummaryService
	userMemoryService          *UserMemoryService
	userSettingsService        *UserSettingsService
	scheduledTaskService       *ScheduledTaskService
	voiceFileService           *VoiceFileService
	wsHub                      WSHub
}

type WSHub interface {
	SendToUser(userID int64, message []byte)
}

type ChatEngineResult struct {
	ReplySentences   []Sentence    `json:"reply_sentences"`
	DetectedLanguage string        `json:"detected_language"`
	Memory           *MemoryInfo   `json:"memory"`
	Summary          *SummaryInfo  `json:"summary"`
	Function         *FunctionInfo `json:"function"`
	WaitForNextMsg   bool          `json:"wait_for_next_message"`
}

type Sentence struct {
	Original    string `json:"original"`
	Translation string `json:"translation"`
}

type MemoryInfo struct {
	ShouldStore     bool    `json:"should_store"`
	SemanticContent string  `json:"semantic_content"`
	Importance      float64 `json:"importance"`
	MemoryType      string  `json:"memory_type"`
	Language        string  `json:"language"`
}

type FunctionInfo struct {
	CallFunction bool                   `json:"call_function"`
	FunctionName string                 `json:"function_name"`
	FunctionArgs map[string]interface{} `json:"function_args"`
}

type SummaryInfo struct {
	ShouldUpdate bool   `json:"should_update"`
	Content      string `json:"content"`
}

func NewChatService(
	messageService *MessageService,
	conversationSummaryService *ConversationSummaryService,
	userMemoryService *UserMemoryService,
	userSettingsService *UserSettingsService,
	scheduledTaskService *ScheduledTaskService,
	voiceFileService *VoiceFileService,
	wsHub WSHub,
) *ChatService {
	return &ChatService{
		messageService:             messageService,
		conversationSummaryService: conversationSummaryService,
		userMemoryService:          userMemoryService,
		userSettingsService:        userSettingsService,
		scheduledTaskService:       scheduledTaskService,
		voiceFileService:           voiceFileService,
		wsHub:                      wsHub,
	}
}

type ChatRequest struct {
	UserID  int64  `json:"user_id" binding:"required"`
	Message string `json:"message" binding:"required"`
}

type ChatResponse struct {
	MessageID int64  `json:"message_id"`
	Message   string `json:"message"`
	Status    string `json:"status"`
}

type memoryWithDistance struct {
	models.UserMemory
	Distance float64 `gorm:"column:distance"`
}

func (cs *ChatService) TranscribeAudio(ctx context.Context, userID int64, audioFile io.Reader) (string, *int64, error) {
	uploadDir := "uploads/voices"
	os.MkdirAll(uploadDir, 0755)

	filename := fmt.Sprintf("%d_%d_%d.mp3", userID, time.Now().Unix(), rand.Intn(10000))
	filepath := filepath.Join(uploadDir, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return "", nil, err
	}

	fileSize, err := io.Copy(file, audioFile)
	if err != nil {
		file.Close()
		return "", nil, err
	}

	if err := file.Sync(); err != nil {
		file.Close()
		return "", nil, err
	}
	if err := file.Close(); err != nil {
		return "", nil, err
	}

	file, err = os.Open(filepath)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	settings, err := cs.userSettingsService.GetUserSettings(userID)
	if err != nil {
		return "", nil, err
	}

	apiKey := settings.STTAPIKey
	apiBaseURL := settings.STTAPIBaseURL

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(apiBaseURL),
	)

	model := settings.STTModel
	if model == "" {
		model = "whisper-1"
	}

	transcription, err := client.Audio.Transcriptions.New(ctx, openai.AudioTranscriptionNewParams{
		Model: openai.AudioModel(model),
		File:  file,
	})
	if err != nil {
		return "", nil, err
	}
	duration, err := detectMP3DurationSeconds(filepath)
	if err != nil {
		log.Printf("failed to detect uploaded voice duration: %v", err)
		duration = 0
	}

	voiceFile := &models.VoiceFile{
		UserID:    userID,
		VoiceRole: "user",
		VoiceURL:  filepath,
		Duration:  duration,
		FileSize:  fileSize,
	}
	if err := cs.voiceFileService.CreateVoiceFile(voiceFile); err != nil {
		return "", nil, err
	}

	return transcription.Text, &voiceFile.ID, nil
}

func (cs *ChatService) TextToSpeech(ctx context.Context, userID int64, text string, settings *models.UserSettings) (*int64, error) {
	apiKey := settings.STTAPIKey
	apiBaseURL := settings.STTAPIBaseURL

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(apiBaseURL),
	)

	model := settings.TTSModel
	if model == "" {
		model = "tts-1"
	}

	voice := settings.TTSVoice
	if voice == "" {
		voice = "alloy"
	}

	res, err := client.Audio.Speech.New(ctx, openai.AudioSpeechNewParams{
		Model:          openai.SpeechModel(model),
		Input:          text,
		Voice:          openai.AudioSpeechNewParamsVoice(voice),
		ResponseFormat: openai.AudioSpeechNewParamsResponseFormatMP3,
	})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	uploadDir := "uploads/voices"
	os.MkdirAll(uploadDir, 0755)

	filename := fmt.Sprintf("%d_%d_%d.mp3", userID, time.Now().Unix(), rand.Intn(10000))
	filepath := filepath.Join(uploadDir, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return nil, err
	}

	fileSize, err := io.Copy(file, res.Body)
	if err != nil {
		file.Close()
		return nil, err
	}

	if err := file.Sync(); err != nil {
		file.Close()
		return nil, err
	}
	if err := file.Close(); err != nil {
		return nil, err
	}

	duration, err := detectMP3DurationSeconds(filepath)
	if err != nil {
		log.Printf("failed to detect generated voice duration: %v", err)
		duration = 0
	}

	voiceFile := &models.VoiceFile{
		UserID:    userID,
		VoiceRole: "assistant",
		VoiceURL:  filepath,
		Duration:  duration,
		FileSize:  fileSize,
	}
	if err := cs.voiceFileService.CreateVoiceFile(voiceFile); err != nil {
		return nil, err
	}

	return &voiceFile.ID, nil
}

func (cs *ChatService) Chat(ctx context.Context, userID int64, userMessage string) (*models.Message, error) {
	userMsg, err := cs.messageService.CreateMessage(userID, "user", userMessage, "", nil, "text", 0)
	if err != nil {
		return nil, err
	}

	if err := database.DB.Preload("VoiceFile").First(userMsg, userMsg.ID).Error; err != nil {
		return nil, err
	}

	cs.scheduledTaskService.CancelUserPendingTasks(userID, "wait_message")

	go cs.processAIResponse(userID, userMessage, userMsg.ID)

	return userMsg, nil
}

func (cs *ChatService) ChatWithVoice(ctx context.Context, userID int64, userMessage string, voiceFileID *int64) (*models.Message, error) {
	userMsg, err := cs.messageService.CreateMessage(userID, "user", userMessage, "", voiceFileID, "audio", 0)
	if err != nil {
		return nil, err
	}

	if err := database.DB.Preload("VoiceFile").First(userMsg, userMsg.ID).Error; err != nil {
		return nil, err
	}

	if userMsg.VoiceFile != nil {
		userMsg.VoiceFile.VoiceURL = ""
	}

	cs.scheduledTaskService.CancelUserPendingTasks(userID, "wait_message")

	go cs.processAIResponse(userID, userMessage, userMsg.ID)

	return userMsg, nil
}

func (cs *ChatService) GetChatHistory(userID int64, beforeID *int64) ([]models.Message, error) {
	var messages []models.Message
	query := database.DB.Preload("VoiceFile").Where("user_id = ?", userID).Order("id DESC").Limit(20)

	if beforeID != nil {
		query = query.Where("id < ?", *beforeID)
	}

	if err := query.Find(&messages).Error; err != nil {
		return nil, err
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (cs *ChatService) getRecentConversation(userID int64) ([]models.Message, error) {
	var allMessages []models.Message
	err := database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(100).
		Find(&allMessages).Error

	if err != nil {
		return nil, err
	}

	if len(allMessages) == 0 {
		return []models.Message{}, nil
	}

	var recentMessages []models.Message
	const maxInterval = 60 * 60

	for i := 0; i < len(allMessages); i++ {
		if i == 0 {
			recentMessages = append(recentMessages, allMessages[i])
		} else {
			interval := allMessages[i-1].CreatedAt.Unix() - allMessages[i].CreatedAt.Unix()
			if interval <= maxInterval {
				recentMessages = append(recentMessages, allMessages[i])
			} else {
				break
			}
		}
	}

	for i, j := 0, len(recentMessages)-1; i < j; i, j = i+1, j-1 {
		recentMessages[i], recentMessages[j] = recentMessages[j], recentMessages[i]
	}

	return recentMessages, nil
}

func (cs *ChatService) processAIResponse(userID int64, userMessage string, userMessageID int64) {
	ctx := context.Background()

	settings, err := cs.userSettingsService.GetUserSettings(userID)
	if err != nil {
		return
	}

	client := cs.createClient(settings)

	summaryObj, _ := cs.conversationSummaryService.GetConversationSummary(userID)
	summary := ""
	if summaryObj != nil {
		summary = summaryObj.Summary
	}

	recentMessages, _ := cs.getRecentConversation(userID)

	longTermMemories, _ := cs.getRelevantMemories(ctx, client, userID, userMessage, settings)

	messages := cs.buildMessages(recentMessages, longTermMemories, userMessage, settings, summary, false)
	result, tokensUsed, err := cs.callOpenAIWithStructuredOutput(ctx, client, messages, settings)
	if err != nil {
		return
	}

	if result.WaitForNextMsg {
		scheduledAtStr := time.Now().Add(30 * time.Second).UTC().Format(time.RFC3339)
		scheduledAt, err := time.Parse(time.RFC3339, scheduledAtStr)
		if err != nil {
			return
		}

		args := WaitMessageArgs{
			UserID:    userID,
			MessageID: userMessageID,
		}
		argsJSON, _ := json.Marshal(args)

		cs.scheduledTaskService.CreateTask(userID, "wait_message", string(argsJSON), scheduledAt)
		return
	}

	if result.Function != nil && result.Function.CallFunction {
		cs.handleFunctionCall(ctx, userID, result, settings)
	}

	totalSentences := len(result.ReplySentences)
	tokensPerSentence := 0
	if totalSentences > 0 {
		tokensPerSentence = tokensUsed / totalSentences
	}

	var allReplies []string
	for _, sentence := range result.ReplySentences {
		voiceFileID, _ := cs.TextToSpeech(ctx, userID, sentence.Original, settings)

		aiMessage, err := cs.messageService.CreateMessage(
			userID,
			"assistant",
			sentence.Original,
			sentence.Translation,
			voiceFileID,
			"text",
			tokensPerSentence,
		)
		if err != nil {
			continue
		}

		database.DB.Preload("VoiceFile").First(aiMessage, aiMessage.ID)

		if aiMessage.VoiceFile != nil {
			aiMessage.VoiceFile.VoiceURL = ""
		}

		messageJSON, _ := json.Marshal(aiMessage)
		cs.wsHub.SendToUser(userID, messageJSON)

		allReplies = append(allReplies, sentence.Original)
	}

	if result.Memory != nil && result.Memory.ShouldStore {
		go cs.updateMemory(userID, result, settings)
	}

	if result.Summary != nil && result.Summary.ShouldUpdate {
		go cs.updateSummary(userID, result.Summary.Content)
	}
}

func (cs *ChatService) processInstantAIResponse(userID int64, userMessageID int64) {
	ctx := context.Background()

	recentMessages, _ := cs.getRecentConversation(userID)

	userMessage := ""
	for i := len(recentMessages) - 1; i >= 0; i-- {
		if recentMessages[i].Role == "user" {
			if recentMessages[i].ID > userMessageID {
				return
			}
			userMessage = recentMessages[i].TextContent
			break
		}
	}

	settings, err := cs.userSettingsService.GetUserSettings(userID)
	if err != nil {
		return
	}

	client := cs.createClient(settings)

	summaryObj, _ := cs.conversationSummaryService.GetConversationSummary(userID)
	summary := ""
	if summaryObj != nil {
		summary = summaryObj.Summary
	}

	longTermMemories, _ := cs.getRelevantMemories(ctx, client, userID, userMessage, settings)

	messages := cs.buildMessages(recentMessages, longTermMemories, userMessage, settings, summary, true)

	result, tokensUsed, err := cs.callOpenAIWithStructuredOutput(ctx, client, messages, settings)
	if err != nil {
		return
	}

	if result.Function != nil && result.Function.CallFunction {
		cs.handleFunctionCall(ctx, userID, result, settings)
	}

	totalSentences := len(result.ReplySentences)
	tokensPerSentence := 0
	if totalSentences > 0 {
		tokensPerSentence = tokensUsed / totalSentences
	}

	var allReplies []string
	for _, sentence := range result.ReplySentences {
		voiceFileID, _ := cs.TextToSpeech(ctx, userID, sentence.Original, settings)

		aiMessage, err := cs.messageService.CreateMessage(
			userID,
			"assistant",
			sentence.Original,
			sentence.Translation,
			voiceFileID,
			"text",
			tokensPerSentence,
		)
		if err != nil {
			continue
		}

		database.DB.Preload("VoiceFile").First(aiMessage, aiMessage.ID)

		if aiMessage.VoiceFile != nil {
			aiMessage.VoiceFile.VoiceURL = ""
		}

		messageJSON, _ := json.Marshal(aiMessage)
		cs.wsHub.SendToUser(userID, messageJSON)

		allReplies = append(allReplies, sentence.Original)
	}

	if result.Memory != nil && result.Memory.ShouldStore {
		go cs.updateMemory(userID, result, settings)
	}

	if result.Summary != nil && result.Summary.ShouldUpdate {
		go cs.updateSummary(userID, result.Summary.Content)
	}
}

func (cs *ChatService) createClient(settings *models.UserSettings) openai.Client {
	apiKey := settings.STTAPIKey
	apiBaseURL := settings.STTAPIBaseURL

	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if apiBaseURL != "" {
		opts = append(opts, option.WithBaseURL(apiBaseURL))
	}
	return openai.NewClient(opts...)
}

func (cs *ChatService) getRelevantMemories(ctx context.Context, client openai.Client, userID int64, message string, settings *models.UserSettings) ([]models.Message, error) {
	embeddingClient := client

	embeddingAPIKey := settings.EmbeddingAPIKey
	embeddingAPIBaseURL := settings.EmbeddingAPIBaseURL

	if embeddingAPIKey != "" {
		opts := []option.RequestOption{option.WithAPIKey(embeddingAPIKey)}
		if embeddingAPIBaseURL != "" {
			opts = append(opts, option.WithBaseURL(embeddingAPIBaseURL))
		}
		embeddingClient = openai.NewClient(opts...)
	}

	embeddingModel := settings.EmbeddingModel
	if embeddingModel == "" {
		embeddingModel = "text-embedding-3-small"
	}

	resp, err := embeddingClient.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model: openai.EmbeddingModel(embeddingModel),
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: []string{message},
		},
	})
	if err != nil {
		return nil, err
	}

	embedding := resp.Data[0].Embedding
	embeddingFloat32 := make([]float32, len(embedding))
	for i, v := range embedding {
		embeddingFloat32[i] = float32(v)
	}

	var memories []memoryWithDistance
	err = database.DB.Raw(`
		SELECT *, embedding <-> ? AS distance
		FROM user_memories
		WHERE user_id = ?
		ORDER BY embedding <-> ?
		LIMIT 3
	`,
		pgvector.NewVector(embeddingFloat32),
		userID,
		pgvector.NewVector(embeddingFloat32),
	).Scan(&memories).Error

	if err != nil {
		return nil, err
	}

	var allRelevantMessages []models.Message
	seenMessageIDs := make(map[int64]bool)

	const similarityThreshold = 0.9
	for _, memory := range memories {
		if memory.Distance > similarityThreshold {
			continue
		}

		var nearbyMessages []models.Message
		err := database.DB.Where(
			"user_id = ? AND created_at BETWEEN ? AND ?",
			userID,
			memory.CreatedAt.Add(-3*time.Hour),
			memory.CreatedAt.Add(3*time.Hour),
		).
			Order("created_at ASC").
			Find(&nearbyMessages).Error

		if err != nil {
			continue
		}

		expanded := cs.expandMessagesWithInterval(nearbyMessages)

		for _, msg := range expanded {
			if !seenMessageIDs[msg.ID] {
				allRelevantMessages = append(allRelevantMessages, msg)
				seenMessageIDs[msg.ID] = true
			}
		}
	}

	sort.Slice(allRelevantMessages, func(i, j int) bool {
		return allRelevantMessages[i].CreatedAt.Before(allRelevantMessages[j].CreatedAt)
	})

	return allRelevantMessages, nil
}

func (cs *ChatService) expandMessagesWithInterval(messages []models.Message) []models.Message {
	if len(messages) == 0 {
		return []models.Message{}
	}

	const maxInterval = 60 * 60
	var result []models.Message

	included := make([]bool, len(messages))

	for i := 0; i < len(messages); i++ {
		included[i] = true

		for j := i - 1; j >= 0; j-- {
			interval := messages[j+1].CreatedAt.Unix() - messages[j].CreatedAt.Unix()
			if interval <= maxInterval {
				included[j] = true
			} else {
				break
			}
		}

		for j := i + 1; j < len(messages); j++ {
			interval := messages[j].CreatedAt.Unix() - messages[j-1].CreatedAt.Unix()
			if interval <= maxInterval {
				included[j] = true
			} else {
				break
			}
		}
	}

	for i, include := range included {
		if include {
			result = append(result, messages[i])
		}
	}

	return result
}

func (cs *ChatService) buildMessages(recentMessages []models.Message, longTermMemories []models.Message, userMessage string, settings *models.UserSettings, summary string, instant bool) []openai.ChatCompletionMessageParamUnion {
	timezone := settings.Timezone
	if timezone == "" {
		timezone = "UTC"
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	var recentMsgs []interface{}
	for _, msg := range recentMessages {
		timeStr := msg.CreatedAt.In(loc).Format("2006-01-02 15:04:05")
		recentMsgs = append(recentMsgs, fmt.Sprintf("[%s] %s: %s", timeStr, msg.Role, msg.TextContent))
	}

	var longTermMsgs []interface{}
	for _, msg := range longTermMemories {
		timeStr := msg.CreatedAt.In(loc).Format("2006-01-02 15:04:05")
		longTermMsgs = append(longTermMsgs, fmt.Sprintf("[%s] %s: %s", timeStr, msg.Role, msg.TextContent))
	}

	currentTime := time.Now().In(loc).Format("2006-01-02 15:04:05")
	systemContent := BuildFullSystemPrompt(settings.NativeLanguage, settings.TargetLanguage, summary, recentMsgs, longTermMsgs, currentTime, timezone)
	systemNotWaitContent := BuildSystemInstantPrompt()

	if instant {
		systemContent += systemNotWaitContent
	}

	return []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemContent),
		openai.UserMessage(userMessage),
	}
}

func (cs *ChatService) callOpenAIWithStructuredOutput(ctx context.Context, client openai.Client, messages []openai.ChatCompletionMessageParamUnion, settings *models.UserSettings) (*ChatEngineResult, int, error) {
	model := settings.Model
	if model == "" {
		model = "gpt-4o-mini"
	}

	schema := GenerateSchema[ChatEngineResult]()

	jsonSchemaParam := shared.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "chat_engine_output",
		Strict:      param.NewOpt(true),
		Description: param.NewOpt("语言学习聊天引擎的结构化输出"),
		Schema:      schema,
	}

	responseFormat := openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{
			JSONSchema: jsonSchemaParam,
			Type:       "json_schema",
		},
	}

	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages:        messages,
		Model:           openai.ChatModel(model),
		ResponseFormat:  responseFormat,
		ReasoningEffort: "none",
	})
	if err != nil {
		return nil, 0, err
	}

	var result ChatEngineResult
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result); err != nil {
		return nil, 0, err
	}

	return &result, int(resp.Usage.TotalTokens), nil
}

func (cs *ChatService) handleFunctionCall(ctx context.Context, userID int64, result *ChatEngineResult, settings *models.UserSettings) {
	switch result.Function.FunctionName {
	case "schedule_message":
		message, _ := result.Function.FunctionArgs["message"].(string)
		translation, _ := result.Function.FunctionArgs["translation"].(string)
		scheduledAtStr, _ := result.Function.FunctionArgs["scheduled_at"].(string)

		scheduledAt, err := time.Parse(time.RFC3339, scheduledAtStr)
		if err != nil {
			return
		}

		args := SendMessageArgs{
			UserID:      userID,
			Message:     message,
			Translation: translation,
		}
		argsJSON, _ := json.Marshal(args)

		cs.scheduledTaskService.CreateTask(userID, "send_message", string(argsJSON), scheduledAt)
	}
}

func (cs *ChatService) updateSummary(userID int64, content string) {
	cs.conversationSummaryService.UpdateConversationSummary(userID, content)
}

func (cs *ChatService) updateMemory(userID int64, result *ChatEngineResult, settings *models.UserSettings) {
	ctx := context.Background()
	client := cs.createClient(settings)

	embeddingClient := client

	embeddingAPIKey := settings.EmbeddingAPIKey
	embeddingAPIBaseURL := settings.EmbeddingAPIBaseURL

	if embeddingAPIKey != "" {
		opts := []option.RequestOption{option.WithAPIKey(embeddingAPIKey)}
		if embeddingAPIBaseURL != "" {
			opts = append(opts, option.WithBaseURL(embeddingAPIBaseURL))
		}
		embeddingClient = openai.NewClient(opts...)
	}

	embeddingModel := settings.EmbeddingModel
	if embeddingModel == "" {
		embeddingModel = "text-embedding-3-small"
	}

	resp, err := embeddingClient.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model: openai.EmbeddingModel(embeddingModel),
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: []string{result.Memory.SemanticContent},
		},
	})
	if err == nil {
		embedding := resp.Data[0].Embedding
		embeddingBytes, _ := json.Marshal(embedding)
		cs.userMemoryService.CreateUserMemory(userID, result.Memory.SemanticContent, string(embeddingBytes), result.Memory.MemoryType, result.Memory.Importance)
	}
}
