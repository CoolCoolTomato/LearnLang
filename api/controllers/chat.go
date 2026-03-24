package controllers

import (
	"learnlang-api/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ChatController struct {
	chatService *services.ChatService
}

func NewChatController(chatService *services.ChatService) *ChatController {
	return &ChatController{chatService: chatService}
}

type ChatRequest struct {
	Message string `json:"message" binding:"required"`
}

func (cc *ChatController) VoiceChat(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	file, err := c.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Audio file required"})
		return
	}

	audioFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open audio file"})
		return
	}
	defer audioFile.Close()

	text, voiceFileID, err := cc.chatService.TranscribeAudio(c.Request.Context(), userID.(int64), audioFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to transcribe audio"})
		return
	}

	response, err := cc.chatService.ChatWithVoice(c.Request.Context(), userID.(int64), text, voiceFileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process chat"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (cc *ChatController) Chat(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	response, err := cc.chatService.Chat(c.Request.Context(), userID.(int64), req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process chat"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (cc *ChatController) GetChatHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	beforeIDStr := c.Query("before_id")
	var beforeID *int64
	if beforeIDStr != "" {
		id, err := strconv.ParseInt(beforeIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid before_id"})
			return
		}
		beforeID = &id
	}

	messages, err := cc.chatService.GetChatHistory(userID.(int64), beforeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chat history"})
		return
	}

	for i := range messages {
		if messages[i].VoiceFile != nil {
			messages[i].VoiceFile.VoiceURL = ""
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": messages})
}
