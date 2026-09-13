package application

import (
	"context"
	"time"

	"github.com/fiqrikm18/quran-app/internal/modules/user/domain"
)

type IUserRepository interface {
	Create(ctx context.Context, user *domain.User, passwordHash string) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id uint64) (*domain.User, error)
	MarkVerified(ctx context.Context, id uint64) error
}

type IUserSessionRepository interface {
	Save(ctx context.Context, userID uint64, refreshJTIHash string, expiresAt time.Time) error
	FindByJTIHash(ctx context.Context, hash string) (*domain.UserSession, error)
	Revoke(ctx context.Context, jtiHash string) error
}

type IPasswordHasher interface {
	Hash(string) (string, error)
	Compare(hash, plain string) error
}
