package routes

import (
	"learnlang-api/config"
	"learnlang-api/controllers"
	"learnlang-api/middleware"
	"learnlang-api/utils"

	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(api *gin.RouterGroup, cfg *config.Config, tokenManager *utils.TokenManager, services *Services) {
	authController := controllers.NewAuthController(services.AuthService)
	profileController := controllers.NewProfileController(services.UserService, services.UserSettingsService)
	chatController := controllers.NewChatController(services.ChatService)
	wsController := controllers.NewWebSocketController(services.Hub)
	voiceFileController := controllers.NewVoiceFileController(services.VoiceFileService)
	modelProviderController := controllers.NewModelProviderController(services.ModelProviderService)

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
}
