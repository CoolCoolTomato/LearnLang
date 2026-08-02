package routes

import (
	"learnlang-api/config"
	"learnlang-api/controllers"
	"learnlang-api/middleware"
	"learnlang-api/utils"

	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(api *gin.RouterGroup, cfg *config.Config, tokenManager *utils.TokenManager, svc *Services) {
	authController := controllers.NewAuthController(svc.AuthService)
	profileController := controllers.NewProfileController(svc.UserService, svc.UserSettingsService)
	chatController := controllers.NewChatController(svc.ChatService)
	translationController := controllers.NewTranslationController(svc.TranslationService)
	wsController := controllers.NewWebSocketController(svc.Hub)
	voiceFileController := controllers.NewVoiceFileController(svc.VoiceFileService)
	modelProviderController := controllers.NewModelProviderController(svc.ModelProviderService)
	vocabularyController := controllers.NewVocabularyController(svc.VocabularyService)
	aiUsageController := controllers.NewAIUsageController(svc.AIUsageService)

	userGroup := api.Group("/user")

	auth := userGroup.Group("/auth")
	{
		auth.POST("/login", authController.Login)
		auth.POST("/register", authController.Register)
		auth.POST("/logout", middleware.AuthMiddleware(cfg.JWT.Secret, tokenManager), authController.Logout)
		auth.POST("/change-password", middleware.AuthMiddleware(cfg.JWT.Secret, tokenManager), authController.ChangePassword)
	}

	chat := userGroup.Group("/chat")
	chat.Use(middleware.AuthMiddleware(cfg.JWT.Secret, tokenManager))
	{
		chat.POST("", chatController.Chat)
		chat.POST("/translate", translationController.Translate)
		chat.POST("/voice", chatController.VoiceChat)
		chat.GET("/history", chatController.GetChatHistory)
	}

	profile := userGroup.Group("/profile")
	profile.Use(middleware.AuthMiddleware(cfg.JWT.Secret, tokenManager))
	{
		profile.GET("", profileController.GetMyProfile)
		profile.PUT("", profileController.UpdateMyProfile)
		profile.POST("/avatar/upload", profileController.UploadAvatar)
		profile.PUT("/avatar", profileController.UpdateAvatar)
		profile.GET("/settings", profileController.GetMySettings)
		profile.PUT("/settings", profileController.UpdateMySettings)
	}

	userGroup.GET("/profile/avatar/:filename", profileController.GetAvatar)

	userSettings := userGroup.Group("/user-settings")
	userSettings.Use(middleware.AuthMiddleware(cfg.JWT.Secret, tokenManager))
	{
		userSettings.POST("/custom-provider-models", modelProviderController.GetCustomProviderModels)
	}

	ws := userGroup.Group("/ws")
	ws.Use(middleware.WebSocketAuthMiddleware(cfg.JWT.Secret, tokenManager))
	{
		ws.GET("/chat", wsController.HandleWebSocket)
	}

	voiceFiles := userGroup.Group("/voice-files")
	voiceFiles.Use(middleware.AuthMiddleware(cfg.JWT.Secret, tokenManager))
	{
		voiceFiles.GET("/:id/content", voiceFileController.GetVoiceFileContent)
	}

	vocabularies := userGroup.Group("/vocabularies")
	vocabularies.Use(middleware.AuthMiddleware(cfg.JWT.Secret, tokenManager))
	{
		vocabularies.GET("", vocabularyController.List)
		vocabularies.POST("", vocabularyController.Create)
		vocabularies.POST("/lookup", vocabularyController.LookupMessage)
		vocabularies.PUT("/:id", vocabularyController.Update)
		vocabularies.DELETE("/:id", vocabularyController.Delete)
		vocabularies.GET("/:id/entries", vocabularyController.GetEntries)
		vocabularies.DELETE("/:id/entries", vocabularyController.ClearEntries)
		vocabularies.POST("/:id/import", vocabularyController.Import)
	}

	usage := userGroup.Group("/usage")
	usage.Use(middleware.AuthMiddleware(cfg.JWT.Secret, tokenManager))
	{
		usage.GET("/events", aiUsageController.List)
		usage.GET("/summary", aiUsageController.Summary)
	}
}
