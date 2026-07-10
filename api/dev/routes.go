package dev

import (
	"learnlang-api/config"
	"learnlang-api/middleware"
	"learnlang-api/services"
	"learnlang-api/utils"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, cfg *config.Config, tokenManager *utils.TokenManager) {
	controller := NewController(services.NewDeveloperDataService())
	group := api.Group("/user/dev")
	group.Use(middleware.AuthMiddleware(cfg.JWT.Secret, tokenManager))
	group.Use(middleware.DeveloperMiddleware(cfg.User.Username))
	{
		group.GET("/dashboard", controller.Dashboard)
		group.GET("/:resource", controller.List)
		group.POST("/:resource", controller.Create)
		group.DELETE("/:resource", controller.DeleteMany)
		group.GET("/:resource/:id", controller.Get)
		group.PUT("/:resource/:id", controller.Update)
		group.DELETE("/:resource/:id", controller.Delete)
	}
}
