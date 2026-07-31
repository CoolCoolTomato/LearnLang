package services

import (
	"bytes"
	"context"
	"encoding/json"
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
	setupServiceTestDB(t)
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
	voice := models.VoiceFile{UserID: 1, VoiceURL: "secret/path.mp3"}
	if err := database.DB.Create(&voice).Error; err != nil {
		t.Fatal(err)
	}
	message, err := runtime.CreateUserMessage(context.Background(), 1, "hello", &voice.ID, "voice")
	if err != nil || message.VoiceFile == nil || message.VoiceFile.VoiceURL != "" {
		t.Fatalf("CreateUserMessage() = %#v, %v", message, err)
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
