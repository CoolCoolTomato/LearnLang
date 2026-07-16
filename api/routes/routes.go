package routes

import (
	"context"
	"learnlang-api/agent"
	"learnlang-api/agent/archive"
	"learnlang-api/config"
	"learnlang-api/database"
	"learnlang-api/dev"
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
	UserProfileSummaryService  *services.UserProfileSummaryService
	UserSettingsService        *services.UserSettingsService
	ConversationArchiveService *services.ConversationArchiveService
	ScheduledTaskService       *services.ScheduledTaskService
	VoiceFileService           *services.VoiceFileService
	TranslationService         *services.TranslationService
	VocabularyService          *services.VocabularyService
	ChatService                *agent.ChatService
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
	userProfileSummaryService := services.NewUserProfileSummaryService()
	userSettingsService := services.NewUserSettingsService()
	conversationArchiveService := services.NewConversationArchiveService()
	scheduledTaskService := services.NewScheduledTaskService()
	voiceFileService := services.NewVoiceFileService()
	translationService := services.NewTranslationService(userSettingsService)
	vocabularyService := services.NewVocabularyService(userSettingsService)

	chatRuntimeService := services.NewChatRuntimeService(messageService, userSettingsService, userProfileSummaryService, scheduledTaskService, voiceFileService, hub)
	memoryStore := agent.NewMemoryStore(cfg.Milvus, database.MilvusClient)
	conversationArchiver := archive.NewService(conversationArchiveService, userSettingsService, memoryStore)
	chatService := agent.NewChatService(chatRuntimeService, memoryStore, conversationArchiver, vocabularyService)

	scheduledTaskService.RegisterHandler("send_message", services.NewSendMessageHandler(chatRuntimeService))

	go scheduledTaskService.StartScheduler(context.Background())

	svc := &Services{
		AuthService:                authService,
		UserService:                userService,
		ModelProviderService:       modelProviderService,
		MessageService:             messageService,
		UserProfileSummaryService:  userProfileSummaryService,
		UserSettingsService:        userSettingsService,
		ConversationArchiveService: conversationArchiveService,
		ScheduledTaskService:       scheduledTaskService,
		VoiceFileService:           voiceFileService,
		TranslationService:         translationService,
		VocabularyService:          vocabularyService,
		ChatService:                chatService,
		Hub:                        hub,
	}

	api := r.Group("/api")

	SetupUserRoutes(api, cfg, tokenManager, svc)
	dev.SetupRoutes(api, cfg, tokenManager, memoryStore, services.NewDeveloperArchiveSearchService(memoryStore, userSettingsService))
}
