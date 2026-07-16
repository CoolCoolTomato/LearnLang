package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"learnlang-api/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type VocabularyController struct {
	vocabularyService *services.VocabularyService
}

func NewVocabularyController(vocabularyService *services.VocabularyService) *VocabularyController {
	return &VocabularyController{vocabularyService: vocabularyService}
}

func (vc *VocabularyController) List(c *gin.Context) {
	result, err := vc.vocabularyService.List(c.Request.Context(), c.GetInt64("user_id"))
	if err != nil {
		vc.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (vc *VocabularyController) Create(c *gin.Context) {
	var input services.VocabularyCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A vocabulary object is required"})
		return
	}
	result, err := vc.vocabularyService.Create(c.Request.Context(), c.GetInt64("user_id"), input)
	if err != nil {
		vc.writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (vc *VocabularyController) Update(c *gin.Context) {
	vocabularyID, ok := vocabularyIDParam(c)
	if !ok {
		return
	}
	var input services.VocabularyUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A vocabulary object is required"})
		return
	}
	result, err := vc.vocabularyService.Update(c.Request.Context(), c.GetInt64("user_id"), vocabularyID, input)
	if err != nil {
		vc.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (vc *VocabularyController) Delete(c *gin.Context) {
	vocabularyID, ok := vocabularyIDParam(c)
	if !ok {
		return
	}
	if err := vc.vocabularyService.Delete(c.Request.Context(), c.GetInt64("user_id"), vocabularyID); err != nil {
		vc.writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (vc *VocabularyController) Import(c *gin.Context) {
	vocabularyID, ok := vocabularyIDParam(c)
	if !ok {
		return
	}
	input, err := decodeVocabularyImport(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A vocabulary import object is required"})
		return
	}

	result, err := vc.vocabularyService.Import(c.Request.Context(), c.GetInt64("user_id"), vocabularyID, *input)
	if err != nil {
		vc.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func decodeVocabularyImport(c *gin.Context) (*services.VocabularyImportInput, error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 20<<20)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(bytes.TrimSpace(body)) == 0 {
		return nil, errors.New("empty import body")
	}

	var input services.VocabularyImportInput
	if bytes.HasPrefix(bytes.TrimSpace(body), []byte("[")) {
		if err := json.Unmarshal(body, &input.Entries); err != nil {
			return nil, err
		}
		return &input, nil
	}
	if err := json.Unmarshal(body, &input); err != nil {
		return nil, err
	}
	if len(input.Entries) > 0 {
		return &input, nil
	}

	var entry services.VocabularyImportEntry
	if err := json.Unmarshal(body, &entry); err != nil || entry.Word == "" {
		return nil, errors.New("invalid import body")
	}
	input.Entries = []services.VocabularyImportEntry{entry}
	return &input, nil
}

func (vc *VocabularyController) GetEntries(c *gin.Context) {
	vocabularyID, ok := vocabularyIDParam(c)
	if !ok {
		return
	}
	page, err := positiveIntQuery(c, "page", 1)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page must be a positive integer"})
		return
	}
	pageSize, err := positiveIntQuery(c, "page_size", 20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page_size must be a positive integer"})
		return
	}

	result, err := vc.vocabularyService.Get(c.Request.Context(), c.GetInt64("user_id"), vocabularyID, page, pageSize)
	if err != nil {
		vc.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (vc *VocabularyController) ClearEntries(c *gin.Context) {
	vocabularyID, ok := vocabularyIDParam(c)
	if !ok {
		return
	}
	deleted, err := vc.vocabularyService.Clear(c.Request.Context(), c.GetInt64("user_id"), vocabularyID)
	if err != nil {
		vc.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": deleted})
}

func (vc *VocabularyController) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrVocabularyNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrVocabularyNameConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrVocabularyInvalidImport),
		errors.Is(err, services.ErrVocabularyInvalidInput),
		errors.Is(err, services.ErrVocabularyLanguageRequired),
		errors.Is(err, services.ErrVocabularyDefaultRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Vocabulary operation failed"})
	}
}

func vocabularyIDParam(c *gin.Context) (int64, bool) {
	vocabularyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || vocabularyID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vocabulary ID"})
		return 0, false
	}
	return vocabularyID, true
}

func positiveIntQuery(c *gin.Context, name string, fallback int) (int, error) {
	value := c.Query(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, errors.New("invalid positive integer")
	}
	return parsed, nil
}
