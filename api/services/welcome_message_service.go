package services

import (
	"context"
	"learnlang-api/database"
	"learnlang-api/models"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type localizedGreeting struct {
	Text string
}

var welcomeGreetings = map[string]localizedGreeting{
	"en": {Text: "Hello! I'm glad you're here. Let's start learning together."},
	"zh": {Text: "你好！很高兴见到你，让我们一起开始学习吧。"},
	"ja": {Text: "こんにちは！お会いできてうれしいです。一緒に学習を始めましょう。"},
	"ko": {Text: "안녕하세요! 만나서 반가워요. 함께 공부를 시작해요."},
	"es": {Text: "¡Hola! Me alegra que estés aquí. Empecemos a aprender juntos."},
	"fr": {Text: "Bonjour ! Je suis ravi de vous accueillir. Commençons à apprendre ensemble."},
}

func normalizeGreetingLanguage(language string) string {
	language = strings.ToLower(strings.TrimSpace(language))
	if separator := strings.IndexAny(language, "-_"); separator >= 0 {
		language = language[:separator]
	}
	if _, exists := welcomeGreetings[language]; !exists {
		return "en"
	}
	return language
}

func welcomeMessage(targetLanguage, nativeLanguage string) (string, string) {
	target := normalizeGreetingLanguage(targetLanguage)
	native := normalizeGreetingLanguage(nativeLanguage)
	original := welcomeGreetings[target].Text
	if target == native {
		return original, ""
	}
	return original, welcomeGreetings[native].Text
}

func (crs *ChatRuntimeService) ensureWelcomeMessage(ctx context.Context, userID int64) error {
	return database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&user, userID).Error; err != nil {
			return err
		}

		var messageCount int64
		if err := tx.Model(&models.Message{}).Where("user_id = ?", userID).Count(&messageCount).Error; err != nil {
			return err
		}
		if messageCount > 0 {
			return nil
		}

		var settings models.UserSettings
		if err := tx.Where("user_id = ?", userID).First(&settings).Error; err != nil {
			return err
		}
		if strings.TrimSpace(settings.Timezone) == "" || strings.TrimSpace(settings.NativeLanguage) == "" || strings.TrimSpace(settings.TargetLanguage) == "" {
			return nil
		}

		original, translation := welcomeMessage(settings.TargetLanguage, settings.NativeLanguage)
		return tx.Create(&models.Message{
			UserID: userID, Role: "assistant", TextContent: original,
			Translation: translation, InputType: "text",
		}).Error
	})
}
