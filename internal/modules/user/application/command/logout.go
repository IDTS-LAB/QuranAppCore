package command

import (
	"context"

	"github.com/fiqrikm18/quran-app/internal/core/security"
	"github.com/fiqrikm18/quran-app/internal/modules/user/application"
)

type LogoutCommand struct {
	RefreshToken string
	UserID       uint64
}

type LogoutCommandHandler struct {
	SessionRepository application.IUserSessionRepository
	TokenProvider     security.TokenGate
}

func NewLogoutCommandHandler(
	sessionRepository application.IUserSessionRepository,
	tokenProvider security.TokenGate,
) *LogoutCommandHandler {
	return &LogoutCommandHandler{
		SessionRepository: sessionRepository,
		TokenProvider:     tokenProvider,
	}
}

func (command *LogoutCommandHandler) Exec(ctx context.Context, cmd *LogoutCommand) error {
	claims, err := command.TokenProvider.ValidateRefreshToken(cmd.RefreshToken)
	if err != nil {
		return err
	}

	if uint64(claims.UserID) != cmd.UserID {
		return application.ErrUnauthorized
	}

	refreshJTIHash := hashJTI(cmd.RefreshToken)

	return command.SessionRepository.Revoke(ctx, refreshJTIHash)
}
