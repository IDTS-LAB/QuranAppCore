package domain

import (
	"net/mail"
	"strings"
	"time"
)

type User struct {
	ID              uint64
	Email           string
	PasswordHash    string
	DisplayName     string
	EmailVerifiedAt time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type UserDevice struct {
	ID         uint64
	UserID     uint64
	DeviceName string
	Platform   string
	AppVersion string
	PushToken  string
	LastSeenAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type UserSession struct {
	ID               uint64
	UserID           uint64
	RefreshTokenHash string
	ExpiredAt        time.Time
	RevokedAt        time.Time
	LastUsedAt       time.Time
	CreatedAt        time.Time
}

func NewUser(email, displayName string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, ErrEmailInvalid
	}

	return &User{
		Email:       email,
		DisplayName: displayName,
	}, nil
}
