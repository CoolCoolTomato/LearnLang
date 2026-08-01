package memory

import (
	"context"
	"strings"
	"testing"
)

func TestNewStoreDefaultsAndEmbeddingValidation(t *testing.T) {
	store := NewStore(Config{}, nil)
	if store.cfg.Collection != "user_memory_summaries" {
		t.Fatalf("NewStore() config = %#v", store.cfg)
	}
	if err := validateEmbedding(nil, 2); err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("empty embedding error = %v", err)
	}
	if err := validateEmbedding([]float32{1}, 2); err == nil || !strings.Contains(err.Error(), "dimension mismatch") {
		t.Fatalf("dimension error = %v", err)
	}
	if err := validateEmbedding([]float32{1, 2}, 2); err != nil {
		t.Fatalf("valid embedding error = %v", err)
	}
	if err := validateEmbedding([]float32{1}, 0); err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("missing dimension error = %v", err)
	}
	if got := NewStore(Config{Collection: "custom"}, nil).collectionName(2); got != "custom_dim_2" {
		t.Fatalf("collectionName() = %q", got)
	}
}

func TestStoreMethodsRejectMissingMilvusClient(t *testing.T) {
	store := NewStore(Config{}, nil)
	if _, err := store.InsertArchive(context.Background(), 1, "summary", []int64{1}, []float32{1, 2}, 2); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("InsertArchive() error = %v", err)
	}
	if _, err := store.Search(context.Background(), 1, []float32{1, 2}, 2, 0); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("Search() error = %v", err)
	}
	if err := store.DeleteArchives(context.Background(), nil, 0); err != nil {
		t.Fatalf("DeleteArchives(empty) error = %v", err)
	}
	if err := store.DeleteArchives(context.Background(), []string{"id"}, 0); err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("DeleteArchives(missing dimension) error = %v", err)
	}
	if err := store.DeleteArchives(context.Background(), []string{"id"}, 2); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("DeleteArchives() error = %v", err)
	}
}
