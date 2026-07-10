package middleware

import (
	"learnlang-api/database"
	"learnlang-api/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// DeveloperMiddleware permits dedicated developer accounts and the configured initial account.
func DeveloperMiddleware(initialUsername string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role == "developer" || role == "admin" {
			c.Next()
			return
		}

		userID, ok := c.Get("user_id")
		if ok && strings.TrimSpace(initialUsername) != "" {
			var user models.User
			if err := database.DB.Select("username").First(&user, userID).Error; err == nil && user.Username == initialUsername {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Developer access required"})
		c.Abort()
	}
}
