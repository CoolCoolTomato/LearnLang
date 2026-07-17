package utils

import "errors"

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserNotFound        = errors.New("user not found")
	ErrEmailExists         = errors.New("email already exists")
	ErrPhoneExists         = errors.New("phone already exists")
	ErrRegistrationContact = errors.New("email or phone is required")
	ErrInvalidEmail        = errors.New("invalid email")
	ErrInvalidPhone        = errors.New("invalid phone")
	ErrProviderNotFound    = errors.New("provider not found")
	ErrProviderInactive    = errors.New("provider is not active")
	ErrAPIKeyNotConfigured = errors.New("API key is not configured")
)
