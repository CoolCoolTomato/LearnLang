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
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type ChatRuntimeService struct {
	messageService        *MessageService
	userSettingsService   *UserSettingsService
	profileSummaryService *UserProfileSummaryService
	scheduledTaskService  *ScheduledTaskService
	voiceFileService      *VoiceFileService
	wsHub                 WSHub
}

type WSHub interface {
	SendToUser(userID int64, message []byte)
}

func NewChatRuntimeService(
	messageService *MessageService,
	userSettingsService *UserSettingsService,
	profileSummaryService *UserProfileSummaryService,
	scheduledTaskService *ScheduledTaskService,
	voiceFileService *VoiceFileService,
	wsHub WSHub,
) *ChatRuntimeService {
	return &ChatRuntimeService{
		messageService:        messageService,
		userSettingsService:   userSettingsService,
		profileSummaryService: profileSummaryService,
		scheduledTaskService:  scheduledTaskService,
		voiceFileService:      voiceFileService,
		wsHub:                 wsHub,
	}
}

func (crs *ChatRuntimeService) TranscribeAudio(ctx context.Context, userID int64, audioFile io.Reader) (string, *int64, error) {
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

	settings, err := crs.userSettingsService.GetUserSettings(userID)
	if err != nil {
		return "", nil, err
	}

	client := openai.NewClient(
		option.WithAPIKey(settings.STTAPIKey),
		option.WithBaseURL(settings.STTAPIBaseURL),
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
	if err := crs.voiceFileService.CreateVoiceFile(voiceFile); err != nil {
		return "", nil, err
	}

	return transcription.Text, &voiceFile.ID, nil
}

func (crs *ChatRuntimeService) TextToSpeech(ctx context.Context, userID int64, text string, settings *models.UserSettings) (*int64, error) {
	client := openai.NewClient(
		option.WithAPIKey(settings.TTSAPIKey),
		option.WithBaseURL(settings.TTSAPIBaseURL),
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
		Model: openai.SpeechModel(model),
		Input: text,
		Voice: openai.AudioSpeechNewParamsVoiceUnion{
			OfAudioSpeechNewsVoiceString2: openai.String(voice),
		},
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
	if err := crs.voiceFileService.CreateVoiceFile(voiceFile); err != nil {
		return nil, err
	}

	return &voiceFile.ID, nil
}

func (crs *ChatRuntimeService) CreateUserMessage(ctx context.Context, userID int64, text string, voiceFileID *int64, messageType string) (*models.Message, error) {
	message, err := crs.messageService.CreateMessage(ctx, userID, "user", text, "", voiceFileID, messageType, 0)
	if err != nil {
		return nil, err
	}

	if err := database.DB.WithContext(ctx).Preload("VoiceFile").First(message, message.ID).Error; err != nil {
		return nil, err
	}

	if message.VoiceFile != nil {
		message.VoiceFile.VoiceURL = ""
	}

	return message, nil
}

func (crs *ChatRuntimeService) SaveAssistantReply(ctx context.Context, userID int64, original, translation string) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	settings, err := crs.userSettingsService.GetUserSettings(userID)
	if err != nil {
		return 0, err
	}

	voiceFileID, _ := crs.TextToSpeech(ctx, userID, original, settings)
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	message, err := crs.messageService.CreateMessage(ctx, userID, "assistant", original, translation, voiceFileID, "text", 0)
	if err != nil {
		return 0, err
	}

	database.DB.WithContext(ctx).Preload("VoiceFile").First(message, message.ID)
	if message.VoiceFile != nil {
		message.VoiceFile.VoiceURL = ""
	}

	messageJSON, _ := json.Marshal(message)
	crs.wsHub.SendToUser(userID, messageJSON)

	return message.ID, nil
}

func (crs *ChatRuntimeService) ScheduleMessage(ctx context.Context, userID int64, message, translation string, scheduledAt time.Time) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	args := SendMessageArgs{
		UserID:      userID,
		Message:     message,
		Translation: translation,
	}
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return 0, err
	}

	task, err := crs.scheduledTaskService.CreateTask(ctx, userID, "send_message", string(argsJSON), scheduledAt)
	if err != nil {
		return 0, err
	}

	return task.ID, nil
}

func (crs *ChatRuntimeService) GetChatHistory(userID int64, beforeID *int64) ([]models.Message, error) {
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

func (crs *ChatRuntimeService) GetShortTermMemory(ctx context.Context, userID, beforeMessageID int64, since time.Time) ([]models.Message, error) {
	var messages []models.Message
	query := database.DB.WithContext(ctx).
		Where("user_id = ? AND created_at >= ?", userID, since).
		Order("created_at ASC, id ASC")
	if beforeMessageID > 0 {
		query = query.Where("id < ?", beforeMessageID)
	}
	if err := query.Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (crs *ChatRuntimeService) UserSettings(userID int64) (*models.UserSettings, error) {
	return crs.userSettingsService.GetUserSettings(userID)
}

func (crs *ChatRuntimeService) UserProfileSummary(userID int64) (*models.UserProfileSummary, error) {
	return crs.profileSummaryService.GetUserProfileSummary(userID)
}
