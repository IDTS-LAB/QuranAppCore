package database

import (
	"context"
	"errors"
	"time"

	"github.com/fiqrikm18/quran-app/internal/modules/user/domain"
	"github.com/fiqrikm18/quran-app/internal/modules/user/infrastructure/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

func (repo *UserRepository) Create(ctx context.Context, user *domain.User, passwordHash string) error {
	userModel := &models.UserModel{
		Email:        user.Email,
		DisplayName:  user.DisplayName,
		PasswordHash: passwordHash,
	}

	return repo.DB.WithContext(ctx).Create(&userModel).Error
}

func (repo *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}
	err := repo.DB.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (repo *UserRepository) FindByID(ctx context.Context, id uint64) (*domain.User, error) {
	user := &domain.User{}
	err := repo.DB.WithContext(ctx).
		Where("id = ?", id).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (repo *UserRepository) MarkVerified(ctx context.Context, id uint64) error {
	return repo.DB.WithContext(ctx).
		Where("id = ?", id).
		Update("email_verified_at", time.Now()).Error
}

type PasswordHasher struct {
}

func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{}
}

func (hasher *PasswordHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (hasher *PasswordHasher) Compare(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
