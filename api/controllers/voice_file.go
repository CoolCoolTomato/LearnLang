package controllers

import (
	"learnlang-api/services"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type VoiceFileController struct {
	voiceFileService *services.VoiceFileService
}

func NewVoiceFileController(voiceFileService *services.VoiceFileService) *VoiceFileController {
	return &VoiceFileController{voiceFileService: voiceFileService}
}

func (vfc *VoiceFileController) GetVoiceFileContent(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid voice file ID"})
		return
	}

	voiceFile, err := vfc.voiceFileService.GetVoiceFile(id)
	if err != nil {
		log.Printf("failed to fetch voice file %d: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Voice file not found"})
		return
	}

	if strings.HasPrefix(voiceFile.VoiceURL, "http://") || strings.HasPrefix(voiceFile.VoiceURL, "https://") {
		resp, err := http.Get(voiceFile.VoiceURL)
		if err != nil {
			log.Printf("failed to fetch remote voice file %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch remote file"})
			return
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Printf("failed to close remote voice response for file %d: %v", id, err)
			}
		}()

		c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
		return
	}

	c.File(voiceFile.VoiceURL)
}
