package memory

import (
	"context"
	"encoding/json"
	"fmt"
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
	Address    string
	Collection string
	Dimension  int
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
	cfg Config
}

func NewStore(cfg Config) *Store {
	if cfg.Collection == "" {
		cfg.Collection = "user_memory_summaries"
	}
	if cfg.Dimension <= 0 {
		cfg.Dimension = 1536
	}
	return &Store{cfg: cfg}
}

func (s *Store) InsertSummary(ctx context.Context, userID int64, summary, memoryType string, importanceScore float64, messageIDs []int64, embedding []float32) (string, error) {
	if err := s.validateEmbedding(embedding); err != nil {
		return "", err
	}

	cli, err := s.client(ctx)
	if err != nil {
		return "", err
	}
	defer cli.Close(ctx)

	if err := s.ensureCollection(ctx, cli); err != nil {
		return "", err
	}

	id := uuid.NewString()
	now := time.Now().UTC()
	messageIDsJSON, err := json.Marshal(messageIDs)
	if err != nil {
		return "", err
	}

	_, err = cli.Insert(ctx, milvusclient.NewColumnBasedInsertOption(s.cfg.Collection).
		WithColumns(
			column.NewColumnVarChar(fieldID, []string{id}),
			column.NewColumnInt64(fieldUserID, []int64{userID}),
			column.NewColumnVarChar(fieldSummary, []string{summary}),
			column.NewColumnVarChar(fieldMemoryType, []string{memoryType}),
			column.NewColumnDouble(fieldImportanceScore, []float64{importanceScore}),
			column.NewColumnVarChar(fieldMessageIDs, []string{string(messageIDsJSON)}),
			column.NewColumnInt64(fieldCreatedAt, []int64{now.Unix()}),
			column.NewColumnInt64(fieldUpdatedAt, []int64{now.Unix()}),
			column.NewColumnFloatVector(fieldEmbedding, s.cfg.Dimension, [][]float32{embedding}),
		))
	if err != nil {
		return "", err
	}

	flushTask, err := cli.Flush(ctx, milvusclient.NewFlushOption(s.cfg.Collection))
	if err == nil {
		_ = flushTask.Await(ctx)
	}

	return id, nil
}

func (s *Store) Search(ctx context.Context, userID int64, embedding []float32, limit int) ([]Summary, error) {
	if limit <= 0 {
		limit = 3
	}
	if err := s.validateEmbedding(embedding); err != nil {
		return nil, err
	}

	cli, err := s.client(ctx)
	if err != nil {
		return nil, err
	}
	defer cli.Close(ctx)

	if err := s.ensureCollection(ctx, cli); err != nil {
		return nil, err
	}

	results, err := cli.Search(ctx, milvusclient.NewSearchOption(
		s.cfg.Collection,
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

func (s *Store) client(ctx context.Context) (*milvusclient.Client, error) {
	if s.cfg.Address == "" {
		return nil, fmt.Errorf("milvus address is required")
	}

	return milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address: s.cfg.Address,
	})
}

func (s *Store) ensureCollection(ctx context.Context, cli *milvusclient.Client) error {
	has, err := cli.HasCollection(ctx, milvusclient.NewHasCollectionOption(s.cfg.Collection))
	if err != nil {
		return err
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
			WithField(entity.NewField().WithName(fieldEmbedding).WithDataType(entity.FieldTypeFloatVector).WithDim(s.cfg.Dimension))

		err = cli.CreateCollection(ctx, milvusclient.NewCreateCollectionOption(s.cfg.Collection, schema).
			WithIndexOptions(
				milvusclient.NewCreateIndexOption(s.cfg.Collection, fieldEmbedding, index.NewAutoIndex(entity.COSINE)),
			))
		if err != nil {
			return err
		}
	}

	loadTask, err := cli.LoadCollection(ctx, milvusclient.NewLoadCollectionOption(s.cfg.Collection))
	if err != nil {
		return err
	}
	return loadTask.Await(ctx)
}

func (s *Store) validateEmbedding(embedding []float32) error {
	if len(embedding) == 0 {
		return fmt.Errorf("embedding is empty")
	}
	if len(embedding) != s.cfg.Dimension {
		return fmt.Errorf("embedding dimension mismatch: got %d, want %d", len(embedding), s.cfg.Dimension)
	}
	return nil
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
		_ = json.Unmarshal([]byte(messageIDsJSON), &linkedMessageIDs)

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
