package middleware

import (
	"learnlang-api/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(secret string, tokenManager *utils.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := utils.ParseToken(token, secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		valid, err := tokenManager.ValidateToken(claims.UserID, token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token validation failed"})
			c.Abort()
			return
		}

		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}
