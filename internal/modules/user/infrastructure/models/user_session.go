package models

import "time"

type UserSessionModel struct {
	ID               uint64    `gorm:"primaryKey;column:id"`
	UserID           uint64    `gorm:"column:user_id"`
	RefreshTokenHash string    `gorm:"column:refresh_token_hash"`
	ExpiresAt        time.Time `gorm:"column:expires_at"`
	RevokedAt        time.Time `gorm:"column:revoked_at"`
	LastUsedAt       time.Time `gorm:"column:last_used_at"`
	CreatedAt        time.Time `gorm:"column:created_at"`
}

func (UserSessionModel) TableName() string {
	return "user_sessions"
}
