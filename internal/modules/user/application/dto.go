package application

import (
	"errors"
	"time"
)

var ErrUnauthorized = errors.New("unauthorized")

type RegisterRequest struct {
	Email       string `json:"email" example:"user@example.com" validate:"required,email"`
	Password    string `json:"password" example:"s3cret-pass" validate:"required,min=8"`
	DisplayName string `json:"display_name" example:"Aisyah" validate:"required,min=2,max=50"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	// ExpiresAt is when the access token stops being valid. Clients should
	// use the refresh token to get a new pair before then.
	ExpiresAt time.Time `json:"expires_at"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type MeResponse struct {
	ID            uint64    `json:"id"`
	Email         string    `json:"email"`
	DisplayName   string    `json:"display_name"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}
