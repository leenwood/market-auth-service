package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrTokenNotFound      = errors.New("token not found")
	ErrUserDeleted        = errors.New("user account deleted")
)
