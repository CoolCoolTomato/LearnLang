package services

import (
	"learnlang-api/config"
	"learnlang-api/database"
	"learnlang-api/models"
	"learnlang-api/utils"
	"time"
)

type AuthService struct {
	cfg          *config.Config
	tokenManager *utils.TokenManager
}

func NewAuthService(cfg *config.Config, tokenManager *utils.TokenManager) *AuthService {
	return &AuthService{
		cfg:          cfg,
		tokenManager: tokenManager,
	}
}

func (as *AuthService) Login(account, password string) (*models.User, string, error) {
	var user models.User
	if err := database.DB.Where("email = ? OR phone = ?", account, account).First(&user).Error; err != nil {
		return nil, "", err
	}

	if !utils.CheckPassword(password, user.PasswordHash) {
		return nil, "", utils.ErrInvalidCredentials
	}

	now := time.Now().UTC()
	user.LastActiveAt = &now
	database.DB.Save(&user)

	token, err := utils.GenerateToken(user.ID, user.Role, as.cfg.JWT.Secret)
	if err != nil {
		return nil, "", err
	}

	if err := as.tokenManager.SaveToken(user.ID, token, 24*time.Hour); err != nil {
		return nil, "", err
	}

	return &user, token, nil
}

func (as *AuthService) Logout(userID int64) error {
	return as.tokenManager.DeleteToken(userID)
}

func (as *AuthService) Register(email, phone *string, username, password string) (*models.User, string, error) {
	userService := NewUserService()
	user, err := userService.CreateUser(email, phone, username, password, "user")
	if err != nil {
		return nil, "", err
	}

	token, err := utils.GenerateToken(user.ID, user.Role, as.cfg.JWT.Secret)
	if err != nil {
		return nil, "", err
	}

	if err := as.tokenManager.SaveToken(user.ID, token, 24*time.Hour); err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (as *AuthService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return utils.ErrUserNotFound
	}

	if !utils.CheckPassword(oldPassword, user.PasswordHash) {
		return utils.ErrInvalidCredentials
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hashedPassword
	return database.DB.Save(&user).Error
}
