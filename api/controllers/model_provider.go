package controllers

import (
	"learnlang-api/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ModelProviderController struct {
	modelProviderService *services.ModelProviderService
}

func NewModelProviderController(modelProviderService *services.ModelProviderService) *ModelProviderController {
	return &ModelProviderController{
		modelProviderService: modelProviderService,
	}
}

type CustomProviderRequest struct {
	APIBaseURL string `json:"api_base_url" binding:"required"`
	APIKey     string `json:"api_key" binding:"required"`
}

func (mpc *ModelProviderController) GetCustomProviderModels(c *gin.Context) {
	var req CustomProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	models, err := mpc.modelProviderService.GetCustomProviderModels(req.APIBaseURL, req.APIKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch models from provider"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": models})
}
