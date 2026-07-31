package services

import (
	"context"
	"encoding/json"
	"errors"
	"learnlang-api/models"
	"learnlang-api/utils"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestWelcomeMessageHelpers(t *testing.T) {
	for input, want := range map[string]string{"zh-CN": "zh", " JA_jp ": "ja", "unknown": "en", "": "en"} {
		if got := normalizeGreetingLanguage(input); got != want {
			t.Errorf("normalizeGreetingLanguage(%q) = %q, want %q", input, got, want)
		}
	}
	original, translation := welcomeMessage("en-US", "zh-CN")
	if original == "" || translation == "" || original == translation {
		t.Fatalf("welcomeMessage(en, zh) = %q, %q", original, translation)
	}
	original, translation = welcomeMessage("fr", "fr-FR")
	if original == "" || translation != "" {
		t.Fatalf("welcomeMessage(fr, fr) = %q, %q", original, translation)
	}
}

func TestValidateArchiveSegments(t *testing.T) {
	candidates := []models.Message{{ID: 1}, {ID: 2}, {ID: 3}}
	valid := []ArchiveSegmentInput{{Summary: "one", MessageIDs: []int64{1, 2}}, {Summary: "two", MessageIDs: []int64{3}}}
	if err := validateArchiveSegments(candidates, valid); err != nil {
		t.Fatalf("valid archive rejected: %v", err)
	}
	tests := []struct {
		name     string
		segments []ArchiveSegmentInput
	}{
		{"blank summary", []ArchiveSegmentInput{{MessageIDs: []int64{1}}}},
		{"no messages", []ArchiveSegmentInput{{Summary: "x"}}},
		{"unknown", []ArchiveSegmentInput{{Summary: "x", MessageIDs: []int64{4}}}},
		{"skip", []ArchiveSegmentInput{{Summary: "x", MessageIDs: []int64{2}}}},
		{"duplicate", []ArchiveSegmentInput{{Summary: "x", MessageIDs: []int64{1, 1}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateArchiveSegments(candidates, tt.segments); err == nil {
				t.Fatal("invalid archive segments accepted")
			}
		})
	}
}

func TestDeveloperHelpers(t *testing.T) {
	for _, tt := range []struct {
		page, size, wantPage, wantSize int
	}{{0, 0, 1, 20}, {2, 101, 2, 100}, {3, 10, 3, 10}} {
		page, size := normalizeDeveloperPage(tt.page, tt.size)
		if page != tt.wantPage || size != tt.wantSize {
			t.Errorf("normalizeDeveloperPage(%d, %d) = %d, %d", tt.page, tt.size, page, size)
		}
	}
	resources := []string{DeveloperResourceMessages, DeveloperResourceScheduledTasks, DeveloperResourceUserProfileSummaries, DeveloperResourceConversationArchives, DeveloperResourceUserSettings, DeveloperResourceUsers, DeveloperResourceVoiceFiles}
	for _, resource := range resources {
		model, err := developerModel(resource)
		if err != nil || model == nil {
			t.Errorf("developerModel(%q) = %v, %v", resource, model, err)
		}
		list, err := developerListTarget(resource)
		if err != nil || list == nil {
			t.Errorf("developerListTarget(%q) = %v, %v", resource, list, err)
		}
	}
	if _, err := developerModel("bad"); err == nil {
		t.Fatal("developerModel() accepted unknown resource")
	}
	if _, err := developerListTarget("bad"); err == nil {
		t.Fatal("developerListTarget() accepted unknown resource")
	}
	filtered := allowedDeveloperValues(DeveloperResourceMessages, map[string]any{" role ": "user", "id": 99, "bad": true})
	if !reflect.DeepEqual(filtered, map[string]any{"role": "user"}) {
		t.Fatalf("allowedDeveloperValues() = %#v", filtered)
	}
	if developerModelID(&models.Message{ID: 9}) != 9 || developerModelID(struct{}{}) != 0 {
		t.Fatal("developerModelID() returned unexpected ID")
	}
}

type fakeArchiveVectors struct {
	ids []string
	err error
}

func (f *fakeArchiveVectors) DeleteArchives(_ context.Context, ids []string) error {
	f.ids = append([]string(nil), ids...)
	return f.err
}

func TestDeleteArchiveVectors(t *testing.T) {
	fake := &fakeArchiveVectors{}
	service := NewDeveloperDataService(fake)
	if err := service.deleteArchiveVectors(context.Background(), nil); err != nil || fake.ids != nil {
		t.Fatalf("empty delete = %#v, %v", fake.ids, err)
	}
	if err := service.deleteArchiveVectors(context.Background(), []string{"a", "b"}); err != nil || !reflect.DeepEqual(fake.ids, []string{"a", "b"}) {
		t.Fatalf("delete = %#v, %v", fake.ids, err)
	}
	fake.err = errors.New("mock failure")
	if err := service.deleteArchiveVectors(context.Background(), []string{"c"}); !errors.Is(err, fake.err) {
		t.Fatalf("delete error = %v", err)
	}
}

func TestRegistrationHelpers(t *testing.T) {
	value := "  User@Example.COM "
	if got := normalizeRegistrationContact(&value, true); got == nil || *got != "user@example.com" {
		t.Fatalf("normalizeRegistrationContact() = %v", got)
	}
	empty := " "
	if normalizeRegistrationContact(&empty, false) != nil || normalizeRegistrationContact(nil, false) != nil {
		t.Fatal("empty/nil registration contact was not nil")
	}
	for _, valid := range []string{"user@example.com", "a+b@example.co.uk"} {
		if !validRegistrationEmail(valid) {
			t.Errorf("validRegistrationEmail(%q) = false", valid)
		}
	}
	for _, invalid := range []string{"Display Name <user@example.com>", "bad", strings.Repeat("a", 256)} {
		if validRegistrationEmail(invalid) {
			t.Errorf("validRegistrationEmail(%q) = true", invalid)
		}
	}
}

func TestModelProviderServiceUsesHTTPProvider(t *testing.T) {
	var authorization string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		if r.URL.Path != "/models" {
			http.Error(w, "wrong path", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"model-a","object":"model","created":12,"owned_by":"owner"}]}`))
	}))
	defer server.Close()

	service := NewModelProviderService()
	models, err := service.GetCustomProviderModels(server.URL, "test-key")
	if err != nil {
		t.Fatalf("GetCustomProviderModels() error = %v", err)
	}
	if len(models) != 1 || models[0].ID != "model-a" || models[0].Created != 12 || models[0].OwnedBy != "owner" {
		t.Fatalf("models = %#v", models)
	}
	if authorization != "Bearer test-key" {
		t.Fatalf("Authorization = %q", authorization)
	}
	if _, err := service.GetCustomProviderModels(server.URL, ""); !errors.Is(err, utils.ErrAPIKeyNotConfigured) {
		t.Fatalf("empty key error = %v", err)
	}
}

func TestDetectMP3DurationErrors(t *testing.T) {
	if _, err := detectMP3DurationSeconds("missing-file.mp3"); err == nil {
		t.Fatal("missing MP3 file did not return an error")
	}
	file, err := os.CreateTemp(t.TempDir(), "empty-*.mp3")
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := detectMP3DurationSeconds(file.Name()); err == nil {
		t.Fatal("empty MP3 file did not return an error")
	}
}

func TestVocabularyRandomHelpers(t *testing.T) {
	if normalizeVocabularyLanguage("  EN_us ") != "en" {
		t.Fatalf("normalizeVocabularyLanguage() = %q", normalizeVocabularyLanguage("  EN_us "))
	}
	items := []models.Vocabulary{{ID: 3}, {ID: 8}}
	if got := vocabularyIDs(items); !reflect.DeepEqual(got, []int64{3, 8}) {
		t.Fatalf("vocabularyIDs() = %#v", got)
	}
	if sql := vocabularyLanguageMatchSQL("target_language"); !strings.Contains(sql, "target_language") || !strings.Contains(sql, "LOWER") {
		t.Fatalf("vocabularyLanguageMatchSQL() = %q", sql)
	}
}

func TestMarshalSendMessageArgs(t *testing.T) {
	data, err := json.Marshal(SendMessageArgs{UserID: 1, Message: "hello", Translation: "你好"})
	if err != nil || !strings.Contains(string(data), `"user_id":1`) {
		t.Fatalf("marshal SendMessageArgs = %s, %v", data, err)
	}
}
