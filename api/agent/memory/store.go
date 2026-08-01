package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/milvus-io/milvus/client/v2/column"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

const (
	fieldID              = "id"
	fieldUserID          = "user_id"
	fieldSummary         = "summary"
	fieldMemoryType      = "memory_type"
	fieldImportanceScore = "importance_score"
	fieldMessageIDs      = "message_ids"
	fieldCreatedAt       = "created_at"
	fieldUpdatedAt       = "updated_at"
	fieldEmbedding       = "embedding"
)

type Config struct {
	Collection string
}

type Summary struct {
	ID              string
	UserID          int64
	Summary         string
	MemoryType      string
	ImportanceScore float64
	MessageIDs      []int64
	Score           float32
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Store struct {
	cfg       Config
	client    *milvusclient.Client
	ready     map[int]string
	readyLock sync.Mutex
}

func NewStore(cfg Config, client *milvusclient.Client) *Store {
	if cfg.Collection == "" {
		cfg.Collection = "user_memory_summaries"
	}
	return &Store{cfg: cfg, client: client, ready: make(map[int]string)}
}

func (s *Store) InsertArchive(ctx context.Context, userID int64, summary string, messageIDs []int64, embedding []float32, dimension int) (string, error) {
	if err := validateEmbedding(embedding, dimension); err != nil {
		return "", err
	}

	collection, err := s.ensureCollection(ctx, dimension)
	if err != nil {
		return "", err
	}

	id := uuid.NewString()
	now := time.Now().UTC()
	messageIDsJSON, err := json.Marshal(messageIDs)
	if err != nil {
		return "", err
	}

	_, err = s.client.Insert(ctx, milvusclient.NewColumnBasedInsertOption(collection).
		WithColumns(
			column.NewColumnVarChar(fieldID, []string{id}),
			column.NewColumnInt64(fieldUserID, []int64{userID}),
			column.NewColumnVarChar(fieldSummary, []string{summary}),
			column.NewColumnVarChar(fieldMemoryType, []string{"conversation_archive"}),
			column.NewColumnDouble(fieldImportanceScore, []float64{1}),
			column.NewColumnVarChar(fieldMessageIDs, []string{string(messageIDsJSON)}),
			column.NewColumnInt64(fieldCreatedAt, []int64{now.Unix()}),
			column.NewColumnInt64(fieldUpdatedAt, []int64{now.Unix()}),
			column.NewColumnFloatVector(fieldEmbedding, dimension, [][]float32{embedding}),
		))
	if err != nil {
		return "", err
	}

	flushTask, err := s.client.Flush(ctx, milvusclient.NewFlushOption(collection))
	if err == nil {
		if err := flushTask.Await(ctx); err != nil {
			log.Printf("failed to await Milvus archive insert flush: %v", err)
		}
	} else {
		log.Printf("failed to flush Milvus archive insert: %v", err)
	}

	return id, nil
}

func (s *Store) DeleteArchives(ctx context.Context, embeddingIDs []string, dimension int) error {
	if len(embeddingIDs) == 0 {
		return nil
	}
	if err := validateDimension(dimension); err != nil {
		return err
	}
	collection, err := s.ensureCollection(ctx, dimension)
	if err != nil {
		return err
	}

	if _, err := s.client.Delete(ctx, milvusclient.NewDeleteOption(collection).WithStringIDs(fieldID, embeddingIDs)); err != nil {
		return err
	}
	flushTask, err := s.client.Flush(ctx, milvusclient.NewFlushOption(collection))
	if err != nil {
		return err
	}
	return flushTask.Await(ctx)
}

func (s *Store) Search(ctx context.Context, userID int64, embedding []float32, dimension, limit int) ([]Summary, error) {
	if limit <= 0 {
		limit = 3
	}
	if err := validateEmbedding(embedding, dimension); err != nil {
		return nil, err
	}

	collection, err := s.ensureCollection(ctx, dimension)
	if err != nil {
		return nil, err
	}

	results, err := s.client.Search(ctx, milvusclient.NewSearchOption(
		collection,
		limit,
		[]entity.Vector{entity.FloatVector(embedding)},
	).WithFilter(fmt.Sprintf("%s == %d", fieldUserID, userID)).
		WithOutputFields(fieldID, fieldUserID, fieldSummary, fieldMemoryType, fieldImportanceScore, fieldMessageIDs, fieldCreatedAt, fieldUpdatedAt))
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return []Summary{}, nil
	}

	return summariesFromResult(results[0])
}

