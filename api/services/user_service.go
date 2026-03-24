package services

import (
	"learnlang-api/database"
	"learnlang-api/models"
	"learnlang-api/utils"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

type UserListResult struct {
	Total int64
	Users []models.User
}

func (us *UserService) CreateUser(email, phone *string, username, password, role string) (*models.User, error) {
	var count int64
	if email != nil && *email != "" {
		database.DB.Model(&models.User{}).Where("email = ?", email).Count(&count)
		if count > 0 {
			return nil, utils.ErrEmailExists
		}
	}
	if phone != nil && *phone != "" {
		database.DB.Model(&models.User{}).Where("phone = ?", phone).Count(&count)
		if count > 0 {
			return nil, utils.ErrPhoneExists
		}
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	if role == "" {
		role = "user"
	}

	user := models.User{
		Email:        email,
		Phone:        phone,
		Username:     username,
		PasswordHash: hashedPassword,
		Role:         role,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	settingsService := NewUserSettingsService()
	if err := settingsService.CreateUserSettings(user.ID); err != nil {
		return nil, err
	}

	summary := models.ConversationSummary{
		UserID:  user.ID,
		Summary: "",
	}
	database.DB.Create(&summary)

	return &user, nil
}

func (us *UserService) GetUser(userID int64) (*models.User, error) {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return nil, utils.ErrUserNotFound
	}
	return &user, nil
}

func (us *UserService) ListUsers(page, size int, email, phone, username string) (*UserListResult, error) {
	query := database.DB.Model(&models.User{})

	if email != "" {
		query = query.Where("email LIKE ?", "%"+email+"%")
	}
	if phone != "" {
		query = query.Where("phone LIKE ?", "%"+phone+"%")
	}
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	var total int64
	query.Count(&total)

	var users []models.User
	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Order("id DESC").Find(&users).Error; err != nil {
		return nil, err
	}

	return &UserListResult{Total: total, Users: users}, nil
}

func (us *UserService) UpdateUser(userID int64, email, phone *string, username, password, role string) (*models.User, error) {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return nil, utils.ErrUserNotFound
	}

	if email != nil && *email != "" {
		var count int64
		database.DB.Model(&models.User{}).Where("email = ? AND id != ?", email, userID).Count(&count)
		if count > 0 {
			return nil, utils.ErrEmailExists
		}
		user.Email = email
	}

	if phone != nil && *phone != "" {
		var count int64
		database.DB.Model(&models.User{}).Where("phone = ? AND id != ?", phone, userID).Count(&count)
		if count > 0 {
			return nil, utils.ErrPhoneExists
		}
		user.Phone = phone
	}

	if username != "" {
		user.Username = username
	}

	if password != "" {
		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = hashedPassword
	}

	if role != "" {
		user.Role = role
	}

	if err := database.DB.Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (us *UserService) DeleteUser(userID int64) error {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return utils.ErrUserNotFound
	}

	return database.DB.Delete(&user).Error
}

func (us *UserService) UpdateProfile(userID int64, email, phone *string, username string) (*models.User, error) {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return nil, utils.ErrUserNotFound
	}

	if email != nil && *email != "" {
		var count int64
		database.DB.Model(&models.User{}).Where("email = ? AND id != ?", email, userID).Count(&count)
		if count > 0 {
			return nil, utils.ErrEmailExists
		}
		user.Email = email
	}

	if phone != nil && *phone != "" {
		var count int64
		database.DB.Model(&models.User{}).Where("phone = ? AND id != ?", phone, userID).Count(&count)
		if count > 0 {
			return nil, utils.ErrPhoneExists
		}
		user.Phone = phone
	}

	if username != "" {
		user.Username = username
	}

	if err := database.DB.Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (us *UserService) UpdateAvatar(userID int64, avatarFilename string) (*models.User, error) {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return nil, utils.ErrUserNotFound
	}

	user.AvatarURL = avatarFilename
	if err := database.DB.Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
