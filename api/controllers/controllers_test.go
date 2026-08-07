package controllers

import (
	"learnlang-api/services"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func controllerContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")
	return context, recorder
}

func TestVocabularyControllerHelpers(t *testing.T) {
	context, recorder := controllerContext(http.MethodGet, "/?page=3", "")
	context.Params = gin.Params{{Key: "id", Value: "12"}}
	if id, ok := vocabularyIDParam(context); !ok || id != 12 {
		t.Fatalf("vocabularyIDParam() = %d, %v", id, ok)
	}
	if page, err := positiveIntQuery(context, "page", 1); err != nil || page != 3 {
		t.Fatalf("positiveIntQuery() = %d, %v", page, err)
	}
	if size, err := positiveIntQuery(context, "size", 20); err != nil || size != 20 {
		t.Fatalf("positiveIntQuery(fallback) = %d, %v", size, err)
	}

	for _, value := range []string{"", "0", "bad"} {
		context, recorder = controllerContext(http.MethodGet, "/", "")
		context.Params = gin.Params{{Key: "id", Value: value}}
		if _, ok := vocabularyIDParam(context); ok || recorder.Code != http.StatusBadRequest {
			t.Errorf("vocabularyIDParam(%q) = ok %v, status %d", value, ok, recorder.Code)
		}
	}
	for _, query := range []string{"0", "bad"} {
		context, _ = controllerContext(http.MethodGet, "/?page="+query, "")
		if _, err := positiveIntQuery(context, "page", 1); err == nil {
			t.Errorf("positiveIntQuery(%q) unexpectedly succeeded", query)
		}
	}
}

func TestVocabularyControllerErrorMapping(t *testing.T) {
	controller := NewVocabularyController(nil)
	tests := []struct {
		err  error
		code int
	}{
		{services.ErrVocabularyNotFound, http.StatusNotFound},
		{services.ErrVocabularyMessageNotFound, http.StatusNotFound},
		{services.ErrVocabularyNameConflict, http.StatusConflict},
		{services.ErrVocabularyInvalidImport, http.StatusBadRequest},
		{services.ErrVocabularyInvalidInput, http.StatusBadRequest},
		{services.ErrVocabularyLanguageRequired, http.StatusBadRequest},
		{http.ErrAbortHandler, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		context, recorder := controllerContext(http.MethodGet, "/", "")
		controller.writeError(context, tt.err)
		if recorder.Code != tt.code {
			t.Errorf("writeError(%v) status = %d, want %d", tt.err, recorder.Code, tt.code)
		}
	}
}

func TestControllerValidationPaths(t *testing.T) {
	tests := []struct {
		name string
		call func(*gin.Context)
		code int
	}{
		{"chat auth", NewChatController(nil).Chat, http.StatusUnauthorized},
		{"transcribe auth", NewChatController(nil).Transcribe, http.StatusUnauthorized},
		{"history auth", NewChatController(nil).GetChatHistory, http.StatusUnauthorized},
		{"translation auth", NewTranslationController(nil).Translate, http.StatusUnauthorized},
		{"profile auth", NewProfileController(nil, nil).GetMyProfile, http.StatusUnauthorized},
		{"profile update auth", NewProfileController(nil, nil).UpdateMyProfile, http.StatusUnauthorized},
		{"avatar update auth", NewProfileController(nil, nil).UpdateAvatar, http.StatusUnauthorized},
		{"settings auth", NewProfileController(nil, nil).GetMySettings, http.StatusUnauthorized},
		{"settings update auth", NewProfileController(nil, nil).UpdateMySettings, http.StatusUnauthorized},
		{"websocket auth", NewWebSocketController(nil).HandleWebSocket, http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, recorder := controllerContext(http.MethodPost, "/", `{}`)
			tt.call(context)
			if recorder.Code != tt.code {
				t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestAuthenticatedControllerBadRequests(t *testing.T) {
	context, recorder := controllerContext(http.MethodPost, "/", `{}`)
	context.Set("user_id", int64(1))
	NewChatController(nil).Chat(context)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("chat bad request = %d %s", recorder.Code, recorder.Body.String())
	}

	context, recorder = controllerContext(http.MethodGet, "/?before_id=bad", "")
	context.Set("user_id", int64(1))
	NewChatController(nil).GetChatHistory(context)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("history bad request = %d %s", recorder.Code, recorder.Body.String())
	}

	context, recorder = controllerContext(http.MethodPost, "/", "")
	context.Set("user_id", int64(1))
	NewChatController(nil).Transcribe(context)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("transcribe bad request = %d %s", recorder.Code, recorder.Body.String())
	}

	context, recorder = controllerContext(http.MethodPost, "/", `{}`)
	context.Set("user_id", int64(1))
	NewTranslationController(nil).Translate(context)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("translation bad request = %d %s", recorder.Code, recorder.Body.String())
	}

	context, recorder = controllerContext(http.MethodGet, "/", "")
	context.Params = gin.Params{{Key: "id", Value: "bad"}}
	NewVoiceFileController(nil).GetVoiceFileContent(context)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("voice file bad request = %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestUpdateSettingsRequiresEmbeddingDimension(t *testing.T) {
	context, recorder := controllerContext(http.MethodPut, "/", `{"embedding_model":"text-embedding"}`)
	context.Set("user_id", int64(1))

	NewProfileController(nil, nil).UpdateMySettings(context)

	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "Embedding dimension") {
		t.Fatalf("missing embedding dimension = %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestUpdateSettingsRejectsInvalidLLMType(t *testing.T) {
	context, recorder := controllerContext(http.MethodPut, "/", `{"llm_type":"claude"}`)
	context.Set("user_id", int64(1))

	NewProfileController(nil, nil).UpdateMySettings(context)

	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "LLM type") {
		t.Fatalf("invalid LLM type = %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestAvatarValidationPaths(t *testing.T) {
	controller := NewProfileController(nil, nil)
	for _, filename := range []string{"", "../secret", "a..b"} {
		context, recorder := controllerContext(http.MethodGet, "/", "")
		context.Params = gin.Params{{Key: "filename", Value: filename}}
		controller.GetAvatar(context)
		if recorder.Code != http.StatusBadRequest {
			t.Errorf("GetAvatar(%q) = %d %s", filename, recorder.Code, recorder.Body.String())
		}
	}
	context, recorder := controllerContext(http.MethodGet, "/", "")
	context.Params = gin.Params{{Key: "filename", Value: "missing.png"}}
	controller.GetAvatar(context)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("missing avatar = %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestModelProviderControllerWithMockHTTP(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"m","object":"model"}]}`))
	}))
	defer provider.Close()
	controller := NewModelProviderController(services.NewModelProviderService())
	context, recorder := controllerContext(http.MethodPost, "/", `{"api_base_url":"`+provider.URL+`","api_key":"key"}`)
	controller.GetCustomProviderModels(context)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"id":"m"`) {
		t.Fatalf("provider response = %d %s", recorder.Code, recorder.Body.String())
	}
	context, recorder = controllerContext(http.MethodPost, "/", `{}`)
	controller.GetCustomProviderModels(context)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("provider bad request = %d %s", recorder.Code, recorder.Body.String())
	}
}
