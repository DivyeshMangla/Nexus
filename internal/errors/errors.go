package errors

import "errors"

var (
	ErrUnauthorized    = errors.New("unauthorized")
	ErrInvalidRequest  = errors.New("invalid request")
	ErrUserNotFound    = errors.New("user not found")
	ErrChannelNotFound = errors.New("channel not found")
	ErrDMCreationFailed = errors.New("failed to create DM")
)