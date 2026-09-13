package database

import (
	"context"
	"time"

	"github.com/fiqrikm18/quran-app/internal/modules/user/domain"
	"github.com/fiqrikm18/quran-app/internal/modules/user/infrastructure/models"
	"gorm.io/gorm"
)

type UserSessionRepository struct {
	DB *gorm.DB
}

func NewUserSessionRepository(db *gorm.DB) *UserSessionRepository {
	return &UserSessionRepository{
		DB: db,
	}
}

func (repo *UserSessionRepository) Save(ctx context.Context, userID uint64, refreshJTIHash string, expiresAt time.Time) error {
	session := &models.UserSessionModel{
		UserID:           userID,
		RefreshTokenHash: refreshJTIHash,
		ExpiresAt:        expiresAt,
		LastUsedAt:       time.Now(),
	}

	return repo.DB.WithContext(ctx).Create(session).Error
}

func (repo *UserSessionRepository) FindByJTIHash(ctx context.Context, hash string) (*domain.UserSession, error) {
	model := &models.UserSessionModel{}
	err := repo.DB.WithContext(ctx).
		Where("refresh_token_hash = ? AND revoked_at IS NULL AND expired_at > ?", hash, time.Now()).
		First(model).Error

	if err != nil {
		return nil, err
	}

	return &domain.UserSession{
		ID:               model.ID,
		UserID:           model.UserID,
		RefreshTokenHash: model.RefreshTokenHash,
		ExpiredAt:        model.ExpiresAt,
		RevokedAt:        model.RevokedAt,
		LastUsedAt:       model.LastUsedAt,
		CreatedAt:        model.CreatedAt,
	}, nil
}

func (repo *UserSessionRepository) Revoke(ctx context.Context, jtiHash string) error {
	return repo.DB.WithContext(ctx).
		Model(&models.UserSessionModel{}).
		Where("refresh_token_hash = ?", jtiHash).
		Update("revoked_at", time.Now()).Error
}
