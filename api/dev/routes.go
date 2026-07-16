package dev

import (
	"learnlang-api/agent/memory"
	"learnlang-api/config"
	"learnlang-api/middleware"
	"learnlang-api/services"
	"learnlang-api/utils"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, cfg *config.Config, tokenManager *utils.TokenManager, memoryStore *memory.Store, archiveSearchService *services.DeveloperArchiveSearchService) {
	controller := NewController(services.NewDeveloperDataService(memoryStore), archiveSearchService)
	group := api.Group("/user/dev")
	group.Use(middleware.AuthMiddleware(cfg.JWT.Secret, tokenManager))
	group.Use(middleware.DeveloperMiddleware(cfg.User.Username))
	{
		group.GET("/dashboard", controller.Dashboard)
		group.POST("/conversation-archives/search", controller.SearchConversationArchives)
		group.GET("/:resource", controller.List)
		group.POST("/:resource", controller.Create)
		group.DELETE("/:resource", controller.DeleteMany)
		group.GET("/:resource/:id", controller.Get)
		group.PUT("/:resource/:id", controller.Update)
		group.DELETE("/:resource/:id", controller.Delete)
	}
}
