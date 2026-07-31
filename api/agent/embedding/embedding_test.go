package embedding

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestCreateEmbeddingWithMockHTTP(t *testing.T) {
	var body, authorization string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		data, _ := io.ReadAll(r.Body)
		body = string(data)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","data":[{"object":"embedding","embedding":[1.5,-2.25],"index":0}],"model":"text-embedding-3-small","usage":{"prompt_tokens":1,"total_tokens":1}}`))
	}))
	defer server.Close()
	got, err := Create(context.Background(), Config{FallbackKey: "fallback-key", FallbackURL: server.URL}, "hello")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !reflect.DeepEqual(got, []float32{1.5, -2.25}) {
		t.Fatalf("embedding = %#v", got)
	}
	if authorization != "Bearer fallback-key" || !strings.Contains(body, "text-embedding-3-small") || !strings.Contains(body, "hello") {
		t.Fatalf("request authorization/body = %q / %s", authorization, body)
	}
}

func TestCreateEmbeddingErrors(t *testing.T) {
	if _, err := Create(context.Background(), Config{}, "hello"); err == nil || !strings.Contains(err.Error(), "api key") {
		t.Fatalf("missing key error = %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","data":[]}`))
	}))
	defer server.Close()
	if _, err := Create(context.Background(), Config{APIKey: "key", APIBaseURL: server.URL}, "hello"); err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("empty response error = %v", err)
	}
}
