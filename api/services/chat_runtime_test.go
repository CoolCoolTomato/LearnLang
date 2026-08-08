package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"learnlang-api/database"
	"learnlang-api/models"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

type fakeWSHub struct {
	userID   int64
	messages [][]byte
}

func (h *fakeWSHub) SendToUser(userID int64, message []byte) {
	h.userID = userID
	h.messages = append(h.messages, append([]byte(nil), message...))
}

func setupChatRuntimeTest(t *testing.T, apiBaseURL string) (*ChatRuntimeService, *fakeWSHub) {
	t.Helper()
	setupServiceTestDB(t, &models.ConversationArchive{})
	settings := models.UserSettings{
		UserID: 1, NativeLanguage: "zh-CN", TargetLanguage: "en-US", Timezone: "Asia/Shanghai",
		STTAPIKey: "key", STTAPIBaseURL: apiBaseURL, STTModel: "whisper-test",
		TTSAPIKey: "key", TTSAPIBaseURL: apiBaseURL, TTSModel: "tts-test", TTSVoice: "alloy",
	}
	if err := database.DB.Create(&settings).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.DB.Create(&models.UserProfileSummary{UserID: 1, Summary: "profile"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.DB.Create(&models.User{ID: 1, Username: "user", PasswordHash: "hash"}).Error; err != nil {
		t.Fatal(err)
	}
	hub := &fakeWSHub{}
	runtime := NewChatRuntimeService(
		NewMessageService(), NewUserSettingsService(), NewUserProfileSummaryService(),
		NewScheduledTaskService(), NewVoiceFileService(), hub,
	)
	return runtime, hub
}

func TestChatRuntimeMessageAndTaskOperations(t *testing.T) {
	runtime, hub := setupChatRuntimeTest(t, "http://127.0.0.1:1")
	voice := models.VoiceFile{UserID: 1, VoiceRole: "user", VoiceURL: "secret/path.mp3"}
	if err := database.DB.Create(&voice).Error; err != nil {
		t.Fatal(err)
	}
	message, err := runtime.CreateUserMessage(context.Background(), 1, "hello", &voice.ID, "voice")
	if err != nil || message.InputType != "voice" || message.VoiceFile == nil || message.VoiceFile.VoiceURL != "" {
		t.Fatalf("CreateUserMessage() = %#v, %v", message, err)
	}
	otherUserVoice := models.VoiceFile{UserID: 2, VoiceRole: "user", VoiceURL: "other.mp3"}
	if err := database.DB.Create(&otherUserVoice).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.CreateUserMessage(context.Background(), 1, "hello", &otherUserVoice.ID, "voice"); !errors.Is(err, ErrUserVoiceFileInvalid) {
		t.Fatalf("cross-user voice error = %v", err)
	}
	assistantVoice := models.VoiceFile{UserID: 1, VoiceRole: "assistant", VoiceURL: "assistant.mp3"}
	if err := database.DB.Create(&assistantVoice).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.CreateUserMessage(context.Background(), 1, "hello", &assistantVoice.ID, "voice"); !errors.Is(err, ErrUserVoiceFileInvalid) {
		t.Fatalf("assistant voice error = %v", err)
	}

	taskID, err := runtime.ScheduleMessage(context.Background(), 1, "later", "稍后", time.Date(2026, 8, 1, 1, 0, 0, 0, time.FixedZone("local", 8*3600)))
	if err != nil || taskID == 0 {
		t.Fatalf("ScheduleMessage() = %d, %v", taskID, err)
	}
	var task models.ScheduledTask
	if err := database.DB.First(&task, taskID).Error; err != nil || task.FunctionName != "send_message" || !strings.Contains(task.Args, "later") || task.ScheduledAt.Location() != time.UTC {
		t.Fatalf("scheduled task = %#v, %v", task, err)
	}

	runtime.SendAgentError(1)
	if hub.userID != 1 || len(hub.messages) != 1 || !strings.Contains(string(hub.messages[0]), "agent_error") {
		t.Fatalf("agent error events = %#v", hub.messages)
	}
	runtime.SendAgentTurnCompleted(1)
	if len(hub.messages) != 2 || !strings.Contains(string(hub.messages[1]), "agent_turn_completed") {
		t.Fatalf("agent completion events = %#v", hub.messages)
	}
	settings, err := runtime.UserSettings(1)
	if err != nil || settings.TargetLanguage != "en-US" {
		t.Fatalf("UserSettings() = %#v, %v", settings, err)
	}
	profile, err := runtime.UserProfileSummary(1)
	if err != nil || profile.Summary != "profile" {
		t.Fatalf("UserProfileSummary() = %#v, %v", profile, err)
	}

	since := time.Now().Add(-time.Hour)
	before := int64(999)
	history, err := runtime.GetChatHistory(context.Background(), 1, &before)
	if err != nil || len(history) != 1 {
		t.Fatalf("GetChatHistory() = %#v, %v", history, err)
	}
	short, err := runtime.GetShortTermMemory(context.Background(), 1, 0, since)
	if err != nil || len(short) != 1 || short[0].ID != message.ID {
		t.Fatalf("GetShortTermMemory() = %#v, %v", short, err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := runtime.ScheduleMessage(canceled, 1, "x", "y", time.Now()); err != context.Canceled {
		t.Fatalf("canceled ScheduleMessage() error = %v", err)
	}
}

func TestGetShortTermMemoryFallsBackToLatestHistoryAfterWindow(t *testing.T) {
	runtime, _ := setupChatRuntimeTest(t, "http://127.0.0.1:1")
	now := time.Now().UTC()
	oldMessage := models.Message{
		UserID:      1,
		Role:        "user",
		TextContent: "old context",
		CreatedAt:   now.Add(-48 * time.Hour),
	}
	currentMessage := models.Message{
		UserID:      1,
		Role:        "user",
		TextContent: "current input",
		CreatedAt:   now,
	}
	if err := database.DB.Create(&oldMessage).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.DB.Create(&currentMessage).Error; err != nil {
		t.Fatal(err)
	}

	got, err := runtime.GetShortTermMemory(context.Background(), 1, currentMessage.ID, now.Add(-24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != oldMessage.ID {
		t.Fatalf("fallback short-term memory = %#v, want latest preceding message", got)
	}
}

func TestGetShortTermMemoryPrefersRecentWindow(t *testing.T) {
	runtime, _ := setupChatRuntimeTest(t, "http://127.0.0.1:1")
	now := time.Now().UTC()
	oldMessage := models.Message{
		UserID:      1,
		Role:        "user",
		TextContent: "old context",
		CreatedAt:   now.Add(-48 * time.Hour),
	}
	recentMessage := models.Message{
		UserID:      1,
		Role:        "user",
		TextContent: "recent context",
		CreatedAt:   now.Add(-time.Hour),
	}
	currentMessage := models.Message{
		UserID:      1,
		Role:        "user",
		TextContent: "current input",
		CreatedAt:   now,
	}
	for _, message := range []*models.Message{&oldMessage, &recentMessage, &currentMessage} {
		if err := database.DB.Create(message).Error; err != nil {
			t.Fatal(err)
		}
	}

	got, err := runtime.GetShortTermMemory(context.Background(), 1, currentMessage.ID, now.Add(-24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != recentMessage.ID {
		t.Fatalf("recent short-term memory = %#v, want only recent window", got)
	}
}

func TestGetShortTermMemoryFallbackIncludesAllUnarchivedMessages(t *testing.T) {
	runtime, _ := setupChatRuntimeTest(t, "http://127.0.0.1:1")
	now := time.Now().UTC()
	for i := 0; i < 21; i++ {
		message := models.Message{
			UserID:      1,
			Role:        "user",
			TextContent: fmt.Sprintf("old context %d", i),
			CreatedAt:   now.Add(-time.Duration(48-i) * time.Hour),
		}
		if err := database.DB.Create(&message).Error; err != nil {
			t.Fatal(err)
		}
	}
	currentMessage := models.Message{
		UserID:      1,
		Role:        "user",
		TextContent: "current input",
		CreatedAt:   now,
	}
	if err := database.DB.Create(&currentMessage).Error; err != nil {
		t.Fatal(err)
	}

	got, err := runtime.GetShortTermMemory(context.Background(), 1, currentMessage.ID, now.Add(-24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 21 {
		t.Fatalf("fallback length = %d, want all 21 unarchived messages", len(got))
	}
	if got[0].TextContent != "old context 0" || got[len(got)-1].TextContent != "old context 20" {
		t.Fatalf("fallback order = %q ... %q, want chronological order", got[0].TextContent, got[len(got)-1].TextContent)
	}
}

func TestGetShortTermMemoryFallbackIncludesUnarchivedAndFillsArchived(t *testing.T) {
	runtime, _ := setupChatRuntimeTest(t, "http://127.0.0.1:1")
	now := time.Now().UTC()
	messages := make([]models.Message, 25)
	messageIDs := make([]int64, 10)
	for i := range messages {
		messages[i] = models.Message{
			UserID:      1,
			Role:        "user",
			TextContent: fmt.Sprintf("history %02d", i),
			CreatedAt:   now.Add(-time.Duration(50-i) * time.Hour),
		}
		if err := database.DB.Create(&messages[i]).Error; err != nil {
			t.Fatal(err)
		}
		if i < len(messageIDs) {
			messageIDs[i] = messages[i].ID
		}
	}
	if err := database.DB.Create(&models.ConversationArchive{
		UserID:       1,
		MessageIDs:   messageIDs,
		Summary:      "archived history",
		MessageCount: len(messageIDs),
	}).Error; err != nil {
		t.Fatal(err)
	}
	currentMessage := models.Message{
		UserID:      1,
		Role:        "user",
		TextContent: "current input",
		CreatedAt:   now,
	}
	if err := database.DB.Create(&currentMessage).Error; err != nil {
		t.Fatal(err)
	}

	got, err := runtime.GetShortTermMemory(context.Background(), 1, currentMessage.ID, now.Add(-24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 20 {
		t.Fatalf("fallback length = %d, want 20", len(got))
	}
	if got[0].TextContent != "history 05" || got[len(got)-1].TextContent != "history 24" {
		t.Fatalf("fallback range = %q ... %q, want five archived plus all unarchived", got[0].TextContent, got[len(got)-1].TextContent)
	}
	for i := 10; i < len(messages); i++ {
		found := false
		for _, message := range got {
			if message.ID == messages[i].ID {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("unarchived message %d was omitted", i)
		}
	}
}

func TestChatRuntimeAudioUsesMockHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/audio/transcriptions":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"text":"transcribed"}`))
		case "/audio/speech":
			w.Header().Set("Content-Type", "audio/mpeg")
			_, _ = w.Write([]byte("mock-mp3-content"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	runtime, hub := setupChatRuntimeTest(t, server.URL)

	oldWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWorkingDirectory) })

	text, voiceID, err := runtime.TranscribeAudio(context.Background(), 1, bytes.NewBufferString("invalid-mp3-but-valid-upload"))
	if err != nil || text != "transcribed" || voiceID == nil || *voiceID == 0 {
		t.Fatalf("TranscribeAudio() = %q, %v, %v", text, voiceID, err)
	}
	settings, _ := runtime.UserSettings(1)
	ttsID, err := runtime.TextToSpeech(context.Background(), 1, "hello", settings)
	if err != nil || ttsID == nil || *ttsID == 0 {
		t.Fatalf("TextToSpeech() = %v, %v", ttsID, err)
	}
	replyID, err := runtime.SaveAssistantReply(context.Background(), 1, "answer", "回答")
	if err != nil || replyID == 0 {
		t.Fatalf("SaveAssistantReply() = %d, %v", replyID, err)
	}
	if len(hub.messages) != 1 {
		t.Fatalf("assistant reply websocket messages = %d", len(hub.messages))
	}
	var sent models.Message
	if err := json.Unmarshal(hub.messages[0], &sent); err != nil || sent.ID != replyID || sent.TextContent != "answer" {
		t.Fatalf("sent message = %#v, %v", sent, err)
	}

	handler := NewSendMessageHandler(runtime)
	if err := handler(`{"user_id":1,"message":"scheduled","translation":"计划"}`); err != nil {
		t.Fatalf("send message handler error = %v", err)
	}
	if err := handler(`{`); err == nil {
		t.Fatal("send message handler accepted invalid JSON")
	}
}

func TestTextToSpeechUsesFishAudioRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/tts" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer fish-key" || r.Header.Get("model") != "s2.1-pro-free" {
			t.Fatalf("Fish Audio headers = %#v", r.Header)
		}
		var payload fishAudioTTSRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.Text != "hello" || payload.ReferenceID != "voice-model-id" || payload.Format != "mp3" || payload.Temperature != 0 || payload.TopP != 0 || payload.ChunkLength != 300 || !payload.Normalize || payload.Latency != "normal" || payload.RepetitionPenalty != 1.2 || !payload.ConditionOnPreviousChunks {
			t.Fatalf("Fish Audio payload = %#v", payload)
		}
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Header().Set("x-request-id", "fish-request")
		_, _ = w.Write([]byte("mock-mp3-content"))
	}))
	defer server.Close()
	runtime, _ := setupChatRuntimeTest(t, server.URL)

	oldWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWorkingDirectory) })

	settings, err := runtime.UserSettings(1)
	if err != nil {
		t.Fatal(err)
	}
	settings.TTSType = models.TTSTypeFishAudio
	settings.TTSAPIKey = "fish-key"
	settings.TTSModel = "s2.1-pro-free"
	settings.TTSVoice = "voice-model-id"

	voiceFileID, err := runtime.TextToSpeech(context.Background(), 1, "hello", settings)
	if err != nil || voiceFileID == nil || *voiceFileID == 0 {
		t.Fatalf("TextToSpeech() = %v, %v", voiceFileID, err)
	}
}

func TestTextToSpeechRejectsFishAudioWithoutReferenceID(t *testing.T) {
	runtime, _ := setupChatRuntimeTest(t, "http://127.0.0.1:1")
	settings, err := runtime.UserSettings(1)
	if err != nil {
		t.Fatal(err)
	}
	settings.TTSType = models.TTSTypeFishAudio
	settings.TTSVoice = ""

	if _, err := runtime.TextToSpeech(context.Background(), 1, "hello", settings); !errors.Is(err, ErrFishAudioReferenceIDRequired) {
		t.Fatalf("TextToSpeech() error = %v", err)
	}
}
