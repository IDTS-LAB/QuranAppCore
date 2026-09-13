package models

import (
	"time"

	"gorm.io/gorm"
)

type UserModel struct {
	ID              uint64         `gorm:"primaryKey;column:id"`
	Email           string         `gorm:"column:email"`
	PasswordHash    string         `gorm:"column:password_hash"`
	DisplayName     string         `gorm:"column:display_name"`
	EmailVerifiedAt time.Time      `gorm:"column:email_verified_at"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index;column:deleted_at"`
}

func (UserModel) TableName() string {
	return "users"
}
