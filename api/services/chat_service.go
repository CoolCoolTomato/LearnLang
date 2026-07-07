package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"learnlang-api/agent"
	"learnlang-api/agent/memory"
	"learnlang-api/config"
	"learnlang-api/database"
	"learnlang-api/models"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type ChatService struct {
	messageService             *MessageService
	conversationSummaryService *ConversationSummaryService
	memoryStore                *memory.Store
	userSettingsService        *UserSettingsService
	scheduledTaskService       *ScheduledTaskService
	voiceFileService           *VoiceFileService
	chatAgent                  *agent.Service
	wsHub                      WSHub
}

type WSHub interface {
	SendToUser(userID int64, message []byte)
}

func NewChatService(
	messageService *MessageService,
	conversationSummaryService *ConversationSummaryService,
	milvusCfg config.MilvusConfig,
	userSettingsService *UserSettingsService,
	scheduledTaskService *ScheduledTaskService,
	voiceFileService *VoiceFileService,
	wsHub WSHub,
) *ChatService {
	memoryStore := memory.NewStore(memory.Config{
		Address:    fmt.Sprintf("%s:%s", milvusCfg.Host, milvusCfg.Port),
		Collection: milvusCfg.Collection,
		Dimension:  milvusCfg.Dimension,
	})

	return &ChatService{
		messageService:             messageService,
		conversationSummaryService: conversationSummaryService,
		memoryStore:                memoryStore,
		userSettingsService:        userSettingsService,
		scheduledTaskService:       scheduledTaskService,
		voiceFileService:           voiceFileService,
		chatAgent: agent.NewService(agent.Config{
			MemoryStore: memoryStore,
		}),
		wsHub: wsHub,
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

	result, err := cs.runAgentChat(ctx, userID, userMessage, settings, false)
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
		tokensPerSentence = result.TokensUsed / totalSentences
	}

	memoryMessageIDs := []int64{userMessageID}
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
		memoryMessageIDs = append(memoryMessageIDs, aiMessage.ID)
	}

	if result.Memory != nil && result.Memory.ShouldStore {
		go cs.updateMemory(userID, result, settings, memoryMessageIDs)
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

	result, err := cs.runAgentChat(ctx, userID, userMessage, settings, true)
	if err != nil {
		return
	}

	if result.Function != nil && result.Function.CallFunction {
		cs.handleFunctionCall(ctx, userID, result, settings)
	}

	totalSentences := len(result.ReplySentences)
	tokensPerSentence := 0
	if totalSentences > 0 {
		tokensPerSentence = result.TokensUsed / totalSentences
	}

	memoryMessageIDs := []int64{userMessageID}
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
		memoryMessageIDs = append(memoryMessageIDs, aiMessage.ID)
	}

	if result.Memory != nil && result.Memory.ShouldStore {
		go cs.updateMemory(userID, result, settings, memoryMessageIDs)
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

func (cs *ChatService) runAgentChat(ctx context.Context, userID int64, userMessage string, settings *models.UserSettings, instant bool) (*agent.ChatResult, error) {
	timezone := settings.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	apiKey := settings.APIKey
	if apiKey == "" {
		apiKey = settings.STTAPIKey
	}

	apiBaseURL := settings.APIBaseURL
	if apiBaseURL == "" {
		apiBaseURL = settings.STTAPIBaseURL
	}

	return cs.chatAgent.RunChat(ctx, agent.ChatRequest{
		UserID:      userID,
		UserInput:   userMessage,
		Instant:     instant,
		CurrentTime: time.Now().In(loc).Format("2006-01-02 15:04:05"),
		Timezone:    timezone,
		Settings: agent.UserSettings{
			APIKey:              apiKey,
			APIBaseURL:          apiBaseURL,
			Model:               settings.Model,
			EmbeddingAPIKey:     settings.EmbeddingAPIKey,
			EmbeddingAPIBaseURL: settings.EmbeddingAPIBaseURL,
			EmbeddingModel:      settings.EmbeddingModel,
			NativeLanguage:      settings.NativeLanguage,
			TargetLanguage:      settings.TargetLanguage,
		},
	})
}

func (cs *ChatService) handleFunctionCall(ctx context.Context, userID int64, result *agent.ChatResult, settings *models.UserSettings) {
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

func (cs *ChatService) updateMemory(userID int64, result *agent.ChatResult, settings *models.UserSettings, messageIDs []int64) {
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
		embedding := make([]float32, 0, len(resp.Data[0].Embedding))
		for _, v := range resp.Data[0].Embedding {
			embedding = append(embedding, float32(v))
		}
		if _, err := cs.memoryStore.InsertSummary(ctx, userID, result.Memory.SemanticContent, result.Memory.MemoryType, result.Memory.Importance, messageIDs, embedding); err != nil {
			log.Printf("failed to store user memory in milvus: %v", err)
		}
	}
}
