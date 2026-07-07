package routes

import (
	"context"
	"learnlang-api/config"
	"learnlang-api/database"
	"learnlang-api/services"
	"learnlang-api/utils"
	"learnlang-api/websocket"

	"github.com/gin-gonic/gin"
)

type Services struct {
	AuthService                *services.AuthService
	UserService                *services.UserService
	ModelProviderService       *services.ModelProviderService
	MessageService             *services.MessageService
	ConversationSummaryService *services.ConversationSummaryService
	UserSettingsService        *services.UserSettingsService
	ScheduledTaskService       *services.ScheduledTaskService
	VoiceFileService           *services.VoiceFileService
	ChatService                *services.ChatService
	Hub                        *websocket.Hub
}

func SetupRoutes(r *gin.Engine, cfg *config.Config) {
	tokenManager := utils.NewTokenManager(database.RedisClient)

	hub := websocket.NewHub()
	go hub.Run()

	authService := services.NewAuthService(cfg, tokenManager)
	userService := services.NewUserService()
	modelProviderService := services.NewModelProviderService()
	messageService := services.NewMessageService()
	conversationSummaryService := services.NewConversationSummaryService()
	userSettingsService := services.NewUserSettingsService()
	scheduledTaskService := services.NewScheduledTaskService()
	voiceFileService := services.NewVoiceFileService()

	chatService := services.NewChatService(messageService, conversationSummaryService, cfg.Milvus, userSettingsService, scheduledTaskService, voiceFileService, hub)

	scheduledTaskService.RegisterHandler("send_message", services.NewSendMessageHandler(messageService, chatService, userSettingsService, hub))
	scheduledTaskService.RegisterHandler("wait_message", services.NewWaitMessageHandler(chatService))

	go scheduledTaskService.StartScheduler(context.Background())

	svc := &Services{
		AuthService:                authService,
		UserService:                userService,
		ModelProviderService:       modelProviderService,
		MessageService:             messageService,
		ConversationSummaryService: conversationSummaryService,
		UserSettingsService:        userSettingsService,
		ScheduledTaskService:       scheduledTaskService,
		VoiceFileService:           voiceFileService,
		ChatService:                chatService,
		Hub:                        hub,
	}

	api := r.Group("/api")

	SetupUserRoutes(api, cfg, tokenManager, svc)
}