func (s *Store) ensureCollection(ctx context.Context, dimension int) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("milvus client is not initialized")
	}

	s.readyLock.Lock()
	defer s.readyLock.Unlock()

	if collection := s.ready[dimension]; collection != "" {
		return collection, nil
	}
	collection := s.collectionName(dimension)
	has, err := s.client.HasCollection(ctx, milvusclient.NewHasCollectionOption(collection))
	if err != nil {
		return "", fmt.Errorf("check milvus collection %s: %w", collection, err)
	}
	if !has {
		schema := entity.NewSchema().
			WithField(entity.NewField().WithName(fieldID).WithDataType(entity.FieldTypeVarChar).WithMaxLength(64).WithIsPrimaryKey(true)).
			WithField(entity.NewField().WithName(fieldUserID).WithDataType(entity.FieldTypeInt64)).
			WithField(entity.NewField().WithName(fieldSummary).WithDataType(entity.FieldTypeVarChar).WithMaxLength(8192)).
			WithField(entity.NewField().WithName(fieldMemoryType).WithDataType(entity.FieldTypeVarChar).WithMaxLength(64)).
			WithField(entity.NewField().WithName(fieldImportanceScore).WithDataType(entity.FieldTypeDouble)).
			WithField(entity.NewField().WithName(fieldMessageIDs).WithDataType(entity.FieldTypeVarChar).WithMaxLength(4096)).
			WithField(entity.NewField().WithName(fieldCreatedAt).WithDataType(entity.FieldTypeInt64)).
			WithField(entity.NewField().WithName(fieldUpdatedAt).WithDataType(entity.FieldTypeInt64)).
			WithField(entity.NewField().WithName(fieldEmbedding).WithDataType(entity.FieldTypeFloatVector).WithDim(int64(dimension)))

		err = s.client.CreateCollection(ctx, milvusclient.NewCreateCollectionOption(collection, schema).
			WithIndexOptions(
				milvusclient.NewCreateIndexOption(collection, fieldEmbedding, index.NewAutoIndex(entity.COSINE)),
			))
		if err != nil {
			has, checkErr := s.client.HasCollection(ctx, milvusclient.NewHasCollectionOption(collection))
			if checkErr != nil {
				return "", fmt.Errorf("create milvus collection %s: %w; recheck collection existence: %v", collection, err, checkErr)
			}
			if !has {
				return "", fmt.Errorf("create milvus collection %s: %w", collection, err)
			}
		}
	}

	loadTask, err := s.client.LoadCollection(ctx, milvusclient.NewLoadCollectionOption(collection))
	if err != nil {
		return "", fmt.Errorf("load milvus collection %s: %w", collection, err)
	}
	if err := loadTask.Await(ctx); err != nil {
		return "", fmt.Errorf("await milvus collection %s load: %w", collection, err)
	}

	s.ready[dimension] = collection
	return collection, nil
}

func validateEmbedding(embedding []float32, dimension int) error {
	if err := validateDimension(dimension); err != nil {
		return err
	}
	if len(embedding) == 0 {
		return fmt.Errorf("embedding is empty")
	}
	if len(embedding) != dimension {
		return fmt.Errorf("embedding dimension mismatch: got %d, want %d", len(embedding), dimension)
	}
	return nil
}

func validateDimension(dimension int) error {
	if dimension <= 0 {
		return fmt.Errorf("embedding dimension is required")
	}
	return nil
}

func (s *Store) collectionName(dimension int) string {
	return fmt.Sprintf("%s_dim_%d", s.cfg.Collection, dimension)
}

func summariesFromResult(result milvusclient.ResultSet) ([]Summary, error) {
	ids := result.GetColumn(fieldID)
	userIDs := result.GetColumn(fieldUserID)
	summaries := result.GetColumn(fieldSummary)
	memoryTypes := result.GetColumn(fieldMemoryType)
	importanceScores := result.GetColumn(fieldImportanceScore)
	messageIDs := result.GetColumn(fieldMessageIDs)
	createdAts := result.GetColumn(fieldCreatedAt)
	updatedAts := result.GetColumn(fieldUpdatedAt)

	count := len(result.Scores)
	items := make([]Summary, 0, count)
	for i := 0; i < count; i++ {
		id, err := ids.GetAsString(i)
		if err != nil {
			return nil, err
		}
		userID, err := userIDs.GetAsInt64(i)
		if err != nil {
			return nil, err
		}
		summary, err := summaries.GetAsString(i)
		if err != nil {
			return nil, err
		}
		memoryType, err := memoryTypes.GetAsString(i)
		if err != nil {
			return nil, err
		}
		importanceScore, err := importanceScores.GetAsDouble(i)
		if err != nil {
			return nil, err
		}
		messageIDsJSON, err := messageIDs.GetAsString(i)
		if err != nil {
			return nil, err
		}
		createdAt, err := createdAts.GetAsInt64(i)
		if err != nil {
			return nil, err
		}
		updatedAt, err := updatedAts.GetAsInt64(i)
		if err != nil {
			return nil, err
		}

		var linkedMessageIDs []int64
		if err := json.Unmarshal([]byte(messageIDsJSON), &linkedMessageIDs); err != nil {
			log.Printf("failed to unmarshal memory message IDs for summary %s: %v", id, err)
		}

		items = append(items, Summary{
			ID:              id,
			UserID:          userID,
			Summary:         summary,
			MemoryType:      memoryType,
			ImportanceScore: importanceScore,
			MessageIDs:      linkedMessageIDs,
			Score:           result.Scores[i],
			CreatedAt:       time.Unix(createdAt, 0).UTC(),
			UpdatedAt:       time.Unix(updatedAt, 0).UTC(),
		})
	}

	return items, nil
}
