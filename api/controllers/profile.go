package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"learnlang-api/services"
	"learnlang-api/utils"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type ProfileController struct {
	userService         *services.UserService
	userSettingsService *services.UserSettingsService
}

func NewProfileController(userService *services.UserService, userSettingsService *services.UserSettingsService) *ProfileController {
	return &ProfileController{
		userService:         userService,
		userSettingsService: userSettingsService,
	}
}

type UpdateProfileRequest struct {
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	Username string  `json:"username"`
}

type UpdateAvatarRequest struct {
	Filename string `json:"filename" binding:"required"`
}

type UpdateMySettingsRequest struct {
	APIBaseURL          string `json:"api_base_url"`
	APIKey              string `json:"api_key"`
	Model               string `json:"model"`
	EmbeddingAPIBaseURL string `json:"embedding_api_base_url"`
	EmbeddingAPIKey     string `json:"embedding_api_key"`
	EmbeddingModel      string `json:"embedding_model"`
	STTAPIBaseURL       string `json:"stt_api_base_url"`
	STTAPIKey           string `json:"stt_api_key"`
	STTModel            string `json:"stt_model"`
	TTSAPIBaseURL       string `json:"tts_api_base_url"`
	TTSAPIKey           string `json:"tts_api_key"`
	TTSModel            string `json:"tts_model"`
	TTSVoice            string `json:"tts_voice"`
	NativeLanguage      string `json:"native_language"`
	TargetLanguage      string `json:"target_language"`
	Timezone            string `json:"timezone"`
}

func (pc *ProfileController) GetMyProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, err := pc.userService.GetUser(userID.(int64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (pc *ProfileController) UpdateMyProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	user, err := pc.userService.UpdateProfile(userID.(int64), req.Email, req.Phone, req.Username)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		case errors.Is(err, utils.ErrEmailExists), errors.Is(err, utils.ErrPhoneExists):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

func (pc *ProfileController) UploadAvatar(c *gin.Context) {
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Avatar file required"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported avatar format"})
		return
	}

	if err := os.MkdirAll("uploads/avatar", 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare avatar directory"})
		return
	}

	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate avatar filename"})
		return
	}

	filename := hex.EncodeToString(random) + ext
	savePath := filepath.Join("uploads/avatar", filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save avatar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"filename": filename})
}

func (pc *ProfileController) GetAvatar(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" || filename != filepath.Base(filename) || strings.Contains(filename, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid avatar filename"})
		return
	}

	path := filepath.Join("uploads/avatar", filename)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Avatar not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access avatar"})
		return
	}

	c.File(path)
}

func (pc *ProfileController) UpdateAvatar(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req UpdateAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	filename := req.Filename
	if filename != filepath.Base(filename) || strings.Contains(filename, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid avatar filename"})
		return
	}

	path := filepath.Join("uploads/avatar", filename)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Avatar file not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access avatar file"})
		return
	}

	user, err := pc.userService.UpdateAvatar(userID.(int64), filename)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update avatar"})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

func (pc *ProfileController) GetMySettings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	settings, err := pc.userSettingsService.GetUserSettings(userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (pc *ProfileController) UpdateMySettings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req UpdateMySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	updates := make(map[string]interface{})
	if req.APIBaseURL != "" {
		updates["api_base_url"] = req.APIBaseURL
	}
	if req.APIKey != "" {
		updates["api_key"] = req.APIKey
	}
	if req.Model != "" {
		updates["model"] = req.Model
	}
	if req.EmbeddingAPIBaseURL != "" {
		updates["embedding_api_base_url"] = req.EmbeddingAPIBaseURL
	}
	if req.EmbeddingAPIKey != "" {
		updates["embedding_api_key"] = req.EmbeddingAPIKey
	}
	if req.EmbeddingModel != "" {
		updates["embedding_model"] = req.EmbeddingModel
	}
	if req.STTAPIBaseURL != "" {
		updates["stt_api_base_url"] = req.STTAPIBaseURL
	}
	if req.STTAPIKey != "" {
		updates["stt_api_key"] = req.STTAPIKey
	}
	if req.STTModel != "" {
		updates["stt_model"] = req.STTModel
	}
	if req.TTSAPIBaseURL != "" {
		updates["tts_api_base_url"] = req.TTSAPIBaseURL
	}
	if req.TTSAPIKey != "" {
		updates["tts_api_key"] = req.TTSAPIKey
	}
	if req.TTSModel != "" {
		updates["tts_model"] = req.TTSModel
	}
	if req.TTSVoice != "" {
		updates["tts_voice"] = req.TTSVoice
	}
	if req.NativeLanguage != "" {
		updates["native_language"] = req.NativeLanguage
	}
	if req.TargetLanguage != "" {
		updates["target_language"] = req.TargetLanguage
	}
	if req.Timezone != "" {
		updates["timezone"] = req.Timezone
	}

	settings, err := pc.userSettingsService.UpdateUserSettings(userID.(int64), updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}
