package auth

import "errors"

var (
	ErrDuplicateEmail      = errors.New("duplicate_email")
	ErrInvalidCredentials  = errors.New("invalid_credentials")
	ErrNotAuthenticated    = errors.New("not_authenticated")
	ErrInvalidEmail        = errors.New("invalid_email")
	ErrInvalidPassword     = errors.New("invalid_password")
	ErrUserNotFound        = errors.New("user_not_found")
)
