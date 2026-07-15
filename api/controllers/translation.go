package controllers

import (
	"errors"
	"learnlang-api/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TranslationController struct {
	translationService *services.TranslationService
}

func NewTranslationController(translationService *services.TranslationService) *TranslationController {
	return &TranslationController{translationService: translationService}
}

type TranslationRequest struct {
	Text string `json:"text" binding:"required"`
}

type TranslationResponse struct {
	Translation string `json:"translation"`
}

func (tc *TranslationController) Translate(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req TranslationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	translation, err := tc.translationService.Translate(c.Request.Context(), userID.(int64), req.Text)
	if err != nil {
		if errors.Is(err, services.ErrTranslationTextRequired) || errors.Is(err, services.ErrTranslationTextTooLong) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to translate text"})
		return
	}

	c.JSON(http.StatusOK, TranslationResponse{Translation: translation})
}
