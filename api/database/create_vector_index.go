package database

import (
	"log"
)

func CreateVectorIndex() error {
	if err := DB.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
		log.Printf("Warning: Failed to create vector extension: %v", err)
		return err
	}

	sql := `
		CREATE INDEX IF NOT EXISTS idx_user_memories_embedding
		ON user_memories
		USING ivfflat (embedding vector_cosine_ops)
		WITH (lists = 100)
	`

	if err := DB.Exec(sql).Error; err != nil {
		log.Printf("Warning: Failed to create vector index: %v", err)
		return err
	}

	log.Println("Vector index created successfully")
	return nil
}
