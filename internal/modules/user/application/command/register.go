package command

import (
	"context"

	"github.com/fiqrikm18/quran-app/internal/modules/user/application"
	"github.com/fiqrikm18/quran-app/internal/modules/user/domain"
)

type RegisterCommand struct {
	Email       string
	Password    string
	DisplayName string
}

type RegisterCommandHandler struct {
	UserRepository application.IUserRepository
	PasswordHashed application.IPasswordHasher
}

func NewRegisterCommandHandler(
	UserRepository application.IUserRepository,
	PasswordHashed application.IPasswordHasher,
) *RegisterCommandHandler {
	return &RegisterCommandHandler{
		UserRepository: UserRepository,
		PasswordHashed: PasswordHashed,
	}
}

func (command *RegisterCommandHandler) Exec(ctx context.Context, cmd *RegisterCommand) error {
	user, err := domain.NewUser(cmd.Email, cmd.DisplayName)
	if err != nil {
		return err
	}

	userExists, err := command.UserRepository.FindByEmail(ctx, user.Email)
	if err != nil {
		return err
	}

	if userExists != nil {
		return domain.ErrEmailAlreadyUsed
	}

	hash, err := command.PasswordHashed.Hash(cmd.Password)
	if err != nil {
		return err
	}

	if err := command.UserRepository.Create(ctx, user, hash); err != nil {
		return err
	}

	return nil
}
