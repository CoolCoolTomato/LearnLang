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
		&models.ConversationSummary{},
		&models.UserMemory{},
		&models.ScheduledTask{},
		&models.VoiceFile{},
	)
	if err != nil {
		return err
	}
	log.Println("Database migration completed")
	return nil
}
