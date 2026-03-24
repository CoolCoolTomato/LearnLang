package controllers

import (
	"errors"
	"learnlang-api/models"
	"learnlang-api/services"
	"learnlang-api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *services.AuthService
}

func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

type LoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	Username string  `json:"username" binding:"required"`
	Password string  `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	user, token, err := ac.authService.Login(req.Account, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token, User: user})
}

func (ac *AuthController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	if (req.Email == nil || *req.Email == "") && (req.Phone == nil || *req.Phone == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email or phone is required"})
		return
	}

	user, token, err := ac.authService.Register(req.Email, req.Phone, req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrEmailExists), errors.Is(err, utils.ErrPhoneExists):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register"})
		}
		return
	}

	c.JSON(http.StatusCreated, LoginResponse{Token: token, User: user})
}

func (ac *AuthController) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	err := ac.authService.ChangePassword(userID.(int64), req.OldPassword, req.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		case errors.Is(err, utils.ErrInvalidCredentials):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Old password is incorrect"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func (ac *AuthController) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := ac.authService.Logout(userID.(int64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
