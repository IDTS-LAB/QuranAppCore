package command

import (
	"context"

	"github.com/fiqrikm18/quran-app/internal/core/security"
	"github.com/fiqrikm18/quran-app/internal/modules/user/application"
	"github.com/fiqrikm18/quran-app/internal/modules/user/domain"
)

type LoginCommand struct {
	Email    string
	Password string
}

type LoginCommandHandler struct {
	UserRepository    application.IUserRepository
	SessionRepository application.IUserSessionRepository
	PasswordHashed    application.IPasswordHasher
	TokenProvider     security.TokenGate
}

func NewLoginCommandHandler(
	userRepository application.IUserRepository,
	sessionRepository application.IUserSessionRepository,
	passwordHashed application.IPasswordHasher,
	tokenProvider security.TokenGate,
) *LoginCommandHandler {
	return &LoginCommandHandler{
		UserRepository:    userRepository,
		SessionRepository: sessionRepository,
		PasswordHashed:    passwordHashed,
		TokenProvider:     tokenProvider,
	}
}

func (command *LoginCommandHandler) Exec(ctx context.Context, cmd *LoginCommand) (*application.LoginResponse, error) {
	user, err := command.UserRepository.FindByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := command.PasswordHashed.Compare(user.PasswordHash, cmd.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if user.EmailVerifiedAt.IsZero() {
		return nil, domain.ErrEmailNotVerified
	}

	tokenPair, err := command.TokenProvider.Issue(uint(user.ID), user.Email, "user")
	if err != nil {
		return nil, err
	}

	refreshJTIHash := hashJTI(tokenPair.RefreshToken)
	expiresAt := tokenPair.RefreshExpiresAt

	if err := command.SessionRepository.Save(ctx, user.ID, refreshJTIHash, expiresAt); err != nil {
		return nil, err
	}

	return &application.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.AccessExpiresAt,
	}, nil
}
