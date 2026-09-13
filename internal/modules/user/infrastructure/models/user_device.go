package models

import "time"

type UserDeviceModel struct {
	ID         uint64    `gorm:"primaryKey;column:id"`
	UserID     uint64    `gorm:"column:user_id"`
	DeviceName string    `gorm:"column:device_name"`
	Platform   string    `gorm:"column:platform"`
	AppVersion string    `gorm:"column:app_version"`
	PushToken  string    `gorm:"column:push_token"`
	LastSeenAt time.Time `gorm:"column:last_ssen_at"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (UserDeviceModel) TableName() string {
	return "user_devices"
}
