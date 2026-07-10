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
	service *services.DeveloperDataService
}

func NewController(service *services.DeveloperDataService) *Controller {
	return &Controller{service: service}
}

func (cc *Controller) Dashboard(c *gin.Context) {
	result, err := cc.service.Dashboard()
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
	} else if err.Error() == "ids are required" || err.Error() == "at least one writable field is required" || len(err.Error()) >= 17 && err.Error()[:17] == "unknown developer" {
		status = http.StatusBadRequest
	}
	c.JSON(status, gin.H{"error": err.Error()})
}
