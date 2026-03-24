package database

import (
	"learnlang-api/config"
	"learnlang-api/models"
	"learnlang-api/utils"
	"log"
)

func InitUser(cfg *config.Config) error {
	if cfg.User.Username == "" || cfg.User.Password == "" {
		log.Println("user account not configured, skipping user initialization")
		return nil
	}

	var count int64
	DB.Model(&models.User{}).Count(&count)

	if count > 0 {
		log.Println("Users already exist, skipping user initialization")
		return nil
	}

	hashedPassword, err := utils.HashPassword(cfg.User.Password)
	if err != nil {
		return err
	}

	var email *string
	if cfg.User.Email != "" {
		email = &cfg.User.Email
	}

	admin := models.User{
		Username:     cfg.User.Username,
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         "user",
	}

	if err := DB.Create(&admin).Error; err != nil {
		return err
	}

	settings := models.UserSettings{
		UserID: admin.ID,
	}

	if err := DB.Create(&settings).Error; err != nil {
		return err
	}

	log.Printf("User user created: %s", cfg.User.Username)
	return nil
}
