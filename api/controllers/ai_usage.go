package controllers

import (
	"errors"
	"learnlang-api/models"
	"learnlang-api/services"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AIUsageController struct {
	usage *services.AIUsageService
}

func NewAIUsageController(usage *services.AIUsageService) *AIUsageController {
	return &AIUsageController{usage: usage}
}

type userAIUsageEvent struct {
	Operation string  `json:"operation"`
	Model     string  `json:"model"`
	Usage     float64 `json:"usage"`
	Unit      string  `json:"unit"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

type userAIUsagePage struct {
	Items    []userAIUsageEvent `json:"items"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

func (c *AIUsageController) List(ctx *gin.Context) {
	query, ok := userAIUsageQuery(ctx)
	if !ok {
		return
	}
	page, err := c.usage.List(ctx.Request.Context(), query)
	if err != nil {
		c.writeError(ctx, err)
		return
	}
	items := make([]userAIUsageEvent, 0, len(page.Items))
	for _, event := range page.Items {
		items = append(items, visibleAIUsageEvent(event))
	}
	ctx.JSON(http.StatusOK, userAIUsagePage{Items: items, Total: page.Total, Page: page.Page, PageSize: page.PageSize})
}

func (c *AIUsageController) Summary(ctx *gin.Context) {
	query, ok := userAIUsageQuery(ctx)
	if !ok {
		return
	}
	result, err := c.usage.Summary(ctx.Request.Context(), query)
	if err != nil {
		c.writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func userAIUsageQuery(ctx *gin.Context) (services.AIUsageQuery, bool) {
	value, exists := ctx.Get("user_id")
	userID, valid := value.(int64)
	if !exists || !valid || userID <= 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return services.AIUsageQuery{}, false
	}
	query := services.AIUsageQuery{UserID: userID, Operation: ctx.Query("operation")}
	for key, target := range map[string]*int{"page": &query.Page, "page_size": &query.PageSize} {
		raw := strings.TrimSpace(ctx.Query(key))
		if raw == "" {
			continue
		}
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid usage query"})
			return services.AIUsageQuery{}, false
		}
		*target = parsed
	}
	return query, true
}

func visibleAIUsageEvent(event models.AIUsageEvent) userAIUsageEvent {
	return userAIUsageEvent{
		Operation: event.Operation, Model: event.Model, Usage: event.Usage,
		Unit: event.Unit, Status: event.Status, CreatedAt: event.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func (c *AIUsageController) writeError(ctx *gin.Context, err error) {
	if errors.Is(err, services.ErrAIUsageInvalidInput) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid usage query"})
		return
	}
	log.Printf("failed to query AI usage: %v", err)
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Usage query failed"})
}
