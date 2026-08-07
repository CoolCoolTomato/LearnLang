package controllers

import (
	"errors"
	"learnlang-api/agent"
	"learnlang-api/services"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ChatController struct {
	chatService *agent.ChatService
}

func NewChatController(chatService *agent.ChatService) *ChatController {
	return &ChatController{chatService: chatService}
}

type ChatRequest struct {
	Message     string `json:"message" binding:"required"`
	VoiceFileID *int64 `json:"voice_file_id"`
}

type TranscriptionResponse struct {
	Text        string `json:"text"`
	VoiceFileID int64  `json:"voice_file_id"`
}

func (cc *ChatController) Transcribe(c *gin.Context) {
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
		log.Printf("failed to open uploaded chat audio file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open audio file"})
		return
	}
	defer func() {
		if err := audioFile.Close(); err != nil {
			log.Printf("failed to close uploaded chat audio file: %v", err)
		}
	}()

	text, voiceFileID, err := cc.chatService.TranscribeAudio(c.Request.Context(), userID.(int64), audioFile)
	if err != nil {
		log.Printf("failed to transcribe chat audio for user %d: %v", userID.(int64), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to transcribe audio"})
		return
	}
	if voiceFileID == nil {
		log.Printf("transcription succeeded without a voice file for user %d", userID.(int64))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save audio file"})
		return
	}

	c.JSON(http.StatusOK, TranscriptionResponse{Text: text, VoiceFileID: *voiceFileID})
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

	response, err := cc.chatService.Chat(c.Request.Context(), userID.(int64), req.Message, req.VoiceFileID)
	if err != nil {
		log.Printf("failed to process chat for user %d: %v", userID.(int64), err)
		if errors.Is(err, services.ErrUserVoiceFileInvalid) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user voice file"})
			return
		}
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

	messages, err := cc.chatService.GetChatHistory(c.Request.Context(), userID.(int64), beforeID)
	if err != nil {
		log.Printf("failed to fetch chat history for user %d: %v", userID.(int64), err)
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
