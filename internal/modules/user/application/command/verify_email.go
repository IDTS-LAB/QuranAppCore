package command

import (
	"context"
	"errors"

	"github.com/fiqrikm18/quran-app/internal/core/security"
	"github.com/fiqrikm18/quran-app/internal/modules/user/application"
	"github.com/fiqrikm18/quran-app/internal/modules/user/domain"
)

type VerifyEmailCommand struct {
	Token string
}

type VerifyEmailCommandHandler struct {
	UserRepository application.IUserRepository
	TokenProvider  security.TokenGate
}

func NewVerifyEmailCommandHandler(
	userRepository application.IUserRepository,
	tokenProvider security.TokenGate,
) *VerifyEmailCommandHandler {
	return &VerifyEmailCommandHandler{
		UserRepository: userRepository,
		TokenProvider:  tokenProvider,
	}
}

func (command *VerifyEmailCommandHandler) Exec(ctx context.Context, cmd *VerifyEmailCommand) error {
	claims, err := command.TokenProvider.ValidateAccessToken(cmd.Token)
	if err != nil {
		if errors.Is(err, security.ErrTokenExpired) {
			return domain.ErrTokenExpired
		}
		return domain.ErrTokenInvalid
	}

	if claims.TokenType != "verify_email" {
		return domain.ErrTokenInvalid
	}

	user, err := command.UserRepository.FindByID(ctx, uint64(claims.UserID))
	if err != nil {
		return err
	}

	if user == nil {
		return domain.ErrTokenInvalid
	}

	if !user.EmailVerifiedAt.IsZero() {
		return nil
	}

	return command.UserRepository.MarkVerified(ctx, user.ID)
}
