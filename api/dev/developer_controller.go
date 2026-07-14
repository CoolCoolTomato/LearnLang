package dev

import (
	"errors"
	"learnlang-api/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Controller struct {
	service              *services.DeveloperDataService
	archiveSearchService *services.DeveloperArchiveSearchService
}

func NewController(service *services.DeveloperDataService, archiveSearchService *services.DeveloperArchiveSearchService) *Controller {
	return &Controller{service: service, archiveSearchService: archiveSearchService}
}

func (cc *Controller) Dashboard(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	result, err := cc.service.Dashboard(userID.(int64))
	if err != nil {
		cc.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (cc *Controller) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	result, err := cc.service.List(c.Param("resource"), page, size)
	if err != nil {
		cc.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (cc *Controller) SearchConversationArchives(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	var request struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A JSON object with query is required"})
		return
	}
	if cc.archiveSearchService == nil {
		cc.writeError(c, errors.New("archive search service is not configured"))
		return
	}
	results, err := cc.archiveSearchService.Search(c.Request.Context(), userID.(int64), request.Query, request.Limit)
	if err != nil {
		cc.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"query": request.Query, "results": results})
}

func (cc *Controller) Get(c *gin.Context) {
	id, ok := cc.id(c)
	if !ok {
		return
	}
	result, err := cc.service.Get(c.Param("resource"), id)
	if err != nil {
		cc.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (cc *Controller) Create(c *gin.Context) {
	var values map[string]any
	if err := c.ShouldBindJSON(&values); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A JSON object is required"})
		return
	}
	result, err := cc.service.Create(c.Param("resource"), values)
	if err != nil {
		cc.writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (cc *Controller) Update(c *gin.Context) {
	id, ok := cc.id(c)
	if !ok {
		return
	}
	var values map[string]any
	if err := c.ShouldBindJSON(&values); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A JSON object is required"})
		return
	}
	result, err := cc.service.Update(c.Param("resource"), id, values)
	if err != nil {
		cc.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (cc *Controller) Delete(c *gin.Context) {
	id, ok := cc.id(c)
	if !ok {
		return
	}
	if err := cc.service.Delete(c.Param("resource"), id); err != nil {
		cc.writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (cc *Controller) DeleteMany(c *gin.Context) {
	var request struct {
		IDs []int64 `json:"ids"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A JSON object with ids is required"})
		return
	}
	deleted, err := cc.service.DeleteMany(c.Param("resource"), request.IDs)
	if err != nil {
		cc.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": deleted})
}

func (cc *Controller) id(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return 0, false
	}
	return id, true
}

func (cc *Controller) writeError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	if errors.Is(err, gorm.ErrRecordNotFound) {
		status = http.StatusNotFound
	} else if err.Error() == "ids are required" || err.Error() == "query is required" || err.Error() == "at least one writable field is required" || len(err.Error()) >= 17 && err.Error()[:17] == "unknown developer" {
		status = http.StatusBadRequest
	}
	c.JSON(status, gin.H{"error": err.Error()})
}
