package database

import (
	"learnlang-api/models"
	"log"
)

func Migrate() error {
	err := DB.AutoMigrate(
		&models.User{},
		&models.UserSettings{},
		&models.Message{},
		&models.ConversationArchive{},
		&models.UserProfileSummary{},
		&models.ScheduledTask{},
		&models.VoiceFile{},
		&models.Vocabulary{},
		&models.VocabularyEntry{},
		&models.VocabularyPronunciation{},
		&models.VocabularyMeaning{},
		&models.VocabularyExample{},
		&models.VocabularyEntryRelation{},
		&models.VocabularyAgentSelection{},
		&models.AIUsageEvent{},
	)
	if err != nil {
		return err
	}
	if err := migrateLegacyVocabularySchema(); err != nil {
		return err
	}
	if err := DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_conversation_archives_message_ids_gin
		ON conversation_archives USING GIN (message_ids jsonb_path_ops)
	`).Error; err != nil {
		return err
	}
	if err := DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_vocabularies_one_default_per_user
		ON vocabularies (user_id)
		WHERE is_default = TRUE
	`).Error; err != nil {
		return err
	}
	log.Println("Database migration completed")
	return nil
}
