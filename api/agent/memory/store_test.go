package memory

import (
	"context"
	"strings"
	"testing"
)

func TestNewStoreDefaultsAndEmbeddingValidation(t *testing.T) {
	store := NewStore(Config{}, nil)
	if store.cfg.Collection != "user_memory_summaries" || store.cfg.Dimension != 1536 {
		t.Fatalf("NewStore() config = %#v", store.cfg)
	}
	if err := store.validateEmbedding(nil); err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("empty embedding error = %v", err)
	}
	if err := store.validateEmbedding([]float32{1}); err == nil || !strings.Contains(err.Error(), "dimension mismatch") {
		t.Fatalf("dimension error = %v", err)
	}
	custom := NewStore(Config{Collection: "custom", Dimension: 2}, nil)
	if err := custom.validateEmbedding([]float32{1, 2}); err != nil {
		t.Fatalf("valid embedding error = %v", err)
	}
}

func TestStoreMethodsRejectMissingMilvusClient(t *testing.T) {
	store := NewStore(Config{Dimension: 2}, nil)
	if _, err := store.InsertArchive(context.Background(), 1, "summary", []int64{1}, []float32{1, 2}); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("InsertArchive() error = %v", err)
	}
	if _, err := store.Search(context.Background(), 1, []float32{1, 2}, 0); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("Search() error = %v", err)
	}
	if err := store.DeleteArchives(context.Background(), nil); err != nil {
		t.Fatalf("DeleteArchives(empty) error = %v", err)
	}
	if err := store.DeleteArchives(context.Background(), []string{"id"}); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("DeleteArchives() error = %v", err)
	}
}
