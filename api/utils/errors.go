package utils

import "errors"

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserNotFound        = errors.New("user not found")
	ErrEmailExists         = errors.New("email already exists")
	ErrPhoneExists         = errors.New("phone already exists")
	ErrProviderNotFound    = errors.New("provider not found")
	ErrProviderInactive    = errors.New("provider is not active")
	ErrAPIKeyNotConfigured = errors.New("API key is not configured")
)
