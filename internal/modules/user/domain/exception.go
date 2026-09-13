package domain

import "errors"

var (
	ErrEmailEmpty       = errors.New("email cannot be empty")
	ErrEmailInvalid     = errors.New("email format is invalid")
	ErrEmailNotVerified = errors.New("email not verified")
	ErrTokenExpired     = errors.New("token expired")
	ErrTokenInvalid     = errors.New("token invalid")
	ErrUserNotFound     = errors.New("user not found")

	// ErrInvalidCredentials covers both unknown emails and wrong passwords.
	// One sentinel on purpose: callers must not reveal which one failed.
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyUsed   = errors.New("email address already used")
)
