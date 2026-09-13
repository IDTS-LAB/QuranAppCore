package query

import (
	"context"

	"github.com/fiqrikm18/quran-app/internal/modules/user/application"
	"github.com/fiqrikm18/quran-app/internal/modules/user/domain"
)

type MeQuery struct {
	UserID uint64
}

type MeQueryHandler struct {
	UserRepository application.IUserRepository
}

func NewMeQueryHandler(userRepository application.IUserRepository) *MeQueryHandler {
	return &MeQueryHandler{
		UserRepository: userRepository,
	}
}

func (q *MeQueryHandler) Exec(ctx context.Context, cmd *MeQuery) (*domain.User, error) {
	user, err := q.UserRepository.FindByID(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	return user, nil
}
