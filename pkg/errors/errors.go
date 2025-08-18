package errors

import "errors"

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUserExists       = errors.New("user already exists")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrTokenGeneration  = errors.New("failed to generate token")
	ErrTokenInvalid     = errors.New("invalid token")
	ErrPasswordHashing  = errors.New("failed to hash password")
)