package command

import (
	"context"
	"errors"
	"time"

	"github.com/fiqrikm18/quran-app/internal/core/security"
	"github.com/fiqrikm18/quran-app/internal/modules/user/application"
	"github.com/fiqrikm18/quran-app/internal/modules/user/domain"
)

type RefreshTokenCommand struct {
	RefreshToken string
}

type RefreshTokenCommandHandler struct {
	UserRepository    application.IUserRepository
	SessionRepository application.IUserSessionRepository
	TokenProvider     security.TokenGate
}

func NewRefreshTokenCommandHandler(
	userRepository application.IUserRepository,
	sessionRepository application.IUserSessionRepository,
	tokenProvider security.TokenGate,
) *RefreshTokenCommandHandler {
	return &RefreshTokenCommandHandler{
		UserRepository:    userRepository,
		SessionRepository: sessionRepository,
		TokenProvider:     tokenProvider,
	}
}

func (command *RefreshTokenCommandHandler) Exec(ctx context.Context, cmd *RefreshTokenCommand) (*application.RefreshTokenResponse, error) {
	claims, err := command.TokenProvider.ValidateRefreshToken(cmd.RefreshToken)
	if err != nil {
		if errors.Is(err, security.ErrTokenExpired) {
			return nil, domain.ErrTokenExpired
		}
		return nil, domain.ErrTokenInvalid
	}

	refreshJTIHash := hashJTI(cmd.RefreshToken)
	session, err := command.SessionRepository.FindByJTIHash(ctx, refreshJTIHash)
	if err != nil {
		return nil, domain.ErrTokenInvalid
	}

	if session.RevokedAt.IsZero() && session.ExpiredAt.Before(time.Now()) {
		return nil, domain.ErrTokenExpired
	}

	user, err := command.UserRepository.FindByID(ctx, uint64(claims.UserID))
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, domain.ErrTokenInvalid
	}

	if user.EmailVerifiedAt.IsZero() {
		return nil, domain.ErrEmailNotVerified
	}

	if err := command.SessionRepository.Revoke(ctx, refreshJTIHash); err != nil {
		return nil, err
	}

	tokenPair, err := command.TokenProvider.Issue(uint(user.ID), user.Email, "user")
	if err != nil {
		return nil, err
	}

	newRefreshJTIHash := hashJTI(tokenPair.RefreshToken)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	if err := command.SessionRepository.Save(ctx, user.ID, newRefreshJTIHash, expiresAt); err != nil {
		return nil, err
	}

	return &application.RefreshTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}
