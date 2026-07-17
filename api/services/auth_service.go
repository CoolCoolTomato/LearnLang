package services

import (
	"learnlang-api/config"
	"learnlang-api/database"
	"learnlang-api/models"
	"learnlang-api/utils"
	"net/mail"
	"strings"
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
	account = strings.TrimSpace(account)
	var user models.User
	if err := database.DB.Where("LOWER(email) = LOWER(?) OR phone = ?", account, account).First(&user).Error; err != nil {
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

func (as *AuthService) Register(email, phone *string, password string) (*models.User, string, error) {
	email = normalizeRegistrationContact(email, true)
	phone = normalizeRegistrationContact(phone, false)
	if email == nil && phone == nil {
		return nil, "", utils.ErrRegistrationContact
	}
	if email != nil && !validRegistrationEmail(*email) {
		return nil, "", utils.ErrInvalidEmail
	}
	if phone != nil && len([]rune(*phone)) > 32 {
		return nil, "", utils.ErrInvalidPhone
	}

	username := ""
	if phone != nil {
		username = *phone
	} else {
		username = strings.SplitN(*email, "@", 2)[0]
	}
	usernameRunes := []rune(username)
	if len(usernameRunes) > 64 {
		username = string(usernameRunes[:64])
	}

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

func normalizeRegistrationContact(value *string, lowercase bool) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil
	}
	if lowercase {
		normalized = strings.ToLower(normalized)
	}
	return &normalized
}

func validRegistrationEmail(email string) bool {
	if len(email) > 255 {
		return false
	}
	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email
}

func (as *AuthService) ChangePassword(userID int64, newPassword string) error {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return utils.ErrUserNotFound
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hashedPassword
	return database.DB.Save(&user).Error
}
